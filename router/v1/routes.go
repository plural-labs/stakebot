package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	distribution "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	"github.com/plural-labs/stakebot/bot"
	"github.com/plural-labs/stakebot/types"
)

func RegisterRoutes(router *mux.Router, bot *bot.AutoStakeBot) {
	h := &Handler{bot: bot}
	router.HandleFunc("/status", h.Status).Methods("GET")
	router.HandleFunc("/chains", h.Chains).Methods("GET")
	router.HandleFunc("/chain", h.ChainById).Methods("GET")
	router.HandleFunc("/address", h.Address).Methods("GET")
	router.HandleFunc("/register", h.RegisterAddress).Methods("GET")
	router.HandleFunc("/restake", h.Restake).Methods("GET")
}

type Handler struct {
	bot *bot.AutoStakeBot
}

func (h Handler) Status(res http.ResponseWriter, req *http.Request) {
	address := req.URL.Query().Get("address")
	if address == "" {
		res.WriteHeader(http.StatusOK)
		return
	} else {
		record, err := h.bot.Store.GetRecord(address)
		if err != nil {
			RespondWithJSON(res, http.StatusOK, err.Error())
		} else {
			RespondWithJSON(res, http.StatusOK, record)
		}
	}
}

func (h Handler) Chains(res http.ResponseWriter, req *http.Request) {
	RespondWithJSON(res, http.StatusOK, h.bot.Chains())
}

func (h Handler) ChainById(res http.ResponseWriter, req *http.Request) {
	id := req.URL.Query().Get("id")
	if id == "" {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, chain := range h.bot.Chains() {
		if chain.Id == id {
			RespondWithJSON(res, http.StatusOK, chain)
			return
		}
	}
	RespondWithJSON(res, http.StatusOK, "Chain not found")
}

func (h Handler) RegisterAddress(res http.ResponseWriter, req *http.Request) {
	log.Info().Msg("Registering new address")
	address := req.URL.Query().Get("address")
	if address == "" {
		RespondWithJSON(res, http.StatusOK, "No address specified")
	}
	frequencyStr := req.URL.Query().Get("frequency")
	toleranceStr := req.URL.Query().Get("tolerance")

	chain, err := h.bot.Chains().FindChainFromAddress(address)
	if err != nil {
		RespondWithJSON(res, http.StatusOK, fmt.Sprintf("No chain saved corresponds with the address %s", address))
	}
	var (
		frequency int32
		tolerance int64
		ok        bool
	)
	if frequencyStr == "" {
		frequency = chain.DefaultFrequency
	} else {
		frequency, ok = types.Frequency_value[strings.ToUpper(frequencyStr)]
		if !ok {
			RespondWithJSON(res, http.StatusBadRequest, fmt.Sprintf("Unknown interval %s", frequencyStr))
			return
		}
	}
	if toleranceStr == "" {
		tolerance = chain.DefaultTolerance
	} else {
		number, err := strconv.Atoi(toleranceStr)
		if err != nil {
			RespondWithJSON(res, http.StatusBadRequest, fmt.Sprintf("Failed to parse tolerance: %s", err.Error()))
			return
		}
		tolerance = int64(number)
	}

	conn, err := grpc.Dial(chain.GRPC, grpc.WithInsecure())
	if err != nil {
		log.Error().Err(err).Msg("Registering address")
		RespondWithJSON(res, http.StatusBadRequest, fmt.Sprintf("Unable to connect to gRPC server (%s)", chain.GRPC))
		return
	}

	bech32Address, err := h.bot.Bech32Address(chain.Id)
	if err != nil {
		panic(err)
	}

	valid, err := ValidateAddress(req.Context(), conn, address, bech32Address)
	if err != nil {
		log.Error().Err(err).Msg("Registering address")
	}
	if err != nil || !valid {
		RespondWithJSON(res, http.StatusBadRequest, fmt.Sprintf("Unable to validate address %s, error: %s", address, err.Error()))
		return
	}

	record := &types.Record{
		Address:   address,
		Frequency: types.Frequency(frequency),
		Tolerance: tolerance,
	}
	err = h.bot.Store.SetRecord(record)
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
		RespondWithJSON(res, http.StatusOK, h.bot.HEXAddress())
		return
	}

	address, err := h.bot.Bech32Address(chainId)
	if err != nil {
		RespondWithJSON(res, http.StatusOK, "Chain not supported")
		return
	}

	RespondWithJSON(res, http.StatusOK, address)
}

