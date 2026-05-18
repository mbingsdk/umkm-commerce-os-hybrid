package product

import (
	"time"

	"github.com/google/uuid"
)

type PublicCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

type PublicListFilters struct {
	Query        string
	CategorySlug string
	InStock      *bool
	Limit        int
	Cursor       *PublicCursor
}

type PublicListItem struct {
	ID              uuid.UUID
	TenantID        uuid.UUID
	StoreID         uuid.UUID
	Name            string
	Slug            string
	Price           int64
	CompareAtPrice  *int64
	Status          string
	PrimaryImageURL string
	Stock           Stock
	CreatedAt       time.Time
}

type PublicCategorySummary struct {
	ID   uuid.UUID
	Name string
	Slug string
}

type PublicProduct struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	Name           string
	Slug           string
	Description    string
	Price          int64
	CompareAtPrice *int64
	WeightGram     int
	Status         string
	Category       *PublicCategorySummary
	Stock          Stock
	Images         []Image
}
