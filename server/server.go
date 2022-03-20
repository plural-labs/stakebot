package server

import (
	"net/http"
	"time"

	"github.com/plural-labs/autostaker/router/v1"
	"github.com/gorilla/mux"
)

func Serve(config Config) error {
	router := mux.NewRouter()

	v1.RegisterRoutes(router)

	server := &http.Server{
		Handler:      router,
		Addr:         config.ListenAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	return server.ListenAndServe()
}
