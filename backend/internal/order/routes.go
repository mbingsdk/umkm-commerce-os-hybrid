package order

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
	r.Route("/orders", func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.OrderRead)).Get("/", handler.List)
		r.With(requirePermission(permission.OrderReadDetail)).Get("/{orderId}", handler.Detail)
		r.With(requirePermission(permission.OrderUpdateStatus)).Patch("/{orderId}/status", handler.UpdateStatus)
	})
}
