package store

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
	r.Route("/stores/current", func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.StoreRead)).Get("/", handler.GetCurrent)
		r.With(requirePermission(permission.StoreUpdate)).Patch("/", handler.UpdateProfile)
		r.With(requirePermission(permission.StorePublish)).Post("/publish", handler.Publish)
		r.With(requirePermission(permission.StorePublish)).Post("/unpublish", handler.Unpublish)
		r.With(requirePermission(permission.StoreUpdateBusinessHours)).Put("/business-hours", handler.ReplaceBusinessHours)
	})
}
