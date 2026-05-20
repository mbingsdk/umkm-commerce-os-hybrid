package courier

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	AuditActionCourierZoneCreated = "courier.zone_created"
	AuditActionCourierZoneUpdated = "courier.zone_updated"
	AuditActionCourierZoneDeleted = "courier.zone_deleted"

	AggregateCourierZone = "courier_zone"
)

var ErrZoneNotFound = errors.New("courier zone not found")

type Zone struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	Name        string
	Description string
	Rate        int64
	IsActive    bool
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

type ListZoneFilters struct {
	IsActive *bool
}

type CreateZoneParams struct {
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	Name        string
	Description string
	Rate        int64
	IsActive    bool
	SortOrder   int
}

type UpdateZoneParams struct {
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	ZoneID      uuid.UUID
	Name        string
	Description string
	Rate        int64
	IsActive    bool
	SortOrder   int
}
