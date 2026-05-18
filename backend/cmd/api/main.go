package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/app"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/config"
	platformlogger "github.com/sdkdev/umkm-commerce-os/backend/internal/platform/logger"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := platformlogger.New(cfg.AppEnv)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	deps, err := app.NewDependencies(ctx, cfg, app.BuildInfo{
		Version:   version,
		Commit:    commit,
		BuildTime: buildTime,
	}, logger)
	if err != nil {
		logger.Error("initialize dependencies", "error", err)
		os.Exit(1)
	}
	defer deps.Close()

	api := app.NewAPI(deps)
	if err := api.Run(ctx); err != nil {
		logger.Error("run api", "error", err)
		os.Exit(1)
	}
}
