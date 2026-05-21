package shipment

type CreateShipmentRequest struct {
	CourierType     string `json:"courier_type"`
	CourierName     string `json:"courier_name"`
	TrackingNumber  string `json:"tracking_number"`
	ShippingCost    int64  `json:"shipping_cost"`
	AssignedToName  string `json:"assigned_to_name"`
	AssignedToPhone string `json:"assigned_to_phone"`
	Note            string `json:"note"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}
