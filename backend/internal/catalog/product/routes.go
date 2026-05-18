package product

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
	r.Route("/products", func(r chi.Router) {
		r.Use(tenantMiddleware)

		r.With(requirePermission(permission.ProductRead)).Get("/", handler.List)
		r.With(requirePermission(permission.ProductCreate)).Post("/", handler.Create)
		r.With(requirePermission(permission.ProductRead)).Get("/{productId}", handler.Get)
		r.With(requirePermission(permission.ProductUpdate)).Patch("/{productId}", handler.Update)
		r.With(requirePermission(permission.ProductDelete)).Delete("/{productId}", handler.Delete)
	})
}
