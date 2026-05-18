package app

import (
	"context"
	"log/slog"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/config"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

type Dependencies struct {
	Config config.Config
	Logger *slog.Logger
	DB     *db.DB
	Build  BuildInfo
}

func NewDependencies(ctx context.Context, cfg config.Config, build BuildInfo, logger *slog.Logger) (*Dependencies, error) {
	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	return &Dependencies{
		Config: cfg,
		Logger: logger,
		DB:     database,
		Build:  build,
	}, nil
}

func (d *Dependencies) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
