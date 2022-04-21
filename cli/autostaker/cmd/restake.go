package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

func init() {
	var tolerance int64
	var restakeCmd = &cobra.Command{
		Use:   "restake [url] [address]",
		Short: "manually restakes the tokens of a registered address",
		Example: `autostaker restake https://autostaker.plural.to
cosmos147l494tccpk7ecr8vmqc67y542tl90659dgvda --tolerance 10000`,
		Args: cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			userAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			addr := args[0]
			if !strings.Contains(addr, "://") {
				addr = "http://" + addr
			}

			query := fmt.Sprintf("%s/v1/restake?address=%s", addr, userAddress.String())

			if tolerance >= 0 {
				query += fmt.Sprintf("&tolerance=%d", tolerance)
			}

			resp, err := http.Get(query)
			if err != nil {
				return fmt.Errorf("http GET error: %w", err)
			}
			if resp.StatusCode != 200 {
				return fmt.Errorf("Received unexpected code %d from url", resp.StatusCode)
			}

			respBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("Reading response: %w", err)
			}
			var message string
			err = json.Unmarshal(respBytes, &message)
			if err != nil {
				return fmt.Errorf("Unmarshalling message: %w", err)
			}

			c.Printf(message)

			return nil
		},
	}
	restakeCmd.Flags().Int64Var(&tolerance, "tolerance", -1, "How many native tokens to remain liquid for fees")
	rootCmd.AddCommand(restakeCmd)
}
