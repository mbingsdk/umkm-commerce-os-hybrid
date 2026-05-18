package app

import (
	"context"
	"log/slog"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/config"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/password"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/token"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

type Dependencies struct {
	Config       config.Config
	Logger       *slog.Logger
	DB           *db.DB
	Build        BuildInfo
	AccessTokens *token.JWTService
	AuthHandler  *auth.Handler
}

func NewDependencies(ctx context.Context, cfg config.Config, build BuildInfo, logger *slog.Logger) (*Dependencies, error) {
	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	accessTokens := token.NewJWTService(cfg.JWTSecret, cfg.AccessTokenTTL)
	refreshTokens := token.NewRefreshTokenService(cfg.RefreshTokenTTL)
	userRepo := auth.NewUserRepository()
	refreshTokenRepo := auth.NewRefreshTokenRepository()
	authService := auth.NewService(
		database,
		userRepo,
		refreshTokenRepo,
		password.NewBcryptHasher(),
		accessTokens,
		refreshTokens,
	)

	return &Dependencies{
		Config:       cfg,
		Logger:       logger,
		DB:           database,
		Build:        build,
		AccessTokens: accessTokens,
		AuthHandler:  auth.NewHandler(authService, logger),
	}, nil
}

func (d *Dependencies) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
