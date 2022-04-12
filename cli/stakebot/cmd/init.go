package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/go-bip39"
	"github.com/plural-labs/autostaker/types"
	"github.com/spf13/cobra"
)

const (
	mnemonicEntropySize   = 256
	keyName               = "autostaker"
	keySigningAlgorithm   = "secp256k1"
	defaultDir            = ".autostaker"
	defaultConfigFileName = "config.toml"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize an instance of the autostaker",
	Long:  "Creates a config, keys and a database needed to run the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Printf("Initializing autostaker account.\n")
		keyInfo, mnemonic, err := initAccount()
		if err != nil {
			return err
		}

		if keyInfo != nil {
			cmd.Printf(`
Generated a new private key for the autostaker server
Pubkey: %X
Mnemonic: %v

Write this mnemonic phrase in a safe place
`, keyInfo.GetPubKey().Bytes(), mnemonic)
		}
		if err := initConfig(); err != nil {
			return err
		}
		cmd.Printf("Initialized config\n")
		return nil
	},
}

func getKeyring() (keyring.Keyring, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(homeDir, defaultDir)
	kb, err := keyring.New(keyName, keyring.BackendOS, filePath, os.Stdin)
	if err != nil {
		return nil, err
	}
	return kb, nil
}

func initAccount() (keyring.Info, string, error) {
	var bip39Passphrase, hdPath string
	kb, err := getKeyring()

	// Check to see if the account already exists
	info, err := kb.Key(keyName)
	if info != nil {
		fmt.Printf("Account already exists\n")
		return nil, "", nil
	}

	keyringAlgos, _ := kb.SupportedAlgorithms()
	algo, err := keyring.NewSigningAlgoFromString(keySigningAlgorithm, keyringAlgos)
	if err != nil {
		return nil, "", err
	}

	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		return nil, "", err
	}

	mnemonic, err := bip39.NewMnemonic(entropySeed)

	k, err := kb.NewAccount(keyName, mnemonic, bip39Passphrase, hdPath, algo)
	if err != nil {
		return nil, "", err
	}

	return k, mnemonic, nil
}

func initConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	rootDir := filepath.Join(homeDir, defaultDir)
	if err := os.Mkdir(rootDir, 0700); err != nil {
		return err
	}
	filePath := filepath.Join(rootDir, defaultConfigFileName)
	_, err = os.Stat(filePath)
	if os.IsExist(err) {
		return nil
	}
	config := types.DefaultConfig()
	return config.Save(filePath)
}