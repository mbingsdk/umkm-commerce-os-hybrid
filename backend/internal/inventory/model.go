package inventory

import "github.com/google/uuid"

const (
	MovementTypeInitial  = "initial"
	MovementTypeReserved = "reserved"
)

type StockSnapshot struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductID         uuid.UUID
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
	LowStockThreshold int
}

type StockMovement struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	ProductID     uuid.UUID
	MovementType  string
	Quantity      int
	BalanceAfter  int
	ReferenceType string
	ReferenceID   *uuid.UUID
	Note          string
	CreatedBy     *uuid.UUID
}

type CreateSnapshotParams struct {
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductID         uuid.UUID
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
	LowStockThreshold int
}

type CreateMovementParams struct {
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	ProductID     uuid.UUID
	MovementType  string
	Quantity      int
	BalanceAfter  int
	ReferenceType string
	ReferenceID   *uuid.UUID
	Note          string
	CreatedBy     *uuid.UUID
}
