package v1

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/plural-labs/autostaker/store"
	"github.com/plural-labs/autostaker/types"
)

func RegisterRoutes(router *mux.Router, store *store.Store, chains []types.Chain, address string) {
	h := &Handler{store: store, chains: chains, address: address}
	router.HandleFunc("/status", h.Status).Methods("GET")
	router.HandleFunc("/chains", h.Chains).Methods("GET")
	router.HandleFunc("/chain", h.ChainById).Methods("GET")
	router.HandleFunc("/address", h.Address).Methods("GET")
	router.HandleFunc("/register", h.RegisterAddress).Methods("POST")
}

type Handler struct {
	store   *store.Store
	chains  []types.Chain
	// hex coded address
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

	valid, err := ValidateAddress(context.Background(), conn, address, h.address)
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
		// TODO: should return a 500 error
		return
	}

	// Success!
	RespondWithJSON(res, http.StatusOK, record)
}

func (h Handler) Address(res http.ResponseWriter, req *http.Request) {
	chainId := req.URL.Query().Get("chain_id")
	if chainId == "" {
		// return hex coded address
		RespondWithJSON(res, http.StatusOK, h.address)
		return
	}
	for _, chain := range h.chains {
		if chain.Id == chainId {
			bz, err := hex.DecodeString(h.address)
			if err != nil {
				log.Error().Err(err).Msg("Decoding address")
				// TODO: should return a 500 error
				return
			}
			bech32Addr, err := bech32.ConvertAndEncode(chain.Prefix, bz)
			if err != nil {
				log.Error().Err(err).Msg("Converting address to Bech32")
				// TODO: should return a 500 error
				return
			}
			RespondWithJSON(res, http.StatusOK, bech32Addr)
			return
		}
	}
	RespondWithJSON(res, http.StatusOK, "Chain not supported")
}

// RespondWithJSON provides an auxiliary function to return an HTTP response
// with JSON content and an HTTP status code.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

// ValidateAddress checks whether a specified address is valid for autostaking. The address must
// have granted authorization of the required messages as well as feegrant
func ValidateAddress(ctx context.Context, conn *grpc.ClientConn, address, authority string) (bool, error) {
	feegrantClient := feegrant.NewQueryClient(conn)
	resp, err := feegrantClient.Allowance(ctx, &feegrant.QueryAllowanceRequest{
		Granter: address,
		Grantee: authority,
	})
	if err != nil {
		return false, err
	}
	if resp.Allowance == nil {
		return false, fmt.Errorf("address %s is not covering the fees for autostaker (%s)", address, authority)
	}
	grant := resp.Allowance.Allowance.GetCachedValue().(feegrant.Grant)
	allowance := grant.Allowance.GetCachedValue().(feegrant.AllowedMsgAllowance)
	contains := false
	for _, msgType := range allowance.AllowedMessages {
		if msgType == sdk.MsgTypeURL(&authz.MsgExec{}) {
			contains = true
		}
	}
	if !contains {
		return false, fmt.Errorf("address %s does not cover authz.MsgExec fees for autostaker (%s)", address, authority)
	}

	authzClient := authz.NewQueryClient(conn)
	grantsResp, err := authzClient.Grants(ctx, &authz.QueryGrantsRequest{
		Granter:    address,
		Grantee:    authority,
		MsgTypeUrl: sdk.MsgTypeURL(&staking.MsgDelegate{}),
	})
	if err != nil {
		return false, err
	}
	if len(grantsResp.Grants) == 0 {
		return false, fmt.Errorf("address %s must authorize the autostaker (%s) to MsgDelegate", address, authority)
	}

	grantsResp, err = authzClient.Grants(ctx, &authz.QueryGrantsRequest{
		Granter:    address,
		Grantee:    authority,
		MsgTypeUrl: sdk.MsgTypeURL(&distribution.MsgWithdrawDelegatorReward{}),
	})
	if err != nil {
		return false, err
	}
	if len(grantsResp.Grants) == 0 {
		return false, fmt.Errorf("grants %s must authorize the autostaker (%s) to MsgWithdrawDelegatorReward", address, authority)
	}

	return true, nil
}
