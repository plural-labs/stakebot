package bot

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	cron "github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/plural-labs/stakebot/client"
	"github.com/plural-labs/stakebot/store"
	"github.com/plural-labs/stakebot/types"
)

type AutoStakeBot struct {
	Store *store.Store

	chains  types.ChainRegistry
	cron    *cron.Cron
	client  *client.Client
	address string
}

func New(homeDir string, key keyring.Keyring, chains []types.Chain) (*AutoStakeBot, error) {
	store, err := store.New(homeDir)
	if err != nil {
		return nil, err
	}

	keys, err := key.List()
	if err != nil {
		return nil, err
	}
	if len(keys) != 1 {
		return nil, fmt.Errorf("expected 1 key, got %d", len(keys))
	}
	address := keys[0].GetAddress().String()

	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		panic(err)
	}
	hexAddress := hex.EncodeToString(bz)
	client := client.New(key, chains)

	return &AutoStakeBot{
		chains:  chains,
		Store:   store,
		cron:    cron.New(),
		client:  client,
		address: hexAddress,
	}, nil
}

func (bot AutoStakeBot) StartJobs() error {
	cronStrings := map[int32]string{
		1: "@hourly",
		2: "@every 6h", // quarter day
		3: "@daily",
		4: "@weekly",
		5: "@monthly",
	}

	// create a cron job for each frequency
	for frequency := int32(1); frequency <= 5; frequency++ {
		id, err := bot.cron.AddFunc(cronStrings[frequency], bot.Job(frequency))
		if err != nil {
			return err
		}

		log.Debug().Str("frequency", types.Frequency_name[frequency]).Str("cron string", cronStrings[frequency]).Int64("Id", int64(id)).Msg("Scheduled cron job")
	}

	// start up the scheduler
	bot.cron.Start()

	records, err := bot.Store.Len()
	if err != nil {
		return err
	}
	log.Info().Int("records", records).Msg("Started cron scheduler")
	return nil
}

func (bot AutoStakeBot) StopJobs() {
	ctx := bot.cron.Stop()

	// wait for all scheduled jobs to terminate
	<-ctx.Done()
}

func (bot AutoStakeBot) Chains() types.ChainRegistry {
	return bot.chains
}

func (bot AutoStakeBot) HEXAddress() string {
	return bot.address
}

func (bot AutoStakeBot) Bech32Address(chainID string) (string, error) {
	chain, err := bot.chains.FindChainById(chainID)
	if err != nil {
		return "", nil
	}
	bz, err := hex.DecodeString(bot.address)
	if err != nil {
		panic(err)
	}
	bech32Address, err := bech32.ConvertAndEncode(chain.Prefix, bz)
	if err != nil {
		panic(err)
	}
	return bech32Address, nil
}

func (bot AutoStakeBot) Job(frequency int32) func() {
	return func() {
		log.Info().Int32("frequency", frequency).Msg("Starting cron job")
		// We don't cache these as records may have been removed or added between cron jobs
		records, err := bot.Store.GetRecordsByFrequency(frequency)
		if err != nil {
			log.Error().Err(err).Str("frequency", types.Frequency_name[frequency]).Msg("Retrieveing records")
		}

		for _, record := range records {
			chain, err := bot.chains.FindChainFromAddress(record.Address)
			if err != nil {
				log.Error().Err(err).Str("address", record.Address).Msg("Finding chain")
			}

			// TODO: consider using a timeout so we don't get stuck on a single user
			rewards, err := bot.Restake(context.TODO(), record.Address, record.Tolerance, sdk.NewInt64Coin(chain.NativeDenom, chain.RestakeFee))
			if err != nil {
				log.Error().Err(err).Str("address", record.Address).Msg("Restaking")
				record.ErrorLogs = err.Error()
				continue
			}
			record.TotalAutostakedRewards += rewards
			record.LastUpdatedUnixTime = time.Now().Unix()
			record.ErrorLogs = ""

			err = bot.Store.SetRecord(record)
			if err != nil {
				log.Error().Err(err).Str("address", record.Address).Msg("Saving record")
			}

		}
		log.Info().Int("records", len(records)).Str("frequency", types.Frequency_name[frequency]).Int32("freq", frequency).Msg("Completed cron job")
	}
}
