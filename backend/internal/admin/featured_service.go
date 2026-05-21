package admin

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type featuredAdminStore interface {
	ListFeaturedItems(ctx context.Context, q db.Queryer, filters FeaturedListFilters) ([]FeaturedItem, error)
	FindFeaturedItemByIDForUpdate(ctx context.Context, q db.Queryer, featuredID uuid.UUID) (*FeaturedItem, error)
	CreateFeaturedItem(ctx context.Context, q db.Queryer, params CreateFeaturedParams) (*FeaturedItem, error)
	UpdateFeaturedItem(ctx context.Context, q db.Queryer, params UpdateFeaturedParams) (*FeaturedItem, error)
	DeleteFeaturedItem(ctx context.Context, q db.Queryer, featuredID uuid.UUID) (*FeaturedItem, error)
	FindFeaturedStoreTarget(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) (*FeaturedStoreTarget, error)
	FindFeaturedProductTarget(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID *uuid.UUID, productID uuid.UUID) (*FeaturedProductTarget, error)
	ListAdminAuditLogs(ctx context.Context, q db.Queryer, filters AuditLogListFilters) ([]AdminAuditLogItem, error)
}

type CreateFeaturedInput struct {
	ActorUserID uuid.UUID
	ItemType    string
	TenantID    uuid.UUID
	StoreID     *uuid.UUID
	ProductID   *uuid.UUID
	Placement   string
	SortOrder   int
	StartsAt    *time.Time
	EndsAt      *time.Time
	IsActive    *bool
	IPAddress   string
	UserAgent   string
}

type UpdateFeaturedInput struct {
	ActorUserID  uuid.UUID
	FeaturedID   uuid.UUID
	ItemType     *string
	TenantID     *uuid.UUID
	StoreID      *uuid.UUID
	StoreIDSet   bool
	ProductID    *uuid.UUID
	ProductIDSet bool
	Placement    *string
	SortOrder    *int
	StartsAt     *time.Time
	EndsAt       *time.Time
	IsActive     *bool
	IPAddress    string
	UserAgent    string
}

type DeleteFeaturedInput struct {
	ActorUserID uuid.UUID
	FeaturedID  uuid.UUID
	IPAddress   string
	UserAgent   string
}

type featuredDraft struct {
	ItemType  string
	TenantID  uuid.UUID
	StoreID   *uuid.UUID
	ProductID *uuid.UUID
	Placement string
	SortOrder int
	StartsAt  *time.Time
	EndsAt    *time.Time
	IsActive  bool
}

func (s *Service) ListFeaturedItems(ctx context.Context, filters FeaturedListFilters) ([]AdminFeaturedResponse, PaginationMeta, error) {
	repo, err := s.featuredRepo()
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	normalized, err := normalizeFeaturedListFilters(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1
	items, err := repo.ListFeaturedItems(ctx, s.db, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeAdminListCursor(items[len(items)-1].CreatedAt, items[len(items)-1].ID)
		if err != nil {
			return nil, PaginationMeta{}, apperror.Internal(err)
		}
		nextCursor = &encoded
	}

	return NewFeaturedResponses(items), PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) CreateFeaturedItem(ctx context.Context, input CreateFeaturedInput) (AdminFeaturedResponse, error) {
	repo, err := s.featuredRepo()
	if err != nil {
		return AdminFeaturedResponse{}, err
	}
	normalized, err := normalizeCreateFeaturedInput(input)
	if err != nil {
		return AdminFeaturedResponse{}, err
	}

	var created *FeaturedItem
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		params, target, err := s.resolveFeaturedTarget(ctx, tx, repo, featuredDraft{
			ItemType:  normalized.ItemType,
			TenantID:  normalized.TenantID,
			StoreID:   normalized.StoreID,
			ProductID: normalized.ProductID,
			Placement: normalized.Placement,
			SortOrder: normalized.SortOrder,
			StartsAt:  normalized.StartsAt,
			EndsAt:    normalized.EndsAt,
			IsActive:  boolValue(normalized.IsActive, true),
		})
		if err != nil {
			return err
		}
		params.CreatedBy = normalized.ActorUserID

		item, err := repo.CreateFeaturedItem(ctx, tx, params)
		if err != nil {
			return apperror.Internal(err)
		}
		enrichFeaturedTarget(item, target)

		if _, err := s.repo.CreateAuditLog(ctx, tx, AuditEntry{
			ActorUserID: normalized.ActorUserID,
			Action:      AuditActionFeaturedCreated,
			TargetType:  AggregateFeaturedDiscovery,
			TargetID:    &item.ID,
			AfterData:   featuredAuditSnapshot(item),
			IPAddress:   normalized.IPAddress,
			UserAgent:   normalized.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}
		created = item
		return nil
	})
	if err != nil {
		return AdminFeaturedResponse{}, err
	}
	return NewFeaturedResponse(*created), nil
}

