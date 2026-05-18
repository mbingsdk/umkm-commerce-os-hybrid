package product

import "github.com/google/uuid"

type CreateRequest struct {
	CategoryID     *uuid.UUID `json:"category_id"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug"`
	Description    string     `json:"description"`
	SKU            string     `json:"sku"`
	Barcode        string     `json:"barcode"`
	Price          int64      `json:"price"`
	CompareAtPrice *int64     `json:"compare_at_price"`
	CostPrice      *int64     `json:"cost_price"`
	WeightGram     int        `json:"weight_gram"`
	LengthCM       *float64   `json:"length_cm"`
	WidthCM        *float64   `json:"width_cm"`
	HeightCM       *float64   `json:"height_cm"`
	Status         string     `json:"status"`
	IsDiscoverable bool       `json:"is_discoverable"`
	TrackInventory bool       `json:"track_inventory"`
	AllowBackorder bool       `json:"allow_backorder"`
	InitialStock   int        `json:"initial_stock"`
}

type UpdateRequest struct {
	CategoryID     *uuid.UUID `json:"category_id"`
	Name           *string    `json:"name"`
	Slug           *string    `json:"slug"`
	Description    *string    `json:"description"`
	SKU            *string    `json:"sku"`
	Barcode        *string    `json:"barcode"`
	Price          *int64     `json:"price"`
	CompareAtPrice *int64     `json:"compare_at_price"`
	CostPrice      *int64     `json:"cost_price"`
	WeightGram     *int       `json:"weight_gram"`
	LengthCM       *float64   `json:"length_cm"`
	WidthCM        *float64   `json:"width_cm"`
	HeightCM       *float64   `json:"height_cm"`
	Status         *string    `json:"status"`
	IsDiscoverable *bool      `json:"is_discoverable"`
	TrackInventory *bool      `json:"track_inventory"`
	AllowBackorder *bool      `json:"allow_backorder"`
}
