package router
import (
	"github.com/gorilla/mux"
	cron "github.com/robfig/cron/v3"

	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/router/v1"
)

func RegisterRoutes(router *mux.Router, store *store.Store, cron *cron.Cron) {
	r := router.PathPrefix("/v1").Subrouter()
	v1.RegisterRoutes(r, store, cron)
}