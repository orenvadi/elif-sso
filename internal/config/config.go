package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env      string // No need for env-default tag, we'll set defaults directly in the code
	Storage  Storage
	TokenTTL time.Duration
	GRPC     GRPC
}

type Storage struct {
	User     string
	Password string
	Host     string
	DbName   string
}

func (s Storage) DSN() string {
	return fmt.Sprintf("%s:%s@%s/%s", s.User, s.Password, s.Host, s.DbName)
}

type GRPC struct {
	Port    int
	Timeout time.Duration
}

func MustLoad() *Config {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %v", err))
	}

	var cfg Config

	// Retrieve environment variables and set defaults
	cfg.Env = os.Getenv("ENV")
	if cfg.Env == "" {
		cfg.Env = "local" // Default value if ENV is not set
	}

	cfg.Storage.User = os.Getenv("STORAGE_USER")
	if cfg.Storage.User == "" {
		cfg.Storage.User = "postgres" // Default value if STORAGE_USER is not set
	}
	cfg.Storage.Password = os.Getenv("STORAGE_PASSWORD")
	if cfg.Storage.Password == "" {
		cfg.Storage.Password = "postgres" // Default value if STORAGE_PASSWORD is not set
	}
	cfg.Storage.Host = os.Getenv("STORAGE_HOST")
	if cfg.Storage.Host == "" {
		cfg.Storage.Host = "localhost" // Default value if STORAGE_HOST is not set
	}
	cfg.Storage.DbName = os.Getenv("STORAGE_DB_NAME")
	if cfg.Storage.DbName == "" {
		panic("STORAGE_DB_NAME environment variable is required")
	}

	tokenTTLStr := os.Getenv("TOKEN_TTL")
	if tokenTTLStr == "" {
		panic("TOKEN_TTL environment variable is required")
	}
	tokenTTL, err := time.ParseDuration(tokenTTLStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TOKEN_TTL: %s", err))
	}
	cfg.TokenTTL = tokenTTL

	grpcPortStr := os.Getenv("GRPC_PORT")
	if grpcPortStr == "" {
		panic("GRPC_PORT environment variable is required")
	}
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse GRPC_PORT: %s", err))
	}
	cfg.GRPC.Port = grpcPort

	grpcTimeoutStr := os.Getenv("GRPC_TIMEOUT")
	if grpcTimeoutStr == "" {
		panic("GRPC_TIMEOUT environment variable is required")
	}
	grpcTimeout, err := time.ParseDuration(grpcTimeoutStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse GRPC_TIMEOUT: %s", err))
	}
	cfg.GRPC.Timeout = grpcTimeout

	return &cfg
}
