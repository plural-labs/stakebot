package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"

	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cobra"

	"github.com/plural-labs/autostaker/client"
	"github.com/plural-labs/autostaker/types"
)

func init() {
	var (
		frequency      string
		tolerance      int64
		keyringDir     string
		appName        string
		keyringBackend string
	)
	var registerCmd = &cobra.Command{
		Use:   "register [url] [address]",
		Short: "Set up an account with a autostaking bot",
		Args:  cobra.ExactArgs(2),
		RunE: func(c *cobra.Command, args []string) error {
			_, err := url.Parse(args[0])
			if err != nil {
				return err
			}
			url := args[0]

			userAddress, err := sdk.AccAddressFromBech32(args[1])
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

			if _, err := signer.KeyByAddress(userAddress); err != nil {
				return err
			}

			chainsResp, err := http.Get(fmt.Sprintf("%s/v1/chains", url))
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
			var chains types.ChainRegistry
			err = json.Unmarshal(chainBytes, &chains)
			if err != nil {
				return err
			}

			chain, err := chains.FindChainFromAddress(args[1])
			if err != nil {
				return fmt.Errorf("autostaker bot does not support any chain with the address %s", args[1])
			}

			addressResp, err := http.Get(fmt.Sprintf("%s/v1/address?chain_id=%s", url, chain.Id))
			if err != nil {
				return err
			}

			if addressResp.StatusCode != 200 {
				return fmt.Errorf("Received unexpected code %d from url with GET /address", chainsResp.StatusCode)
			}

			addressBytes, err := ioutil.ReadAll(addressResp.Body)
			if err != nil {
				return err
			}

			var address string
			err = json.Unmarshal(addressBytes, &address)
			if err != nil {
				return err
			}
			botAddress, err := sdk.AccAddressFromBech32(address)
			if err != nil {
				return fmt.Errorf("Autostaking bot provided incorrect address %s, %w", address, err)
			}

			c.Printf("Authorizing autostaking bot (%s) with address %s on %s\n", botAddress.String(), userAddress.String(), chain.Id)

			client := client.New(signer, chains)
			if err := AuthorizeRestaking(c.Context(), client, userAddress, botAddress); err != nil {
				return err
			}

			queryStr := fmt.Sprintf("%s/v1/register?address=%s", url, userAddress.String())
			if tolerance >= 0 {
				queryStr += fmt.Sprintf("&tolerance=%d", tolerance)
			}
			if frequency != "" {
				queryStr += fmt.Sprintf("&frequency=%s", frequency)
			}

			registerResp, err := http.Get(queryStr)
			if err != nil {
				return err
			}
			if registerResp.StatusCode != 200 {
				body, err := ioutil.ReadAll(registerResp.Body)
				if err != nil {
					return fmt.Errorf("Failed to read body from GET /register: %w", err)
				}
				return fmt.Errorf("Received unexpected code %d from url with GET /register: %s", registerResp.StatusCode, body)
			}

			c.Printf("Successfully registered %s\n", userAddress.String())

			return nil
		},
	}

	registerCmd.Flags().Int64Var(&tolerance, "tolerance", -1, "How many native tokens to remain liquid for fees")
	registerCmd.Flags().StringVar(&frequency, "frequency", "", "How often to restake (quarterday|daily|weekly|monthly)")
	registerCmd.Flags().StringVar(&appName, "app", "", "Name of the application")
	registerCmd.Flags().StringVar(&keyringDir, "keyring-dir", "", "Directory where the keyring is stored")
	registerCmd.Flags().StringVar(&keyringBackend, "keyring-backend", keyring.BackendOS, "Select keyring's backend (os|file|test)")

	rootCmd.AddCommand(registerCmd)
}

func AuthorizeRestaking(ctx context.Context, c *client.Client, userAddress, botAddress sdk.AccAddress) error {
	delegateAuth := authz.NewGenericAuthorization(sdk.MsgTypeURL(&staking.MsgDelegate{}))
	claimAuth := authz.NewGenericAuthorization(sdk.MsgTypeURL(&distribution.MsgWithdrawDelegatorReward{}))
	inTenYears := time.Now().Add(10 * 365 * 24 * time.Hour)

	authorizeDelegationsMsg, err := authz.NewMsgGrant(userAddress, botAddress, delegateAuth, inTenYears)
	if err != nil {
		return err
	}

	authorizeClaimMsg, err := authz.NewMsgGrant(userAddress, botAddress, claimAuth, inTenYears)
	if err != nil {
		return err
	}

	allowedMsg, err := feegrant.NewAllowedMsgAllowance(&feegrant.BasicAllowance{SpendLimit: nil, Expiration: &inTenYears}, []string{sdk.MsgTypeURL(&authz.MsgExec{})})
	if err != nil {
		return err
	}
	feegrantMsg, err := feegrant.NewMsgGrantAllowance(allowedMsg, userAddress, botAddress)
	if err != nil {
		return err
	}

	resp, err := c.Send(ctx, []sdk.Msg{authorizeDelegationsMsg, authorizeClaimMsg, feegrantMsg}, client.WithFee(sdk.NewCoin("stake", sdk.NewInt(10))))
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("failed to submit transaction: %s", resp.RawLog)
	}

	return nil
}

// TODO: Implement ability to revoke restaking
func RevokeRestaking(ctx context.Context, client *client.Client, userAddress, botAddress sdk.AccAddress) error {
	panic("Not Implemented")
	return nil
}
