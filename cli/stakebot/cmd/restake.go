package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/plural-labs/autostaker/bot"
	"github.com/plural-labs/autostaker/types"
	"github.com/spf13/cobra"
)

func init() {
	var tolerance int64
	var restakeCmd = &cobra.Command{
		Use:   "restake [address]",
		Short: "manually restakes the tokens of a registered address",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			filePath := filepath.Join(homeDir, defaultDir, defaultConfigFileName)
			config, err := types.LoadConfig(filePath)
			if err != nil {
				return err
			}

			chain, err := types.FindChainFromAddress(config.Chains, args[0])
			if err != nil {
				return fmt.Errorf("autostakebot does not support chain with address %s", args[0])
			}

			keyring, err := getKeyring()
			if err != nil {
				return err
			}

			stakingBot, err := bot.New(config, filepath.Join(homeDir, defaultDir), keyring)
			if err != nil {
				return err
			}

			if tolerance < 0 {
				tolerance = chain.DefaultTolerance
			}

			value, err := stakingBot.Restake(c.Context(), args[0], tolerance)
			if err != nil {
				return err
			}

			c.Printf("Successfully restaked %d tokens\n", value)

			return nil
		},
	}
	restakeCmd.Flags().Int64Var(&tolerance, "tolerance", -1, "How many native tokens to remain liquid for fees")
	rootCmd.AddCommand(restakeCmd)
}
