package middleware

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
)

func TestTenantResolverRejectsMissingTenantHeader(t *testing.T) {
	userID := uuid.New()
	handler := TenantResolver(fakeTenantValidator{}, discardLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: userID}))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusBadRequest, apperror.CodeValidation)
}

func TestTenantResolverRejectsTenantNotOwnedByUser(t *testing.T) {
	userID := uuid.New()
	tenantID := uuid.New()
	handler := TenantResolver(fakeTenantValidator{
		err: apperror.TenantAccessDenied("Tenant access denied"),
	}, discardLogger())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	req.Header.Set("X-Tenant-ID", tenantID.String())
	req = req.WithContext(auth.WithContext(req.Context(), auth.AuthContext{UserID: userID}))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusForbidden, apperror.CodeTenantAccessDenied)
}

type fakeTenantValidator struct {
	err error
}

func (f fakeTenantValidator) ValidateAccess(context.Context, uuid.UUID, uuid.UUID) (tenantctx.TenantContext, error) {
	if f.err != nil {
		return tenantctx.TenantContext{}, f.err
	}
	return tenantctx.TenantContext{TenantID: uuid.New(), StoreID: uuid.New(), UserID: uuid.New(), Role: "owner"}, nil
}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func assertErrorCode(t *testing.T, res *httptest.ResponseRecorder, status int, code apperror.Code) {
	t.Helper()

	if res.Code != status {
		t.Fatalf("status = %d, want %d; body=%s", res.Code, status, res.Body.String())
	}

	var body struct {
		Error struct {
			Code apperror.Code `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error.Code != code {
		t.Fatalf("error code = %s, want %s", body.Error.Code, code)
	}
}
