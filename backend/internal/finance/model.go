package finance

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	PaymentMethodCash         = "cash"
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodQRISManual   = "qris_manual"
	PaymentMethodOther        = "other"

	AuditActionExpenseCreated = "finance.expense_created"
	AuditActionExpenseUpdated = "finance.expense_updated"
	AuditActionExpenseDeleted = "finance.expense_deleted"

	EventExpenseChanged = "ExpenseChanged"
	AggregateExpense    = "expense"
)

var (
	ErrExpenseNotFound  = errors.New("expense not found")
	ErrCategoryNotFound = errors.New("expense category not found")
	errInvalidCursor    = errors.New("invalid cursor")
)

type ExpenseCategory struct {
	ID        uuid.UUID
	TenantID  *uuid.UUID
	StoreID   *uuid.UUID
	Name      string
	Slug      string
	IsSystem  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Expense struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	CategoryID    *uuid.UUID
	CategoryName  string
	CategorySlug  string
	Title         string
	Amount        int64
	ExpenseDate   time.Time
	PaymentMethod string
	Note          string
	CreatedBy     *uuid.UUID
	CreatedByName string
	UpdatedBy     *uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

type ListExpenseFilters struct {
	DateFrom     *time.Time
	DateTo       *time.Time
	CategoryID   *uuid.UUID
	CategorySlug string
	Query        string
	Limit        int
	Cursor       *Cursor
}

type Cursor struct {
	ExpenseDate time.Time `json:"expense_date"`
	CreatedAt   time.Time `json:"created_at"`
	ID          uuid.UUID `json:"id"`
}

type CreateExpenseParams struct {
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	CategoryID    *uuid.UUID
	Title         string
	Amount        int64
	ExpenseDate   time.Time
	PaymentMethod string
	Note          string
	CreatedBy     uuid.UUID
}

type UpdateExpenseParams struct {
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	ExpenseID     uuid.UUID
	CategoryID    *uuid.UUID
	Title         string
	Amount        int64
	ExpenseDate   time.Time
	PaymentMethod string
	Note          string
	UpdatedBy     uuid.UUID
}

func EncodeExpenseCursor(item Expense) (string, error) {
	payload, err := json.Marshal(Cursor{
		ExpenseDate: item.ExpenseDate,
		CreatedAt:   item.CreatedAt,
		ID:          item.ID,
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
	if cursor.ID == uuid.Nil || cursor.CreatedAt.IsZero() || cursor.ExpenseDate.IsZero() {
		return nil, errInvalidCursor
	}
	return &cursor, nil
}