func (s *Service) UpdateFeaturedItem(ctx context.Context, input UpdateFeaturedInput) (AdminFeaturedResponse, error) {
	repo, err := s.featuredRepo()
	if err != nil {
		return AdminFeaturedResponse{}, err
	}
	normalized, err := normalizeUpdateFeaturedInput(input)
	if err != nil {
		return AdminFeaturedResponse{}, err
	}

	var updated *FeaturedItem
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := repo.FindFeaturedItemByIDForUpdate(ctx, tx, normalized.FeaturedID)
		if err != nil {
			if errors.Is(err, ErrFeaturedItemNotFound) {
				return apperror.NotFound("Featured item not found")
			}
			return apperror.Internal(err)
		}

		draft := featuredDraft{
			ItemType:  current.ItemType,
			TenantID:  current.TenantID,
			StoreID:   current.StoreID,
			ProductID: current.ProductID,
			Placement: current.Placement,
			SortOrder: current.SortOrder,
			StartsAt:  current.StartsAt,
			EndsAt:    current.EndsAt,
			IsActive:  current.IsActive,
		}
		if normalized.ItemType != nil {
			draft.ItemType = *normalized.ItemType
		}
		if normalized.TenantID != nil {
			draft.TenantID = *normalized.TenantID
		}
		if normalized.StoreIDSet {
			draft.StoreID = normalized.StoreID
		}
		if normalized.ProductIDSet {
			draft.ProductID = normalized.ProductID
		}
		if normalized.Placement != nil {
			draft.Placement = *normalized.Placement
		}
		if normalized.SortOrder != nil {
			draft.SortOrder = *normalized.SortOrder
		}
		if normalized.StartsAt != nil {
			draft.StartsAt = normalized.StartsAt
		}
		if normalized.EndsAt != nil {
			draft.EndsAt = normalized.EndsAt
		}
		if normalized.IsActive != nil {
			draft.IsActive = *normalized.IsActive
		}

		params, target, err := s.resolveFeaturedTarget(ctx, tx, repo, draft)
		if err != nil {
			return err
		}

		item, err := repo.UpdateFeaturedItem(ctx, tx, UpdateFeaturedParams{
			ID:        normalized.FeaturedID,
			ItemType:  params.ItemType,
			ItemID:    params.ItemID,
			TenantID:  params.TenantID,
			StoreID:   params.StoreID,
			ProductID: params.ProductID,
			Placement: params.Placement,
			SortOrder: params.SortOrder,
			StartsAt:  params.StartsAt,
			EndsAt:    params.EndsAt,
			IsActive:  params.IsActive,
		})
		if err != nil {
			if errors.Is(err, ErrFeaturedItemNotFound) {
				return apperror.NotFound("Featured item not found")
			}
			return apperror.Internal(err)
		}
		enrichFeaturedTarget(item, target)

		if _, err := s.repo.CreateAuditLog(ctx, tx, AuditEntry{
			ActorUserID: normalized.ActorUserID,
			Action:      AuditActionFeaturedUpdated,
			TargetType:  AggregateFeaturedDiscovery,
			TargetID:    &normalized.FeaturedID,
			BeforeData:  featuredAuditSnapshot(current),
			AfterData:   featuredAuditSnapshot(item),
			IPAddress:   normalized.IPAddress,
			UserAgent:   normalized.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}
		updated = item
		return nil
	})
	if err != nil {
		return AdminFeaturedResponse{}, err
	}
	return NewFeaturedResponse(*updated), nil
}

