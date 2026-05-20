package pos

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	SessionStatusOpen      = "open"
	SessionStatusClosed    = "closed"
	SessionStatusCancelled = "cancelled"

	PaymentMethodCash       = "cash"
	TransactionStatusDone   = "completed"
	AuditActionSessionOpen  = "pos.session_opened"
	AuditActionSessionClose = "pos.session_closed"

	EventCashierSessionOpened = "CashierSessionOpened"
	EventCashierSessionClosed = "CashierSessionClosed"
	AggregateCashierSession   = "cashier_session"
)

var (
	ErrSessionNotFound    = errors.New("cashier session not found")
	ErrOpenSessionExists  = errors.New("cashier already has an open session")
	ErrSessionAlreadyDone = errors.New("cashier session is already closed")
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
