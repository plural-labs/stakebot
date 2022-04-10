package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"

	"github.com/plural-labs/autostaker/client"
	"github.com/plural-labs/autostaker/types"
)

var (
	frequency      string
	tolerance      int64
	keyringDir     string
	appName        string
	keyringBackend string
)

func init() {
	var registerCmd = &cobra.Command{
		Use:   "register [url] [address]",
		Short: "Set up an account with a autostaking bot",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			_, err := url.Parse(args[0])
			if err != nil {
				return err
			}

			if keyringDir == "" {
				keyringDir, err = os.UserHomeDir()
				if err != nil {
					return err
				}
			}

			signer, err := keyring.New(appName, keyringBackend, keyringDir, os.Stdin)
			if err != nil {
				return err
			}

			chainsResp, err := http.Get(fmt.Sprintf("%s/v1/chains", args[0]))
			if err != nil {
				return err
			}
			if chainsResp.StatusCode != 200 {
				return fmt.Errorf("Received unexpected code %d from url", chainsResp.StatusCode)
			}

			chainBytes, err := ioutil.ReadAll(chainsResp.Body)
			if err != nil {
				return err
			}
			var chains []types.Chain
			err = json.Unmarshal(chainBytes, &chains)
			if err != nil {
				return err
			}

			chain, err := types.FindChainFromAddress(chains, args[1])
			if err != nil {
				return fmt.Errorf("autostaker bot does not support any chain with the address %s", args[1])
			}

			addressResp, err := http.Get(fmt.Sprintf("%s/v1/address?chain_id=%s", args[0], chain.Id))
			if err != nil {
				return err
			}

			client := client.New(signer, chains)

			return nil
		},
	}

	registerCmd.Flags().Int64Var(&tolerance, "tolerance", -1, "How many native tokens to remain liquid for fees")
	registerCmd.Flags().StringVar(&frequency, "freuency", "daily", "How often to restake")
	registerCmd.Flags().StringVar(&appName, "app", "", "Name of the application")
	registerCmd.Flags().StringVar(&keyringDir, "keyring-dir", "", "Directory where the keyring is stored")
	registerCmd.Flags().StringVar(&keyringBackend, "keyring-backend", keyring.BackendOS, "Select keyring's backend (os|file|test)")

	rootCmd.AddCommand(registerCmd)
}

// func Authorize(ctx context.Context, signer keyring.Keyring, conn *grpc.ClientConn, userAddress, botAddress sdk.AccAddress) error {
// 	delegateAuth := authz.NewGenericAuthorization(sdk.MsgTypeURL(&staking.MsgDelegate{}))
// 	claimAuth := authz.NewGenericAuthorization(sdk.MsgTypeURL(&distribution.MsgWithdrawDelegatorReward{}))
// 	inTenYears := time.Now().Add(10 * 365 * 24 * time.Hour)

// 	authorizeDelegationsMsg, err := authz.NewMsgGrant(userAddress, botAddress, delegateAuth, inTenYears)
// 	if err != nil {
// 		return err
// 	}

// 	authorizeClaimMsg, err := authz.NewMsgGrant(userAddress, botAddress, claimAuth, inTenYears)
// 	if err != nil {
// 		return err
// 	}

// 	allowedMsg, err := feegrant.NewAllowedMsgAllowance(&feegrant.BasicAllowance{SpendLimit: nil, Expiration: &inTenYears}, []string{sdk.MsgTypeURL(&authz.MsgExec{})})
// 	if err != nil {
// 		return err
// 	}
// 	feegrantMsg, err := feegrant.NewMsgGrantAllowance(allowedMsg, userAddress, botAddress)
// 	if err != nil {
// 		return err
// 	}

// 	bot.client.Send(ctx, []sdk.Msg{authorizeDelegationsMsg, authorizeClaimMsg, feegrantMsg})

// }

// func (bot AutoStakeBot) Cancel() error {
// 	return nil
// }
