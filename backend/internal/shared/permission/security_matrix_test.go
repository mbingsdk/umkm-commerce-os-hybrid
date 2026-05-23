package permission_test

import (
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/testutil"
)

func TestSecurityPermissionMatrixByRole(t *testing.T) {
	fixtures := testutil.NewSecurityFixtures()

	tests := []struct {
		name       string
		role       string
		permission permission.Permission
		want       bool
	}{
		{name: "owner can create products", role: fixtures.Users.Owner.TenantRole, permission: permission.ProductCreate, want: true},
		{name: "owner can adjust inventory", role: fixtures.Users.Owner.TenantRole, permission: permission.InventoryAdjust, want: true},
		{name: "owner can cancel orders", role: fixtures.Users.Owner.TenantRole, permission: permission.OrderCancel, want: true},
		{name: "owner can read finance", role: fixtures.Users.Owner.TenantRole, permission: permission.FinanceReadSummary, want: true},
		{name: "owner can use POS", role: fixtures.Users.Owner.TenantRole, permission: permission.POSCreateTransaction, want: true},
		{name: "manager can process orders", role: fixtures.Users.Manager.TenantRole, permission: permission.OrderUpdateStatus, want: true},
		{name: "manager can read finance", role: fixtures.Users.Manager.TenantRole, permission: permission.FinanceReadSummary, want: true},
		{name: "manager cannot delete products", role: fixtures.Users.Manager.TenantRole, permission: permission.ProductDelete, want: false},
		{name: "manager cannot publish store", role: fixtures.Users.Manager.TenantRole, permission: permission.StorePublish, want: false},
		{name: "staff can read orders", role: fixtures.Users.Staff.TenantRole, permission: permission.OrderRead, want: true},
		{name: "staff cannot read finance", role: fixtures.Users.Staff.TenantRole, permission: permission.FinanceReadSummary, want: false},
		{name: "staff cannot use POS transaction", role: fixtures.Users.Staff.TenantRole, permission: permission.POSCreateTransaction, want: false},
		{name: "cashier can create POS transaction", role: fixtures.Users.Cashier.TenantRole, permission: permission.POSCreateTransaction, want: true},
		{name: "cashier cannot adjust inventory", role: fixtures.Users.Cashier.TenantRole, permission: permission.InventoryAdjust, want: false},
		{name: "cashier cannot read finance", role: fixtures.Users.Cashier.TenantRole, permission: permission.FinanceReadSummary, want: false},
		{name: "cashier cannot read tenant orders", role: fixtures.Users.Cashier.TenantRole, permission: permission.OrderRead, want: false},
		{name: "super admin platform role is not a tenant dashboard role", role: auth.PlatformRoleSuperAdmin, permission: permission.DashboardReadSummary, want: false},
		{name: "unknown future permission defaults deny", role: fixtures.Users.Owner.TenantRole, permission: permission.Permission("future.admin.manage"), want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := permission.Allowed(tt.role, tt.permission); got != tt.want {
				t.Fatalf("Allowed(%q, %q) = %v, want %v", tt.role, tt.permission, got, tt.want)
			}
		})
	}
}
