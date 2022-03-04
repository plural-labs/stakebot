package v1

import (
	"net/http"

	"github.com/gorilla/mux"
)

func RegisterRoutes(router *mux.Router) {
	r := router.PathPrefix("/v1").Subrouter()
	r.HandleFunc("/status", StatusHandler)
}

func StatusHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
}
