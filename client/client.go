package client

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"

	"github.com/plural-labs/stakebot/types"
)

type Client struct {
	signer keyring.Keyring
	chains types.ChainRegistry
}

func New(signer keyring.Keyring, chains types.ChainRegistry) *Client {
	return &Client{signer: signer, chains: chains}
}
