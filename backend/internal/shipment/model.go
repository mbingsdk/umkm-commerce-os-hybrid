package shipment

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	CourierTypeInternal = "internal"
	CourierTypeManual   = "manual"

	StatusPending        = "pending"
	StatusReadyForPickup = "ready_for_pickup"
	StatusPickedUp       = "picked_up"
	StatusOnDelivery     = "on_delivery"
	StatusDelivered      = "delivered"
	StatusFailed         = "failed"
	StatusCancelled      = "cancelled"

	EventShipmentCreated       = "ShipmentCreated"
	EventShipmentStatusUpdated = "ShipmentStatusUpdated"
	EventOrderDelivered        = "OrderDelivered"
	EventNotificationRequested = "NotificationRequested"
	AggregateShipment          = "shipment"
	AggregateOrder             = "order"
)

var (
	ErrShipmentNotFound = errors.New("shipment not found")
	ErrOrderNotFound    = errors.New("order not found")
	errInvalidCursor    = errors.New("invalid cursor")
)

type Shipment struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	StoreID         uuid.UUID
	OrderID         uuid.UUID
	OrderNumber     string
	CustomerName    string
	CustomerPhone   string
	CourierType     string
	CourierName     string
	TrackingNumber  string
	Status          string
	ShippingCost    int64
	AssignedToName  string
	AssignedToPhone string
	Note            string
	ShippedAt       *time.Time
	DeliveredAt     *time.Time
	CreatedBy       *uuid.UUID
	UpdatedBy       *uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type StatusLog struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	ShipmentID uuid.UUID
	FromStatus string
	ToStatus   string
	Note       string
	CreatedBy  *uuid.UUID
	CreatedAt  time.Time
}

type ListFilters struct {
	Status *string
	Query  string
	Limit  int
	Cursor *Cursor
}

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

type CreateShipmentParams struct {
	TenantID        uuid.UUID
	StoreID         uuid.UUID
	OrderID         uuid.UUID
	CourierType     string
	CourierName     string
	TrackingNumber  string
	ShippingCost    int64
	AssignedToName  string
	AssignedToPhone string
	Note            string
	CreatedBy       uuid.UUID
}

type UpdateShipmentStatusParams struct {
	TenantID   uuid.UUID
	StoreID    uuid.UUID
	ShipmentID uuid.UUID
	Status     string
	UpdatedBy  uuid.UUID
}

type CreateStatusLogParams struct {
	TenantID   uuid.UUID
	ShipmentID uuid.UUID
	FromStatus string
	ToStatus   string
	Note       string
	CreatedBy  uuid.UUID
}

func EncodeCursor(item Shipment) (string, error) {
	payload, err := json.Marshal(Cursor{
		CreatedAt: item.CreatedAt,
		ID:        item.ID,
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
