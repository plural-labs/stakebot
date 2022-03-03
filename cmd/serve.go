package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use: "serve",
	Short: "Run the autostaker server",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyInfo, mnemonic, err := initAccount()
		if err != nil {
			return err
		}

		cmd.Printf(`
Generated an account: %d
Pubkey: %X
Mnemonic: %v
Write this mnemonic phrase in a safe place
		`, keyInfo.GetAddress().String(), keyInfo.GetPubKey().Bytes(), mnemonic)
		return initConfig()
	},
}