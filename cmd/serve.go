package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/plural-labs/autostaker/server"
	"github.com/plural-labs/autostaker/types"
)

func init() {
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run the autostaker server",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		filePath := filepath.Join(homeDir, defaultDir, defaultConfigFileName)
		config, err := types.LoadConfig(filePath)
		if err != nil {
			return err
		}

		return server.Serve(config, filepath.Join(homeDir, defaultDir))
	},
}
