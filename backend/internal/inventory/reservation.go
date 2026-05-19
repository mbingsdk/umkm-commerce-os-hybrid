package inventory

import (
	"time"

	"github.com/google/uuid"
)

const (
	ReservationStatusActive    = "active"
	ReservationStatusConfirmed = "confirmed"
	ReservationStatusReleased  = "released"
	ReservationStatusExpired   = "expired"
	ReservationStatusCancelled = "cancelled"
)

type StockReservation struct {
	ID         uuid.UUID
	TenantID   uuid.UUID
	StoreID    uuid.UUID
	ProductID  uuid.UUID
	OrderID    uuid.UUID
	Quantity   int
	Status     string
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	ReleasedAt *time.Time
}