func (h Handler) Restake(res http.ResponseWriter, req *http.Request) {
	log.Info().Msg("Restake")
	address := req.URL.Query().Get("address")
	if address == "" {
		RespondWithJSON(res, http.StatusBadRequest, "No address specified")
		return
	}
	var (
		tolerance int64
		err       error
	)

	record, err := h.bot.Store.GetRecord(address)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("Getting record")
		RespondWithJSON(res, http.StatusOK, err.Error())
		return
	}

	toleranceStr := req.URL.Query().Get("tolerance")
	if toleranceStr == "" {
		tolerance = record.Tolerance
	} else {
		tolerance, err = strconv.ParseInt(toleranceStr, 10, 64)
		if err != nil {
			RespondWithJSON(res, http.StatusOK, err.Error())
			return
		}
	}

	value, err := h.bot.Restake(context.Background(), address, tolerance)
	record.LastUpdatedUnixTime = time.Now().Unix()
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("Restaking")
		record.ErrorLogs = err.Error()
		if err := h.bot.Store.SetRecord(record); err != nil {
			log.Error().Err(err).Str("address", address).Msg("Saving record")
		}

		RespondWithJSON(res, http.StatusOK, err.Error())
		return
	}

	record.TotalAutostakedRewards += value
	if err := h.bot.Store.SetRecord(record); err != nil {
		log.Error().Err(err).Str("address", address).Msg("Saving record")
	}

	RespondWithJSON(res, http.StatusOK, fmt.Sprintf("Successfully restaked %d tokens\n", value))
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
		return false, fmt.Errorf("feegrant allowance query: %w", err)
	}
	if resp.Allowance == nil {
		return false, fmt.Errorf("address %s is not covering the fees for stakebot (%s)", address, authority)
	}

	allowance := &feegrant.AllowedMsgAllowance{}
	err = proto.Unmarshal(resp.Allowance.Allowance.Value, allowance)
	if err != nil {
		return false, fmt.Errorf("expected to umarshal AllowedMsgAllowance: %w", err)
	}

	contains := false
	for _, msgType := range allowance.AllowedMessages {
		if msgType == sdk.MsgTypeURL(&authz.MsgExec{}) {
			contains = true
		}
	}
	if !contains {
		return false, fmt.Errorf("address %s does not cover authz.MsgExec fees for stakebot (%s)", address, authority)
	}

	authzClient := authz.NewQueryClient(conn)
	grantsResp, err := authzClient.Grants(ctx, &authz.QueryGrantsRequest{
		Granter:    address,
		Grantee:    authority,
		MsgTypeUrl: sdk.MsgTypeURL(&staking.MsgDelegate{}),
	})
	if err != nil {
		return false, fmt.Errorf("authorization MsgDelegate query: %w", err)
	}
	if len(grantsResp.Grants) == 0 {
		return false, fmt.Errorf("address %s must authorize the stakebot (%s) to MsgDelegate", address, authority)
	}

	grantsResp, err = authzClient.Grants(ctx, &authz.QueryGrantsRequest{
		Granter:    address,
		Grantee:    authority,
		MsgTypeUrl: sdk.MsgTypeURL(&distribution.MsgWithdrawDelegatorReward{}),
	})
	if err != nil {
		return false, fmt.Errorf("authorization MsgWithdrawDelegatorReward query: %w", err)
	}
	if len(grantsResp.Grants) == 0 {
		return false, fmt.Errorf("grants %s must authorize the stakebot (%s) to MsgWithdrawDelegatorReward", address, authority)
	}

	return true, nil
}
