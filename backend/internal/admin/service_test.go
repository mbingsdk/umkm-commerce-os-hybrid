package admin

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/token"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	sharedmw "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/middleware"
)

var (
	testSuperAdminID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testOwnerID      = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testManagerID    = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testTargetID     = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

func TestAdminRoutesRejectUnauthenticatedRequest(t *testing.T) {
	router, _, _ := newAdminTestRouter(t, map[uuid.UUID]User{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusUnauthorized)
	}
}

func TestAdminRoutesRejectTenantRoles(t *testing.T) {
	users := map[uuid.UUID]User{
		testOwnerID: {
			ID:           testOwnerID,
			Name:         "Tenant Owner",
			Email:        "owner@example.test",
			PlatformRole: auth.PlatformRoleUser,
			Status:       auth.UserStatusActive,
		},
		testManagerID: {
			ID:           testManagerID,
			Name:         "Tenant Manager",
			Email:        "manager@example.test",
			PlatformRole: auth.PlatformRoleUser,
			Status:       auth.UserStatusActive,
		},
	}
	router, tokens, _ := newAdminTestRouter(t, users)

	for _, tc := range []struct {
		name   string
		userID uuid.UUID
	}{
		{name: "tenant owner", userID: testOwnerID},
		{name: "tenant manager", userID: testManagerID},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, tc.userID, auth.PlatformRoleSuperAdmin))
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != http.StatusForbidden {
				t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
			}
		})
	}
}

func TestAdminRoutesAcceptSuperAdminWithoutTenantHeader(t *testing.T) {
	router, tokens, _ := newAdminTestRouter(t, map[uuid.UUID]User{
		testSuperAdminID: {
			ID:           testSuperAdminID,
			Name:         "Platform Admin",
			Email:        "admin@example.test",
			PlatformRole: auth.PlatformRoleSuperAdmin,
			Status:       auth.UserStatusActive,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, testSuperAdminID, auth.PlatformRoleSuperAdmin))
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			User AdminUserResponse `json:"user"`
		} `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Data.User.PlatformRole != auth.PlatformRoleSuperAdmin {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestAdminRoutesRejectInactiveSuperAdmin(t *testing.T) {
	router, tokens, _ := newAdminTestRouter(t, map[uuid.UUID]User{
		testSuperAdminID: {
			ID:           testSuperAdminID,
			Name:         "Inactive Admin",
			Email:        "inactive-admin@example.test",
			PlatformRole: auth.PlatformRoleSuperAdmin,
			Status:       "suspended",
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, testSuperAdminID, auth.PlatformRoleSuperAdmin))
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
}

func TestRecordAuditCreatesAdminAuditLog(t *testing.T) {
	repo := &fakeAdminRepository{
		users: map[uuid.UUID]User{},
	}
	service := NewService(nil, repo)

	log, err := service.RecordAudit(context.Background(), AuditEntry{
		ActorUserID: testSuperAdminID,
		Action:      "admin.tenant.update_status",
		TargetType:  "tenant",
		TargetID:    &testTargetID,
		BeforeData:  map[string]string{"status": "active"},
		AfterData:   map[string]string{"status": "suspended"},
		IPAddress:   "127.0.0.1",
		UserAgent:   "admin-test",
	})
	if err != nil {
		t.Fatalf("record audit returned error: %v", err)
	}
	if len(repo.auditLogs) != 1 {
		t.Fatalf("audit logs = %d, want 1", len(repo.auditLogs))
	}
	if log.Action != "admin.tenant.update_status" || log.TargetID == nil || *log.TargetID != testTargetID {
		t.Fatalf("unexpected audit log: %#v", log)
	}
}

func TestRecordAuditRequiresAction(t *testing.T) {
	service := NewService(nil, &fakeAdminRepository{})

	_, err := service.RecordAudit(context.Background(), AuditEntry{ActorUserID: testSuperAdminID})
	assertAppErrorCode(t, err, apperror.CodeValidation)
}

func newAdminTestRouter(t *testing.T, users map[uuid.UUID]User) (http.Handler, *token.JWTService, *fakeAdminRepository) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokens := token.NewJWTService("test-secret", time.Hour)
	repo := &fakeAdminRepository{users: users}
	service := NewService(nil, repo)
	handler := NewHandler(service, logger)

	router := chi.NewRouter()
	router.Route("/api/v1", func(r chi.Router) {
		RegisterRoutes(r, handler, sharedmw.Auth(tokens, logger), Guard(service, logger))
	})

	return router, tokens, repo
}

func mustAccessToken(t *testing.T, service *token.JWTService, userID uuid.UUID, platformRole string) string {
	t.Helper()

	raw, err := service.Generate(userID, platformRole)
	if err != nil {
		t.Fatalf("generate access token: %v", err)
	}
	return raw
}

func assertAppErrorCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error %s, got %T: %v", code, err, err)
	}
	if appErr.Code != code {
		t.Fatalf("error code = %s, want %s", appErr.Code, code)
	}
}

type fakeAdminRepository struct {
	users     map[uuid.UUID]User
	auditLogs []AuditEntry
}

func (f *fakeAdminRepository) FindUserByID(_ context.Context, _ db.Queryer, userID uuid.UUID) (*User, error) {
	user, ok := f.users[userID]
	if !ok {
		return nil, ErrAdminUserNotFound
	}
	return &user, nil
}

func (f *fakeAdminRepository) CreateAuditLog(_ context.Context, _ db.Queryer, entry AuditEntry) (*AuditLog, error) {
	f.auditLogs = append(f.auditLogs, entry)
	return &AuditLog{
		ID:          uuid.MustParse("55555555-5555-5555-5555-555555555555"),
		ActorUserID: entry.ActorUserID,
		Action:      entry.Action,
		TargetType:  entry.TargetType,
		TargetID:    entry.TargetID,
		CreatedAt:   time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
	}, nil
}
