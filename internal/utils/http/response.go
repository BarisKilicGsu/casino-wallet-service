package http

import (
	"encoding/json"
	"net/http"

	"github.com/BarisKilicGsu/casino-wallet-service/models"
	"go.uber.org/zap"
)

// JSONResponse sends a JSON response with the given status code and data
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		zap.L().Error("Error encoding JSON response", zap.Error(err))
	}
}

func JSONResponseNoData(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(models.SuccessResponse{
		Success: true,
	}); err != nil {
		zap.L().Error("Error encoding JSON response", zap.Error(err))
	}
}

// ErrorResponse sends a JSON error response with the given status code and error message
func ErrorResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := models.SuccessResponse{
		Success: false,
		Error:   err.Error(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		zap.L().Error("Error encoding JSON error response", zap.Error(err))
	}
}
