package permission

var rolePermissions = map[Role]map[Permission]bool{
	RoleOwner: {
		TenantRead:               true,
		TenantUpdate:             true,
		TenantManageMembers:      true,
		TenantManagePlan:         true,
		StoreRead:                true,
		StoreUpdate:              true,
		StorePublish:             true,
		StoreUpdateBusinessHours: true,
		StoreUpdateSEO:           true,
		StoreUpdateDiscovery:     true,
	},
	RoleManager: {
		TenantRead:               true,
		StoreRead:                true,
		StoreUpdateBusinessHours: true,
		StoreUpdateSEO:           true,
		StoreUpdateDiscovery:     true,
	},
	RoleStaff: {
		TenantRead: true,
		StoreRead:  true,
	},
	RoleCashier: {
		TenantRead: true,
		StoreRead:  true,
	},
	RoleInventoryStaff: {
		TenantRead: true,
		StoreRead:  true,
	},
	RoleCourierAdmin: {
		TenantRead: true,
		StoreRead:  true,
	},
}

func Allowed(role string, permission Permission) bool {
	permissions, ok := rolePermissions[Role(role)]
	if !ok {
		return false
	}

	return permissions[permission]
}
