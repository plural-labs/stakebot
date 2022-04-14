package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/plural-labs/autostaker/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status [url] [address]",
	Short: "Queries the current autostaking status of an address",
	Args:  cobra.ExactArgs(2),
	RunE: func(c *cobra.Command, args []string) error {
		_, err := url.Parse(args[0])
		if err != nil {
			return err
		}
		url := args[0]

		resp, err := http.Get(fmt.Sprintf("%s/v1/status?address=%s", url, args[1]))
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
		var record types.Record
		err = json.Unmarshal(respBytes, &record)
		if err != nil {
			// If the record doesn't exist we need to still catch the message
			var message string
			err = json.Unmarshal(respBytes, &message)
			if err != nil {
				return err
			}
			c.Printf("%s\n", message)
			return err
		}

		c.Printf(`Status:
Address: %s
Tolerance: %d
Frequency: %s
Last Restaked: %s ago
Total Rewards Restaked: %d
Errors: %s
`, record.Address, record.Tolerance, types.Frequency_name[int32(record.Frequency)], time.Unix(record.LastUpdatedUnixTime, 0).String(), record.TotalAutostakedRewards, record.ErrorLogs)

		return nil
	},
}
