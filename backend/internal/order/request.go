package order

type UpdateStatusRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

type CancelRequest struct {
	Reason string `json:"reason"`
	Note   string `json:"note"`
}
