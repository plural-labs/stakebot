package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/cmwaters/autostaker/router/v1"
)

func Serve(config Config) error {
	router := mux.NewRouter()

	v1.RegisterRoutes(router)

	server := &http.Server{
		Handler: router,
		Addr: config.ListenAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout: 10 * time.Second,
	}

	return server.ListenAndServe()
}