package bot

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"google.golang.org/grpc"

	"github.com/plural-labs/autostaker/client"
	"github.com/plural-labs/autostaker/types"
)

// Restake queries an addresses' delegations. It executes a claim call on all delegations. It then calculates
// a users liquid balance in the staking denom. It divides the balance proportionally to the delegated validators
// and bundles together delegate msgs to effectively restake all available tokens above the specified tolerance.
// This is a blocking function.
// NOTE: This only allows staking of the native token. I haven't seen a chain yet where you can stake other tokens
// but correct me if I'm wrong.
func (bot AutoStakeBot) Restake(ctx context.Context, address string, tolerance int64) (int64, error) {
	chain, err := types.FindChainFromAddress(bot.config.Chains, address)
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
	if delegations.Total.AmountOf(chain.NativeDenom).BigInt().Int64() == 0 {
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
	stakableBalance := resp.Balance.Amount.Int64() + delegations.Total.AmountOf(chain.NativeDenom).BigInt().Int64() - tolerance
	for idx, delegation := range delegations.Rewards {
		amount := stakableBalance / delegation.Reward.AmountOf(chain.NativeDenom).BigInt().Int64()
		delegateMsg := &staking.MsgDelegate{
			DelegatorAddress: address,
			ValidatorAddress: delegation.ValidatorAddress,
			Amount:           sdk.NewInt64Coin(chain.NativeDenom, amount),
		}
		msgs[idx+len(delegations.Rewards)] = delegateMsg
	}

	// wrap all the messages in an auth exec message
	authzMsg := authz.NewMsgExec(sdk.AccAddress(bot.address), msgs)

	// TODO: Might be helpful to catch the results and log them to INFO for debugging
	txResp, err := bot.client.Send(ctx, []sdk.Msg{&authzMsg}, client.WithGranter(address))
	if err != nil {
		return 0, err
	}

	if txResp.Code != 0 {
		return 0, fmt.Errorf("failed to submit restake transaction: %v", txResp.RawLog)
	}

	return delegations.Total.AmountOf(chain.NativeDenom).BigInt().Int64(), nil
}
