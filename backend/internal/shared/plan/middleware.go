package plan

import (
	"log/slog"
	"net/http"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
)

func RequireFeature(service *Service, feature Feature, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if service == nil {
				httpserver.WriteError(w, r, logger, apperror.Internal(nil))
				return
			}

			currentTenant, ok := tenantctx.FromContext(r.Context())
			if !ok {
				httpserver.WriteError(w, r, logger, apperror.Validation("Validation failed", []map[string]string{
					{"field": "tenant", "message": "Tenant context is required"},
				}))
				return
			}

			if err := service.RequireFeature(r.Context(), service.db, currentTenant.TenantID, feature); err != nil {
				httpserver.WriteError(w, r, logger, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
