package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/plural-labs/autostaker/bot"
)

func Serve(ctx context.Context, listenAddr string, stakebot *bot.AutoStakeBot) error {
	r := mux.NewRouter()
	RegisterRoutes(r, stakebot)

	server := &http.Server{
		Handler:      r,
		Addr:         listenAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("Server error")
		}
	}()

	log.Info().Str("ListenAddr", listenAddr).Msg("Started server")

	select {
	case <-ctx.Done():
		log.Info().Msg("Shutting down server")
		err := server.Close()
		if err != nil {
			return err
		}
		return nil

	}
}
