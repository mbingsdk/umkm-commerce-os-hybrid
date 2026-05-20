package order

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var errInvalidCursor = errors.New("invalid cursor")

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

	EventOrderStatusUpdated = "OrderStatusUpdated"
	AggregateOrder          = "order"
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

type ReservationSummary struct {
	Status   string
	Quantity int
	Count    int
}

type ListFilters struct {
	Status        *string
	PaymentStatus *string
	Source        *string
	Query         string
	DateFrom      *time.Time
	DateTo        *time.Time
	Limit         int
	Cursor        *Cursor
}

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

type ListParams struct {
	TenantID uuid.UUID
	StoreID  uuid.UUID
	Filters  ListFilters
}

type UpdateStatusParams struct {
	TenantID uuid.UUID
	StoreID  uuid.UUID
	OrderID  uuid.UUID
	Status   string
}

type CreateStatusLogParams struct {
	TenantID   uuid.UUID
	OrderID    uuid.UUID
	FromStatus string
	ToStatus   string
	Note       string
	CreatedBy  uuid.UUID
}

func EncodeCursor(order Order) (string, error) {
	payload, err := json.Marshal(Cursor{
		CreatedAt: order.CreatedAt,
		ID:        order.ID,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeCursor(raw string) (*Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var cursor Cursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, err
	}
	if cursor.ID == uuid.Nil || cursor.CreatedAt.IsZero() {
		return nil, errInvalidCursor
	}
	return &cursor, nil
}
