package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

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

			_, err = config.Chains.FindChainFromAddress(args[0])
			if err != nil {
				return fmt.Errorf("autostakebot does not support chain with address %s", args[0])
			}

			addr, err := url.Parse(config.ListenAddr)
			if err != nil {
				return err
			}

			query := fmt.Sprintf("%s/v1/restake?address=%s", addr.String(), args[0])

			if tolerance >= 0 {
				query += fmt.Sprintf("&tolerance=%s", tolerance)
			}

			resp, err := http.Get(query)
			if err != nil {
				return err
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("Received unexpected code %d from url", resp.StatusCode)
			}

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			var message string
			err = json.Unmarshal(respBytes, &message)
			if err != nil {
				return err
			}

			c.Printf(message)

			return nil
		},
	}
	restakeCmd.Flags().Int64Var(&tolerance, "tolerance", -1, "How many native tokens to remain liquid for fees")
	rootCmd.AddCommand(restakeCmd)
}
