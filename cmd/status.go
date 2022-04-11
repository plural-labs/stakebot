package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status [url] [address]",
	Short: "Queries the current autostaking status of an address",
	Args:  cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		return nil
	},
}
