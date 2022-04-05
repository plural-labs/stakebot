package cmd

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/plural-labs/autostaker/bot"
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

		stakingBot, err := bot.New(config, filepath.Join(homeDir, defaultDir))
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGTERM)
		defer cancel()

		err = stakingBot.Start(ctx)
		if err != nil {
			return err
		}

		<-ctx.Done()

		return nil
	},
}
