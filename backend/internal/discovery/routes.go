package discovery

import "github.com/go-chi/chi/v5"

func RegisterPublicRoutes(r chi.Router, handler *Handler) {
	r.Get("/public/discovery/home", handler.Home)
	r.Get("/public/discovery/stores", handler.ListStores)
	r.Get("/public/discovery/products", handler.ListProducts)
	r.Get("/public/discovery/search", handler.Search)
}
