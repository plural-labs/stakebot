package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/cmwaters/autostaker/server"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use: "serve",
	Short: "Run the autostaker server",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filePath := filepath.Join(homeDir, defaultDir, defaultConfigFileName)
		config, err := server.LoadConfig(filePath)
		if err != nil {
			return err
		}

		return server.Serve(config)
	},
}