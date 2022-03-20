package server

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/plural-labs/autostaker/types"
)

type Config struct {
	Chains     []types.Chain
	ListenAddr string
}

func DefaultConfig() Config {
	return Config{Chains: make([]types.Chain, 0), ListenAddr: "http://localhost:8080"}
}

func (cfg Config) Save(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to save manifest file %q: %w", file, err)
	}
	return toml.NewEncoder(f).Encode(cfg)
}

func LoadConfig(file string) (Config, error) {
	config := Config{}
	_, err := toml.Decode(file, &config)
	if err != nil {
		return config, fmt.Errorf("failed to load config from %q: %w", file, err)
	}
	return config, nil
}