func (s *Service) DeleteFeaturedItem(ctx context.Context, input DeleteFeaturedInput) (AdminFeaturedResponse, error) {
	repo, err := s.featuredRepo()
	if err != nil {
		return AdminFeaturedResponse{}, err
	}
	if input.ActorUserID == uuid.Nil {
		return AdminFeaturedResponse{}, invalidField("actor_user_id", "Actor is required")
	}
	if input.FeaturedID == uuid.Nil {
		return AdminFeaturedResponse{}, invalidField("featuredId", "featuredId must be a valid UUID")
	}

	var deleted *FeaturedItem
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := repo.FindFeaturedItemByIDForUpdate(ctx, tx, input.FeaturedID)
		if err != nil {
			if errors.Is(err, ErrFeaturedItemNotFound) {
				return apperror.NotFound("Featured item not found")
			}
			return apperror.Internal(err)
		}

		item, err := repo.DeleteFeaturedItem(ctx, tx, input.FeaturedID)
		if err != nil {
			if errors.Is(err, ErrFeaturedItemNotFound) {
				return apperror.NotFound("Featured item not found")
			}
			return apperror.Internal(err)
		}

		if _, err := s.repo.CreateAuditLog(ctx, tx, AuditEntry{
			ActorUserID: input.ActorUserID,
			Action:      AuditActionFeaturedDeleted,
			TargetType:  AggregateFeaturedDiscovery,
			TargetID:    &input.FeaturedID,
			BeforeData:  featuredAuditSnapshot(current),
			AfterData:   map[string]any{"deleted": true, "id": input.FeaturedID.String()},
			IPAddress:   strings.TrimSpace(input.IPAddress),
			UserAgent:   strings.TrimSpace(input.UserAgent),
		}); err != nil {
			return apperror.Internal(err)
		}
		deleted = item
		return nil
	})
	if err != nil {
		return AdminFeaturedResponse{}, err
	}
	return NewFeaturedResponse(*deleted), nil
}

