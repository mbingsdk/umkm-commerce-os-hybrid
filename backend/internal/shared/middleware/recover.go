package middleware

import (
	"log/slog"
	"net/http"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered",
						"request_id", httpserver.RequestIDFromContext(r.Context()),
						"panic", recovered,
					)
					httpserver.WriteError(w, r, logger, apperror.New(
						apperror.CodeInternal,
						"Internal server error",
					))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
