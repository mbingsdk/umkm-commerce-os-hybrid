package product

import "github.com/google/uuid"

type StockResponse struct {
	QuantityOnHand    int `json:"quantity_on_hand"`
	QuantityReserved  int `json:"quantity_reserved"`
	QuantityAvailable int `json:"quantity_available"`
	LowStockThreshold int `json:"low_stock_threshold,omitempty"`
}

type ImageResponse struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	AltText   string    `json:"alt_text,omitempty"`
	IsPrimary bool      `json:"is_primary"`
	SortOrder int       `json:"sort_order"`
}

type ListItemResponse struct {
	ID              uuid.UUID     `json:"id"`
	Name            string        `json:"name"`
	Slug            string        `json:"slug"`
	SKU             string        `json:"sku,omitempty"`
	Price           int64         `json:"price"`
	CompareAtPrice  *int64        `json:"compare_at_price,omitempty"`
	Status          string        `json:"status"`
	IsDiscoverable  bool          `json:"is_discoverable"`
	PrimaryImageURL string        `json:"primary_image_url,omitempty"`
	Stock           StockResponse `json:"stock"`
}

type SummaryResponse struct {
	ID     uuid.UUID     `json:"id"`
	Name   string        `json:"name"`
	Slug   string        `json:"slug"`
	Status string        `json:"status"`
	Stock  StockResponse `json:"stock"`
}

type DetailResponse struct {
	ID             uuid.UUID       `json:"id"`
	CategoryID     *uuid.UUID      `json:"category_id"`
	Name           string          `json:"name"`
	Slug           string          `json:"slug"`
	Description    string          `json:"description,omitempty"`
	SKU            string          `json:"sku,omitempty"`
	Barcode        string          `json:"barcode,omitempty"`
	Price          int64           `json:"price"`
	CompareAtPrice *int64          `json:"compare_at_price,omitempty"`
	CostPrice      *int64          `json:"cost_price,omitempty"`
	WeightGram     int             `json:"weight_gram"`
	LengthCM       *float64        `json:"length_cm,omitempty"`
	WidthCM        *float64        `json:"width_cm,omitempty"`
	HeightCM       *float64        `json:"height_cm,omitempty"`
	Status         string          `json:"status"`
	IsDiscoverable bool            `json:"is_discoverable"`
	TrackInventory bool            `json:"track_inventory"`
	AllowBackorder bool            `json:"allow_backorder"`
	Images         []ImageResponse `json:"images"`
	Stock          StockResponse   `json:"stock"`
}

func NewListItemResponse(product *Product) ListItemResponse {
	return ListItemResponse{
		ID:              product.ID,
		Name:            product.Name,
		Slug:            product.Slug,
		SKU:             product.SKU,
		Price:           product.Price,
		CompareAtPrice:  product.CompareAtPrice,
		Status:          product.Status,
		IsDiscoverable:  product.IsDiscoverable,
		PrimaryImageURL: product.PrimaryImageURL,
		Stock:           newStockResponse(product.Stock),
	}
}

func NewSummaryResponse(product *Product) SummaryResponse {
	return SummaryResponse{
		ID:     product.ID,
		Name:   product.Name,
		Slug:   product.Slug,
		Status: product.Status,
		Stock:  newStockResponse(product.Stock),
	}
}

func NewDetailResponse(product *Product, includeCostPrice bool) DetailResponse {
	images := make([]ImageResponse, 0, len(product.Images))
	for _, image := range product.Images {
		images = append(images, ImageResponse{
			ID:        image.ID,
			URL:       image.URL,
			AltText:   image.AltText,
			IsPrimary: image.IsPrimary,
			SortOrder: image.SortOrder,
		})
	}

	var costPrice *int64
	if includeCostPrice {
		costPrice = product.CostPrice
	}

	return DetailResponse{
		ID:             product.ID,
		CategoryID:     product.CategoryID,
		Name:           product.Name,
		Slug:           product.Slug,
		Description:    product.Description,
		SKU:            product.SKU,
		Barcode:        product.Barcode,
		Price:          product.Price,
		CompareAtPrice: product.CompareAtPrice,
		CostPrice:      costPrice,
		WeightGram:     product.WeightGram,
		LengthCM:       product.LengthCM,
		WidthCM:        product.WidthCM,
		HeightCM:       product.HeightCM,
		Status:         product.Status,
		IsDiscoverable: product.IsDiscoverable,
		TrackInventory: product.TrackInventory,
		AllowBackorder: product.AllowBackorder,
		Images:         images,
		Stock:          newStockResponse(product.Stock),
	}
}

func newStockResponse(stock Stock) StockResponse {
	return StockResponse{
		QuantityOnHand:    stock.QuantityOnHand,
		QuantityReserved:  stock.QuantityReserved,
		QuantityAvailable: stock.QuantityAvailable,
		LowStockThreshold: stock.LowStockThreshold,
	}
}

func CanReadCostPrice(role string) bool {
	return role == "owner" || role == "manager"
}