func (s *Service) ListAuditLogs(ctx context.Context, filters AuditLogListFilters) ([]AdminAuditLogResponse, PaginationMeta, error) {
	repo, err := s.featuredRepo()
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	normalized, err := normalizeAuditLogFilters(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1
	items, err := repo.ListAdminAuditLogs(ctx, s.db, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeAdminListCursor(items[len(items)-1].CreatedAt, items[len(items)-1].ID)
		if err != nil {
			return nil, PaginationMeta{}, apperror.Internal(err)
		}
		nextCursor = &encoded
	}

	return NewAuditLogResponses(items), PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) featuredRepo() (featuredAdminStore, error) {
	repo, ok := s.repo.(featuredAdminStore)
	if !ok {
		return nil, apperror.Internal(errors.New("admin repository does not support featured discovery"))
	}
	return repo, nil
}

func (s *Service) resolveFeaturedTarget(ctx context.Context, q db.Queryer, repo featuredAdminStore, draft featuredDraft) (CreateFeaturedParams, any, error) {
	if err := validateFeaturedDraft(draft); err != nil {
		return CreateFeaturedParams{}, nil, err
	}

	params := CreateFeaturedParams{
		ItemType:  draft.ItemType,
		TenantID:  draft.TenantID,
		Placement: draft.Placement,
		SortOrder: draft.SortOrder,
		StartsAt:  draft.StartsAt,
		EndsAt:    draft.EndsAt,
		IsActive:  draft.IsActive,
	}

	switch draft.ItemType {
	case FeaturedItemTypeStore:
		if draft.StoreID == nil || *draft.StoreID == uuid.Nil {
			return CreateFeaturedParams{}, nil, invalidField("store_id", "store_id is required for featured store")
		}
		target, err := repo.FindFeaturedStoreTarget(ctx, q, draft.TenantID, *draft.StoreID)
		if err != nil {
			if errors.Is(err, ErrFeaturedStoreNotFound) {
				return CreateFeaturedParams{}, nil, apperror.NotFound("Store not found")
			}
			return CreateFeaturedParams{}, nil, apperror.Internal(err)
		}
		if !tenantPublicEligible(target.TenantStatus) {
			return CreateFeaturedParams{}, nil, invalidField("tenant_id", "tenant must be active or trialing")
		}
		if target.Status != "published" {
			return CreateFeaturedParams{}, nil, invalidField("store_id", "store must be published")
		}
		if !target.IsDiscoverable {
			return CreateFeaturedParams{}, nil, invalidField("store_id", "store must be discoverable")
		}
		params.ItemID = target.ID
		params.StoreID = &target.ID
		params.ProductID = nil
		return params, target, nil

	case FeaturedItemTypeProduct:
		if draft.ProductID == nil || *draft.ProductID == uuid.Nil {
			return CreateFeaturedParams{}, nil, invalidField("product_id", "product_id is required for featured product")
		}
		target, err := repo.FindFeaturedProductTarget(ctx, q, draft.TenantID, draft.StoreID, *draft.ProductID)
		if err != nil {
			if errors.Is(err, ErrFeaturedProductNotFound) {
				return CreateFeaturedParams{}, nil, apperror.NotFound("Product not found")
			}
			return CreateFeaturedParams{}, nil, apperror.Internal(err)
		}
		if !tenantPublicEligible(target.TenantStatus) {
			return CreateFeaturedParams{}, nil, invalidField("tenant_id", "tenant must be active or trialing")
		}
		if target.StoreStatus != "published" || !target.StoreDiscoverable {
			return CreateFeaturedParams{}, nil, invalidField("store_id", "product store must be published and discoverable")
		}
		if target.Status != "active" || !target.IsDiscoverable {
			return CreateFeaturedParams{}, nil, invalidField("product_id", "product must be active and discoverable")
		}
		params.ItemID = target.ID
		params.StoreID = &target.StoreID
		params.ProductID = &target.ID
		return params, target, nil
	default:
		return CreateFeaturedParams{}, nil, invalidField("item_type", "item_type must be store or product")
	}
}

func normalizeCreateFeaturedInput(input CreateFeaturedInput) (CreateFeaturedInput, error) {
	input.ItemType = strings.ToLower(strings.TrimSpace(input.ItemType))
	input.Placement = strings.ToLower(strings.TrimSpace(input.Placement))
	if input.Placement == "" {
		input.Placement = FeaturedPlacementHome
	}
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)

	var details []map[string]string
	if input.ActorUserID == uuid.Nil {
		details = append(details, map[string]string{"field": "actor_user_id", "message": "Actor is required"})
	}
	if input.TenantID == uuid.Nil {
		details = append(details, map[string]string{"field": "tenant_id", "message": "tenant_id must be a valid UUID"})
	}
	if input.ItemType != FeaturedItemTypeStore && input.ItemType != FeaturedItemTypeProduct {
		details = append(details, map[string]string{"field": "item_type", "message": "item_type must be store or product"})
	}
	if !allowedFeaturedPlacement(input.Placement) {
		details = append(details, map[string]string{"field": "placement", "message": "placement must be home, stores, products, category, or city"})
	}
	if input.StartsAt != nil && input.EndsAt != nil && !input.EndsAt.After(*input.StartsAt) {
		details = append(details, map[string]string{"field": "ends_at", "message": "ends_at must be after starts_at"})
	}
	if len(details) > 0 {
		return CreateFeaturedInput{}, apperror.Validation("Validation failed", details)
	}
	return input, nil
}

func normalizeUpdateFeaturedInput(input UpdateFeaturedInput) (UpdateFeaturedInput, error) {
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)
	if input.ItemType != nil {
		value := strings.ToLower(strings.TrimSpace(*input.ItemType))
		input.ItemType = &value
	}
	if input.Placement != nil {
		value := strings.ToLower(strings.TrimSpace(*input.Placement))
		input.Placement = &value
	}

	var details []map[string]string
	if input.ActorUserID == uuid.Nil {
		details = append(details, map[string]string{"field": "actor_user_id", "message": "Actor is required"})
	}
	if input.FeaturedID == uuid.Nil {
		details = append(details, map[string]string{"field": "featuredId", "message": "featuredId must be a valid UUID"})
	}
	if input.ItemType != nil && *input.ItemType != FeaturedItemTypeStore && *input.ItemType != FeaturedItemTypeProduct {
		details = append(details, map[string]string{"field": "item_type", "message": "item_type must be store or product"})
	}
	if input.Placement != nil && !allowedFeaturedPlacement(*input.Placement) {
		details = append(details, map[string]string{"field": "placement", "message": "placement must be home, stores, products, category, or city"})
	}
	if input.StartsAt != nil && input.EndsAt != nil && !input.EndsAt.After(*input.StartsAt) {
		details = append(details, map[string]string{"field": "ends_at", "message": "ends_at must be after starts_at"})
	}
	if len(details) > 0 {
		return UpdateFeaturedInput{}, apperror.Validation("Validation failed", details)
	}
	return input, nil
}

