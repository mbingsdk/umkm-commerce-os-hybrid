package checkout

import "github.com/google/uuid"

const PaymentMethodManualTransfer = "manual_transfer"

type CheckoutRequest struct {
	Items           []CheckoutItemRequest    `json:"items"`
	Customer        CheckoutCustomerRequest  `json:"customer"`
	ShippingAddress CheckoutAddressRequest   `json:"shipping_address"`
	PaymentMethod   string                   `json:"payment_method,omitempty"`
	CustomerNote    string                   `json:"customer_note,omitempty"`
	Shipping        *CheckoutShippingRequest `json:"shipping,omitempty"`
}

type CheckoutItemRequest struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     *int64    `json:"price,omitempty"`
}

type CheckoutCustomerRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email,omitempty"`
}

type CheckoutAddressRequest struct {
	Label          string `json:"label,omitempty"`
	RecipientName  string `json:"recipient_name,omitempty"`
	RecipientPhone string `json:"recipient_phone,omitempty"`
	Address        string `json:"address"`
	City           string `json:"city,omitempty"`
	Province       string `json:"province,omitempty"`
	PostalCode     string `json:"postal_code,omitempty"`
}

type CheckoutShippingRequest struct {
	CourierZoneID *uuid.UUID `json:"courier_zone_id,omitempty"`
	CourierCode   string     `json:"courier_code,omitempty"`
	ServiceCode   string     `json:"service_code,omitempty"`
}
