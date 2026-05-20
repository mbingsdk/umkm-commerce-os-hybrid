package order

import (
	"time"

	"github.com/google/uuid"
)

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

type ListItemResponse struct {
	ID             uuid.UUID `json:"id"`
	OrderNumber    string    `json:"order_number"`
	Source         string    `json:"source"`
	Status         string    `json:"status"`
	PaymentStatus  string    `json:"payment_status"`
	ShipmentStatus *string   `json:"shipment_status"`
	GrandTotal     int64     `json:"grand_total"`
	CustomerName   string    `json:"customer_name"`
	CustomerPhone  string    `json:"customer_phone"`
	ItemCount      int       `json:"item_count"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DetailResponse struct {
	Order             OrderResponse                `json:"order"`
	Items             []ItemResponse               `json:"order_items"`
	Customer          CustomerSnapshotResponse     `json:"customer_snapshot"`
	ShippingAddress   ShippingAddressResponse      `json:"shipping_address"`
	StatusLogs        []StatusLogResponse          `json:"status_logs"`
	PaymentSummary    PaymentSummaryResponse       `json:"payment_summary"`
	StockReservations []ReservationSummaryResponse `json:"stock_reservations"`
}

type OrderResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrderNumber    string     `json:"order_number"`
	Source         string     `json:"source"`
	Status         string     `json:"status"`
	PaymentStatus  string     `json:"payment_status"`
	ShipmentStatus *string    `json:"shipment_status"`
	Subtotal       int64      `json:"subtotal"`
	DiscountTotal  int64      `json:"discount_total"`
	ShippingCost   int64      `json:"shipping_cost"`
	TaxTotal       int64      `json:"tax_total"`
	GrandTotal     int64      `json:"grand_total"`
	ConfirmedAt    *time.Time `json:"confirmed_at"`
	PaidAt         *time.Time `json:"paid_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	CancelledAt    *time.Time `json:"cancelled_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ItemResponse struct {
	ID            uuid.UUID  `json:"id"`
	ProductID     *uuid.UUID `json:"product_id"`
	ProductName   string     `json:"product_name"`
	SKU           string     `json:"sku,omitempty"`
	Quantity      int        `json:"quantity"`
	UnitPrice     int64      `json:"unit_price"`
	DiscountTotal int64      `json:"discount_total"`
	Subtotal      int64      `json:"subtotal"`
}

type CustomerSnapshotResponse struct {
	ID    *uuid.UUID `json:"id"`
	Name  string     `json:"name"`
	Phone string     `json:"phone"`
	Email string     `json:"email,omitempty"`
}

type ShippingAddressResponse struct {
	Address    string `json:"address"`
	City       string `json:"city,omitempty"`
	Province   string `json:"province,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
}

type StatusLogResponse struct {
	ID         uuid.UUID  `json:"id"`
	FromStatus *string    `json:"from_status"`
	ToStatus   string     `json:"to_status"`
	Note       string     `json:"note,omitempty"`
	CreatedBy  *uuid.UUID `json:"created_by"`
	CreatedAt  time.Time  `json:"created_at"`
}

type PaymentSummaryResponse struct {
	PaymentStatus string     `json:"payment_status"`
	Subtotal      int64      `json:"subtotal"`
	DiscountTotal int64      `json:"discount_total"`
	ShippingCost  int64      `json:"shipping_cost"`
	TaxTotal      int64      `json:"tax_total"`
	GrandTotal    int64      `json:"grand_total"`
	PaidAt        *time.Time `json:"paid_at"`
}

type ReservationSummaryResponse struct {
	Status   string `json:"status"`
	Quantity int    `json:"quantity"`
	Count    int    `json:"count"`
}

type UpdateStatusResponse struct {
	ID          uuid.UUID `json:"id"`
	OrderNumber string    `json:"order_number"`
	Status      string    `json:"status"`
}

type CancelResponse struct {
	ID                   uuid.UUID `json:"id"`
	OrderNumber          string    `json:"order_number"`
	Status               string    `json:"status"`
	ReleasedReservations int       `json:"released_reservations"`
	ReleasedQuantity     int       `json:"released_quantity"`
}

func NewListItemResponse(order Order, itemCount int) ListItemResponse {
	return ListItemResponse{
		ID:             order.ID,
		OrderNumber:    order.OrderNumber,
		Source:         order.Source,
		Status:         order.Status,
		PaymentStatus:  order.PaymentStatus,
		ShipmentStatus: stringPtrFromNonEmpty(order.ShipmentStatus),
		GrandTotal:     order.GrandTotal,
		CustomerName:   order.CustomerName,
		CustomerPhone:  order.CustomerPhone,
		ItemCount:      itemCount,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

func NewDetailResponse(order Order, items []Item, logs []StatusLog, reservations []ReservationSummary) DetailResponse {
	itemResponses := make([]ItemResponse, 0, len(items))
	for idx := range items {
		itemResponses = append(itemResponses, NewItemResponse(items[idx]))
	}

	logResponses := make([]StatusLogResponse, 0, len(logs))
	for idx := range logs {
		logResponses = append(logResponses, NewStatusLogResponse(logs[idx]))
	}

	reservationResponses := make([]ReservationSummaryResponse, 0, len(reservations))
	for idx := range reservations {
		reservationResponses = append(reservationResponses, ReservationSummaryResponse(reservations[idx]))
	}

	return DetailResponse{
		Order:             NewOrderResponse(order),
		Items:             itemResponses,
		Customer:          NewCustomerSnapshotResponse(order),
		ShippingAddress:   NewShippingAddressResponse(order),
		StatusLogs:        logResponses,
		PaymentSummary:    NewPaymentSummaryResponse(order),
		StockReservations: reservationResponses,
	}
}

func NewOrderResponse(order Order) OrderResponse {
	return OrderResponse{
		ID:             order.ID,
		OrderNumber:    order.OrderNumber,
		Source:         order.Source,
		Status:         order.Status,
		PaymentStatus:  order.PaymentStatus,
		ShipmentStatus: stringPtrFromNonEmpty(order.ShipmentStatus),
		Subtotal:       order.Subtotal,
		DiscountTotal:  order.DiscountTotal,
		ShippingCost:   order.ShippingCost,
		TaxTotal:       order.TaxTotal,
		GrandTotal:     order.GrandTotal,
		ConfirmedAt:    order.ConfirmedAt,
		PaidAt:         order.PaidAt,
		CompletedAt:    order.CompletedAt,
		CancelledAt:    order.CancelledAt,
		CreatedAt:      order.CreatedAt,
		UpdatedAt:      order.UpdatedAt,
	}
}

func NewItemResponse(item Item) ItemResponse {
	return ItemResponse{
		ID:            item.ID,
		ProductID:     item.ProductID,
		ProductName:   item.ProductName,
		SKU:           item.SKU,
		Quantity:      item.Quantity,
		UnitPrice:     item.UnitPrice,
		DiscountTotal: item.DiscountTotal,
		Subtotal:      item.Subtotal,
	}
}

func NewCustomerSnapshotResponse(order Order) CustomerSnapshotResponse {
	return CustomerSnapshotResponse{
		ID:    order.CustomerID,
		Name:  order.CustomerName,
		Phone: order.CustomerPhone,
		Email: order.CustomerEmail,
	}
}

func NewShippingAddressResponse(order Order) ShippingAddressResponse {
	return ShippingAddressResponse{
		Address:    order.ShippingAddress,
		City:       order.ShippingCity,
		Province:   order.ShippingProvince,
		PostalCode: order.ShippingPostalCode,
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

func NewPaymentSummaryResponse(order Order) PaymentSummaryResponse {
	return PaymentSummaryResponse{
		PaymentStatus: order.PaymentStatus,
		Subtotal:      order.Subtotal,
		DiscountTotal: order.DiscountTotal,
		ShippingCost:  order.ShippingCost,
		TaxTotal:      order.TaxTotal,
		GrandTotal:    order.GrandTotal,
		PaidAt:        order.PaidAt,
	}
}

func NewUpdateStatusResponse(order Order) UpdateStatusResponse {
	return UpdateStatusResponse{
		ID:          order.ID,
		OrderNumber: order.OrderNumber,
		Status:      order.Status,
	}
}

func NewCancelResponse(order Order, releasedReservations int, releasedQuantity int) CancelResponse {
	return CancelResponse{
		ID:                   order.ID,
		OrderNumber:          order.OrderNumber,
		Status:               order.Status,
		ReleasedReservations: releasedReservations,
		ReleasedQuantity:     releasedQuantity,
	}
}

func stringPtrFromNonEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
