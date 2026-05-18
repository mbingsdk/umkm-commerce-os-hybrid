package tenant

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware func(http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/onboarding/create-store", handler.CreateStore)
		r.Get("/tenants", handler.ListMyTenants)
	})
}
