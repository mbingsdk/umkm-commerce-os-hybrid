package checkout

import (
	"time"

	"github.com/google/uuid"
)

const (
	EventOrderCreated          = "OrderCreated"
	EventStockReserved         = "StockReserved"
	EventNotificationRequested = "NotificationRequested"

	AggregateOrder = "order"
)

type ProductForCheckout struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	Name           string
	SKU            string
	Price          int64
	Status         string
	TrackInventory bool
	AllowBackorder bool
}

type StockSnapshot struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductID         uuid.UUID
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
}

type CourierZoneForCheckout struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	StoreID  uuid.UUID
	Name     string
	Rate     int64
}

type CustomerRecord struct {
	ID       uuid.UUID
	TenantID uuid.UUID
	StoreID  uuid.UUID
	Name     string
	Phone    string
	Email    string
}

type AddressRecord struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	CustomerID uuid.UUID
}

type OrderRecord struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	OrderNumber   string
	Status        string
	PaymentStatus string
	Subtotal      int64
	DiscountTotal int64
	ShippingCost  int64
	TaxTotal      int64
	GrandTotal    int64
	CreatedAt     time.Time
}

type FindOrCreateCustomerParams struct {
	TenantID uuid.UUID
	StoreID  uuid.UUID
	Name     string
	Phone    string
	Email    string
}

type CreateAddressParams struct {
	TenantID       uuid.UUID
	CustomerID     uuid.UUID
	Label          string
	RecipientName  string
	RecipientPhone string
	Address        string
	City           string
	Province       string
	PostalCode     string
}

type CreateOrderParams struct {
	TenantID           uuid.UUID
	StoreID            uuid.UUID
	CustomerID         uuid.UUID
	OrderNumber        string
	Source             string
	Status             string
	PaymentStatus      string
	Subtotal           int64
	DiscountTotal      int64
	ShippingCost       int64
	TaxTotal           int64
	GrandTotal         int64
	CustomerName       string
	CustomerPhone      string
	CustomerEmail      string
	ShippingAddress    string
	ShippingCity       string
	ShippingProvince   string
	ShippingPostalCode string
	CustomerNote       string
}

type CreateOrderItemParams struct {
	TenantID      uuid.UUID
	OrderID       uuid.UUID
	ProductID     uuid.UUID
	ProductName   string
	SKU           string
	Quantity      int
	UnitPrice     int64
	DiscountTotal int64
	Subtotal      int64
}

type CreateReservationParams struct {
	TenantID  uuid.UUID
	StoreID   uuid.UUID
	ProductID uuid.UUID
	OrderID   uuid.UUID
	Quantity  int
	ExpiresAt *time.Time
}

type UpdateSnapshotParams struct {
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductID         uuid.UUID
	QuantityReserved  int
	QuantityAvailable int
}

type CreateStockMovementParams struct {
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	ProductID     uuid.UUID
	MovementType  string
	Quantity      int
	BalanceAfter  int
	ReferenceType string
	ReferenceID   uuid.UUID
	Note          string
}
