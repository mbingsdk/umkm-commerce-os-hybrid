package pos

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	plans "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/plan"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	tenantMiddleware func(http.Handler) http.Handler,
	requirePermission func(permission.Permission) func(http.Handler) http.Handler,
	requireFeature func(plans.Feature) func(http.Handler) http.Handler,
) {
	r.Route("/pos", func(r chi.Router) {
		r.Use(tenantMiddleware)
		r.Use(requireFeature(plans.FeaturePOS))

		r.With(requirePermission(permission.POSOpenSession)).Post("/sessions/open", handler.OpenSession)
		r.With(requirePermission(permission.POSReadSession)).Get("/sessions/current", handler.CurrentSession)
		r.With(requirePermission(permission.POSCloseSession)).Post("/sessions/{sessionId}/close", handler.CloseSession)
		r.With(requirePermission(permission.POSReadProduct)).Get("/products", handler.ListProducts)
		r.With(requirePermission(permission.POSCreateTransaction)).Post("/transactions", handler.CreateTransaction)
		r.With(requirePermission(permission.POSReadTransaction)).Get("/transactions", handler.ListTransactions)
		r.With(requirePermission(permission.POSReadTransaction)).Get("/transactions/{transactionId}", handler.TransactionDetail)
	})
}
