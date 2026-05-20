package inventory

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	MovementTypeInitial       = "initial"
	MovementTypeReserved      = "reserved"
	MovementTypeReleased      = "released"
	MovementTypeAdjustmentIn  = "adjustment_in"
	MovementTypeAdjustmentOut = "adjustment_out"

	AdjustmentTypeIn  = "in"
	AdjustmentTypeOut = "out"

	ReferenceTypeManualAdjustment = "manual_adjustment"

	AuditActionStockAdjusted    = "inventory.stock_adjusted"
	AuditActionThresholdUpdated = "inventory.threshold_updated"

	EventStockAdjusted = "StockAdjusted"
	AggregateProduct   = "product"
)

var (
	ErrProductNotFound       = errors.New("product not found")
	ErrStockSnapshotNotFound = errors.New("stock snapshot not found")
	errInvalidCursor         = errors.New("invalid cursor")
)

type ProductRef struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	CategoryID     *uuid.UUID
	Name           string
	SKU            string
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
	LowStockThreshold int
	UpdatedAt         time.Time
}

type StockListItem struct {
	ProductID         uuid.UUID
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductName       string
	SKU               string
	CategoryID        *uuid.UUID
	CategoryName      string
	PrimaryImageURL   string
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
	LowStockThreshold int
	UpdatedAt         time.Time
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
	Reason        string
	Note          string
	CreatedBy     *uuid.UUID
	CreatedByName string
	CreatedAt     time.Time
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
	Reason        string
	Note          string
	CreatedBy     *uuid.UUID
}

type ListStockFilters struct {
	Query      string
	LowStock   *bool
	OutOfStock *bool
	CategoryID *uuid.UUID
	Limit      int
	Cursor     *Cursor
}

type ListMovementFilters struct {
	Limit  int
	Cursor *Cursor
}

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

func EncodeStockCursor(item StockListItem) (string, error) {
	return encodeCursor(Cursor{CreatedAt: item.UpdatedAt, ID: item.ProductID})
}

func EncodeMovementCursor(item StockMovement) (string, error) {
	return encodeCursor(Cursor{CreatedAt: item.CreatedAt, ID: item.ID})
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

func encodeCursor(cursor Cursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}
