package product

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/catalog/category"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type productStore interface {
	List(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]Product, error)
	Create(ctx context.Context, q db.Queryer, params CreateParams) (*Product, error)
	FindByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) (*Product, error)
	ListImages(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) ([]Image, error)
	Update(ctx context.Context, q db.Queryer, params UpdateParams) error
	SoftDelete(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) error
}

type categoryReader interface {
	FindByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, categoryID uuid.UUID) (*category.Category, error)
}

type stockWriter interface {
	CreateSnapshot(ctx context.Context, q db.Queryer, params inventory.CreateSnapshotParams) (*inventory.StockSnapshot, error)
	CreateMovement(ctx context.Context, q db.Queryer, params inventory.CreateMovementParams) (*inventory.StockMovement, error)
}

type Service struct {
	db         database
	products   productStore
	categories categoryReader
	stocks     stockWriter
}

type CreateInput struct {
	CategoryID     *uuid.UUID
	Name           string
	Slug           string
	Description    string
	SKU            string
	Barcode        string
	Price          int64
	CompareAtPrice *int64
	CostPrice      *int64
	WeightGram     int
	LengthCM       *float64
	WidthCM        *float64
	HeightCM       *float64
	Status         string
	IsDiscoverable bool
	TrackInventory bool
	AllowBackorder bool
	InitialStock   int
}

type UpdateInput struct {
	CategoryID     *uuid.UUID
	Name           *string
	Slug           *string
	Description    *string
	SKU            *string
	Barcode        *string
	Price          *int64
	CompareAtPrice *int64
	CostPrice      *int64
	WeightGram     *int
	LengthCM       *float64
	WidthCM        *float64
	HeightCM       *float64
	Status         *string
	IsDiscoverable *bool
	TrackInventory *bool
	AllowBackorder *bool
}

func NewService(
	database database,
	products productStore,
	categories categoryReader,
	stocks stockWriter,
) *Service {
	return &Service{
		db:         database,
		products:   products,
		categories: categories,
		stocks:     stocks,
	}
}

func (s *Service) List(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters ListFilters,
) ([]ListItemResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, err
	}
	filters.Query = strings.TrimSpace(filters.Query)

	items, err := s.products.List(ctx, s.db, tenantID, storeID, filters)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	response := make([]ListItemResponse, 0, len(items))
	for idx := range items {
		response = append(response, NewListItemResponse(&items[idx]))
	}

	return response, nil
}

func (s *Service) Create(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	actorUserID uuid.UUID,
	input CreateInput,
) (SummaryResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return SummaryResponse{}, err
	}
	normalized, err := validateCreate(input)
	if err != nil {
		return SummaryResponse{}, err
	}

	var response SummaryResponse
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		if err := s.validateCategory(ctx, tx, tenantID, storeID, normalized.CategoryID); err != nil {
			return err
		}

		created, err := s.products.Create(ctx, tx, CreateParams{
			TenantID:       tenantID,
			StoreID:        storeID,
			CategoryID:     normalized.CategoryID,
			Name:           normalized.Name,
			Slug:           normalized.Slug,
			Description:    normalized.Description,
			SKU:            normalized.SKU,
			Barcode:        normalized.Barcode,
			Price:          normalized.Price,
			CompareAtPrice: normalized.CompareAtPrice,
			CostPrice:      normalized.CostPrice,
			WeightGram:     normalized.WeightGram,
			LengthCM:       normalized.LengthCM,
			WidthCM:        normalized.WidthCM,
			HeightCM:       normalized.HeightCM,
			Status:         normalized.Status,
			IsDiscoverable: normalized.IsDiscoverable,
			TrackInventory: normalized.TrackInventory,
			AllowBackorder: normalized.AllowBackorder,
		})
		if err != nil {
			if errors.Is(err, ErrProductSlugAlreadyInUse) {
				return invalidField("slug", "Product slug is already in use")
			}
			return apperror.Internal(err)
		}

		snapshot, err := s.stocks.CreateSnapshot(ctx, tx, inventory.CreateSnapshotParams{
			TenantID:          tenantID,
			StoreID:           storeID,
			ProductID:         created.ID,
			QuantityOnHand:    normalized.InitialStock,
			QuantityReserved:  0,
			QuantityAvailable: normalized.InitialStock,
			LowStockThreshold: 5,
		})
		if err != nil {
			return apperror.Internal(err)
		}

		if normalized.InitialStock > 0 {
			productID := created.ID
			createdBy := actorUserID
			if _, err := s.stocks.CreateMovement(ctx, tx, inventory.CreateMovementParams{
				TenantID:      tenantID,
				StoreID:       storeID,
				ProductID:     created.ID,
				MovementType:  inventory.MovementTypeInitial,
				Quantity:      normalized.InitialStock,
				BalanceAfter:  normalized.InitialStock,
				ReferenceType: "product",
				ReferenceID:   &productID,
				Note:          "Initial stock",
				CreatedBy:     &createdBy,
			}); err != nil {
				return apperror.Internal(err)
			}
		}

		created.Stock = Stock{
			QuantityOnHand:    snapshot.QuantityOnHand,
			QuantityReserved:  snapshot.QuantityReserved,
			QuantityAvailable: snapshot.QuantityAvailable,
			LowStockThreshold: snapshot.LowStockThreshold,
		}
		response = NewSummaryResponse(created)
		return nil
	})
	if err != nil {
		return SummaryResponse{}, err
	}

	return response, nil
}

