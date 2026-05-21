package courier

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
	r.Route("/courier", func(r chi.Router) {
		r.Use(tenantMiddleware)
		r.Use(requireFeature(plans.FeatureCourier))

		r.With(requirePermission(permission.CourierReadZone)).Get("/zones", handler.ListZones)
		r.With(requirePermission(permission.CourierCreateZone)).Post("/zones", handler.CreateZone)
		r.With(requirePermission(permission.CourierUpdateZone)).Patch("/zones/{zoneId}", handler.UpdateZone)
		r.With(requirePermission(permission.CourierDeleteZone)).Delete("/zones/{zoneId}", handler.DeleteZone)
	})
}

func RegisterPublicRoutes(r chi.Router, handler *Handler) {
	r.Get("/public/stores/{storeSlug}/courier/zones", handler.ListPublicZones)
}
