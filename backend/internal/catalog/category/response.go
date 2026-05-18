package category

import "github.com/google/uuid"

type Response struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
}

func NewResponse(category *Category) Response {
	return Response{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ParentID:    category.ParentID,
		SortOrder:   category.SortOrder,
		IsActive:    category.IsActive,
	}
}
