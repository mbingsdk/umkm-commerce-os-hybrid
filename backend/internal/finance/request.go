package finance

import "github.com/google/uuid"

type CreateExpenseRequest struct {
	CategoryID    *uuid.UUID `json:"category_id"`
	Category      string     `json:"category"`
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Amount        int64      `json:"amount"`
	ExpenseDate   string     `json:"expense_date"`
	PaymentMethod string     `json:"payment_method"`
	Note          string     `json:"note"`
}

type UpdateExpenseRequest struct {
	CategoryID    *uuid.UUID `json:"category_id"`
	Category      *string    `json:"category"`
	Title         *string    `json:"title"`
	Description   *string    `json:"description"`
	Amount        *int64     `json:"amount"`
	ExpenseDate   *string    `json:"expense_date"`
	PaymentMethod *string    `json:"payment_method"`
	Note          *string    `json:"note"`
}
