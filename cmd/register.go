package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
)

var (
	interval  string
	tolerance int64
	keyringDir string
	keyringBackend string
)

func init() {
	var registerCmd = &cobra.Command{
		Use:   "register [url] [key]",
		Short: "Set up an account with a autostaking bot",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			_, err := url.Parse(args[0])
			if err != nil {
				return err
			}

			homeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}

			rootDir := filepath.Join(homeDir, ".gaia")

			keys, err := keyring.New(args[1], keyring.BackendOS, rootDir, os.Stdin)
			if err != nil {
				return err
			}

			infos, err := keys.List()
			if err != nil {
				return err
			}

			fmt.Print(infos[0].GetName())
			return nil
		},
	}

	registerCmd.Flags().Int64VarP(&tolerance, "tolerance", "t", -1, "How many native tokens to remain liquid for fees")
	registerCmd.Flags().StringVarP(&interval, "interval", "i", "daily", "How often to restake")
	registerCmd.Flags().StringVar(&keyringDir, "keyring-dir", "", "")
	registerCmd.Flags().StringVar(&keyringBackend, "keyring-backend", keyring.BackendOS, "Select keyring's backend (os|file|test)")

	rootCmd.AddCommand(registerCmd)
}
