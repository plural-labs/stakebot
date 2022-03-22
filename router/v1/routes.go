package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	cron "github.com/robfig/cron/v3"

	"github.com/plural-labs/autostaker/store"
)

func RegisterRoutes(router *mux.Router, store *store.Store, cron *cron.Cron) {
	h := &Handler{}
	router.HandleFunc("/status/{address}", h.StatusHandler).Methods("GET")
	router.HandleFunc("/chains", h.ChainsHandler).Methods("GET")
}

type Handler struct {
	store store.Store
}

func (h Handler) StatusHandler(res http.ResponseWriter, req *http.Request) {

	res.WriteHeader(http.StatusOK)
}

func (h Handler) ChainsHandler(res http.ResponseWriter, req *http.Request) {

}


// RespondWithJSON provides an auxiliary function to return an HTTP response
// with JSON content and an HTTP status code.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}