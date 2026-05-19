package order

import (
	"time"

	"github.com/google/uuid"
)

const (
	SourceStorefront = "storefront"

	StatusPending     = "pending"
	StatusConfirmed   = "confirmed"
	StatusProcessing  = "processing"
	StatusReadyToShip = "ready_to_ship"
	StatusShipped     = "shipped"
	StatusDelivered   = "delivered"
	StatusCompleted   = "completed"
	StatusCancelled   = "cancelled"
	StatusReturned    = "returned"
	StatusRefunded    = "refunded"

	PaymentStatusUnpaid              = "unpaid"
	PaymentStatusWaitingConfirmation = "waiting_confirmation"
	PaymentStatusPaid                = "paid"
	PaymentStatusFailed              = "failed"
	PaymentStatusRefunded            = "refunded"
)

type Order struct {
	ID                 uuid.UUID
	TenantID           uuid.UUID
	StoreID            uuid.UUID
	CustomerID         *uuid.UUID
	OrderNumber        string
	Source             string
	Status             string
	PaymentStatus      string
	ShipmentStatus     string
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
	InternalNote       string
	ConfirmedAt        *time.Time
	PaidAt             *time.Time
	CompletedAt        *time.Time
	CancelledAt        *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Item struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	OrderID       uuid.UUID
	ProductID     *uuid.UUID
	ProductName   string
	SKU           string
	Quantity      int
	UnitPrice     int64
	DiscountTotal int64
	Subtotal      int64
	CreatedAt     time.Time
}

type StatusLog struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	OrderID    uuid.UUID
	FromStatus string
	ToStatus   string
	Note       string
	CreatedBy  *uuid.UUID
	CreatedAt  time.Time
}
