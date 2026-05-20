package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/config"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	platformlogger "github.com/sdkdev/umkm-commerce-os/backend/internal/platform/logger"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
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

	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("initialize worker database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	worker := outbox.NewWorker(
		database,
		outbox.NewRepository(),
		outbox.NewDefaultDispatcher(),
		logger,
		outbox.WorkerConfig{
			PollInterval: cfg.WorkerPollInterval,
			BatchSize:    cfg.WorkerBatchSize,
			MaxAttempts:  cfg.WorkerMaxAttempts,
			RetryDelay:   cfg.WorkerRetryDelay,
		},
	)

	if err := worker.Run(ctx); err != nil {
		logger.Error("run outbox worker", "error", err)
		os.Exit(1)
	}
}
