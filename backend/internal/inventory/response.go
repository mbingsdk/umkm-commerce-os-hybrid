package inventory

import "github.com/google/uuid"

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type StockResponse struct {
	ProductID         string  `json:"product_id"`
	Name              string  `json:"name"`
	SKU               string  `json:"sku,omitempty"`
	CategoryID        *string `json:"category_id,omitempty"`
	CategoryName      string  `json:"category_name,omitempty"`
	PrimaryImageURL   string  `json:"primary_image_url,omitempty"`
	QuantityOnHand    int     `json:"quantity_on_hand"`
	QuantityReserved  int     `json:"quantity_reserved"`
	QuantityAvailable int     `json:"quantity_available"`
	LowStockThreshold int     `json:"low_stock_threshold"`
	IsLowStock        bool    `json:"is_low_stock"`
	IsOutOfStock      bool    `json:"is_out_of_stock"`
	UpdatedAt         string  `json:"updated_at"`
}

type StockMovementResponse struct {
	ID            string         `json:"id"`
	ProductID     string         `json:"product_id"`
	MovementType  string         `json:"movement_type"`
	Quantity      int            `json:"quantity"`
	BalanceAfter  int            `json:"balance_after"`
	ReferenceType string         `json:"reference_type,omitempty"`
	ReferenceID   *string        `json:"reference_id,omitempty"`
	Reason        string         `json:"reason,omitempty"`
	Note          string         `json:"note,omitempty"`
	CreatedBy     *ActorResponse `json:"created_by,omitempty"`
	CreatedAt     string         `json:"created_at"`
}

type ActorResponse struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

type AdjustStockResponse struct {
	ProductID         string `json:"product_id"`
	QuantityOnHand    int    `json:"quantity_on_hand"`
	QuantityReserved  int    `json:"quantity_reserved"`
	QuantityAvailable int    `json:"quantity_available"`
	LowStockThreshold int    `json:"low_stock_threshold"`
	MovementID        string `json:"movement_id"`
}

type UpdateThresholdResponse struct {
	ProductID         string `json:"product_id"`
	LowStockThreshold int    `json:"low_stock_threshold"`
}

func NewStockResponse(item StockListItem) StockResponse {
	var categoryID *string
	if item.CategoryID != nil {
		value := item.CategoryID.String()
		categoryID = &value
	}

	return StockResponse{
		ProductID:         item.ProductID.String(),
		Name:              item.ProductName,
		SKU:               item.SKU,
		CategoryID:        categoryID,
		CategoryName:      item.CategoryName,
		PrimaryImageURL:   item.PrimaryImageURL,
		QuantityOnHand:    item.QuantityOnHand,
		QuantityReserved:  item.QuantityReserved,
		QuantityAvailable: item.QuantityAvailable,
		LowStockThreshold: item.LowStockThreshold,
		IsLowStock:        item.QuantityAvailable <= item.LowStockThreshold,
		IsOutOfStock:      item.QuantityAvailable == 0,
		UpdatedAt:         item.UpdatedAt.Format(timeFormatRFC3339),
	}
}

func NewStockMovementResponse(item StockMovement) StockMovementResponse {
	var referenceID *string
	if item.ReferenceID != nil {
		value := item.ReferenceID.String()
		referenceID = &value
	}

	var actor *ActorResponse
	if item.CreatedBy != nil && *item.CreatedBy != uuid.Nil {
		actor = &ActorResponse{
			ID:   item.CreatedBy.String(),
			Name: item.CreatedByName,
		}
	}

	return StockMovementResponse{
		ID:            item.ID.String(),
		ProductID:     item.ProductID.String(),
		MovementType:  item.MovementType,
		Quantity:      item.Quantity,
		BalanceAfter:  item.BalanceAfter,
		ReferenceType: item.ReferenceType,
		ReferenceID:   referenceID,
		Reason:        item.Reason,
		Note:          item.Note,
		CreatedBy:     actor,
		CreatedAt:     item.CreatedAt.Format(timeFormatRFC3339),
	}
}

func NewAdjustStockResponse(snapshot StockSnapshot, movement StockMovement) AdjustStockResponse {
	return AdjustStockResponse{
		ProductID:         snapshot.ProductID.String(),
		QuantityOnHand:    snapshot.QuantityOnHand,
		QuantityReserved:  snapshot.QuantityReserved,
		QuantityAvailable: snapshot.QuantityAvailable,
		LowStockThreshold: snapshot.LowStockThreshold,
		MovementID:        movement.ID.String(),
	}
}

func NewUpdateThresholdResponse(snapshot StockSnapshot) UpdateThresholdResponse {
	return UpdateThresholdResponse{
		ProductID:         snapshot.ProductID.String(),
		LowStockThreshold: snapshot.LowStockThreshold,
	}
}

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
