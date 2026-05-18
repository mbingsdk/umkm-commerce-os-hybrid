package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
)

type tenantAccessValidator interface {
	ValidateAccess(ctx context.Context, userID uuid.UUID, tenantID uuid.UUID) (tenantctx.TenantContext, error)
}

func TenantResolver(validator tenantAccessValidator, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authCtx, ok := auth.FromContext(r.Context())
			if !ok {
				httpserver.WriteError(w, r, logger, apperror.Unauthorized("Authentication required"))
				return
			}

			rawTenantID := strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
			if rawTenantID == "" {
				httpserver.WriteError(w, r, logger, apperror.Validation("Validation failed", []map[string]string{
					{"field": "X-Tenant-ID", "message": "X-Tenant-ID header is required"},
				}))
				return
			}

			tenantID, err := uuid.Parse(rawTenantID)
			if err != nil {
				httpserver.WriteError(w, r, logger, apperror.Validation("Validation failed", []map[string]string{
					{"field": "X-Tenant-ID", "message": "X-Tenant-ID must be a valid UUID"},
				}))
				return
			}

			resolvedTenant, err := validator.ValidateAccess(r.Context(), authCtx.UserID, tenantID)
			if err != nil {
				httpserver.WriteError(w, r, logger, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(tenantctx.WithContext(r.Context(), resolvedTenant)))
		})
	}
}
