package types

import (
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Chains     ChainRegistry
	ListenAddr string `toml:"listen_addr"`
}

func DefaultConfig() Config {
	return Config{ListenAddr: "localhost:8000", Chains: DefaultChains()}
}

func DefaultChains() []Chain {
	return []Chain{
		{
			GRPC:             "localhost:9090",
			RPC:              "localhost:26657",
			Id:               "cosmoshub-4",
			Prefix:           "cosmos",
			DefaultFrequency: int32(Frequency_DAILY),
			DefaultTolerance: 1000000,
			NativeDenom:      "uatom",
			AppName:          "gaia",
		},
	}
}

func (cfg Config) Save(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to save manifest file %q: %w", file, err)
	}
	return toml.NewEncoder(f).Encode(cfg)
}

func LoadConfig(file string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(file, &config)
	if err != nil {
		return config, fmt.Errorf("failed to load config from %q: %w", file, err)
	}
	return config, nil
}

type Chain struct {
	GRPC             string `toml:"grpc"`
	RPC              string `toml:"rpc"`
	Id               string `toml:"chain_id"`
	Prefix           string `toml:"chain_prefix"`
	DefaultFrequency int32  `toml:"default_interval"`
	DefaultTolerance int64  `toml:"default_tolerance"`
	NativeDenom      string `toml:"native_denom"`
	AppName          string `toml:"app_name"`
}

type ChainRegistry []Chain

func (r ChainRegistry) FindChainFromAddress(address string) (Chain, error) {
	for _, chain := range r {
		if strings.HasPrefix(address, chain.Prefix) {
			return chain, nil
		}
	}
	return Chain{}, fmt.Errorf("no chain found for address %s", address)
}

func (r ChainRegistry) FindChainById(id string) (Chain, error) {
	for _, chain := range r {
		if id == chain.Id {
			return chain, nil
		}
	}
	return Chain{}, fmt.Errorf("no chain found with id %s", id)
}
