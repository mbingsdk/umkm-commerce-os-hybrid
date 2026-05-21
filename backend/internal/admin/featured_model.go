package admin

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	FeaturedItemTypeStore   = "store"
	FeaturedItemTypeProduct = "product"

	FeaturedPlacementHome     = "home"
	FeaturedPlacementStores   = "stores"
	FeaturedPlacementProducts = "products"
	FeaturedPlacementCategory = "category"
	FeaturedPlacementCity     = "city"

	AuditActionFeaturedCreated = "admin.discovery.featured.create"
	AuditActionFeaturedUpdated = "admin.discovery.featured.update"
	AuditActionFeaturedDeleted = "admin.discovery.featured.delete"

	AggregateFeaturedDiscovery = "discovery_featured_item"
)

type FeaturedItem struct {
	ID          uuid.UUID
	ItemType    string
	TenantID    uuid.UUID
	StoreID     *uuid.UUID
	ProductID   *uuid.UUID
	Placement   string
	SortOrder   int
	StartsAt    *time.Time
	EndsAt      *time.Time
	IsActive    bool
	CreatedBy   *uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	StoreName   string
	StoreSlug   string
	ProductName string
	ProductSlug string
}

type FeaturedStoreTarget struct {
	ID             uuid.UUID
	TenantID       uuid.UUID
	Name           string
	Slug           string
	TenantStatus   string
	Status         string
	IsDiscoverable bool
}

type FeaturedProductTarget struct {
	ID                uuid.UUID
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	Name              string
	Slug              string
	TenantStatus      string
	Status            string
	IsDiscoverable    bool
	StoreStatus       string
	StoreDiscoverable bool
}

type CreateFeaturedParams struct {
	ItemType  string
	ItemID    uuid.UUID
	TenantID  uuid.UUID
	StoreID   *uuid.UUID
	ProductID *uuid.UUID
	Placement string
	SortOrder int
	StartsAt  *time.Time
	EndsAt    *time.Time
	IsActive  bool
	CreatedBy uuid.UUID
}

type UpdateFeaturedParams struct {
	ID        uuid.UUID
	ItemType  string
	ItemID    uuid.UUID
	TenantID  uuid.UUID
	StoreID   *uuid.UUID
	ProductID *uuid.UUID
	Placement string
	SortOrder int
	StartsAt  *time.Time
	EndsAt    *time.Time
	IsActive  bool
}

type FeaturedListFilters struct {
	ItemType  string
	Placement string
	TenantID  *uuid.UUID
	IsActive  *bool
	Limit     int
	Cursor    *AdminListCursor
}

type AuditLogListFilters struct {
	ActorUserID *uuid.UUID
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	DateFrom    *time.Time
	DateTo      *time.Time
	Limit       int
	Cursor      *AdminListCursor
}

type AdminListCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

type AdminAuditLogItem struct {
	ID          uuid.UUID
	ActorUserID uuid.UUID
	ActorName   string
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	BeforeData  json.RawMessage
	AfterData   json.RawMessage
	IPAddress   string
	UserAgent   string
	CreatedAt   time.Time
}

func EncodeAdminListCursor(createdAt time.Time, id uuid.UUID) (string, error) {
	payload, err := json.Marshal(AdminListCursor{
		CreatedAt: createdAt,
		ID:        id,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeAdminListCursor(raw string) (*AdminListCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var cursor AdminListCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, err
	}
	if cursor.ID == uuid.Nil || cursor.CreatedAt.IsZero() {
		return nil, errInvalidCursor
	}
	return &cursor, nil
}
