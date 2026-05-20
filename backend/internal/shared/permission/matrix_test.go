package permission

import "testing"

func TestAllowedDefaultsToDeny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		role       string
		permission Permission
		want       bool
	}{
		{name: "owner can publish", role: string(RoleOwner), permission: StorePublish, want: true},
		{name: "manager cannot publish", role: string(RoleManager), permission: StorePublish, want: false},
		{name: "manager cannot use broad profile update without field level policy", role: string(RoleManager), permission: StoreUpdate, want: false},
		{name: "manager can update business hours", role: string(RoleManager), permission: StoreUpdateBusinessHours, want: true},
		{name: "owner can create products", role: string(RoleOwner), permission: ProductCreate, want: true},
		{name: "owner can read orders", role: string(RoleOwner), permission: OrderRead, want: true},
		{name: "manager can update order status", role: string(RoleManager), permission: OrderUpdateStatus, want: true},
		{name: "manager can update payment status", role: string(RoleManager), permission: OrderUpdatePaymentStatus, want: true},
		{name: "manager can cancel orders", role: string(RoleManager), permission: OrderCancel, want: true},
		{name: "staff can read order detail", role: string(RoleStaff), permission: OrderReadDetail, want: true},
		{name: "staff cannot update payment status", role: string(RoleStaff), permission: OrderUpdatePaymentStatus, want: false},
		{name: "staff cannot cancel orders while limited policy is not modeled", role: string(RoleStaff), permission: OrderCancel, want: false},
		{name: "cashier cannot read orders while POS order policy is not modeled", role: string(RoleCashier), permission: OrderRead, want: false},
		{name: "inventory staff cannot update order status while limited policy is not modeled", role: string(RoleInventoryStaff), permission: OrderUpdateStatus, want: false},
		{name: "owner can create expenses", role: string(RoleOwner), permission: FinanceCreateExpense, want: true},
		{name: "manager can read expenses", role: string(RoleManager), permission: FinanceReadExpense, want: true},
		{name: "staff cannot read finance expenses", role: string(RoleStaff), permission: FinanceReadExpense, want: false},
		{name: "cashier cannot create finance expenses", role: string(RoleCashier), permission: FinanceCreateExpense, want: false},
		{name: "inventory staff cannot delete finance expenses", role: string(RoleInventoryStaff), permission: FinanceDeleteExpense, want: false},
		{name: "manager can upload product images", role: string(RoleManager), permission: ProductUploadImage, want: true},
		{name: "manager cannot delete products while limited delete policy is not modeled", role: string(RoleManager), permission: ProductDelete, want: false},
		{name: "staff cannot create products while limited policy is not modeled", role: string(RoleStaff), permission: ProductCreate, want: false},
		{name: "staff cannot upload product images while limited policy is not modeled", role: string(RoleStaff), permission: ProductUploadImage, want: false},
		{name: "driver denied while limited policy is not modeled", role: string(RoleDriver), permission: StoreRead, want: false},
		{name: "unknown role denied", role: "unknown", permission: StoreRead, want: false},
		{name: "unknown permission denied", role: string(RoleOwner), permission: Permission("future.scope"), want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := Allowed(tt.role, tt.permission); got != tt.want {
				t.Fatalf("Allowed(%q, %q) = %v, want %v", tt.role, tt.permission, got, tt.want)
			}
		})
	}
}
