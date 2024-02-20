package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/orenvadi/auth-grpc/internal/app"
	"github.com/orenvadi/auth-grpc/internal/config"
)

func main() {
	// DONE init Config object
	cfg := config.MustLoad()

	fmt.Printf("\nServer config: \n%+v\n\n\n", cfg)

	// DONE init Logger
	log := setupLogger(cfg.Env)

	log.Info("starting application", slog.String("env", cfg.Env))

	// DONE init Application (app)

	application := app.New(log, cfg.GRPC.Port, cfg.Storage.DSN(), cfg.TokenTTL)

	// DONE run gRPC-server of the app

	// GraceFull shutdouwn
	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	// reading from chan is blocking operaiton
	sgnl := <-stop

	log.Info("stopping application", slog.String("signal", sgnl.String()))

	application.GRPCSrv.Stop()

	log.Info("application stopped")
}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
