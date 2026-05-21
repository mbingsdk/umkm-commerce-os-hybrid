package app

import (
	"context"
	"log/slog"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/admin"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/catalog/category"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/catalog/product"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/checkout"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/config"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/courier"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/dashboard"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/discovery"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/finance"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/payment"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/password"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/storage"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/token"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/pos"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shipment"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/tenant"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/upload"
)

type BuildInfo struct {
	Version   string
	Commit    string
	BuildTime string
}

type Dependencies struct {
	Config           config.Config
	Logger           *slog.Logger
	DB               *db.DB
	Build            BuildInfo
	AccessTokens     *token.JWTService
	AuthHandler      *auth.Handler
	TenantService    *tenant.Service
	TenantHandler    *tenant.Handler
	StoreHandler     *store.Handler
	PublicStore      *store.PublicHandler
	CategoryHandler  *category.Handler
	PublicCategory   *category.PublicHandler
	ProductHandler   *product.Handler
	PublicProduct    *product.PublicHandler
	UploadHandler    *upload.Handler
	CheckoutHandler  *checkout.Handler
	OrderHandler     *order.Handler
	PaymentHandler   *payment.Handler
	InventoryHandler *inventory.Handler
	POSHandler       *pos.Handler
	FinanceHandler   *finance.Handler
	DashboardHandler *dashboard.Handler
	CourierHandler   *courier.Handler
	ShipmentHandler  *shipment.Handler
	DiscoveryHandler *discovery.Handler
	AdminService     *admin.Service
	AdminHandler     *admin.Handler
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
	categoryRepo := category.NewRepository()
	productRepo := product.NewRepository()
	productImageRepo := product.NewImageRepository()
	inventoryRepo := inventory.NewRepository()
	checkoutRepo := checkout.NewRepository()
	orderRepo := order.NewRepository()
	paymentRepo := payment.NewRepository()
	posRepo := pos.NewRepository()
	financeRepo := finance.NewRepository()
	dashboardRepo := dashboard.NewRepository()
	courierRepo := courier.NewRepository()
	shipmentRepo := shipment.NewRepository()
	discoveryRepo := discovery.NewRepository()
	adminRepo := admin.NewRepository()
	idempotencyRepo := idempotency.NewRepository()
	outboxRepo := outbox.NewRepository()
	assetStore := storage.NewLocal(cfg.StorageLocalDir, cfg.StoragePublicURL, cfg.UploadMaxBytes)
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
	publicStoreService := store.NewPublicService(database, storeRepo)
	categoryService := category.NewService(database, categoryRepo)
	publicCategoryService := category.NewPublicService(database, categoryRepo, publicStoreService)
	productService := product.NewService(database, productRepo, categoryRepo, inventoryRepo, productImageRepo, assetStore)
	publicProductService := product.NewPublicService(database, productRepo, publicStoreService)
	uploadService := upload.NewService(assetStore)
	checkoutService := checkout.NewService(database, publicStoreService, checkoutRepo, idempotencyRepo, outboxRepo)
	orderService := order.NewService(database, orderRepo, outboxRepo)
	paymentService := payment.NewService(database, publicStoreService, paymentRepo, idempotencyRepo, outboxRepo)
	inventoryService := inventory.NewService(database, inventoryRepo, auditRepo, outboxRepo)
	posService := pos.NewService(database, posRepo, auditRepo, idempotencyRepo, outboxRepo)
	financeService := finance.NewService(database, financeRepo, auditRepo, outboxRepo)
	dashboardService := dashboard.NewService(database, dashboardRepo)
	courierService := courier.NewService(database, courierRepo, publicStoreService, auditRepo)
	shipmentService := shipment.NewService(database, shipmentRepo, publicStoreService, outboxRepo)
	discoveryService := discovery.NewService(database, discoveryRepo)
	adminService := admin.NewService(database, adminRepo, outboxRepo)

	return &Dependencies{
		Config:           cfg,
		Logger:           logger,
		DB:               database,
		Build:            build,
		AccessTokens:     accessTokens,
		AuthHandler:      auth.NewHandler(authService, logger),
		TenantService:    tenantService,
		TenantHandler:    tenant.NewHandler(tenantService, logger),
		StoreHandler:     store.NewHandler(storeService, logger),
		PublicStore:      store.NewPublicHandler(publicStoreService, logger),
		CategoryHandler:  category.NewHandler(categoryService, logger),
		PublicCategory:   category.NewPublicHandler(publicCategoryService, logger),
		ProductHandler:   product.NewHandler(productService, logger, cfg.UploadMaxBytes),
		PublicProduct:    product.NewPublicHandler(publicProductService, logger),
		UploadHandler:    upload.NewHandler(uploadService, logger, cfg.UploadMaxBytes),
		CheckoutHandler:  checkout.NewHandler(checkoutService, logger),
		OrderHandler:     order.NewHandler(orderService, logger),
		PaymentHandler:   payment.NewHandler(paymentService, logger),
		InventoryHandler: inventory.NewHandler(inventoryService, logger),
		POSHandler:       pos.NewHandler(posService, logger),
		FinanceHandler:   finance.NewHandler(financeService, logger),
		DashboardHandler: dashboard.NewHandler(dashboardService, logger),
		CourierHandler:   courier.NewHandler(courierService, logger),
		ShipmentHandler:  shipment.NewHandler(shipmentService, logger),
		DiscoveryHandler: discovery.NewHandler(discoveryService, logger),
		AdminService:     adminService,
		AdminHandler:     admin.NewHandler(adminService, logger),
	}, nil
}

func (d *Dependencies) Close() {
	if d.DB != nil {
		d.DB.Close()
	}
}
