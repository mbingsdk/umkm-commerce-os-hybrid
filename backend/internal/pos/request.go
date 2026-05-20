package pos

import "github.com/google/uuid"

type OpenSessionRequest struct {
	OpeningCashAmount *int64 `json:"opening_cash_amount"`
	OpeningCash       *int64 `json:"opening_cash"`
	Note              string `json:"note"`
}

type CreateTransactionRequest struct {
	SessionID        uuid.UUID                      `json:"session_id"`
	CashierSessionID uuid.UUID                      `json:"cashier_session_id"`
	Items            []CreateTransactionItemRequest `json:"items"`
	PaymentMethod    string                         `json:"payment_method"`
	AmountPaid       *int64                         `json:"amount_paid"`
	PaymentAmount    *int64                         `json:"payment_amount"`
	Note             string                         `json:"note"`
}

type CreateTransactionItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}

type CloseSessionRequest struct {
	ClosingCashAmount *int64 `json:"closing_cash_amount"`
	ClosingCash       *int64 `json:"closing_cash"`
	Note              string `json:"note"`
}
