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
		CategoryRead:             true,
		CategoryCreate:           true,
		CategoryUpdate:           true,
		CategoryDelete:           true,
		ProductRead:              true,
		ProductCreate:            true,
		ProductUpdate:            true,
		ProductDelete:            true,
		ProductUploadImage:       true,
		InventoryRead:            true,
		InventoryReadMovement:    true,
		InventoryAdjust:          true,
		InventoryUpdateThreshold: true,
		OrderRead:                true,
		OrderReadDetail:          true,
		OrderUpdateStatus:        true,
	},
	RoleManager: {
		TenantRead:               true,
		StoreRead:                true,
		StoreUpdateBusinessHours: true,
		StoreUpdateSEO:           true,
		StoreUpdateDiscovery:     true,
		CategoryRead:             true,
		CategoryCreate:           true,
		CategoryUpdate:           true,
		CategoryDelete:           true,
		ProductRead:              true,
		ProductCreate:            true,
		ProductUpdate:            true,
		ProductUploadImage:       true,
		InventoryRead:            true,
		InventoryReadMovement:    true,
		InventoryAdjust:          true,
		InventoryUpdateThreshold: true,
		OrderRead:                true,
		OrderReadDetail:          true,
		OrderUpdateStatus:        true,
	},
	RoleStaff: {
		TenantRead:        true,
		StoreRead:         true,
		CategoryRead:      true,
		ProductRead:       true,
		InventoryRead:     true,
		OrderRead:         true,
		OrderReadDetail:   true,
		OrderUpdateStatus: true,
	},
	RoleCashier: {
		TenantRead:   true,
		StoreRead:    true,
		CategoryRead: true,
	},
	RoleInventoryStaff: {
		TenantRead:               true,
		StoreRead:                true,
		CategoryRead:             true,
		ProductRead:              true,
		InventoryRead:            true,
		InventoryReadMovement:    true,
		InventoryAdjust:          true,
		InventoryUpdateThreshold: true,
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
