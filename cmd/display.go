package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(displayCmd)
}

var displayCmd = &cobra.Command{
	Use:   "address",
	Short: "Return the address of the autostaker server",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyring, err := getKeyring()
		if err != nil {
			return err
		}

		key, err := keyring.Key(keyName)
		if err != nil {
			return err
		}

		fmt.Println(key.GetAddress().String())

		return nil
	},
}
