package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/token"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type accessTokenParser interface {
	Parse(raw string) (*token.AccessClaims, error)
}

func Auth(parser accessTokenParser, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken, ok := bearerToken(r.Header.Get("Authorization"))
			if !ok {
				httpserver.WriteError(w, r, logger, apperror.Unauthorized("Authentication required"))
				return
			}

			claims, err := parser.Parse(rawToken)
			if err != nil {
				httpserver.WriteError(w, r, logger, apperror.Unauthorized("Invalid access token"))
				return
			}

			userID, err := uuid.Parse(claims.UserID)
			if err != nil {
				httpserver.WriteError(w, r, logger, apperror.Unauthorized("Invalid access token"))
				return
			}

			ctx := auth.WithContext(r.Context(), auth.AuthContext{
				UserID:       userID,
				PlatformRole: claims.PlatformRole,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	return parts[1], true
}
