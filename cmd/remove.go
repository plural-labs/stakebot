package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/plural-labs/stakebot/store"
)

func init() {
	rootCmd.AddCommand(removeCmd)
}

var removeCmd = &cobra.Command{
	Use:   "remove [address]",
	Short: "Maually removes an address",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		rootDir := filepath.Join(homeDir, defaultDir)

		store, err := store.New(rootDir)
		if err != nil {
			return err
		}

		if err := store.DeleteRecord(args[0]); err != nil {
			return err
		}

		cmd.Printf("Removed address %s from store\n", args[0])

		return nil
	},
}
