package v1

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
)

func RegisterRoutes(router *mux.Router, store *store.Store, chains []types.Chain) {
	h := &Handler{store: store, chains: chains}
	router.HandleFunc("/status/{address}", h.StatusHandler).Methods("GET")
	router.HandleFunc("/chains", h.ChainsHandler).Methods("GET")
	router.HandleFunc("/chains/{id}", h.ChainByIdHandler).Methods("GET")
}

type Handler struct {
	store  *store.Store
	chains []types.Chain
}

func (h Handler) StatusHandler(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(http.StatusOK)
}

func (h Handler) ChainsHandler(res http.ResponseWriter, req *http.Request) {
	RespondWithJSON(res, http.StatusOK, h.chains)
}

func (h Handler) ChainByIdHandler(res http.ResponseWriter, req *http.Request) {
	for _, chain := range h.chains {
		if chain.Id == req.URL.Query()["id"][0] {
			RespondWithJSON(res, http.StatusOK, chain)
			return
		}
	}
	res.WriteHeader(http.StatusNotFound)
}

// RespondWithJSON provides an auxiliary function to return an HTTP response
// with JSON content and an HTTP status code.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
