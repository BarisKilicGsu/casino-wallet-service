package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLogger(level string) *zap.Logger {
	var logger *zap.Logger
	var err error

	// Log level
	var logLevel zapcore.Level
	switch level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	default:
		logLevel = zapcore.InfoLevel
	}

	// Config
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err = config.Build()
	if err != nil {
		panic("Logger could not initialized")
	}
	zap.ReplaceGlobals(logger)
	return logger
}