func (s *Service) Get(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
	includeCostPrice bool,
) (DetailResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return DetailResponse{}, err
	}
	if productID == uuid.Nil {
		return DetailResponse{}, invalidField("product_id", "Product is required")
	}

	item, err := s.products.FindByID(ctx, s.db, tenantID, storeID, productID)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return DetailResponse{}, apperror.NotFound("Product not found")
		}
		return DetailResponse{}, apperror.Internal(err)
	}

	images, err := s.products.ListImages(ctx, s.db, tenantID, storeID, productID)
	if err != nil {
		return DetailResponse{}, apperror.Internal(err)
	}
	item.Images = images

	return NewDetailResponse(item, includeCostPrice), nil
}

func (s *Service) Update(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
	input UpdateInput,
) (SummaryResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return SummaryResponse{}, err
	}
	if productID == uuid.Nil {
		return SummaryResponse{}, invalidField("product_id", "Product is required")
	}

	current, err := s.products.FindByID(ctx, s.db, tenantID, storeID, productID)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return SummaryResponse{}, apperror.NotFound("Product not found")
		}
		return SummaryResponse{}, apperror.Internal(err)
	}

	merged := CreateInput{
		CategoryID:     current.CategoryID,
		Name:           current.Name,
		Slug:           current.Slug,
		Description:    current.Description,
		SKU:            current.SKU,
		Barcode:        current.Barcode,
		Price:          current.Price,
		CompareAtPrice: current.CompareAtPrice,
		CostPrice:      current.CostPrice,
		WeightGram:     current.WeightGram,
		LengthCM:       current.LengthCM,
		WidthCM:        current.WidthCM,
		HeightCM:       current.HeightCM,
		Status:         current.Status,
		IsDiscoverable: current.IsDiscoverable,
		TrackInventory: current.TrackInventory,
		AllowBackorder: current.AllowBackorder,
	}
	if input.CategoryID != nil {
		merged.CategoryID = input.CategoryID
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
	if input.SKU != nil {
		merged.SKU = *input.SKU
	}
	if input.Barcode != nil {
		merged.Barcode = *input.Barcode
	}
	if input.Price != nil {
		merged.Price = *input.Price
	}
	if input.CompareAtPrice != nil {
		merged.CompareAtPrice = input.CompareAtPrice
	}
	if input.CostPrice != nil {
		merged.CostPrice = input.CostPrice
	}
	if input.WeightGram != nil {
		merged.WeightGram = *input.WeightGram
	}
	if input.LengthCM != nil {
		merged.LengthCM = input.LengthCM
	}
	if input.WidthCM != nil {
		merged.WidthCM = input.WidthCM
	}
	if input.HeightCM != nil {
		merged.HeightCM = input.HeightCM
	}
	if input.Status != nil {
		merged.Status = *input.Status
	}
	if input.IsDiscoverable != nil {
		merged.IsDiscoverable = *input.IsDiscoverable
	}
	if input.TrackInventory != nil {
		merged.TrackInventory = *input.TrackInventory
	}
	if input.AllowBackorder != nil {
		merged.AllowBackorder = *input.AllowBackorder
	}

	normalized, err := validateCreate(merged)
	if err != nil {
		return SummaryResponse{}, err
	}
	if err := s.validateCategory(ctx, s.db, tenantID, storeID, normalized.CategoryID); err != nil {
		return SummaryResponse{}, err
	}

	if err := s.products.Update(ctx, s.db, UpdateParams{
		TenantID:       tenantID,
		StoreID:        storeID,
		ProductID:      productID,
		CategoryID:     normalized.CategoryID,
		Name:           normalized.Name,
		Slug:           normalized.Slug,
		Description:    normalized.Description,
		SKU:            normalized.SKU,
		Barcode:        normalized.Barcode,
		Price:          normalized.Price,
		CompareAtPrice: normalized.CompareAtPrice,
		CostPrice:      normalized.CostPrice,
		WeightGram:     normalized.WeightGram,
		LengthCM:       normalized.LengthCM,
		WidthCM:        normalized.WidthCM,
		HeightCM:       normalized.HeightCM,
		Status:         normalized.Status,
		IsDiscoverable: normalized.IsDiscoverable,
		TrackInventory: normalized.TrackInventory,
		AllowBackorder: normalized.AllowBackorder,
	}); err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			return SummaryResponse{}, apperror.NotFound("Product not found")
		case errors.Is(err, ErrProductSlugAlreadyInUse):
			return SummaryResponse{}, invalidField("slug", "Product slug is already in use")
		default:
			return SummaryResponse{}, apperror.Internal(err)
		}
	}

	updated, err := s.products.FindByID(ctx, s.db, tenantID, storeID, productID)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return SummaryResponse{}, apperror.NotFound("Product not found")
		}
		return SummaryResponse{}, apperror.Internal(err)
	}

	return NewSummaryResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) error {
	if err := validateScope(tenantID, storeID); err != nil {
		return err
	}
	if productID == uuid.Nil {
		return invalidField("product_id", "Product is required")
	}

	if err := s.products.SoftDelete(ctx, s.db, tenantID, storeID, productID); err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return apperror.NotFound("Product not found")
		}
		return apperror.Internal(err)
	}

	return nil
}

