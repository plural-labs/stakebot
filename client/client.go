package client

import (
	"context"

	"google.golang.org/grpc"

	codec "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Restake queries an addresses, delegations. It executes a claim call on all delegations. It then calculates
// a users liquid balance in the staking token. It divides the balance proportionally to the delegated validators
// and bundles together delegate msgs to effectively restake all available tokens greater than a specified tolerance.
// This is a blocking function.
func Restake(ctx context.Context, grpc *grpc.ClientConn, address, authority string, tolerance int64) error {
	authzClient := authz.NewMsgClient(grpc)
	stakingClient := staking.NewQueryClient(grpc)
	var totalStaked int64 = 0
	var stakeDenom string
	delegations, err := stakingClient.DelegatorDelegations(ctx, &staking.QueryDelegatorDelegationsRequest{DelegatorAddr: address})
	if err != nil {
		return err
	}
	claimMsgs := make([]*codec.Any, 0, len(delegations.DelegationResponses))
	for idx, delegation := range delegations.DelegationResponses {
		claimMsg := &distribution.MsgWithdrawDelegatorReward{
			DelegatorAddress: delegation.Delegation.DelegatorAddress,
			ValidatorAddress: delegation.Delegation.ValidatorAddress,
		}
		anyMsg, err := codec.NewAnyWithValue(claimMsg)
		if err != nil {
			return err
		}
		claimMsgs[idx] = anyMsg
		totalStaked += delegation.Balance.Amount.Int64()
		stakeDenom = delegation.Balance.Denom
	}

	// TODO: It may be possible to shrink this all into a single message instead of two different ones
	// TODO: we should detect whether an account has revoked authority to restake and update the database
	_, err = authzClient.Exec(ctx, &authz.MsgExec{Grantee: authority, Msgs: claimMsgs})
	if err != nil {
		return err
	}

	bankClient := bank.NewQueryClient(grpc)
	resp, err := bankClient.Balance(ctx, &bank.QueryBalanceRequest{Address: address, Denom: stakeDenom})
	
	stakableBalance := resp.Balance.Amount.Int64() - tolerance
	if stakableBalance <= 0 {
		return nil
	}

	delegateMsgs := make([]*codec.Any, len(delegations.DelegationResponses))
	for idx, delegation := range delegations.DelegationResponses {
		amount := delegation.Balance.Amount.Int64() * stakableBalance / totalStaked
		delegateMsg := &staking.MsgDelegate{
			DelegatorAddress: address,
			ValidatorAddress: delegation.Delegation.ValidatorAddress,
			Amount: types.NewInt64Coin(stakeDenom, amount), 
		}
		
		anyMsg, err := codec.NewAnyWithValue(delegateMsg)
		if err != nil {
			return err
		}

		delegateMsgs[idx] = anyMsg
	}

	// TODO: Might be helpful to catch the results and log them to INFO for debugging
	_, err = authzClient.Exec(ctx, &authz.MsgExec{Grantee: authority, Msgs: delegateMsgs})

	return err
}
