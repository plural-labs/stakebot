package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cmwaters/autostaker/server"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/go-bip39"
	"github.com/spf13/cobra"
)

const (
	mnemonicEntropySize = 256
	keyName = "autostaker"
	keySigningAlgorithm = "secp256k1"
	defaultDir = ".autostaker"
	defaultConfigFileName = "config.toml"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Initialize an instance of the autostaker",
	Long: "Creates a config, keys and a database needed to run the server",
	RunE: func(cmd *cobra.Command, args []string) error {
		keyInfo, mnemonic, err := initAccount()
		if err != nil {
			return err
		}

		cmd.Printf(`
Generated an account: %d
Pubkey: %X
Mnemonic: %v
Write this mnemonic phrase in a safe place
		`, keyInfo.GetAddress().String(), keyInfo.GetPubKey().Bytes(), mnemonic)
		return initConfig()
	},
}

func getKeyring() (keyring.Keyring, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(homeDir, defaultDir)
	kb, err := keyring.New(keyName, keyring.BackendFile, filePath, os.Stdin)
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
	if err != nil {
		return nil, "", err
	}
	if info != nil {
		return nil, "", errors.New("account already initialized") 
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
	config := server.DefaultConfig()
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filePath := filepath.Join(homeDir, defaultDir, defaultConfigFileName)
	return config.Save(filePath)
}