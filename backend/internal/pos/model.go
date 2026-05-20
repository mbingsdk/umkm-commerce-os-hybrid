package pos

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	SessionStatusOpen      = "open"
	SessionStatusClosed    = "closed"
	SessionStatusCancelled = "cancelled"

	PaymentMethodCash       = "cash"
	PaymentMethodQRISManual = "qris_manual"
	TransactionStatusDone   = "completed"
	AuditActionSessionOpen  = "pos.session_opened"
	AuditActionSessionClose = "pos.session_closed"

	EventCashierSessionOpened  = "CashierSessionOpened"
	EventCashierSessionClosed  = "CashierSessionClosed"
	EventPOSTransactionCreated = "POSTransactionCreated"
	EventStockReduced          = "StockReduced"
	EventNotificationRequested = "NotificationRequested"
	AggregateCashierSession    = "cashier_session"
	AggregatePOSTransaction    = "pos_transaction"
)

var (
	ErrSessionNotFound       = errors.New("cashier session not found")
	ErrOpenSessionExists     = errors.New("cashier already has an open session")
	ErrSessionAlreadyDone    = errors.New("cashier session is already closed")
	ErrTransactionNotFound   = errors.New("pos transaction not found")
	ErrStockSnapshotNotFound = errors.New("stock snapshot not found")
	errInvalidCursor         = errors.New("invalid cursor")
)

type CashierSession struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	CashierID     uuid.UUID
	SessionNumber string
	OpeningCash   int64
	ClosingCash   *int64
	ExpectedCash  *int64
	Difference    *int64
	Status        string
	OpenedAt      time.Time
	ClosedAt      *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateSessionParams struct {
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	CashierID     uuid.UUID
	SessionNumber string
	OpeningCash   int64
}

type CloseSessionParams struct {
	TenantID     uuid.UUID
	StoreID      uuid.UUID
	SessionID    uuid.UUID
	ClosingCash  int64
	ExpectedCash int64
	Difference   int64
}

type POSProduct struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	StoreID         uuid.UUID
	CategoryID      *uuid.UUID
	CategoryName    string
	Name            string
	SKU             string
	Barcode         string
	Price           int64
	Status          string
	PrimaryImageURL string
	StockAvailable  int
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

type POSTransaction struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	CashierSessionID  uuid.UUID
	CashierID         uuid.UUID
	TransactionNumber string
	Subtotal          int64
	DiscountTotal     int64
	TaxTotal          int64
	GrandTotal        int64
	PaymentMethod     string
	PaymentAmount     int64
	ChangeAmount      int64
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type POSTransactionItem struct {
	ID               uuid.UUID
	TenantID         uuid.UUID
	POSTransactionID uuid.UUID
	ProductID        *uuid.UUID
	ProductName      string
	SKU              string
	Quantity         int
	UnitPrice        int64
	DiscountTotal    int64
	Subtotal         int64
	CreatedAt        time.Time
}

type ProductSearchFilters struct {
	Query   string
	Barcode string
	Limit   int
}

type TransactionListFilters struct {
	DateFrom      *time.Time
	DateTo        *time.Time
	PaymentMethod *string
	CashierID     *uuid.UUID
	Limit         int
	Cursor        *Cursor
}

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

type CreateTransactionParams struct {
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	CashierSessionID  uuid.UUID
	CashierID         uuid.UUID
	TransactionNumber string
	Subtotal          int64
	DiscountTotal     int64
	TaxTotal          int64
	GrandTotal        int64
	PaymentMethod     string
	PaymentAmount     int64
	ChangeAmount      int64
}

type CreateTransactionItemParams struct {
	TenantID         uuid.UUID
	POSTransactionID uuid.UUID
	ProductID        uuid.UUID
	ProductName      string
	SKU              string
	Quantity         int
	UnitPrice        int64
	DiscountTotal    int64
	Subtotal         int64
}

type UpdateStockSnapshotParams struct {
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductID         uuid.UUID
	QuantityOnHand    int
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
	CreatedBy     uuid.UUID
}

func EncodeTransactionCursor(transaction POSTransaction) (string, error) {
	payload, err := json.Marshal(Cursor{CreatedAt: transaction.CreatedAt, ID: transaction.ID})
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
