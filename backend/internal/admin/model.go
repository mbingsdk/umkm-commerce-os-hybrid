package admin

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	TenantStatusActive    = "active"
	TenantStatusTrialing  = "trialing"
	TenantStatusSuspended = "suspended"
	TenantStatusCancelled = "cancelled"

	AuditActionTenantStatusUpdated = "admin.tenant.update_status"
	AuditActionTenantPlanUpdated   = "admin.tenant.update_plan"

	EventTenantStatusChanged = "TenantStatusChanged"
	EventTenantPlanChanged   = "TenantPlanChanged"

	AggregateTenant = "tenant"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
	ErrPlanNotFound   = errors.New("plan not found")
	errInvalidCursor  = errors.New("invalid cursor")
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PlatformRole string
	Status       string
}

type Tenant struct {
	ID        uuid.UUID
	PlanID    *uuid.UUID
	Name      string
	Slug      string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Plan struct {
	ID                 uuid.UUID
	Code               string
	Name               string
	PriceMonthly       int64
	ProductLimit       int
	StaffLimit         int
	CanUsePOS          bool
	CanUseDiscovery    bool
	CanUseCourier      bool
	CanUseCustomDomain bool
	IsActive           bool
}

type StoreSummary struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	Status    string
	City      string
	CreatedAt time.Time
}

type OwnerSummary struct {
	ID     uuid.UUID
	Name   string
	Email  string
	Status string
}

type TenantCounts struct {
	StoreCount          int64
	ProductCount        int64
	OrderCount          int64
	UserCount           int64
	POSTransactionCount int64
}

type TenantListItem struct {
	Tenant       Tenant
	Plan         *Plan
	PrimaryStore *StoreSummary
	Owner        *OwnerSummary
	Counts       TenantCounts
}

type TenantDetail struct {
	Tenant       Tenant
	Plan         *Plan
	PrimaryStore *StoreSummary
	Owner        *OwnerSummary
	Counts       TenantCounts
	LatestAudits []AuditSnippet
}

type AuditSnippet struct {
	ID          uuid.UUID
	ActorUserID uuid.UUID
	ActorName   string
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	CreatedAt   time.Time
}

type TenantListFilters struct {
	Status      string
	PlanID      *uuid.UUID
	Query       string
	CreatedFrom *time.Time
	CreatedTo   *time.Time
	Limit       int
	Cursor      *TenantCursor
}

type TenantCursor struct {
	CreatedAt time.Time `json:"created_at"`
	ID        uuid.UUID `json:"id"`
}

type AuditEntry struct {
	ActorUserID uuid.UUID
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	BeforeData  any
	AfterData   any
	IPAddress   string
	UserAgent   string
}

type AuditLog struct {
	ID          uuid.UUID
	ActorUserID uuid.UUID
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	CreatedAt   time.Time
}

func EncodeTenantCursor(item TenantListItem) (string, error) {
	payload, err := json.Marshal(TenantCursor{
		CreatedAt: item.Tenant.CreatedAt,
		ID:        item.Tenant.ID,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodeTenantCursor(raw string) (*TenantCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var cursor TenantCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, err
	}
	if cursor.ID == uuid.Nil || cursor.CreatedAt.IsZero() {
		return nil, errInvalidCursor
	}
	return &cursor, nil
}
