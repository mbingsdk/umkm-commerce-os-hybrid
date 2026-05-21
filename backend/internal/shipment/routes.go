package shipment

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
	r.Route("/", func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.ShipmentCreate)).Post("/orders/{orderId}/shipments", handler.Create)
		r.With(requirePermission(permission.ShipmentRead)).Get("/shipments", handler.List)
		r.With(requirePermission(permission.ShipmentRead)).Get("/shipments/{shipmentId}", handler.Detail)
		r.With(requirePermission(permission.ShipmentUpdateStatus)).Patch("/shipments/{shipmentId}/status", handler.UpdateStatus)
	})
}

func RegisterPublicRoutes(r chi.Router, handler *Handler) {
	r.Get("/public/stores/{storeSlug}/orders/{orderNumber}/tracking", handler.PublicTracking)
}
