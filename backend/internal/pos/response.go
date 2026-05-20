package pos

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"

type SessionResponse struct {
	ID            string `json:"id"`
	SessionNumber string `json:"session_number"`
	CashierID     string `json:"cashier_id"`
	OpeningCash   int64  `json:"opening_cash"`
	ClosingCash   *int64 `json:"closing_cash,omitempty"`
	ExpectedCash  *int64 `json:"expected_cash,omitempty"`
	Difference    *int64 `json:"difference,omitempty"`
	Status        string `json:"status"`
	OpenedAt      string `json:"opened_at"`
	ClosedAt      string `json:"closed_at,omitempty"`
}

func NewSessionResponse(session CashierSession) SessionResponse {
	response := SessionResponse{
		ID:            session.ID.String(),
		SessionNumber: session.SessionNumber,
		CashierID:     session.CashierID.String(),
		OpeningCash:   session.OpeningCash,
		ClosingCash:   session.ClosingCash,
		ExpectedCash:  session.ExpectedCash,
		Difference:    session.Difference,
		Status:        session.Status,
		OpenedAt:      session.OpenedAt.Format(timeFormatRFC3339),
	}
	if session.ClosedAt != nil {
		response.ClosedAt = session.ClosedAt.Format(timeFormatRFC3339)
	}
	return response
}
