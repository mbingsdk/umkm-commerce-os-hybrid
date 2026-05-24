package product

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

const (
	defaultPublicLimit = 20
	maxPublicLimit     = 100
)

type publicStoreResolver interface {
	Resolve(ctx context.Context, slug string) (store.PublicContext, error)
}

type publicProductStore interface {
	ListPublic(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters PublicListFilters) ([]PublicListItem, error)
	FindPublicBySlug(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, slug string) (*PublicProduct, error)
	ListImages(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) ([]Image, error)
}

type PublicService struct {
	db       db.Queryer
	products publicProductStore
	stores   publicStoreResolver
}

func NewPublicService(database db.Queryer, products publicProductStore, stores publicStoreResolver) *PublicService {
	return &PublicService{
		db:       database,
		products: products,
		stores:   stores,
	}
}

func (s *PublicService) List(
	ctx context.Context,
	storeSlug string,
	filters PublicListFilters,
) ([]PublicListItemResponse, PublicPaginationMeta, error) {
	currentStore, err := s.stores.Resolve(ctx, storeSlug)
	if err != nil {
		return nil, PublicPaginationMeta{}, err
	}

	filters = normalizePublicFilters(filters)
	queryFilters := filters
	queryFilters.Limit = filters.Limit + 1

	items, err := s.products.ListPublic(ctx, s.db, currentStore.TenantID, currentStore.StoreID, queryFilters)
	if err != nil {
		return nil, PublicPaginationMeta{}, apperror.Internal(err)
	}

	visibleItems := make([]PublicListItem, 0, len(items))
	for _, item := range items {
		if item.TenantID != currentStore.TenantID ||
			item.StoreID != currentStore.StoreID ||
			item.Status != StatusActive {
			continue
		}
		visibleItems = append(visibleItems, item)
	}

	hasMore := len(visibleItems) > filters.Limit
	if hasMore {
		visibleItems = visibleItems[:filters.Limit]
	}

	response := make([]PublicListItemResponse, 0, len(visibleItems))
	for idx := range visibleItems {
		response = append(response, NewPublicListItemResponse(&visibleItems[idx]))
	}

	var nextCursor *string
	if hasMore && len(visibleItems) > 0 {
		encoded, err := EncodePublicCursor(visibleItems[len(visibleItems)-1])
		if err != nil {
			return nil, PublicPaginationMeta{}, apperror.Internal(err)
		}
		nextCursor = &encoded
	}

	return response, PublicPaginationMeta{
		Pagination: PublicPagination{
			Limit:      filters.Limit,
			NextCursor: nextCursor,
			HasMore:    hasMore,
		},
	}, nil
}

func (s *PublicService) Get(ctx context.Context, storeSlug string, productSlug string) (PublicDetailResponse, error) {
	currentStore, err := s.stores.Resolve(ctx, storeSlug)
	if err != nil {
		return PublicDetailResponse{}, err
	}

	productSlug = strings.TrimSpace(productSlug)
	if productSlug == "" {
		return PublicDetailResponse{}, apperror.NotFound("Product not found")
	}

	item, err := s.products.FindPublicBySlug(ctx, s.db, currentStore.TenantID, currentStore.StoreID, productSlug)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return PublicDetailResponse{}, apperror.NotFound("Product not found")
		}
		return PublicDetailResponse{}, apperror.Internal(err)
	}
	if item.TenantID != currentStore.TenantID ||
		item.StoreID != currentStore.StoreID ||
		item.Status != StatusActive {
		return PublicDetailResponse{}, apperror.NotFound("Product not found")
	}

	images, err := s.products.ListImages(ctx, s.db, currentStore.TenantID, currentStore.StoreID, item.ID)
	if err != nil {
		return PublicDetailResponse{}, apperror.Internal(err)
	}
	item.Images = images

	return NewPublicDetailResponse(item, currentStore), nil
}

func normalizePublicFilters(filters PublicListFilters) PublicListFilters {
	filters.Query = querytext.NormalizeSearch(filters.Query)
	filters.CategorySlug = querytext.NormalizeSearch(filters.CategorySlug)
	if filters.Limit <= 0 {
		filters.Limit = defaultPublicLimit
	}
	if filters.Limit > maxPublicLimit {
		filters.Limit = maxPublicLimit
	}
	return filters
}
