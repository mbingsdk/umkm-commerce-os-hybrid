package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
)

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *responseWriter) WriteHeader(status int) {
	if w.wroteHeader {
		return
	}

	w.status = status
	w.wroteHeader = true
	w.ResponseWriter.WriteHeader(status)
}

func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			latency := time.Since(start)
			route := routePattern(r)
			attrs := []any{
				"request_id", httpserver.RequestIDFromContext(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"route", route,
				"path_template", route,
				"status", wrapped.status,
				"duration_ms", latency.Milliseconds(),
				"ip", requestIP(r),
				"user_agent", r.UserAgent(),
			}
			if latency >= slowRequestThreshold(r) {
				logger.Warn("slow http request", attrs...)
				return
			}

			logger.Info("http request", attrs...)
		})
	}
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		if pattern := rctx.RoutePattern(); pattern != "" {
			return pattern
		}
	}
	return r.URL.Path
}

func slowRequestThreshold(r *http.Request) time.Duration {
	path := r.URL.Path
	if r.Method == http.MethodPost && strings.Contains(path, "/checkout") {
		return 1500 * time.Millisecond
	}
	if r.Method == http.MethodPost && path == "/api/v1/pos/transactions" {
		return 1000 * time.Millisecond
	}
	if strings.HasPrefix(path, "/api/v1/admin/") {
		return 1000 * time.Millisecond
	}
	if r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v1/public/") {
		return 500 * time.Millisecond
	}
	return 800 * time.Millisecond
}

func requestIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
