package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/testutil"
)

func TestTenantDashboardRouteFamiliesRequireTenantID(t *testing.T) {
	fixtures := testutil.NewSecurityFixtures()

	for _, tc := range tenantRouteSecurityCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			router := newTenantSecurityRouter(fixtures, fixtures.Users.Owner, tc)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			assertErrorCode(t, res, http.StatusBadRequest, apperror.CodeValidation)
		})
	}
}

func TestTenantDashboardRouteFamiliesRejectOtherTenant(t *testing.T) {
	fixtures := testutil.NewSecurityFixtures()

	for _, tc := range tenantRouteSecurityCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			router := newTenantSecurityRouter(fixtures, fixtures.Users.Owner, tc)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("X-Tenant-ID", fixtures.Tenants.B.ID.String())
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			assertErrorCode(t, res, http.StatusForbidden, apperror.CodeTenantAccessDenied)
		})
	}
}

func TestTenantDashboardRouteFamiliesAllowTenantMemberWithPermission(t *testing.T) {
	fixtures := testutil.NewSecurityFixtures()

	for _, tc := range tenantRouteSecurityCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			router := newTenantSecurityRouter(fixtures, fixtures.Users.Owner, tc)
			req := httptest.NewRequest(tc.method, tc.path, nil)
			req.Header.Set("X-Tenant-ID", fixtures.Tenants.A.ID.String())
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d; body=%s", res.Code, http.StatusNoContent, res.Body.String())
			}
		})
	}
}

func TestSuperAdminPlatformRoleDoesNotGrantTenantDashboardPermission(t *testing.T) {
	fixtures := testutil.NewSecurityFixtures()
	route := tenantRouteSecurityCase{
		name:       "dashboard summary",
		method:     http.MethodGet,
		path:       "/api/v1/dashboard/summary",
		permission: permission.DashboardReadSummary,
	}
	superAdminAsTenantRole := fixtures.Users.SuperAdmin
	superAdminAsTenantRole.TenantRole = auth.PlatformRoleSuperAdmin
	router := newTenantSecurityRouter(fixtures, superAdminAsTenantRole, route)

	req := httptest.NewRequest(route.method, route.path, nil)
	req.Header.Set("X-Tenant-ID", fixtures.Tenants.A.ID.String())
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	assertErrorCode(t, res, http.StatusForbidden, apperror.CodeForbidden)
}

type tenantRouteSecurityCase struct {
	name       string
	method     string
	path       string
	permission permission.Permission
}

func tenantRouteSecurityCases() []tenantRouteSecurityCase {
	return []tenantRouteSecurityCase{
		{name: "product list", method: http.MethodGet, path: "/api/v1/products", permission: permission.ProductRead},
		{name: "category list", method: http.MethodGet, path: "/api/v1/categories", permission: permission.CategoryRead},
		{name: "inventory stocks", method: http.MethodGet, path: "/api/v1/inventory/stocks", permission: permission.InventoryRead},
		{name: "order list", method: http.MethodGet, path: "/api/v1/orders", permission: permission.OrderRead},
		{name: "pos products", method: http.MethodGet, path: "/api/v1/pos/products", permission: permission.POSReadProduct},
		{name: "finance summary", method: http.MethodGet, path: "/api/v1/finance/summary", permission: permission.FinanceReadSummary},
		{name: "courier zones", method: http.MethodGet, path: "/api/v1/courier/zones", permission: permission.CourierReadZone},
		{name: "shipments", method: http.MethodGet, path: "/api/v1/shipments", permission: permission.ShipmentRead},
		{name: "dashboard summary", method: http.MethodGet, path: "/api/v1/dashboard/summary", permission: permission.DashboardReadSummary},
	}
}

func newTenantSecurityRouter(fixtures testutil.SecurityFixtures, user testutil.UserFixture, route tenantRouteSecurityCase) http.Handler {
	logger := discardLogger()
	router := chi.NewRouter()
	tenantMiddleware := TenantResolver(routeTenantValidator{
		fixtures: fixtures,
		user:     user,
	}, logger)
	requirePermission := RequirePermission(route.permission, logger)

	router.With(
		authContextMiddleware(user),
		tenantMiddleware,
		requirePermission,
	).Method(route.method, route.path, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	return router
}

func authContextMiddleware(user testutil.UserFixture) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := auth.WithContext(r.Context(), auth.AuthContext{
				UserID:       user.ID,
				PlatformRole: user.PlatformRole,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type routeTenantValidator struct {
	fixtures testutil.SecurityFixtures
	user     testutil.UserFixture
}

func (v routeTenantValidator) ValidateAccess(_ context.Context, userID uuid.UUID, tenantID uuid.UUID) (tenantctx.TenantContext, error) {
	if userID != v.user.ID || tenantID != v.fixtures.Tenants.A.ID {
		return tenantctx.TenantContext{}, apperror.TenantAccessDenied("Tenant access denied")
	}

	return tenantctx.TenantContext{
		TenantID: v.fixtures.Tenants.A.ID,
		StoreID:  v.fixtures.Stores.A.ID,
		UserID:   v.user.ID,
		Role:     v.user.TenantRole,
	}, nil
}
