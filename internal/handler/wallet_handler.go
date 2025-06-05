package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/BarisKilicGsu/casino-wallet-service/internal/entities"
	"github.com/BarisKilicGsu/casino-wallet-service/internal/service"
	httpUtils "github.com/BarisKilicGsu/casino-wallet-service/internal/utils/http"
	"github.com/BarisKilicGsu/casino-wallet-service/models"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type WalletHandler struct {
	walletService service.IWalletService
}

func NewWalletHandler(walletService service.IWalletService) *WalletHandler {
	return &WalletHandler{
		walletService: walletService,
	}
}

func (h *WalletHandler) GetPlayerBalance(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("Received get player balance request")

	playerID := mux.Vars(r)["player_id"]
	if playerID == "" {
		zap.L().Warn("Missing player_id parameter in request")
		httpUtils.ErrorResponse(w, http.StatusBadRequest, service.ErrInvalidRequest)
		return
	}

	player, err := h.walletService.GetPlayerBalance(playerID)
	if err != nil {
		zap.L().Error("Error while getting player balance",
			zap.String("player_id", playerID),
			zap.Error(err))
		httpUtils.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	httpUtils.JSONResponse(w, http.StatusOK, player.ToApiResponse())
	zap.L().Info("Successfully returned player balance",
		zap.String("player_id", playerID))
}

func (h *WalletHandler) GetAllPlayers(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("Received get all players request")

	players, err := h.walletService.GetAllPlayers()
	if err != nil {
		zap.L().Error("Error while getting all players", zap.Error(err))
		httpUtils.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	apiPlayers := make([]*models.PlayerResponse, len(players))
	for i, player := range players {
		apiPlayers[i] = player.ToApiResponse()
	}

	httpUtils.JSONResponse(w, http.StatusOK, models.AllPlayersResponse{
		Players: apiPlayers,
	})
	zap.L().Info("Successfully returned all players",
		zap.Int("player_count", len(players)))
}

func (h *WalletHandler) ProcessEvent(w http.ResponseWriter, r *http.Request) {
	zap.L().Debug("Received process event request")

	var transactionRequest models.EventRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		zap.L().Info("Failed to read request body",
			zap.String("url path", r.URL.Path), zap.Error(err))
		httpUtils.ErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	bodyReader := bytes.NewReader(body)
	if err := json.NewDecoder(bodyReader).Decode(&transactionRequest); err != nil {
		zap.L().Info("Failed to read request body because transactionRequest could not be decoded",
			zap.ByteString("request body", body),
			zap.String("url path", r.URL.Path),
			zap.Error(err),
		)
		httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	err = transactionRequest.Validate(strfmt.Default)
	if err != nil {
		zap.L().Info("Failed to read request body because validation failed on transactionRequest",
			zap.Any("Request", transactionRequest),
			zap.String("url path", r.URL.Path),
			zap.Error(err),
		)
		httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	transaction := entities.Transaction{}
	transaction.CreateFromEventRequest(transactionRequest)

	err = h.walletService.ProcessTransaction(&transaction)
	if err != nil {
		zap.L().Error("Error while processing transaction",
			zap.String("req_id", transaction.ReqID),
			zap.String("type", string(transaction.Type)),
			zap.Error(err))
		switch err {
		case service.ErrInsufficientBalance:
			httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrDuplicateRequest:
			httpUtils.ErrorResponse(w, http.StatusConflict, err)
		case service.ErrBetNotFound:
			httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrGameCodeMismatch:
			httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrWalletIDMismatch:
			httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrPlayerIDMismatch:
			httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		case service.ErrDuplicateRound:
			httpUtils.ErrorResponse(w, http.StatusBadRequest, err)
		default:
			httpUtils.ErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	httpUtils.JSONResponseNoData(w, http.StatusOK)
	zap.L().Info("Successfully processed transaction",
		zap.String("req_id", transaction.ReqID),
		zap.String("type", string(transaction.Type)))
}
