package permission

type Role string

const (
	RoleOwner          Role = "owner"
	RoleManager        Role = "manager"
	RoleStaff          Role = "staff"
	RoleCashier        Role = "cashier"
	RoleInventoryStaff Role = "inventory_staff"
	RoleCourierAdmin   Role = "courier_admin"
	RoleDriver         Role = "driver"
)
