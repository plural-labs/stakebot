package bot

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/plural-labs/stakebot/client"
)

// Restake queries an addresses' delegations. It executes a claim call on all delegations. It then calculates
// a users liquid balance in the staking denom. It divides the balance proportionally to the delegated validators
// and bundles together delegate msgs to effectively restake all available tokens above the specified tolerance.
// This is a blocking function.
// NOTE: This only allows staking of the native token. I haven't seen a chain yet where you can stake other tokens
// but correct me if I'm wrong.
func (bot AutoStakeBot) Restake(ctx context.Context, address string, tolerance int64, fee sdk.Coin) (int64, error) {
	chain, err := bot.chains.FindChainFromAddress(address)
	if err != nil {
		return 0, err
	}
	conn, err := grpc.Dial(chain.GRPC, grpc.WithInsecure())
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	distributionClient := distribution.NewQueryClient(conn)
	bankClient := bank.NewQueryClient(conn)
	delegations, err := distributionClient.DelegationTotalRewards(
		ctx,
		&distribution.QueryDelegationTotalRewardsRequest{
			DelegatorAddress: address,
		},
	)

	// Check if there are any rewards to claim
	totalRewards := delegations.Total.AmountOf(chain.NativeDenom).RoundInt64()
	log.Info().Interface("rewards", delegations).Str("address", address).Int64("totalRewards", totalRewards).Msg("Total rewards")
	if totalRewards <= 0 {
		return 0, nil
	}

	msgs := make([]sdk.Msg, len(delegations.Rewards)*2)
	for idx, delegation := range delegations.Rewards {
		claimMsg := &distribution.MsgWithdrawDelegatorReward{
			DelegatorAddress: address,
			ValidatorAddress: delegation.ValidatorAddress,
		}
		msgs[idx] = claimMsg
	}

	resp, err := bankClient.Balance(ctx, &bank.QueryBalanceRequest{Address: address, Denom: chain.NativeDenom})
	if err != nil {
		return 0, err
	}

	// Caclulate how much native token after claiming can be restaked
	stakableBalance := resp.Balance.Amount.Int64() + totalRewards - tolerance
	for idx, delegation := range delegations.Rewards {
		amount := (stakableBalance * delegation.Reward.AmountOf(chain.NativeDenom).RoundInt64()) / totalRewards
		log.Info().Int64("stakableBalance", stakableBalance).Int64("amount", amount).Msg("restake")

		delegateMsg := &staking.MsgDelegate{
			DelegatorAddress: address,
			ValidatorAddress: delegation.ValidatorAddress,
			Amount:           sdk.NewInt64Coin(chain.NativeDenom, amount),
		}
		log.Info().Str("amount", delegateMsg.Amount.String()).Str("delegator", delegateMsg.DelegatorAddress).Str("validator", delegateMsg.ValidatorAddress).Msg("Delegate")
		msgs[idx+len(delegations.Rewards)] = delegateMsg
	}

	// wrap all the messages in an auth exec message
	botBech32Addr, err := bot.Bech32Address(chain.Id)
	if err != nil {
		panic(err)
	}
	accAddress, err := sdk.AccAddressFromBech32(botBech32Addr)
	if err != nil {
		panic(err)
	}
	authzMsg := authz.NewMsgExec(accAddress, msgs)
	log.Info().Str("botAddress", botBech32Addr).Str("userAddress", address).Msg("Prepared messages")

	// TODO: Might be helpful to catch the results and log them to INFO for debugging
	txResp, err := bot.client.Send(ctx, []sdk.Msg{&authzMsg}, client.WithGranter(address), client.WithPubKey(), client.WithFee(fee))
	if err != nil {
		return 0, fmt.Errorf("error sending messages: %w", err)
	}

	if txResp.Code != 0 {
		return 0, fmt.Errorf("failed to submit restake transaction: %v", txResp.RawLog)
	}

	log.Info().Str("data", txResp.Data).Str("logs", txResp.RawLog).Msg("Succesfully sumbitted transaction")

	return totalRewards, nil
}
