package client

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/plural-labs/autostaker/types"
)

type Client struct {
	signer keyring.Keyring
	chains []types.Chain
}

func New(signer keyring.Keyring, chains []types.Chain) *Client {
	return &Client{signer: signer, chains: chains}
}
