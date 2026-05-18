package category

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type categoryStore interface {
	List(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]Category, error)
	Create(ctx context.Context, q db.Queryer, params CreateParams) (*Category, error)
	FindByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, categoryID uuid.UUID) (*Category, error)
	Update(ctx context.Context, q db.Queryer, params UpdateParams) (*Category, error)
	SoftDelete(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, categoryID uuid.UUID) error
}

type Service struct {
	db         db.Queryer
	categories categoryStore
}

type CreateInput struct {
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Description string
	SortOrder   int
	IsActive    bool
}

type UpdateInput struct {
	ParentID    *uuid.UUID
	Name        *string
	Slug        *string
	Description *string
	SortOrder   *int
	IsActive    *bool
}

func NewService(database db.Queryer, categories categoryStore) *Service {
	return &Service{
		db:         database,
		categories: categories,
	}
}

func (s *Service) List(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters ListFilters,
) ([]Response, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, err
	}

	items, err := s.categories.List(ctx, s.db, tenantID, storeID, filters)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	response := make([]Response, 0, len(items))
	for idx := range items {
		response = append(response, NewResponse(&items[idx]))
	}

	return response, nil
}

func (s *Service) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	input CreateInput,
) (Response, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return Response{}, err
	}

	normalized, err := validateCreate(input)
	if err != nil {
		return Response{}, err
	}

	if err := s.validateParent(ctx, tenantID, storeID, normalized.ParentID, uuid.Nil); err != nil {
		return Response{}, err
	}

	created, err := s.categories.Create(ctx, s.db, CreateParams{
		TenantID:    tenantID,
		StoreID:     storeID,
		ParentID:    normalized.ParentID,
		Name:        normalized.Name,
		Slug:        normalized.Slug,
		Description: normalized.Description,
		SortOrder:   normalized.SortOrder,
		IsActive:    normalized.IsActive,
	})
	if err != nil {
		if errors.Is(err, ErrCategorySlugAlreadyInUse) {
			return Response{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "slug", "message": "Category slug is already in use"},
			})
		}
		return Response{}, apperror.Internal(err)
	}

	return NewResponse(created), nil
}

func (s *Service) Update(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	categoryID uuid.UUID,
	input UpdateInput,
) (Response, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return Response{}, err
	}
	if categoryID == uuid.Nil {
		return Response{}, invalidField("category_id", "Category is required")
	}

	current, err := s.categories.FindByID(ctx, s.db, tenantID, storeID, categoryID)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return Response{}, apperror.NotFound("Category not found")
		}
		return Response{}, apperror.Internal(err)
	}

	merged := CreateInput{
		ParentID:    current.ParentID,
		Name:        current.Name,
		Slug:        current.Slug,
		Description: current.Description,
		SortOrder:   current.SortOrder,
		IsActive:    current.IsActive,
	}
	if input.ParentID != nil {
		merged.ParentID = input.ParentID
	}
	if input.Name != nil {
		merged.Name = *input.Name
	}
	if input.Slug != nil {
		merged.Slug = *input.Slug
	}
	if input.Description != nil {
		merged.Description = *input.Description
	}
	if input.SortOrder != nil {
		merged.SortOrder = *input.SortOrder
	}
	if input.IsActive != nil {
		merged.IsActive = *input.IsActive
	}

	normalized, err := validateCreate(merged)
	if err != nil {
		return Response{}, err
	}
	if err := s.validateParent(ctx, tenantID, storeID, normalized.ParentID, categoryID); err != nil {
		return Response{}, err
	}

	updated, err := s.categories.Update(ctx, s.db, UpdateParams{
		TenantID:    tenantID,
		StoreID:     storeID,
		CategoryID:  categoryID,
		ParentID:    normalized.ParentID,
		Name:        normalized.Name,
		Slug:        normalized.Slug,
		Description: normalized.Description,
		SortOrder:   normalized.SortOrder,
		IsActive:    normalized.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrCategoryNotFound):
			return Response{}, apperror.NotFound("Category not found")
		case errors.Is(err, ErrCategorySlugAlreadyInUse):
			return Response{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "slug", "message": "Category slug is already in use"},
			})
		default:
			return Response{}, apperror.Internal(err)
		}
	}

	return NewResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, categoryID uuid.UUID) error {
	if err := validateScope(tenantID, storeID); err != nil {
		return err
	}
	if categoryID == uuid.Nil {
		return invalidField("category_id", "Category is required")
	}

	if err := s.categories.SoftDelete(ctx, s.db, tenantID, storeID, categoryID); err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return apperror.NotFound("Category not found")
		}
		return apperror.Internal(err)
	}

	return nil
}

func (s *Service) validateParent(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	parentID *uuid.UUID,
	categoryID uuid.UUID,
) error {
	if parentID == nil {
		return nil
	}
	if *parentID == categoryID {
		return invalidField("parent_id", "Category cannot be its own parent")
	}
	if _, err := s.categories.FindByID(ctx, s.db, tenantID, storeID, *parentID); err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return invalidField("parent_id", "Parent category must belong to the current tenant and store")
		}
		return apperror.Internal(err)
	}
	return nil
}

func validateCreate(input CreateInput) (CreateInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)

	var details []map[string]string
	if input.Name == "" {
		details = append(details, map[string]string{"field": "name", "message": "Category name is required"})
	}
	if !slugPattern.MatchString(input.Slug) {
		details = append(details, map[string]string{"field": "slug", "message": "Category slug is invalid"})
	}
	if input.SortOrder < 0 {
		details = append(details, map[string]string{"field": "sort_order", "message": "Sort order must be zero or greater"})
	}

	if len(details) > 0 {
		return CreateInput{}, apperror.Validation("Validation failed", details)
	}

	return input, nil
}

func validateScope(tenantID uuid.UUID, storeID uuid.UUID) error {
	var details []map[string]string
	if tenantID == uuid.Nil {
		details = append(details, map[string]string{"field": "tenant_id", "message": "Tenant is required"})
	}
	if storeID == uuid.Nil {
		details = append(details, map[string]string{"field": "store_id", "message": "Store is required"})
	}
	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}
	return nil
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{
		{"field": field, "message": message},
	})
}
