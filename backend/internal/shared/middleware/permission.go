package middleware

import (
	"log/slog"
	"net/http"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
)

func RequirePermission(required permission.Permission, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			currentTenant, ok := tenantctx.FromContext(r.Context())
			if !ok {
				httpserver.WriteError(w, r, logger, apperror.Forbidden("Tenant context is required"))
				return
			}

			if !permission.Allowed(currentTenant.Role, required) {
				httpserver.WriteError(w, r, logger, apperror.Forbidden("Insufficient permission"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
