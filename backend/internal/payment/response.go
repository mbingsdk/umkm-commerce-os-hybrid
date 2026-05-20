package payment

import (
	"time"

	"github.com/google/uuid"
)

type ConfirmationResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrderID        uuid.UUID  `json:"order_id"`
	PayerName      string     `json:"payer_name"`
	BankName       string     `json:"bank_name"`
	TransferAmount int64      `json:"transfer_amount"`
	TransferDate   time.Time  `json:"transfer_date"`
	ProofURL       string     `json:"proof_url,omitempty"`
	Note           string     `json:"note,omitempty"`
	Status         string     `json:"status"`
	ReviewedBy     *uuid.UUID `json:"reviewed_by"`
	ReviewedAt     *time.Time `json:"reviewed_at"`
	ReviewNote     string     `json:"review_note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

type PublicConfirmationResponse struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	OrderNumber string    `json:"order_number"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
}

type ConfirmPaymentResponse struct {
	OrderID       uuid.UUID `json:"order_id"`
	OrderNumber   string    `json:"order_number"`
	OrderStatus   string    `json:"order_status"`
	PaymentStatus string    `json:"payment_status"`
	PaymentID     uuid.UUID `json:"payment_id"`
}

type RejectPaymentResponse struct {
	OrderID        uuid.UUID `json:"order_id"`
	OrderNumber    string    `json:"order_number"`
	PaymentStatus  string    `json:"payment_status"`
	ConfirmationID uuid.UUID `json:"payment_confirmation_id"`
	Status         string    `json:"status"`
}

func NewConfirmationResponse(item Confirmation) ConfirmationResponse {
	return ConfirmationResponse{
		ID:             item.ID,
		OrderID:        item.OrderID,
		PayerName:      item.PayerName,
		BankName:       item.BankName,
		TransferAmount: item.TransferAmount,
		TransferDate:   item.TransferDate,
		ProofURL:       item.ProofURL,
		Note:           item.Note,
		Status:         item.Status,
		ReviewedBy:     item.ReviewedBy,
		ReviewedAt:     item.ReviewedAt,
		ReviewNote:     item.ReviewNote,
		CreatedAt:      item.CreatedAt,
	}
}
