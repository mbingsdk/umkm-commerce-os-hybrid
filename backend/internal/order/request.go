package order

type UpdateStatusRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}
