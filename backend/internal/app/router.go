package app

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	sharedmw "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/middleware"
)

func NewRouter(deps *Dependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(sharedmw.RequestID)
	r.Use(sharedmw.Logger(deps.Logger))
	r.Use(sharedmw.Recover(deps.Logger))
	r.Use(sharedmw.CORS(deps.Config.CORSAllowedOrigins))
	r.Use(sharedmw.RateLimitPlaceholder)

	r.Get("/health/live", func(w http.ResponseWriter, _ *http.Request) {
		httpserver.WriteOK(w, "API process is alive", map[string]string{
			"status": "ok",
		})
	})

	r.Get("/health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := deps.DB.Ping(ctx); err != nil {
			httpserver.WriteError(w, r, deps.Logger, apperror.ServiceUnavailable("API dependencies are not ready", err))
			return
		}

		httpserver.WriteOK(w, "API dependencies are ready", map[string]string{
			"status": "ok",
		})
	})

	r.Get("/version", func(w http.ResponseWriter, _ *http.Request) {
		httpserver.WriteJSON(w, http.StatusOK, map[string]string{
			"app":        deps.Config.AppName,
			"version":    deps.Build.Version,
			"commit":     deps.Build.Commit,
			"build_time": deps.Build.BuildTime,
		})
	})

	return r
}
