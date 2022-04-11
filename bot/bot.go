package bot

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	badger "github.com/dgraph-io/badger/v3"
	"github.com/gorilla/mux"
	cron "github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"

	"github.com/plural-labs/autostaker/client"
	"github.com/plural-labs/autostaker/router"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
)

type AutoStakeBot struct {
	config  types.Config
	store   *store.Store
	server  *http.Server
	cron    *cron.Cron
	client  *client.Client
	address string
}

func New(config types.Config, homeDir string, key keyring.Keyring) (*AutoStakeBot, error) {
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

	r := mux.NewRouter()
	_, bz, err := bech32.DecodeAndConvert(address)
	if err != nil {
		panic(err)
	}
	hexAddress := hex.EncodeToString(bz)
	router.RegisterRoutes(r, store, config.Chains, hexAddress)

	client := client.New(key, config.Chains)

	return &AutoStakeBot{
		config: config,
		store:  store,
		server: &http.Server{
			Handler:      r,
			Addr:         config.ListenAddr,
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
		cron:    cron.New(),
		client:  client,
		address: address,
	}, nil
}

// Starts the bot. Is blocking. Cancel the context to gracefully shut the bot down.
// This function only errors on start up else it will log.
func (bot AutoStakeBot) Start(ctx context.Context) error {
	err := bot.StartJobs()
	if err != nil {
		return err
	}

	go func() {
		log.Info().Str("ListenAddress", bot.config.ListenAddr).Msg("Starting server...")
		err := bot.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server")
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("Shutting down auto staking bot...")
		err := bot.server.Close()
		if err != nil {
			log.Error().Err(err)
		}
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
		job, err := bot.store.GetJob(frequency)
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
			records, err := bot.store.GetRecordsByFrequency(frequency)
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

				err = bot.store.SetRecord(record)
				if err != nil {
					log.Error().Err(err).Str("address", record.Address).Msg("Saving record")
				}
			}
		})
		if err != nil {
			return err
		}

		log.Info().Str("frequency", types.Frequency_name[frequency]).Msg("Scheduled cron job")
		// persist the job to disk
		bot.store.SetJob(&types.Job{
			Id:        int64(id),
			Frequency: types.Frequency(frequency),
		})

	}

	// start up the scheduler
	bot.cron.Start()
	log.Info().Msg("Started cron scheduler")
	return nil
}

func (bot AutoStakeBot) StopJobs() error {
	ctx := bot.cron.Stop()

	// wait for all scheduled jobs to terminate
	<-ctx.Done()

	return bot.store.DeleteAllJobs()
}

func (bot AutoStakeBot) findChain(address string) (types.Chain, error) {
	return types.FindChainFromAddress(bot.config.Chains, address)
}
