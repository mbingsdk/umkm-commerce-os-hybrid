package product

import "github.com/google/uuid"

const (
	StatusDraft    = "draft"
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusArchived = "archived"
)

type Product struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	StoreID         uuid.UUID
	CategoryID      *uuid.UUID
	Name            string
	Slug            string
	Description     string
	SKU             string
	Barcode         string
	Price           int64
	CompareAtPrice  *int64
	CostPrice       *int64
	WeightGram      int
	LengthCM        *float64
	WidthCM         *float64
	HeightCM        *float64
	Status          string
	IsDiscoverable  bool
	TrackInventory  bool
	AllowBackorder  bool
	PrimaryImageURL string
	Stock           Stock
	Images          []Image
}

type Stock struct {
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
	LowStockThreshold int
}

type Image struct {
	ID        uuid.UUID
	URL       string
	AltText   string
	IsPrimary bool
	SortOrder int
}

type CreateParams struct {
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	CategoryID     *uuid.UUID
	Name           string
	Slug           string
	Description    string
	SKU            string
	Barcode        string
	Price          int64
	CompareAtPrice *int64
	CostPrice      *int64
	WeightGram     int
	LengthCM       *float64
	WidthCM        *float64
	HeightCM       *float64
	Status         string
	IsDiscoverable bool
	TrackInventory bool
	AllowBackorder bool
}

type UpdateParams struct {
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	ProductID      uuid.UUID
	CategoryID     *uuid.UUID
	Name           string
	Slug           string
	Description    string
	SKU            string
	Barcode        string
	Price          int64
	CompareAtPrice *int64
	CostPrice      *int64
	WeightGram     int
	LengthCM       *float64
	WidthCM        *float64
	HeightCM       *float64
	Status         string
	IsDiscoverable bool
	TrackInventory bool
	AllowBackorder bool
}

type ListFilters struct {
	Query      string
	Status     *string
	CategoryID *uuid.UUID
}
