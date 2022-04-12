package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"

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

	return server.ListenAndServe()
}
