package checkout

import "github.com/go-chi/chi/v5"

func RegisterPublicRoutes(r chi.Router, handler *Handler) {
	r.Post("/public/stores/{storeSlug}/checkout", handler.Checkout)
}
