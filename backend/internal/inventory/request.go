package inventory

type AdjustStockRequest struct {
	AdjustmentType string `json:"adjustment_type"`
	Type           string `json:"type"`
	Quantity       int    `json:"quantity"`
	Reason         string `json:"reason"`
	Note           string `json:"note"`
}

type UpdateThresholdRequest struct {
	LowStockThreshold int `json:"low_stock_threshold"`
}
