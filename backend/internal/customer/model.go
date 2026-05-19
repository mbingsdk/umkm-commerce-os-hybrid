package customer

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	Name        string
	Phone       string
	Email       string
	Notes       string
	TotalOrders int
	TotalSpent  int64
	LastOrderAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type Address struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	CustomerID     uuid.UUID
	Label          string
	RecipientName  string
	RecipientPhone string
	Address        string
	City           string
	Province       string
	PostalCode     string
	Latitude       *float64
	Longitude      *float64
	IsDefault      bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
