package pos

type OpenSessionRequest struct {
	OpeningCashAmount *int64 `json:"opening_cash_amount"`
	OpeningCash       *int64 `json:"opening_cash"`
	Note              string `json:"note"`
}

type CloseSessionRequest struct {
	ClosingCashAmount *int64 `json:"closing_cash_amount"`
	ClosingCash       *int64 `json:"closing_cash"`
	Note              string `json:"note"`
}
