package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/catalog/category"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/catalog/product"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/checkout"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/courier"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/dashboard"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/finance"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/payment"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/pos"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	sharedmw "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/middleware"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shipment"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/tenant"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/upload"
)

func NewRouter(deps *Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(sharedmw.RequestID)
	r.Use(sharedmw.Logger(deps.Logger))
	r.Use(sharedmw.Recover(deps.Logger))
	r.Use(sharedmw.CORS(deps.Config.CORSAllowedOrigins))
	r.Use(sharedmw.RateLimitPlaceholder)

	r.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		httpserver.WriteOK(w, "API process is alive", map[string]string{
			"status": "ok",
		})
	})

	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := deps.DB.Ping(ctx); err != nil {
			httpserver.WriteError(w, r, deps.Logger, apperror.ServiceUnavailable("API dependencies are not ready", err))
			return
		}

		httpserver.WriteOK(w, "API dependencies are ready", map[string]string{
			"status": "ok",
		})
	})

	r.Get("/version", func(w http.ResponseWriter, _ *http.Request) {
		httpserver.WriteJSON(w, http.StatusOK, map[string]string{
			"app":        deps.Config.AppName,
			"version":    deps.Build.Version,
			"commit":     deps.Build.Commit,
			"build_time": deps.Build.BuildTime,
		})
	})

	if deps.Config.AppEnv == "development" && deps.Config.StorageDriver == "local" {
		r.Mount("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(deps.Config.StorageLocalDir))))
	}

	r.Route("/api/v1", func(r chi.Router) {
		authMiddleware := sharedmw.Auth(deps.AccessTokens, deps.Logger)
		tenantMiddleware := sharedmw.TenantResolver(deps.TenantService, deps.Logger)
		requirePermission := func(required permission.Permission) func(http.Handler) http.Handler {
			return sharedmw.RequirePermission(required, deps.Logger)
		}

		auth.RegisterRoutes(r, deps.AuthHandler, authMiddleware)
		tenant.RegisterRoutes(r, deps.TenantHandler, authMiddleware)
		store.RegisterPublicRoutes(r, deps.PublicStore)
		category.RegisterPublicRoutes(r, deps.PublicCategory)
		product.RegisterPublicRoutes(r, deps.PublicProduct)
		checkout.RegisterPublicRoutes(r, deps.CheckoutHandler)
		payment.RegisterPublicRoutes(r, deps.PaymentHandler)
		courier.RegisterPublicRoutes(r, deps.CourierHandler)
		shipment.RegisterPublicRoutes(r, deps.ShipmentHandler)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			store.RegisterRoutes(r, deps.StoreHandler, tenantMiddleware, requirePermission)
			category.RegisterRoutes(r, deps.CategoryHandler, tenantMiddleware, requirePermission)
			product.RegisterRoutes(r, deps.ProductHandler, tenantMiddleware, requirePermission)
			upload.RegisterRoutes(r, deps.UploadHandler, tenantMiddleware, requirePermission)
			inventory.RegisterRoutes(r, deps.InventoryHandler, tenantMiddleware, requirePermission)
			pos.RegisterRoutes(r, deps.POSHandler, tenantMiddleware, requirePermission)
			finance.RegisterRoutes(r, deps.FinanceHandler, tenantMiddleware, requirePermission)
			dashboard.RegisterRoutes(r, deps.DashboardHandler, tenantMiddleware, requirePermission)
			courier.RegisterRoutes(r, deps.CourierHandler, tenantMiddleware, requirePermission)
			shipment.RegisterRoutes(r, deps.ShipmentHandler, tenantMiddleware, requirePermission)
			order.RegisterRoutes(r, deps.OrderHandler, tenantMiddleware, requirePermission)
			payment.RegisterRoutes(r, deps.PaymentHandler, tenantMiddleware, requirePermission)
		})
	})

	return r
}
