package app

import (
	"fmt"
	"log/slog"
	"time"

	grpcapp "github.com/orenvadi/auth-grpc/internal/app/grpc"
	"github.com/orenvadi/auth-grpc/internal/services/auth"
	"github.com/orenvadi/auth-grpc/internal/storage/postgres"
	// "github.com/orenvadi/auth-grpc/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storageDSN string, tokenTTL time.Duration) *App {
	// DONE init storage

	// storage, err := sqlite.New(storagePath)
	storage, err := postgres.New(fmt.Sprintf("postgres://%s?sslmode=require", storageDSN))
	if err != nil {
		panic(err)
	}

	// DONE init auth server (auth)

	authService := auth.New(log, storage, storage, storage, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, storage, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
