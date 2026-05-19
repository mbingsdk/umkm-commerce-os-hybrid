package checkout

import "github.com/google/uuid"

type CheckoutResponse struct {
	OrderID            uuid.UUID                   `json:"order_id"`
	OrderNumber        string                      `json:"order_number"`
	Status             string                      `json:"status"`
	PaymentStatus      string                      `json:"payment_status"`
	Totals             CheckoutTotalsResponse      `json:"totals"`
	PaymentInstruction CheckoutPaymentInstructions `json:"payment_instruction"`
}

type CheckoutTotalsResponse struct {
	Subtotal      int64 `json:"subtotal"`
	DiscountTotal int64 `json:"discount_total"`
	ShippingCost  int64 `json:"shipping_cost"`
	TaxTotal      int64 `json:"tax_total"`
	GrandTotal    int64 `json:"grand_total"`
}

type CheckoutPaymentInstructions struct {
	Method  string `json:"method"`
	Message string `json:"message"`
}

type CheckoutResult struct {
	Response   CheckoutResponse
	StatusCode int
}
