package tenant

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive   = "active"
	StatusTrialing = "trialing"

	MembershipRoleOwner    = "owner"
	MembershipStatusActive = "active"
)

type Tenant struct {
	ID     uuid.UUID
	Name   string
	Slug   string
	Status string
}

type Membership struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	TenantID uuid.UUID
	Role     string
	Status   string
	JoinedAt time.Time
}

type StoreSummary struct {
	ID     uuid.UUID
	Name   string
	Slug   string
	Status string
}

type AccessRecord struct {
	TenantID     uuid.UUID
	TenantStatus string
	StoreID      uuid.UUID
	Role         string
}

type TenantListItem struct {
	ID     uuid.UUID
	Name   string
	Slug   string
	Role   string
	Status string
	Store  StoreSummary
}

type CreateTenantParams struct {
	PlanID uuid.UUID
	Name   string
	Slug   string
}

type CreateMembershipParams struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
	Role     string
	Status   string
	JoinedAt time.Time
}
