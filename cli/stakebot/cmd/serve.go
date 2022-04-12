package cmd

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/plural-labs/autostaker/bot"
	"github.com/plural-labs/autostaker/router"
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

		keyring, err := getKeyring()
		if err != nil {
			return err
		}

		stakingBot, err := bot.New(filepath.Join(homeDir, defaultDir), keyring, config.Chains)
		if err != nil {
			return err
		}

		ctx, cancel := signal.NotifyContext(cmd.Context(), syscall.SIGTERM, syscall.SIGINT)
		defer cancel()

		err = stakingBot.Run(ctx)
		if err != nil {
			return err
		}

		err = router.Serve(ctx, config.ListenAddr, stakingBot)

		<-ctx.Done()

		return nil
	},
}
