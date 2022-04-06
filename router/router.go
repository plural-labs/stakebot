package router

import (
	"github.com/gorilla/mux"

	"github.com/plural-labs/autostaker/router/v1"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
)

func RegisterRoutes(router *mux.Router, store *store.Store, chains []types.Chain, address string) {
	r := router.PathPrefix("/v1").Subrouter()
	v1.RegisterRoutes(r, store, chains, address)
}
