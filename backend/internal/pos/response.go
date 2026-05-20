package pos

import "github.com/google/uuid"

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

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

type POSProductResponse struct {
	ProductID      string  `json:"product_id"`
	Name           string  `json:"name"`
	SKU            string  `json:"sku,omitempty"`
	Barcode        string  `json:"barcode,omitempty"`
	Price          int64   `json:"price"`
	Image          string  `json:"image,omitempty"`
	AvailableStock int     `json:"available_stock"`
	CategoryID     *string `json:"category_id,omitempty"`
	CategoryName   string  `json:"category_name,omitempty"`
}

type TransactionResponse struct {
	ID                string `json:"id"`
	TransactionNumber string `json:"transaction_number"`
	SessionID         string `json:"session_id"`
	CashierID         string `json:"cashier_id"`
	Subtotal          int64  `json:"subtotal"`
	DiscountTotal     int64  `json:"discount_total"`
	TaxTotal          int64  `json:"tax_total"`
	GrandTotal        int64  `json:"grand_total"`
	PaymentMethod     string `json:"payment_method"`
	AmountPaid        int64  `json:"amount_paid"`
	ChangeAmount      int64  `json:"change_amount"`
	Status            string `json:"status"`
	CreatedAt         string `json:"created_at"`
}

type TransactionDetailResponse struct {
	Transaction TransactionResponse       `json:"transaction"`
	Items       []TransactionItemResponse `json:"items"`
}

type TransactionItemResponse struct {
	ID            string  `json:"id"`
	ProductID     *string `json:"product_id,omitempty"`
	ProductName   string  `json:"product_name"`
	SKU           string  `json:"sku,omitempty"`
	Quantity      int     `json:"quantity"`
	UnitPrice     int64   `json:"unit_price"`
	DiscountTotal int64   `json:"discount_total"`
	Subtotal      int64   `json:"subtotal"`
}

func NewPOSProductResponse(item POSProduct) POSProductResponse {
	var categoryID *string
	if item.CategoryID != nil {
		value := item.CategoryID.String()
		categoryID = &value
	}
	return POSProductResponse{
		ProductID:      item.ID.String(),
		Name:           item.Name,
		SKU:            item.SKU,
		Barcode:        item.Barcode,
		Price:          item.Price,
		Image:          item.PrimaryImageURL,
		AvailableStock: item.StockAvailable,
		CategoryID:     categoryID,
		CategoryName:   item.CategoryName,
	}
}

func NewTransactionResponse(transaction POSTransaction) TransactionResponse {
	return TransactionResponse{
		ID:                transaction.ID.String(),
		TransactionNumber: transaction.TransactionNumber,
		SessionID:         transaction.CashierSessionID.String(),
		CashierID:         transaction.CashierID.String(),
		Subtotal:          transaction.Subtotal,
		DiscountTotal:     transaction.DiscountTotal,
		TaxTotal:          transaction.TaxTotal,
		GrandTotal:        transaction.GrandTotal,
		PaymentMethod:     transaction.PaymentMethod,
		AmountPaid:        transaction.PaymentAmount,
		ChangeAmount:      transaction.ChangeAmount,
		Status:            transaction.Status,
		CreatedAt:         transaction.CreatedAt.Format(timeFormatRFC3339),
	}
}

func NewTransactionDetailResponse(transaction POSTransaction, items []POSTransactionItem) TransactionDetailResponse {
	responseItems := make([]TransactionItemResponse, 0, len(items))
	for _, item := range items {
		responseItems = append(responseItems, NewTransactionItemResponse(item))
	}
	return TransactionDetailResponse{
		Transaction: NewTransactionResponse(transaction),
		Items:       responseItems,
	}
}

func NewTransactionItemResponse(item POSTransactionItem) TransactionItemResponse {
	var productID *string
	if item.ProductID != nil && *item.ProductID != uuid.Nil {
		value := item.ProductID.String()
		productID = &value
	}
	return TransactionItemResponse{
		ID:            item.ID.String(),
		ProductID:     productID,
		ProductName:   item.ProductName,
		SKU:           item.SKU,
		Quantity:      item.Quantity,
		UnitPrice:     item.UnitPrice,
		DiscountTotal: item.DiscountTotal,
		Subtotal:      item.Subtotal,
	}
}
