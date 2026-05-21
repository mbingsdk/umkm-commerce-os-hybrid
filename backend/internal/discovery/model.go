package discovery

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	tenantStatusActive   = "active"
	tenantStatusTrialing = "trialing"
	storeStatusPublished = "published"
	productStatusActive  = "active"

	SearchTypeAll      = "all"
	SearchTypeStores   = "stores"
	SearchTypeProducts = "products"
)

var errInvalidCursor = errors.New("invalid cursor")

type Store struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	LogoURL     string
	BannerURL   string
	City        string
	Province    string
	CreatedAt   time.Time
}

type Product struct {
	ID              uuid.UUID
	Name            string
	Slug            string
	Description     string
	Price           int64
	PrimaryImageURL string
	CategoryName    string
	CategorySlug    string
	StoreID         uuid.UUID
	StoreName       string
	StoreSlug       string
	StoreCity       string
	StoreProvince   string
	CreatedAt       time.Time
}

type CategoryAggregate struct {
	Name  string
	Slug  string
	Count int
}

type CityAggregate struct {
	City  string
	Count int
}

type ListStoresFilters struct {
	Query    string
	City     string
	Category string
	Limit    int
	Cursor   *Cursor
}

type ListProductsFilters struct {
	Query    string
	City     string
	Category string
	PriceMin *int64
	PriceMax *int64
	Limit    int
	Cursor   *Cursor
}

type SearchFilters struct {
	Query    string
	Type     string
	City     string
	Category string
	Limit    int
}

type Cursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

func EncodeStoreCursor(item Store) (string, error) {
	return encodeCursor(Cursor{CreatedAt: item.CreatedAt, ID: item.ID})
}

func EncodeProductCursor(item Product) (string, error) {
	return encodeCursor(Cursor{CreatedAt: item.CreatedAt, ID: item.ID})
}

func DecodeCursor(raw string) (*Cursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var cursor Cursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, err
	}
	if cursor.ID == uuid.Nil || cursor.CreatedAt.IsZero() {
		return nil, errInvalidCursor
	}
	return &cursor, nil
}

func encodeCursor(cursor Cursor) (string, error) {
	payload, err := json.Marshal(cursor)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}
