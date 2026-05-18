package permission

type Permission string

const (
	TenantRead          Permission = "tenant.read"
	TenantCreate        Permission = "tenant.create"
	TenantUpdate        Permission = "tenant.update"
	TenantManageMembers Permission = "tenant.manage_members"
	TenantManagePlan    Permission = "tenant.manage_plan"

	StoreRead                Permission = "store.read"
	StoreUpdate              Permission = "store.update"
	StorePublish             Permission = "store.publish"
	StoreUpdateBusinessHours Permission = "store.update_business_hours"
	StoreUpdateSEO           Permission = "store.update_seo"
	StoreUpdateDiscovery     Permission = "store.update_discovery"
)
