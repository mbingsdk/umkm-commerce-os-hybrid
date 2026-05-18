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
