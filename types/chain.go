package types

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Chains     []Chain
	ListenAddr string `toml:"listen_addr"`
}

func DefaultConfig() Config {
	return Config{ListenAddr: "localhost:8000"}
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
	Id               string `toml:"chain_id"`
	Prefix           string `toml:"chain_prefix"`
	DefaultInterval  uint64 `toml:"default_interval"`
	DefaultTolerance uint64 `toml:"default_tolerance"`
}
