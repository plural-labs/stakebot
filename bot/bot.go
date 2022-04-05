package bot

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cron "github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/plural-labs/autostaker/client"
	"github.com/plural-labs/autostaker/router"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
)

type AutoStakeBot struct {
	config types.Config
	store  *store.Store
	server *http.Server
	cron   *cron.Cron
}

func New(config types.Config, homeDir string) (*AutoStakeBot, error) {
	store, err := store.New(homeDir)
	if err != nil {
		return nil, err
	}

	r := mux.NewRouter()
	c := cron.New()
	router.RegisterRoutes(r, store, c)

	return &AutoStakeBot{
		config: config,
		store:  store,
		server: &http.Server{
			Handler:      r,
			Addr:         config.ListenAddr,
			WriteTimeout: 10 * time.Second,
			ReadTimeout:  10 * time.Second,
		},
		cron: c,
	}, nil
}

func (bot AutoStakeBot) Start(ctx context.Context) error {
	records, err := bot.store.GetAll()
	if err != nil {
		return err
	}
	for _, record := range records {
		id, err := bot.cron.AddFunc(record.Frequency, RestakeJob(record))
		if err != nil {
			return err
		}
		record.Id = int64(id)
	}

	go func() {
		err := bot.server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Starting server")
		} else {
			log.Info().Str("ListenAddress", bot.config.ListenAddr).Msg("Starting server")
		}
	}()

	select {
	case <-ctx.Done():
		return bot.server.Close()
	}
}

func (bot AutoStakeBot) CreateRestakeJob(record *types.Record) error {
	if record.Id >= 0 {
		return nil
	}

	id, err := bot.cron.AddFunc(record.Frequency, func() {
		conn := grpc.Dial(bot.chain)
		client.Restake(ctx)
	})
	if err != nil {
		return err
	}
	record.Id = int64(id)
	return nil
}
