package inventory

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	tenantMiddleware func(http.Handler) http.Handler,
	requirePermission func(permission.Permission) func(http.Handler) http.Handler,
) {
	r.Route("/inventory", func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.InventoryRead)).Get("/stocks", handler.ListStocks)
		r.With(requirePermission(permission.InventoryReadMovement)).Get("/products/{productId}/movements", handler.ListProductMovements)
		r.With(requirePermission(permission.InventoryAdjust)).Post("/products/{productId}/adjust", handler.AdjustStock)
		r.With(requirePermission(permission.InventoryUpdateThreshold)).Patch("/products/{productId}/threshold", handler.UpdateThreshold)
	})
}
