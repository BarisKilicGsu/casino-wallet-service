package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BarisKilicGsu/casino-wallet-service/internal/config"
	"github.com/BarisKilicGsu/casino-wallet-service/internal/handler"
	"github.com/BarisKilicGsu/casino-wallet-service/internal/repository"
	"github.com/BarisKilicGsu/casino-wallet-service/internal/seed"
	"github.com/BarisKilicGsu/casino-wallet-service/internal/service"
	"github.com/BarisKilicGsu/casino-wallet-service/internal/utils/logger"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

func connectWithRetry(dsn string, maxRetries int) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: gormLogger.Default.LogMode(gormLogger.Silent),
		})
		if err == nil {
			return db, nil
		}
		zap.L().Warn("Failed to connect to PostgreSQL, retrying...",
			zap.Error(err),
			zap.Int("attempt", i+1),
			zap.Int("max_attempts", maxRetries))
		time.Sleep(5 * time.Second)
	}

	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}

func main() {
	// Load configuration
	cfg := config.NewConfig()

	// Initialize logger
	logger.InitLogger(cfg.LogLevel)
	zap.L().Info("Application starting")

	// PostgreSQL connection with retry
	db, err := connectWithRetry(cfg.GetDSN(), 5)
	if err != nil {
		zap.L().Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}

	// Get underlying *sql.DB for health check
	sqlDB, err := db.DB()
	if err != nil {
		zap.L().Fatal("Failed to get underlying *sql.DB", zap.Error(err))
	}

	// Add sample players
	if err := seed.SeedPlayers(db); err != nil {
		zap.L().Error("Error during seed operation", zap.Error(err))
	}

	gormRepository := repository.NewGormRepository(db)

	// Create repositories
	playerRepo := repository.NewPlayerRepository(gormRepository)
	transactionRepo := repository.NewTransactionRepository(gormRepository)

	// Create service
	walletService := service.NewWalletService(playerRepo, transactionRepo, gormRepository)

	// Create handlers
	walletHandler := handler.NewWalletHandler(walletService)
	healthHandler := handler.NewHealthHandler(sqlDB)

	// Set up router
	router := InitRouter(walletHandler, healthHandler)

	// Start HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", cfg.ApplicationPort),
		Handler: router,
	}

	// Create channel for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a separate goroutine
	go func() {
		zap.L().Info("Starting HTTP server", zap.String("port", fmt.Sprintf(":%v", cfg.ApplicationPort)))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-stop
	zap.L().Info("Shutdown signal received")

	// Create context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		zap.L().Error("Error while shutting down server", zap.Error(err))
	}

	zap.L().Info("Application closed")
}
