package main

import (
	"net/http"

	"github.com/BarisKilicGsu/casino-wallet-service/internal/handler"
	"github.com/gorilla/mux"
)

func InitRouter(walletHandler *handler.WalletHandler, healthHandler *handler.HealthHandler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	router.HandleFunc("/wallet/{player_id}", walletHandler.GetPlayerBalance).Methods(http.MethodGet)
	router.HandleFunc("/players", walletHandler.GetAllPlayers).Methods(http.MethodGet)
	router.HandleFunc("/event", walletHandler.ProcessEvent).Methods(http.MethodPost)
	router.HandleFunc("/health", healthHandler.HealthCheck).Methods(http.MethodGet)

	return router
}
