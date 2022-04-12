package bot

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	badger "github.com/dgraph-io/badger/v3"
	cron "github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/plural-labs/autostaker/client"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
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

// Run runs the bot. It is blocking. Cancel the context to gracefully shut the bot down.
// This function only errors on start up else it will log.
func (bot AutoStakeBot) Run(ctx context.Context) error {
	err := bot.StartJobs()
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		log.Info().Msg("Shutting down auto staking bot...")
		err = bot.StopJobs()
		if err != nil {
			log.Error().Err(err)
		}
		return nil
	}
}

func (bot AutoStakeBot) StartJobs() error {
	cronStrings := map[int32]string{
		1: "@hourly",
		2: "0 */6 * * *", // quarter day
		3: "@daily",
		4: "@weekly",
		5: "@monthly",
	}

	// create a cron job for each frequency
	for frequency, cronString := range cronStrings {
		job, err := bot.Store.GetJob(frequency)
		if err != nil && err != badger.ErrKeyNotFound {
			return err
		}
		if job != nil {
			// a cron job at this frequency is already running. Perhaps we forgot to stop the previous one
			log.Info().Str("frequency", types.Frequency_name[frequency]).Msg("Cron job already running")
			continue
		}

		id, err := bot.cron.AddFunc(cronString, func() {
			// We don't cache these as records may have been removed or added between cron jobs
			records, err := bot.Store.GetRecordsByFrequency(frequency)
			if err != nil {
				log.Error().Err(err).Str("frequency", types.Frequency_name[frequency]).Msg("Retrieveing records")
			}

			for _, record := range records {
				// TODO: consider using a timeout so we don't get stuck on a single user
				rewards, err := bot.Restake(context.TODO(), record.Address, record.Tolerance)
				if err != nil {
					log.Error().Err(err).Str("address", record.Address).Msg("Restaking")
					record.ErrorLogs = err.Error()
					continue
				}
				record.TotalAutostakedRewards += rewards
				record.LastUpdatedUnixTime = time.Now().Unix()

				err = bot.Store.SetRecord(record)
				if err != nil {
					log.Error().Err(err).Str("address", record.Address).Msg("Saving record")
				}
			}
		})
		if err != nil {
			return err
		}

		// persist the job to disk
		bot.Store.SetJob(&types.Job{
			Id:        int64(id),
			Frequency: types.Frequency(frequency),
		})

		log.Info().Str("frequency", types.Frequency_name[frequency]).Msg("Scheduled cron job")
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

func (bot AutoStakeBot) StopJobs() error {
	ctx := bot.cron.Stop()

	// wait for all scheduled jobs to terminate
	<-ctx.Done()

	return bot.Store.DeleteAllJobs()
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
