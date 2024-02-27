package config

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env      string
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
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error reading config file: %s", err))
	}

	var cfg Config

	cfg.Env = viper.GetString("ENV")
	if cfg.Env == "" {
		cfg.Env = "local"
	}

	cfg.Storage.User = viper.GetString("STORAGE_USER")
	if cfg.Storage.User == "" {
		cfg.Storage.User = "postgres"
	}

	cfg.Storage.Password = viper.GetString("STORAGE_PASSWORD")
	if cfg.Storage.Password == "" {
		cfg.Storage.Password = "postgres"
	}

	cfg.Storage.Host = viper.GetString("STORAGE_HOST")
	if cfg.Storage.Host == "" {
		cfg.Storage.Host = "localhost"
	}

	cfg.Storage.DbName = viper.GetString("STORAGE_DB_NAME")
	if cfg.Storage.DbName == "" {
		panic("STORAGE_DB_NAME environment variable is required")
	}

	tokenTTLStr := viper.GetString("TOKEN_TTL")
	if tokenTTLStr == "" {
		panic("TOKEN_TTL environment variable is required")
	}
	tokenTTL, err := time.ParseDuration(tokenTTLStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse TOKEN_TTL: %s", err))
	}
	cfg.TokenTTL = tokenTTL

	grpcPortStr := viper.GetString("GRPC_PORT")
	if grpcPortStr == "" {
		panic("GRPC_PORT environment variable is required")
	}
	grpcPort, err := strconv.Atoi(grpcPortStr)
	if err != nil {
		panic(fmt.Sprintf("failed to parse GRPC_PORT: %s", err))
	}
	cfg.GRPC.Port = grpcPort

	grpcTimeoutStr := viper.GetString("GRPC_TIMEOUT")
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
