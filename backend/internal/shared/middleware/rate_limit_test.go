package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestRateLimiterTriggersForLogin(t *testing.T) {
	limiter := NewRateLimiter(RateLimitPolicy{
		Name:    "login_test",
		Method:  http.MethodPost,
		Pattern: regexp.MustCompile(`^/api/v1/auth/login$`),
		Limit:   2,
		Window:  time.Minute,
	})
	limiter.now = func() time.Time {
		return time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	}

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < 2; i++ {
		res := performRateLimitedRequest(handler, http.MethodPost, "/api/v1/auth/login", "203.0.113.10:12345")
		if res.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, want 200", i+1, res.Code)
		}
	}

	res := performRateLimitedRequest(handler, http.MethodPost, "/api/v1/auth/login", "203.0.113.10:12345")
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("third request status = %d, want 429", res.Code)
	}

	var body struct {
		Error struct {
			Code apperror.Code `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != apperror.CodeRateLimited {
		t.Fatalf("error code = %s, want %s", body.Error.Code, apperror.CodeRateLimited)
	}
}

func TestRateLimiterTriggersForPublicCheckoutPattern(t *testing.T) {
	limiter := NewRateLimiter(RateLimitPolicy{
		Name:    "checkout_test",
		Method:  http.MethodPost,
		Pattern: regexp.MustCompile(`^/api/v1/public/stores/[^/]+/checkout$`),
		Limit:   1,
		Window:  time.Minute,
	})
	limiter.now = func() time.Time {
		return time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	}

	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	first := performRateLimitedRequest(handler, http.MethodPost, "/api/v1/public/stores/toko-bunga/checkout", "203.0.113.20:12345")
	if first.Code != http.StatusCreated {
		t.Fatalf("first checkout status = %d, want 201", first.Code)
	}

	second := performRateLimitedRequest(handler, http.MethodPost, "/api/v1/public/stores/toko-bunga/checkout", "203.0.113.20:12345")
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second checkout status = %d, want 429", second.Code)
	}
}

func TestCORSDoesNotAllowUnknownOrigin(t *testing.T) {
	handler := CORS([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want empty", got)
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	handler := CORS([]string{"https://app.example.com"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/auth/login", nil)
	req.Header.Set("Origin", "https://app.example.com")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q, want configured origin", got)
	}
}

func performRateLimitedRequest(handler http.Handler, method string, path string, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = remoteAddr
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}
