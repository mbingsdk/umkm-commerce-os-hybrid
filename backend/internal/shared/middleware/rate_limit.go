package middleware

import (
	"math"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type RateLimitPolicy struct {
	Name    string
	Method  string
	Pattern *regexp.Regexp
	Limit   int
	Window  time.Duration
}

type RateLimiter struct {
	mu       sync.Mutex
	now      func() time.Time
	buckets  map[string]rateLimitBucket
	policies []RateLimitPolicy
}

type rateLimitBucket struct {
	WindowStart time.Time
	Count       int
}

func CriticalRateLimitPolicies() []RateLimitPolicy {
	return []RateLimitPolicy{
		{
			Name:    "auth_login",
			Method:  http.MethodPost,
			Pattern: regexp.MustCompile(`^/api/v1/auth/login$`),
			Limit:   5,
			Window:  time.Minute,
		},
		{
			Name:    "auth_register",
			Method:  http.MethodPost,
			Pattern: regexp.MustCompile(`^/api/v1/auth/register$`),
			Limit:   3,
			Window:  time.Minute,
		},
		{
			Name:    "public_checkout",
			Method:  http.MethodPost,
			Pattern: regexp.MustCompile(`^/api/v1/public/stores/[^/]+/checkout$`),
			Limit:   20,
			Window:  time.Minute,
		},
		{
			Name:    "pos_transaction",
			Method:  http.MethodPost,
			Pattern: regexp.MustCompile(`^/api/v1/pos/transactions$`),
			Limit:   60,
			Window:  time.Minute,
		},
	}
}

func NewRateLimiter(policies ...RateLimitPolicy) *RateLimiter {
	return &RateLimiter{
		now:      time.Now,
		buckets:  make(map[string]rateLimitBucket),
		policies: policies,
	}
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		policy, ok := l.match(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		allowed, retryAfter := l.allow(policy, clientAddress(r))
		if !allowed {
			if retryAfter > 0 {
				w.Header().Set("Retry-After", strconv.Itoa(int(math.Ceil(retryAfter.Seconds()))))
			}
			httpserver.WriteJSON(w, http.StatusTooManyRequests, httpserver.ErrorResponse{
				Success: false,
				Message: "Too many requests. Please try again later.",
				Error: httpserver.ErrorPayload{
					Code: apperror.CodeRateLimited,
				},
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (l *RateLimiter) match(r *http.Request) (RateLimitPolicy, bool) {
	for _, policy := range l.policies {
		if policy.Limit <= 0 || policy.Window <= 0 || policy.Pattern == nil {
			continue
		}
		if r.Method == policy.Method && policy.Pattern.MatchString(r.URL.Path) {
			return policy, true
		}
	}
	return RateLimitPolicy{}, false
}

func (l *RateLimiter) allow(policy RateLimitPolicy, client string) (bool, time.Duration) {
	now := l.now().UTC()
	key := policy.Name + ":" + client

	l.mu.Lock()
	defer l.mu.Unlock()

	bucket, ok := l.buckets[key]
	if !ok || now.Sub(bucket.WindowStart) >= policy.Window {
		l.buckets[key] = rateLimitBucket{
			WindowStart: now,
			Count:       1,
		}
		l.cleanup(now)
		return true, 0
	}

	if bucket.Count >= policy.Limit {
		return false, policy.Window - now.Sub(bucket.WindowStart)
	}

	bucket.Count++
	l.buckets[key] = bucket
	return true, 0
}

func (l *RateLimiter) cleanup(now time.Time) {
	for key, bucket := range l.buckets {
		if now.Sub(bucket.WindowStart) > 10*time.Minute {
			delete(l.buckets, key)
		}
	}
}

func clientAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	if r.RemoteAddr != "" {
		return r.RemoteAddr
	}
	return "unknown"
}
