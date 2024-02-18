package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/orenvadi/auth-grpc/internal/app/grpc"
	"github.com/orenvadi/auth-grpc/internal/services/auth"
	"github.com/orenvadi/auth-grpc/internal/storage/sqlite"
)

type App struct {
	GRPCSrv *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, tokenTTL time.Duration) *App {
	// DONE init storage

	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	// DONE init auth servire (auth)

	authService := auth.New(log, storage, storage, storage, tokenTTL)

	grpcApp := grpcapp.New(log, authService, grpcPort)

	return &App{
		GRPCSrv: grpcApp,
	}
}
