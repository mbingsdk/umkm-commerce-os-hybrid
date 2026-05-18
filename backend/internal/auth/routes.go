package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(r chi.Router, handler *Handler, authMiddleware func(http.Handler) http.Handler) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", handler.Register)
		r.Post("/login", handler.Login)
		r.Post("/refresh", handler.Refresh)

		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/logout", handler.Logout)
			r.Get("/me", handler.Me)
		})
	})
}
