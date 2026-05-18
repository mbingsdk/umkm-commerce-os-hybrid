package category

import "github.com/google/uuid"

type CreateRequest struct {
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
}

type UpdateRequest struct {
	Name        *string    `json:"name"`
	Slug        *string    `json:"slug"`
	Description *string    `json:"description"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   *int       `json:"sort_order"`
	IsActive    *bool      `json:"is_active"`
}
