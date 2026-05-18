package store

import "github.com/go-chi/chi/v5"

func RegisterPublicRoutes(r chi.Router, handler *PublicHandler) {
	r.Get("/public/stores/{storeSlug}", handler.Get)
}
