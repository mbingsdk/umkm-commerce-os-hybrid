package app

import (
	"context"
	"log/slog"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/config"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/password"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/token"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/tenant"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

type Dependencies struct {
	Config        config.Config
	Logger        *slog.Logger
	DB            *db.DB
	Build         BuildInfo
	AccessTokens  *token.JWTService
	AuthHandler   *auth.Handler
	TenantService *tenant.Service
	TenantHandler *tenant.Handler
	StoreHandler  *store.Handler
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
	tenantRepo := tenant.NewRepository()
	userTenantRepo := tenant.NewUserTenantRepository()
	storeRepo := store.NewRepository()
	auditRepo := audit.NewRepository()
	authService := auth.NewService(
		database,
		userRepo,
		refreshTokenRepo,
		password.NewBcryptHasher(),
		accessTokens,
		refreshTokens,
	)
	tenantService := tenant.NewService(database, tenantRepo, userTenantRepo, storeRepo, auditRepo)
	storeService := store.NewService(database, storeRepo, auditRepo)

	return &Dependencies{
		Config:        cfg,
		Logger:        logger,
		DB:            database,
		Build:         build,
		AccessTokens:  accessTokens,
		AuthHandler:   auth.NewHandler(authService, logger),
		TenantService: tenantService,
		TenantHandler: tenant.NewHandler(tenantService, logger),
		StoreHandler:  store.NewHandler(storeService, logger),
	}, nil
}

func (d *Dependencies) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
