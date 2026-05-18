package category

import "github.com/google/uuid"

type Category struct {
	ID          uuid.UUID
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Description string
	SortOrder   int
	IsActive    bool
}

type CreateParams struct {
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Description string
	SortOrder   int
	IsActive    bool
}

type UpdateParams struct {
	TenantID    uuid.UUID
	StoreID     uuid.UUID
	CategoryID  uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Description string
	SortOrder   int
	IsActive    bool
}

type ListFilters struct {
	IsActive *bool
}
