package admin

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func RegisterRoutes(
	r chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
	adminGuard func(http.Handler) http.Handler,
) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(authMiddleware)
		r.Use(adminGuard)

		r.Get("/me", handler.Me)
		r.Get("/tenants", handler.ListTenants)
		r.Get("/tenants/{tenantId}", handler.GetTenant)
		r.Patch("/tenants/{tenantId}/status", handler.UpdateTenantStatus)
		r.Patch("/tenants/{tenantId}/plan", handler.UpdateTenantPlan)
	})
}

func Guard(service *Service, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := auth.FromContext(r.Context())
			if !ok {
				httpserver.WriteError(w, r, logger, apperror.Unauthorized("Authentication required"))
				return
			}

			adminCtx, err := service.ValidateSuperAdmin(r.Context(), authCtx.UserID)
			if err != nil {
				httpserver.WriteError(w, r, logger, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(withContext(r.Context(), adminCtx)))
		})
	}
}
