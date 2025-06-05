package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/BarisKilicGsu/casino-wallet-service/models"
)

type HealthHandler struct {
	db *sql.DB
}

func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := models.HealthResponse{
		Status:    "ok",
		Database:  "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Veritabanı bağlantısını kontrol et
	err := h.db.Ping()
	if err != nil {
		response.Status = "error"
		response.Database = "error"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