func (s *Service) validateCategory(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	categoryID *uuid.UUID,
) error {
	if categoryID == nil {
		return nil
	}
	if _, err := s.categories.FindByID(ctx, q, tenantID, storeID, *categoryID); err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			return invalidField("category_id", "Category must belong to the current tenant and store")
		}
		return apperror.Internal(err)
	}
	return nil
}

func validateCreate(input CreateInput) (CreateInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Slug = strings.TrimSpace(input.Slug)
	input.Description = strings.TrimSpace(input.Description)
	input.SKU = strings.TrimSpace(input.SKU)
	input.Barcode = strings.TrimSpace(input.Barcode)
	input.Status = strings.TrimSpace(input.Status)
	if input.Status == "" {
		input.Status = StatusDraft
	}

	var details []map[string]string
	if input.Name == "" {
		details = append(details, map[string]string{"field": "name", "message": "Product name is required"})
	}
	if !slugPattern.MatchString(input.Slug) {
		details = append(details, map[string]string{"field": "slug", "message": "Product slug is invalid"})
	}
	if input.Price < 0 {
		details = append(details, map[string]string{"field": "price", "message": "Price must be zero or greater"})
	}
	if input.CompareAtPrice != nil && *input.CompareAtPrice < input.Price {
		details = append(details, map[string]string{"field": "compare_at_price", "message": "Compare at price must be equal to or greater than price"})
	}
	if input.CostPrice != nil && *input.CostPrice < 0 {
		details = append(details, map[string]string{"field": "cost_price", "message": "Cost price must be zero or greater"})
	}
	if input.InitialStock < 0 {
		details = append(details, map[string]string{"field": "initial_stock", "message": "Initial stock must be zero or greater"})
	}
	if input.WeightGram < 0 {
		details = append(details, map[string]string{"field": "weight_gram", "message": "Weight must be zero or greater"})
	}
	if input.LengthCM != nil && *input.LengthCM < 0 {
		details = append(details, map[string]string{"field": "length_cm", "message": "Length must be zero or greater"})
	}
	if input.WidthCM != nil && *input.WidthCM < 0 {
		details = append(details, map[string]string{"field": "width_cm", "message": "Width must be zero or greater"})
	}
	if input.HeightCM != nil && *input.HeightCM < 0 {
		details = append(details, map[string]string{"field": "height_cm", "message": "Height must be zero or greater"})
	}
	if !validStatus(input.Status) {
		details = append(details, map[string]string{"field": "status", "message": "Product status is invalid"})
	}

	if len(details) > 0 {
		return CreateInput{}, apperror.Validation("Validation failed", details)
	}

	return input, nil
}

func validStatus(status string) bool {
	switch status {
	case StatusDraft, StatusActive, StatusInactive, StatusArchived:
		return true
	default:
		return false
	}
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
