package cmd

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Chains []Chain
	ListenAddr string
}

type Chain struct {
	RPC string
	REST string
	ChainId string
	ChainName string
}

func DefaultConfig() Config {
	return Config{ Chains: make([]Chain, 0), ListenAddr: "http://localhost:8080"}
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
	_, err := toml.Decode(file, &config)
	if err != nil {
		return config, fmt.Errorf("failed to load config from %q: %w", file, err)
	}
	return config, nil 
}