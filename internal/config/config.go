package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type Config struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	Debug            bool
	ApplicationPort  string
	LogLevel         string
}

func NewConfig() *Config {
	return &Config{
		PostgresHost:     ParseEnv("POSTGRES_HOST", true, "localhost"),
		PostgresPort:     ParseEnv("POSTGRES_PORT", true, "5432"),
		PostgresUser:     ParseEnv("POSTGRES_USER", true, "postgres"),
		PostgresPassword: ParseEnv("POSTGRES_PASSWORD", true, "postgres"),
		PostgresDB:       ParseEnv("POSTGRES_DB", true, "casino_wallet"),
		Debug:            ParseEnv("DEBUG", false, "false") == "true",
		ApplicationPort:  ParseEnv("APPLICATION_PORT", false, "8080"),
		LogLevel:         ParseEnv("LOG_LEVEL", false, "info"),
	}
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		c.PostgresHost, c.PostgresPort, c.PostgresUser, c.PostgresPassword, c.PostgresDB)
}

func ParseEnv(key string, required bool, dft string) string {
	_ = godotenv.Load()
	value := os.Getenv(key)
	if value == "" && required {
		zap.L().Panic("Environment variable not found",
			zap.String("variable name", key),
		)
	} else if value == "" {
		return dft
	}
	return value
}
