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

	CategoryRead   Permission = "category.read"
	CategoryCreate Permission = "category.create"
	CategoryUpdate Permission = "category.update"
	CategoryDelete Permission = "category.delete"

	ProductRead        Permission = "product.read"
	ProductCreate      Permission = "product.create"
	ProductUpdate      Permission = "product.update"
	ProductDelete      Permission = "product.delete"
	ProductUploadImage Permission = "product.upload_image"

	InventoryRead            Permission = "inventory.read"
	InventoryReadMovement    Permission = "inventory.read_movement"
	InventoryAdjust          Permission = "inventory.adjust"
	InventoryUpdateThreshold Permission = "inventory.update_threshold"

	OrderRead                Permission = "order.read"
	OrderReadDetail          Permission = "order.read_detail"
	OrderUpdateStatus        Permission = "order.update_status"
	OrderUpdatePaymentStatus Permission = "order.update_payment_status"
	OrderCancel              Permission = "order.cancel"

	POSReadProduct       Permission = "pos.read_product"
	POSOpenSession       Permission = "pos.open_session"
	POSReadSession       Permission = "pos.read_session"
	POSCreateTransaction Permission = "pos.create_transaction"
	POSReadTransaction   Permission = "pos.read_transaction"
	POSCloseSession      Permission = "pos.close_session"
	POSRefundTransaction Permission = "pos.refund_transaction"
)
