package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/plural-labs/autostaker/types"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(findCmd)
}

var findCmd = &cobra.Command{
	Use:   "find [address]",
	Short: "Search for an address",
	Args:  cobra.ExactArgs(1),
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

		addr := config.ListenAddr
		if !strings.Contains(config.ListenAddr, "://") {
			addr = "http://" + addr
		}

		resp, err := http.Get(fmt.Sprintf("%s/v1/status?address=%s", addr, args[0]))
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
			return err
		}

		lastUpdated := "-"
		if record.LastUpdatedUnixTime != 0 {
			lastUpdated = time.Now().Sub(time.UnixMicro(record.LastUpdatedUnixTime)).String()
		}

		cmd.Printf(`Status:
Address: %s
Tolerance: %d
Frequency: %s
Last Restaked: %s ago
Total Rewards Restaked: %d
Errors: %s
`, record.Address, record.Tolerance, types.Frequency_name[int32(record.Frequency)], lastUpdated, record.TotalAutostakedRewards, record.ErrorLogs)

		return nil
	},
}