func normalizeFeaturedListFilters(filters FeaturedListFilters) (FeaturedListFilters, error) {
	filters.ItemType = strings.ToLower(strings.TrimSpace(filters.ItemType))
	filters.Placement = strings.ToLower(strings.TrimSpace(filters.Placement))
	if filters.ItemType != "" && filters.ItemType != FeaturedItemTypeStore && filters.ItemType != FeaturedItemTypeProduct {
		return FeaturedListFilters{}, invalidField("item_type", "item_type must be store or product")
	}
	if filters.Placement != "" && !allowedFeaturedPlacement(filters.Placement) {
		return FeaturedListFilters{}, invalidField("placement", "placement must be home, stores, products, category, or city")
	}
	if filters.Limit <= 0 {
		filters.Limit = defaultTenantListLimit
	}
	if filters.Limit > maxTenantListLimit {
		filters.Limit = maxTenantListLimit
	}
	return filters, nil
}

func normalizeAuditLogFilters(filters AuditLogListFilters) (AuditLogListFilters, error) {
	filters.Action = strings.TrimSpace(filters.Action)
	filters.TargetType = strings.TrimSpace(filters.TargetType)
	if filters.Limit <= 0 {
		filters.Limit = defaultTenantListLimit
	}
	if filters.Limit > maxTenantListLimit {
		filters.Limit = maxTenantListLimit
	}
	return filters, nil
}

func validateFeaturedDraft(draft featuredDraft) error {
	var details []map[string]string
	if draft.TenantID == uuid.Nil {
		details = append(details, map[string]string{"field": "tenant_id", "message": "tenant_id must be a valid UUID"})
	}
	if draft.ItemType != FeaturedItemTypeStore && draft.ItemType != FeaturedItemTypeProduct {
		details = append(details, map[string]string{"field": "item_type", "message": "item_type must be store or product"})
	}
	if !allowedFeaturedPlacement(draft.Placement) {
		details = append(details, map[string]string{"field": "placement", "message": "placement must be home, stores, products, category, or city"})
	}
	if draft.StartsAt != nil && draft.EndsAt != nil && !draft.EndsAt.After(*draft.StartsAt) {
		details = append(details, map[string]string{"field": "ends_at", "message": "ends_at must be after starts_at"})
	}
	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}
	return nil
}

func allowedFeaturedPlacement(value string) bool {
	switch value {
	case FeaturedPlacementHome, FeaturedPlacementStores, FeaturedPlacementProducts, FeaturedPlacementCategory, FeaturedPlacementCity:
		return true
	default:
		return false
	}
}

func tenantPublicEligible(status string) bool {
	return status == TenantStatusActive || status == TenantStatusTrialing
}

func enrichFeaturedTarget(item *FeaturedItem, target any) {
	switch typed := target.(type) {
	case *FeaturedStoreTarget:
		item.StoreName = typed.Name
		item.StoreSlug = typed.Slug
	case *FeaturedProductTarget:
		item.ProductName = typed.Name
		item.ProductSlug = typed.Slug
	}
}

func featuredAuditSnapshot(item *FeaturedItem) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":         item.ID.String(),
		"item_type":  item.ItemType,
		"tenant_id":  item.TenantID.String(),
		"store_id":   uuidPtrString(item.StoreID),
		"product_id": uuidPtrString(item.ProductID),
		"placement":  item.Placement,
		"sort_order": item.SortOrder,
		"starts_at":  optionalFormattedTime(item.StartsAt),
		"ends_at":    optionalFormattedTime(item.EndsAt),
		"is_active":  item.IsActive,
	}
}
