package shipment

import (
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
)

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type ShipmentResponse struct {
	ID              uuid.UUID  `json:"id"`
	OrderID         uuid.UUID  `json:"order_id"`
	OrderNumber     string     `json:"order_number"`
	CourierType     string     `json:"courier_type"`
	CourierName     string     `json:"courier_name,omitempty"`
	TrackingNumber  string     `json:"tracking_number,omitempty"`
	Status          string     `json:"status"`
	ShippingCost    int64      `json:"shipping_cost"`
	AssignedToName  string     `json:"assigned_to_name,omitempty"`
	AssignedToPhone string     `json:"assigned_to_phone,omitempty"`
	Note            string     `json:"note,omitempty"`
	ShippedAt       *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt     *time.Time `json:"delivered_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type DetailResponse struct {
	Shipment ShipmentResponse    `json:"shipment"`
	Timeline []StatusLogResponse `json:"timeline"`
}

type StatusLogResponse struct {
	ID         uuid.UUID  `json:"id"`
	FromStatus *string    `json:"from_status,omitempty"`
	ToStatus   string     `json:"to_status"`
	Note       string     `json:"note,omitempty"`
	CreatedBy  *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateShipmentResponse struct {
	ID             uuid.UUID `json:"id"`
	OrderID        uuid.UUID `json:"order_id"`
	TrackingNumber string    `json:"tracking_number,omitempty"`
	Status         string    `json:"status"`
	ShippingCost   int64     `json:"shipping_cost"`
}

type UpdateStatusResponse struct {
	ID     uuid.UUID `json:"id"`
	Status string    `json:"status"`
}

type PublicTrackingResponse struct {
	OrderNumber    string                     `json:"order_number"`
	Status         string                     `json:"status"`
	PaymentStatus  string                     `json:"payment_status"`
	ShipmentStatus string                     `json:"shipment_status,omitempty"`
	CustomerName   string                     `json:"customer_name"`
	Items          []PublicTrackingItem       `json:"items"`
	Totals         PublicTrackingTotals       `json:"totals"`
	Shipment       *PublicTrackingShipment    `json:"shipment,omitempty"`
	Timeline       []PublicTrackingStatusItem `json:"timeline"`
}

type PublicTrackingItem struct {
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Subtotal    int64  `json:"subtotal"`
}

type PublicTrackingTotals struct {
	Subtotal     int64 `json:"subtotal"`
	ShippingCost int64 `json:"shipping_cost"`
	GrandTotal   int64 `json:"grand_total"`
}

type PublicTrackingShipment struct {
	CourierType    string     `json:"courier_type"`
	CourierName    string     `json:"courier_name,omitempty"`
	TrackingNumber string     `json:"tracking_number,omitempty"`
	Status         string     `json:"status"`
	ShippingCost   int64      `json:"shipping_cost"`
	ShippedAt      *time.Time `json:"shipped_at,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
}

type PublicTrackingStatusItem struct {
	Status    string    `json:"status"`
	Note      string    `json:"note,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func NewShipmentResponse(item Shipment) ShipmentResponse {
	return ShipmentResponse{
		ID:              item.ID,
		OrderID:         item.OrderID,
		OrderNumber:     item.OrderNumber,
		CourierType:     item.CourierType,
		CourierName:     item.CourierName,
		TrackingNumber:  item.TrackingNumber,
		Status:          item.Status,
		ShippingCost:    item.ShippingCost,
		AssignedToName:  item.AssignedToName,
		AssignedToPhone: item.AssignedToPhone,
		Note:            item.Note,
		ShippedAt:       item.ShippedAt,
		DeliveredAt:     item.DeliveredAt,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	}
}

func NewShipmentResponses(items []Shipment) []ShipmentResponse {
	responses := make([]ShipmentResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, NewShipmentResponse(item))
	}
	return responses
}

func NewDetailResponse(item Shipment, logs []StatusLog) DetailResponse {
	logResponses := make([]StatusLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponses = append(logResponses, NewStatusLogResponse(log))
	}
	return DetailResponse{
		Shipment: NewShipmentResponse(item),
		Timeline: logResponses,
	}
}

func NewStatusLogResponse(log StatusLog) StatusLogResponse {
	return StatusLogResponse{
		ID:         log.ID,
		FromStatus: stringPtrFromNonEmpty(log.FromStatus),
		ToStatus:   log.ToStatus,
		Note:       log.Note,
		CreatedBy:  log.CreatedBy,
		CreatedAt:  log.CreatedAt,
	}
}

func NewCreateShipmentResponse(item Shipment) CreateShipmentResponse {
	return CreateShipmentResponse{
		ID:             item.ID,
		OrderID:        item.OrderID,
		TrackingNumber: item.TrackingNumber,
		Status:         item.Status,
		ShippingCost:   item.ShippingCost,
	}
}

func NewUpdateStatusResponse(item Shipment) UpdateStatusResponse {
	return UpdateStatusResponse{
		ID:     item.ID,
		Status: item.Status,
	}
}

func NewPublicTrackingResponse(orderRecord order.Order, items []order.Item, shipmentRecord *Shipment, logs []StatusLog) PublicTrackingResponse {
	publicItems := make([]PublicTrackingItem, 0, len(items))
	for _, item := range items {
		publicItems = append(publicItems, PublicTrackingItem{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Subtotal:    item.Subtotal,
		})
	}

	timeline := make([]PublicTrackingStatusItem, 0, len(logs))
	for _, log := range logs {
		timeline = append(timeline, PublicTrackingStatusItem{
			Status:    log.ToStatus,
			Note:      publicTrackingNote(log.Note),
			CreatedAt: log.CreatedAt,
		})
	}

	var publicShipment *PublicTrackingShipment
	if shipmentRecord != nil {
		publicShipment = &PublicTrackingShipment{
			CourierType:    shipmentRecord.CourierType,
			CourierName:    shipmentRecord.CourierName,
			TrackingNumber: shipmentRecord.TrackingNumber,
			Status:         shipmentRecord.Status,
			ShippingCost:   shipmentRecord.ShippingCost,
			ShippedAt:      shipmentRecord.ShippedAt,
			DeliveredAt:    shipmentRecord.DeliveredAt,
		}
	}

	return PublicTrackingResponse{
		OrderNumber:    orderRecord.OrderNumber,
		Status:         orderRecord.Status,
		PaymentStatus:  orderRecord.PaymentStatus,
		ShipmentStatus: orderRecord.ShipmentStatus,
		CustomerName:   orderRecord.CustomerName,
		Items:          publicItems,
		Totals: PublicTrackingTotals{
			Subtotal:     orderRecord.Subtotal,
			ShippingCost: orderRecord.ShippingCost,
			GrandTotal:   orderRecord.GrandTotal,
		},
		Shipment: publicShipment,
		Timeline: timeline,
	}
}

func publicTrackingNote(note string) string {
	switch note {
	case "Shipment created":
		return "Pengiriman dibuat"
	case "Shipment delivered":
		return "Pesanan telah diterima"
	default:
		return ""
	}
}

func stringPtrFromNonEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
