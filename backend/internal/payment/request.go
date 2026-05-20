package payment

type PublicConfirmationRequest struct {
	CustomerPhone  string `json:"customer_phone"`
	PayerName      string `json:"payer_name"`
	BankName       string `json:"bank_name"`
	TransferAmount int64  `json:"transfer_amount"`
	TransferDate   string `json:"transfer_date"`
	ProofURL       string `json:"proof_url"`
	Note           string `json:"note"`
}

type ReviewPaymentRequest struct {
	PaymentConfirmationID string `json:"payment_confirmation_id"`
	Note                  string `json:"note"`
}
