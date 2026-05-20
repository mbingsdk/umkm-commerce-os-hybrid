package payment

import (
	"time"

	"github.com/google/uuid"
)

const (
	MethodManualTransfer = "manual_transfer"

	ConfirmationStatusPending   = "pending"
	ConfirmationStatusConfirmed = "confirmed"
	ConfirmationStatusRejected  = "rejected"

	PaymentStatusWaitingConfirmation = "waiting_confirmation"
	PaymentStatusPaid                = "paid"
	PaymentStatusFailed              = "failed"

	EventPaymentConfirmed      = "PaymentConfirmed"
	EventOrderPaid             = "OrderPaid"
	EventNotificationRequested = "NotificationRequested"

	AggregatePayment = "payment"
	AggregateOrder   = "order"
)

type Confirmation struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	OrderID        uuid.UUID
	PayerName      string
	BankName       string
	TransferAmount int64
	TransferDate   time.Time
	ProofURL       string
	Note           string
	Status         string
	ReviewedBy     *uuid.UUID
	ReviewedAt     *time.Time
	ReviewNote     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Payment struct {
	ID                    uuid.UUID
	TenantID              uuid.UUID
	StoreID               uuid.UUID
	OrderID               uuid.UUID
	PaymentConfirmationID *uuid.UUID
	Method                string
	Status                string
	Amount                int64
	PayerName             string
	BankName              string
	ProofURL              string
	Note                  string
	PaidAt                *time.Time
	ConfirmedBy           *uuid.UUID
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type CreateConfirmationParams struct {
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	OrderID        uuid.UUID
	PayerName      string
	BankName       string
	TransferAmount int64
	TransferDate   time.Time
	ProofURL       string
	Note           string
}

type ReviewConfirmationParams struct {
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	OrderID        uuid.UUID
	ConfirmationID uuid.UUID
	Status         string
	ReviewedBy     uuid.UUID
	ReviewNote     string
}

type CreatePaymentParams struct {
	TenantID              uuid.UUID
	StoreID               uuid.UUID
	OrderID               uuid.UUID
	PaymentConfirmationID uuid.UUID
	Method                string
	Status                string
	Amount                int64
	PayerName             string
	BankName              string
	ProofURL              string
	Note                  string
	ConfirmedBy           uuid.UUID
}
