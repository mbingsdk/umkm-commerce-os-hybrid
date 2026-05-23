package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/testutil"
)

func TestAdminSeparationAllowsOnlySuperAdminWithoutTenantHeader(t *testing.T) {
	fixtures := testutil.NewSecurityFixtures()
	users := map[uuid.UUID]User{
		fixtures.Users.Owner.ID:      adminUserFromFixture(fixtures.Users.Owner),
		fixtures.Users.Manager.ID:    adminUserFromFixture(fixtures.Users.Manager),
		fixtures.Users.Staff.ID:      adminUserFromFixture(fixtures.Users.Staff),
		fixtures.Users.Cashier.ID:    adminUserFromFixture(fixtures.Users.Cashier),
		fixtures.Users.SuperAdmin.ID: adminUserFromFixture(fixtures.Users.SuperAdmin),
	}
	router, tokens, _ := newAdminTestRouter(t, users)

	tests := []struct {
		name       string
		user       testutil.UserFixture
		wantStatus int
	}{
		{name: "tenant owner rejected", user: fixtures.Users.Owner, wantStatus: http.StatusForbidden},
		{name: "tenant manager rejected", user: fixtures.Users.Manager, wantStatus: http.StatusForbidden},
		{name: "tenant staff rejected", user: fixtures.Users.Staff, wantStatus: http.StatusForbidden},
		{name: "tenant cashier rejected", user: fixtures.Users.Cashier, wantStatus: http.StatusForbidden},
		{name: "super admin accepted without tenant header", user: fixtures.Users.SuperAdmin, wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/me", nil)
			req.Header.Set("Authorization", "Bearer "+mustAccessToken(t, tokens, tt.user.ID, tt.user.PlatformRole))
			res := httptest.NewRecorder()

			router.ServeHTTP(res, req)

			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", res.Code, tt.wantStatus, res.Body.String())
			}
		})
	}
}

func adminUserFromFixture(user testutil.UserFixture) User {
	return User{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PlatformRole: user.PlatformRole,
		Status:       auth.UserStatusActive,
	}
}
