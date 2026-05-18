package upload

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
	r.Route("/uploads", func(r chi.Router) {
		r.Use(tenantMiddleware)
		r.With(requirePermission(permission.ProductUploadImage)).Post("/", handler.Create)
	})
}
