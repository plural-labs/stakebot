package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/plural-labs/autostaker/client"
	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
)

func RegisterRoutes(router *mux.Router, store *store.Store, chains []types.Chain, address string) {
	h := &Handler{store: store, chains: chains, address: address}
	router.HandleFunc("/status", h.Status).Methods("GET")
	router.HandleFunc("/chains", h.Chains).Methods("GET")
	router.HandleFunc("/chain", h.ChainById).Methods("GET")
	router.HandleFunc("/register", h.RegisterAddress).Methods("POST")
}

type Handler struct {
	store   *store.Store
	chains  []types.Chain
	address string
}

func (h Handler) Status(res http.ResponseWriter, req *http.Request) {
	address := req.URL.Query().Get("address")
	if address == "" {
		res.WriteHeader(http.StatusOK)
		return
	} else {
		record, err := h.store.GetRecord(address)
		if err != nil {
			RespondWithJSON(res, http.StatusOK, err)
		} else {
			RespondWithJSON(res, http.StatusOK, record)
		}
	}
}

func (h Handler) Chains(res http.ResponseWriter, req *http.Request) {
	RespondWithJSON(res, http.StatusOK, h.chains)
}

func (h Handler) ChainById(res http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	if id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, chain := range h.chains {
		if chain.Id == id {
			RespondWithJSON(res, http.StatusOK, chain)
			return
		}
	}
	RespondWithJSON(res, http.StatusOK, "Chain not found")
}

func (h Handler) RegisterAddress(res http.ResponseWriter, req *http.Request) {
	address := req.URL.Query().Get("address")
	if address == "" {
		RespondWithJSON(res, http.StatusOK, "No address specified")
	}
	freuencyStr := req.URL.Query().Get("frequency")
	toleranceStr := req.URL.Query().Get("tolerance")

	chain, err := types.FindChainFromAddress(h.chains, address)
	if err != nil {
		RespondWithJSON(res, http.StatusOK, fmt.Sprintf("No chain saved corresponds with the address %s", address))
	}
	var (
		frequency int32
		tolerance int64
		ok        bool
	)
	if freuencyStr == "" {
		frequency = chain.DefaultFrequency
	} else {
		frequency, ok = types.Frequency_value[freuencyStr]
		if !ok {
			RespondWithJSON(res, http.StatusBadRequest, fmt.Sprintf("Unknown interval %s", freuencyStr))
			return
		}
	}
	if toleranceStr == "" {
		tolerance = chain.DefaultTolerance
	} else {
		// TODO: convert string to integer (I forgot how to do that)
		tolerance = 1000000
	}

	conn, err := grpc.Dial(chain.GRPC, grpc.WithInsecure())
	if err != nil {
		log.Error().Err(err).Msg("Registering address")
		RespondWithJSON(res, http.StatusBadRequest, fmt.Sprintf("Unable to connect to gRPC server (%s)", chain.GRPC))
		return
	}

	valid, err := client.ValidateAddress(context.Background(), conn, address, h.address)
	if err != nil {
		log.Error().Err(err).Msg("Registering address")
	}
	if err != nil || !valid {
		RespondWithJSON(res, http.StatusBadRequest, fmt.Sprintf("Unable to validate address: %s", chain.GRPC))
		return
	}

	record := &types.Record{
		Address:   address,
		Frequency: types.Frequency(frequency),
		Tolerance: tolerance,
	}
	err = h.store.SetRecord(record)
	if err != nil {
		log.Error().Err(err).Msg("Saving new record")
		return
	}

	// Success!
	RespondWithJSON(res, http.StatusOK, record)
}

// RespondWithJSON provides an auxiliary function to return an HTTP response
// with JSON content and an HTTP status code.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}
