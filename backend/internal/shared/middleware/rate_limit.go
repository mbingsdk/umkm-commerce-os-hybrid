package middleware

import "net/http"

// RateLimitPlaceholder keeps the middleware slot explicit without enforcing limits yet.
// Real policies for login, checkout, and public routes belong in a later sprint.
func RateLimitPlaceholder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}
