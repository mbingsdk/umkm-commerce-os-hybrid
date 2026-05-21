package admin

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/token"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	sharedmw "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/middleware"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
)

var (
	testSuperAdminID = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testOwnerID      = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testManagerID    = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testTargetID     = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testPlanID       = uuid.MustParse("66666666-6666-6666-6666-666666666666")
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

func TestAdminTenantListRequiresSuperAdmin(t *testing.T) {
	router, tokens, _ := newAdminTestRouter(t, map[uuid.UUID]User{
		testOwnerID: {
			ID:           testOwnerID,
			Name:         "Tenant Owner",
			Email:        "owner@example.test",
			PlatformRole: auth.PlatformRoleUser,
			Status:       auth.UserStatusActive,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, testOwnerID, auth.PlatformRoleUser))
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
}

func TestAdminPlanManagementRequiresSuperAdmin(t *testing.T) {
	router, tokens, _ := newAdminTestRouter(t, map[uuid.UUID]User{
		testOwnerID: {
			ID:           testOwnerID,
			Name:         "Tenant Owner",
			Email:        "owner@example.test",
			PlatformRole: auth.PlatformRoleUser,
			Status:       auth.UserStatusActive,
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/plans", nil)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, testOwnerID, auth.PlatformRoleUser))
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusForbidden)
	}
}

func TestAdminTenantListAllowsSuperAdminWithoutTenantHeader(t *testing.T) {
	tenantID := uuid.New()
	repo := &fakeAdminRepository{
		users: map[uuid.UUID]User{
			testSuperAdminID: {
				ID:           testSuperAdminID,
				Name:         "Platform Admin",
				Email:        "admin@example.test",
				PlatformRole: auth.PlatformRoleSuperAdmin,
				Status:       auth.UserStatusActive,
			},
		},
		listItems: []TenantListItem{{
			Tenant: Tenant{
				ID:        tenantID,
				Name:      "Toko Bunga Ayu",
				Slug:      "toko-bunga-ayu",
				Status:    TenantStatusActive,
				CreatedAt: time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
			},
			Counts: TenantCounts{StoreCount: 1},
		}},
	}
	router, tokens := newAdminTestRouterWithRepo(t, repo, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/tenants", nil)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, testSuperAdminID, auth.PlatformRoleSuperAdmin))
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. body=%s", res.Code, http.StatusOK, res.Body.String())
	}

	var body struct {
		Success bool                      `json:"success"`
		Data    []AdminTenantListResponse `json:"data"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || len(body.Data) != 1 || body.Data[0].ID != tenantID.String() {
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

func TestCreatePlanRejectsDuplicateCode(t *testing.T) {
	repo := &fakeAdminRepository{
		users: map[uuid.UUID]User{},
		plans: map[uuid.UUID]Plan{
			testPlanID: {ID: testPlanID, Code: "starter", Name: "Starter", IsActive: true},
		},
	}
	service := NewService(fakeTxDatabase{}, repo, &fakeOutboxRepository{})

	_, err := service.CreatePlan(context.Background(), CreatePlanInput{
		ActorUserID:  testSuperAdminID,
		Code:         "starter",
		Name:         "Starter Baru",
		PriceMonthly: 0,
	})
	assertAppErrorCode(t, err, apperror.CodeValidation)
}

func TestCreatePlanCreatesAdminAuditLogAndOutbox(t *testing.T) {
	repo := &fakeAdminRepository{
		users: map[uuid.UUID]User{},
		plans: map[uuid.UUID]Plan{},
	}
	outboxRepo := &fakeOutboxRepository{}
	service := NewService(fakeTxDatabase{}, repo, outboxRepo)
	productLimit := 100
	staffLimit := 3

	result, err := service.CreatePlan(context.Background(), CreatePlanInput{
		ActorUserID:     testSuperAdminID,
		Code:            "growth",
		Name:            "Growth",
		Description:     "Untuk UMKM bertumbuh",
		PriceMonthly:    99000,
		ProductLimit:    &productLimit,
		StaffLimit:      &staffLimit,
		CanUsePOS:       boolPtr(true),
		CanUseDiscovery: boolPtr(true),
		CanUseCourier:   boolPtr(true),
		IsActive:        boolPtr(true),
		IPAddress:       "127.0.0.1",
		UserAgent:       "admin-test",
	})
	if err != nil {
		t.Fatalf("CreatePlan error = %v", err)
	}
	if result.Code != "growth" || result.ProductLimit == nil || *result.ProductLimit != productLimit {
		t.Fatalf("unexpected plan result: %#v", result)
	}
	if len(repo.auditLogs) != 1 || repo.auditLogs[0].Action != AuditActionPlanCreated {
		t.Fatalf("audit logs = %#v", repo.auditLogs)
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventPlanChanged {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestUpdateTenantStatusCreatesAdminAuditLogAndOutbox(t *testing.T) {
	tenantID := uuid.New()
	repo := &fakeAdminRepository{
		users: map[uuid.UUID]User{},
		tenants: map[uuid.UUID]Tenant{
			tenantID: {
				ID:     tenantID,
				Name:   "Toko Bunga Ayu",
				Slug:   "toko-bunga-ayu",
				Status: TenantStatusActive,
			},
		},
		plans: map[uuid.UUID]Plan{},
	}
	outboxRepo := &fakeOutboxRepository{}
	service := NewService(fakeTxDatabase{}, repo, outboxRepo)

	result, err := service.UpdateTenantStatus(context.Background(), UpdateTenantStatusInput{
		ActorUserID: testSuperAdminID,
		TenantID:    tenantID,
		Status:      TenantStatusSuspended,
		Reason:      "Pelanggaran kebijakan platform",
		IPAddress:   "127.0.0.1",
		UserAgent:   "admin-test",
	})
	if err != nil {
		t.Fatalf("UpdateTenantStatus error = %v", err)
	}
	if result.Status != TenantStatusSuspended || repo.tenants[tenantID].Status != TenantStatusSuspended {
		t.Fatalf("tenant status = %s, want %s", result.Status, TenantStatusSuspended)
	}
	if len(repo.auditLogs) != 1 || repo.auditLogs[0].Action != AuditActionTenantStatusUpdated {
		t.Fatalf("audit logs = %#v", repo.auditLogs)
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventTenantStatusChanged {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestUpdateTenantPlanValidatesPlan(t *testing.T) {
	tenantID := uuid.New()
	repo := &fakeAdminRepository{
		users: map[uuid.UUID]User{},
		tenants: map[uuid.UUID]Tenant{
			tenantID: {
				ID:     tenantID,
				Name:   "Toko Bunga Ayu",
				Slug:   "toko-bunga-ayu",
				Status: TenantStatusActive,
			},
		},
		plans: map[uuid.UUID]Plan{},
	}
	service := NewService(fakeTxDatabase{}, repo, &fakeOutboxRepository{})

	_, err := service.UpdateTenantPlan(context.Background(), UpdateTenantPlanInput{
		ActorUserID: testSuperAdminID,
		TenantID:    tenantID,
		PlanID:      testPlanID,
	})
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestAdminTenantMutationDoesNotRequireTenantHeader(t *testing.T) {
	tenantID := uuid.New()
	repo := &fakeAdminRepository{
		users: map[uuid.UUID]User{
			testSuperAdminID: {
				ID:           testSuperAdminID,
				Name:         "Platform Admin",
				Email:        "admin@example.test",
				PlatformRole: auth.PlatformRoleSuperAdmin,
				Status:       auth.UserStatusActive,
			},
		},
		tenants: map[uuid.UUID]Tenant{
			tenantID: {
				ID:        tenantID,
				Name:      "Toko Bunga Ayu",
				Slug:      "toko-bunga-ayu",
				Status:    TenantStatusTrialing,
				CreatedAt: time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
			},
		},
	}
	outboxRepo := &fakeOutboxRepository{}
	router, tokens := newAdminTestRouterWithRepo(t, repo, outboxRepo)

	req := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/admin/tenants/"+tenantID.String()+"/status",
		strings.NewReader(`{"status":"active","reason":"Verifikasi selesai"}`),
	)
	req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, testSuperAdminID, auth.PlatformRoleSuperAdmin))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. body=%s", res.Code, http.StatusOK, res.Body.String())
	}
	if repo.tenants[tenantID].Status != TenantStatusActive {
		t.Fatalf("tenant status = %s, want %s", repo.tenants[tenantID].Status, TenantStatusActive)
	}
	if len(repo.auditLogs) != 1 || len(outboxRepo.events) != 1 {
		t.Fatalf("audit logs=%d outbox events=%d, want 1/1", len(repo.auditLogs), len(outboxRepo.events))
	}
}

func TestRecordAuditRequiresAction(t *testing.T) {
	service := NewService(nil, &fakeAdminRepository{})

	_, err := service.RecordAudit(context.Background(), AuditEntry{ActorUserID: testSuperAdminID})
	assertAppErrorCode(t, err, apperror.CodeValidation)
}

func newAdminTestRouter(t *testing.T, users map[uuid.UUID]User) (http.Handler, *token.JWTService, *fakeAdminRepository) {
	t.Helper()

	repo := &fakeAdminRepository{users: users}
	router, tokens := newAdminTestRouterWithRepo(t, repo, nil)
	return router, tokens, repo
}

func newAdminTestRouterWithRepo(t *testing.T, repo *fakeAdminRepository, outboxRepo *fakeOutboxRepository) (http.Handler, *token.JWTService) {
	t.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	tokens := token.NewJWTService("test-secret", time.Hour)
	service := NewService(fakeTxDatabase{}, repo, outboxRepo)
	handler := NewHandler(service, logger)

	router := chi.NewRouter()
	router.Route("/api/v1", func(r chi.Router) {
		RegisterRoutes(r, handler, sharedmw.Auth(tokens, logger), Guard(service, logger))
	})

	return router, tokens
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
	listItems []TenantListItem
	detail    *TenantDetail
	tenants   map[uuid.UUID]Tenant
	plans     map[uuid.UUID]Plan
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

func (f *fakeAdminRepository) ListTenants(_ context.Context, _ db.Queryer, _ TenantListFilters) ([]TenantListItem, error) {
	return f.listItems, nil
}

func (f *fakeAdminRepository) GetTenantDetail(_ context.Context, _ db.Queryer, tenantID uuid.UUID) (*TenantDetail, error) {
	if f.detail != nil && f.detail.Tenant.ID == tenantID {
		return f.detail, nil
	}
	tenant, ok := f.tenants[tenantID]
	if !ok {
		return nil, ErrTenantNotFound
	}
	return &TenantDetail{Tenant: tenant}, nil
}

func (f *fakeAdminRepository) FindTenantByIDForUpdate(_ context.Context, _ db.Queryer, tenantID uuid.UUID) (*Tenant, error) {
	tenant, ok := f.tenants[tenantID]
	if !ok {
		return nil, ErrTenantNotFound
	}
	return &tenant, nil
}

func (f *fakeAdminRepository) UpdateTenantStatus(_ context.Context, _ db.Queryer, tenantID uuid.UUID, status string) (*Tenant, error) {
	tenant, ok := f.tenants[tenantID]
	if !ok {
		return nil, ErrTenantNotFound
	}
	tenant.Status = status
	tenant.UpdatedAt = time.Date(2026, 5, 21, 11, 0, 0, 0, time.UTC)
	f.tenants[tenantID] = tenant
	return &tenant, nil
}

func (f *fakeAdminRepository) FindActivePlanByID(_ context.Context, _ db.Queryer, planID uuid.UUID) (*Plan, error) {
	plan, ok := f.plans[planID]
	if !ok || !plan.IsActive {
		return nil, ErrPlanNotFound
	}
	return &plan, nil
}

func (f *fakeAdminRepository) UpdateTenantPlan(_ context.Context, _ db.Queryer, tenantID uuid.UUID, planID uuid.UUID) (*Tenant, error) {
	tenant, ok := f.tenants[tenantID]
	if !ok {
		return nil, ErrTenantNotFound
	}
	tenant.PlanID = &planID
	tenant.UpdatedAt = time.Date(2026, 5, 21, 11, 0, 0, 0, time.UTC)
	f.tenants[tenantID] = tenant
	return &tenant, nil
}

func (f *fakeAdminRepository) ListPlans(context.Context, db.Queryer) ([]Plan, error) {
	items := make([]Plan, 0, len(f.plans))
	for _, item := range f.plans {
		items = append(items, item)
	}
	return items, nil
}

func (f *fakeAdminRepository) FindPlanByIDForUpdate(_ context.Context, _ db.Queryer, planID uuid.UUID) (*Plan, error) {
	plan, ok := f.plans[planID]
	if !ok {
		return nil, ErrPlanNotFound
	}
	return &plan, nil
}

func (f *fakeAdminRepository) CreatePlan(_ context.Context, _ db.Queryer, params CreatePlanParams) (*Plan, error) {
	for _, existing := range f.plans {
		if existing.Code == params.Code {
			return nil, ErrPlanCodeAlreadyInUse
		}
	}
	plan := Plan{
		ID:                 uuid.New(),
		Code:               params.Code,
		Name:               params.Name,
		Description:        params.Description,
		PriceMonthly:       params.PriceMonthly,
		ProductLimit:       params.ProductLimit,
		StaffLimit:         params.StaffLimit,
		CanUsePOS:          params.CanUsePOS,
		CanUseDiscovery:    params.CanUseDiscovery,
		CanUseCourier:      params.CanUseCourier,
		CanUseCustomDomain: params.CanUseCustomDomain,
		IsActive:           params.IsActive,
	}
	if f.plans == nil {
		f.plans = map[uuid.UUID]Plan{}
	}
	f.plans[plan.ID] = plan
	return &plan, nil
}

func (f *fakeAdminRepository) UpdatePlan(_ context.Context, _ db.Queryer, params UpdatePlanParams) (*Plan, error) {
	if _, ok := f.plans[params.PlanID]; !ok {
		return nil, ErrPlanNotFound
	}
	for id, existing := range f.plans {
		if id != params.PlanID && existing.Code == params.Code {
			return nil, ErrPlanCodeAlreadyInUse
		}
	}
	plan := Plan{
		ID:                 params.PlanID,
		Code:               params.Code,
		Name:               params.Name,
		Description:        params.Description,
		PriceMonthly:       params.PriceMonthly,
		ProductLimit:       params.ProductLimit,
		StaffLimit:         params.StaffLimit,
		CanUsePOS:          params.CanUsePOS,
		CanUseDiscovery:    params.CanUseDiscovery,
		CanUseCourier:      params.CanUseCourier,
		CanUseCustomDomain: params.CanUseCustomDomain,
		IsActive:           params.IsActive,
	}
	f.plans[params.PlanID] = plan
	return &plan, nil
}

func boolPtr(value bool) *bool {
	return &value
}

type fakeOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakeOutboxRepository) Insert(_ context.Context, _ db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error) {
	f.events = append(f.events, params)
	return &outbox.Event{
		ID:            uuid.New(),
		TenantID:      params.TenantID,
		EventType:     params.EventType,
		AggregateType: params.AggregateType,
		AggregateID:   params.AggregateID,
		Payload:       params.Payload,
		Status:        outbox.StatusPending,
		CreatedAt:     time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
	}, nil
}

type fakeTxDatabase struct{}

func (fakeTxDatabase) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (fakeTxDatabase) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (fakeTxDatabase) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func (fakeTxDatabase) WithTx(ctx context.Context, fn func(db.Tx) error) error {
	return fn(fakeTxDatabase{})
}
