package router

import (
	"github.com/gorilla/mux"

	"github.com/plural-labs/autostaker/bot"
	"github.com/plural-labs/autostaker/router/v1"
)

func RegisterRoutes(router *mux.Router, bot *bot.AutoStakeBot) {
	r := router.PathPrefix("/v1").Subrouter()
	v1.RegisterRoutes(r, bot)
}
