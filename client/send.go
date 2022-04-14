package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"

	codec "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func WithGranter(granter string) SendOptionsFn {
	return func(opts SendOptions) SendOptions {
		opts.Granter = granter
		return opts
	}
}

func WithFee(fee sdk.Coin) SendOptionsFn {
	return func(opts SendOptions) SendOptions {
		opts.Fee = fee
		return opts
	}
}

func WithPubKey() SendOptionsFn {
	return func(opts SendOptions) SendOptions {
		opts.PubKey = true
		return opts
	}
}

func (c *Client) Send(ctx context.Context, msgs []sdk.Msg, opts ...SendOptionsFn) (*sdk.TxResponse, error) {
	anyMsgs := make([]*codec.Any, len(msgs))
	for idx, msg := range msgs {
		err := msg.ValidateBasic()
		if err != nil {
			return nil, err
		}

		anyMsgs[idx], err = codec.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
	}

	options := SendOptions{}
	for _, opt := range opts {
		options = opt(options)
	}

	Tx := tx.Tx{
		Body:     &tx.TxBody{Messages: anyMsgs},
		AuthInfo: &tx.AuthInfo{Fee: &tx.Fee{}},
	}
	signers := Tx.GetSigners()
	chain, err := c.chains.FindChainFromAddress(signers[0].String())
	if err != nil {
		return nil, err
	}

	for _, signer := range signers {
		_, err := c.signer.KeyByAddress(signer)
		if err != nil {
			return nil, fmt.Errorf("checking keys: %w", err)
		}
	}

	conn, err := grpc.Dial(
		chain.GRPC,
		grpc.WithInsecure(),
	)
	defer conn.Close()
	if err != nil {
		return nil, err
	}

	accountQuerier := auth.NewQueryClient(conn)
	signerInfos := make([]*tx.SignerInfo, len(signers))
	accountNumbers := make([]uint64, len(signers))
	for idx, signer := range signers {
		acc, err := accountQuerier.Account(ctx, &auth.QueryAccountRequest{Address: signer.String()})
		if err != nil {
			return nil, fmt.Errorf("retrieving account info for %d: %w", signer, err)
		}
		account := &auth.BaseAccount{}
		err = proto.Unmarshal(acc.Account.Value, account)
		if err != nil {
			return nil, err
		}

		signerInfos[idx] = &tx.SignerInfo{
			ModeInfo: &tx.ModeInfo{
				Sum: &tx.ModeInfo_Single_{
					Single: &tx.ModeInfo_Single{Mode: signing.SignMode_SIGN_MODE_DIRECT},
				},
			},
			Sequence: account.Sequence,
		}
		if options.PubKey {
			info, _ := c.signer.KeyByAddress(signer)
			pk := info.GetPubKey()
			pkAny, err := codec.NewAnyWithValue(pk)
			if err != nil {
				return nil, err
			}
			signerInfos[idx].PublicKey = pkAny
		}
		accountNumbers[idx] = account.AccountNumber
	}

	Tx.AuthInfo.SignerInfos = signerInfos

	txClient := tx.NewServiceClient(conn)

	// TODO: add gas estimation that ideally doesn't require signing (soft estimation)
	// txBytes, err := Tx.Marshal()
	// if err != nil {
	// 	return nil, err
	// }
	// simresp, err := txClient.Simulate(ctx, &tx.SimulateRequest{
	// 	TxBytes: txBytes,
	// })
	// if err != nil {
	// 	return nil, err
	// }

	Tx.AuthInfo.Fee = &tx.Fee{GasLimit: 2000000}

	if options.Granter != "" {
		Tx.AuthInfo.Fee.Granter = options.Granter
	}

	if !options.Fee.IsNil() {
		Tx.AuthInfo.Fee.Amount = sdk.NewCoins(options.Fee)
	}

	bodyBytes, err := Tx.Body.Marshal()
	if err != nil {
		return nil, err
	}
	authInfoBytes, err := Tx.AuthInfo.Marshal()
	if err != nil {
		return nil, err
	}
	signatures := make([][]byte, len(signers))
	for idx, signer := range signers {
		signDoc := &tx.SignDoc{
			BodyBytes:     bodyBytes,
			AuthInfoBytes: authInfoBytes,
			ChainId:       chain.Id,
			AccountNumber: accountNumbers[idx],
		}
		signedBytes, err := signDoc.Marshal()
		if err != nil {
			return nil, err
		}

		sig, _, err := c.signer.SignByAddress(signer, signedBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to sign message: %w", err)
		}
		signatures[idx] = sig
	}

	Tx.Signatures = signatures

	if err := Tx.ValidateBasic(); err != nil {
		return nil, err
	}

	raw := &tx.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    signatures,
	}
	txBytes, err := proto.Marshal(raw)
	if err != nil {
		return nil, err
	}

	txResp, err := txClient.BroadcastTx(ctx, &tx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    tx.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	if err != nil {
		return nil, err
	}

	if txResp.TxResponse.Code != 0 {
		return txResp.TxResponse, nil
	}

	for {
		select {
		case <-time.After(time.Millisecond * 100):
			resTx, err := txClient.GetTx(ctx, &tx.GetTxRequest{Hash: txResp.TxResponse.TxHash})
			if err != nil {
				if strings.Contains(err.Error(), "tx not found") {
					// retry
					continue
				}
				return nil, err
			}
			return resTx.TxResponse, nil

		case <-ctx.Done():
			return nil, nil
		}
	}
}

type SendOptionsFn func(opts SendOptions) SendOptions

type SendOptions struct {
	Granter string
	Fee     sdk.Coin
	PubKey  bool
}
