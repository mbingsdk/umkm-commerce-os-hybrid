package category

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
	r.Route("/categories", func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.CategoryRead)).Get("/", handler.List)
		r.With(requirePermission(permission.CategoryCreate)).Post("/", handler.Create)
		r.With(requirePermission(permission.CategoryUpdate)).Patch("/{categoryId}", handler.Update)
		r.With(requirePermission(permission.CategoryDelete)).Delete("/{categoryId}", handler.Delete)
	})
}
