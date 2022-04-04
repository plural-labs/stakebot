package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	cron "github.com/robfig/cron/v3"

	"github.com/plural-labs/autostaker/router"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
)

func Serve(config types.Config, homeDir string) error {
	store, err := store.New(homeDir, config.Chains)
	if err != nil {
		return fmt.Errorf("error creating store: %v", err)
	}
	defer store.Close()

	r := mux.NewRouter()
	c := cron.New()
	router.RegisterRoutes(r, store, c)

	server := &http.Server{
		Handler:      r,
		Addr:         config.ListenAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	records, err := store.GetAll()
	if err != nil {
		return err
	}
	for _, record := range records {
		id, err := c.AddFunc(record.Frequency, RestakeJob(record))
		if err != nil {
			return err
		}
		record.Id = int64(id)
	}

	fmt.Printf("Running server at %s...\n", config.ListenAddr)
	return server.ListenAndServe()
}

func RestakeJob(record *types.Record) func() {
	return func() {
		
	}
}