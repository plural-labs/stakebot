package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/plural-labs/autostaker/router/v1"
	"github.com/plural-labs/autostaker/types"
)

func Serve(config types.Config) error {
	router := mux.NewRouter()

	v1.RegisterRoutes(router)

	server := &http.Server{
		Handler:      router,
		Addr:         config.ListenAddr,
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	fmt.Printf("Running server at %s...\n", config.ListenAddr)
	return server.ListenAndServe()
}
