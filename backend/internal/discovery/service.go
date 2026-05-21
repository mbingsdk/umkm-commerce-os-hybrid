package discovery

import (
	"context"
	"strconv"
	"strings"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
	homeStoreLimit   = 6
	homeProductLimit = 8
	aggregateLimit   = 10
)

type discoveryStore interface {
	ListFeaturedStores(context.Context, db.Queryer, int) ([]Store, error)
	ListFeaturedProducts(context.Context, db.Queryer, int) ([]Product, error)
	ListStores(context.Context, db.Queryer, ListStoresFilters) ([]Store, error)
	ListProducts(context.Context, db.Queryer, ListProductsFilters) ([]Product, error)
	PopularCategories(context.Context, db.Queryer, int) ([]CategoryAggregate, error)
	PopularCities(context.Context, db.Queryer, int) ([]CityAggregate, error)
}

type Service struct {
	db        db.Queryer
	discovery discoveryStore
}

func NewService(database db.Queryer, discoveryRepo discoveryStore) *Service {
	return &Service{
		db:        database,
		discovery: discoveryRepo,
	}
}

func (s *Service) Home(ctx context.Context) (HomeResponse, error) {
	featuredStores, err := s.discovery.ListFeaturedStores(ctx, s.db, homeStoreLimit)
	if err != nil {
		return HomeResponse{}, apperror.Internal(err)
	}
	featuredProducts, err := s.discovery.ListFeaturedProducts(ctx, s.db, homeProductLimit)
	if err != nil {
		return HomeResponse{}, apperror.Internal(err)
	}
	latestStores, err := s.discovery.ListStores(ctx, s.db, ListStoresFilters{Limit: homeStoreLimit})
	if err != nil {
		return HomeResponse{}, apperror.Internal(err)
	}
	latestProducts, err := s.discovery.ListProducts(ctx, s.db, ListProductsFilters{Limit: homeProductLimit})
	if err != nil {
		return HomeResponse{}, apperror.Internal(err)
	}
	categories, err := s.discovery.PopularCategories(ctx, s.db, aggregateLimit)
	if err != nil {
		return HomeResponse{}, apperror.Internal(err)
	}
	cities, err := s.discovery.PopularCities(ctx, s.db, aggregateLimit)
	if err != nil {
		return HomeResponse{}, apperror.Internal(err)
	}

	return HomeResponse{
		FeaturedStores:    NewStoreResponses(featuredStores),
		FeaturedProducts:  NewProductResponses(featuredProducts),
		LatestStores:      NewStoreResponses(latestStores),
		LatestProducts:    NewProductResponses(latestProducts),
		PopularCategories: NewCategoryResponses(categories),
		PopularCities:     NewCityResponses(cities),
	}, nil
}

func (s *Service) ListStores(ctx context.Context, filters ListStoresFilters) ([]StoreResponse, PaginationMeta, error) {
	normalized := normalizeStoreFilters(filters)
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.discovery.ListStores(ctx, s.db, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	nextCursor, err := nextStoreCursor(items, hasMore)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	return NewStoreResponses(items), PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) ListProducts(ctx context.Context, filters ListProductsFilters) ([]ProductResponse, PaginationMeta, error) {
	normalized, err := normalizeProductFilters(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.discovery.ListProducts(ctx, s.db, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	nextCursor, err := nextProductCursor(items, hasMore)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	return NewProductResponses(items), PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) Search(ctx context.Context, filters SearchFilters) (SearchResponse, error) {
	normalized, err := normalizeSearchFilters(filters)
	if err != nil {
		return SearchResponse{}, err
	}

	var response SearchResponse
	if normalized.Type == SearchTypeAll || normalized.Type == SearchTypeStores {
		stores, err := s.discovery.ListStores(ctx, s.db, ListStoresFilters{
			Query:    normalized.Query,
			City:     normalized.City,
			Category: normalized.Category,
			Limit:    normalized.Limit,
		})
		if err != nil {
			return SearchResponse{}, apperror.Internal(err)
		}
		response.Stores = NewStoreResponses(stores)
	}
	if normalized.Type == SearchTypeAll || normalized.Type == SearchTypeProducts {
		products, err := s.discovery.ListProducts(ctx, s.db, ListProductsFilters{
			Query:    normalized.Query,
			City:     normalized.City,
			Category: normalized.Category,
			Limit:    normalized.Limit,
		})
		if err != nil {
			return SearchResponse{}, apperror.Internal(err)
		}
		response.Products = NewProductResponses(products)
	}

	return response, nil
}

func normalizeStoreFilters(filters ListStoresFilters) ListStoresFilters {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.City = strings.TrimSpace(filters.City)
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Limit = normalizeLimit(filters.Limit)
	return filters
}

func normalizeProductFilters(filters ListProductsFilters) (ListProductsFilters, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.City = strings.TrimSpace(filters.City)
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Limit = normalizeLimit(filters.Limit)
	if filters.PriceMin != nil && *filters.PriceMin < 0 {
		return ListProductsFilters{}, invalidField("price_min", "price_min must be greater than or equal to zero")
	}
	if filters.PriceMax != nil && *filters.PriceMax < 0 {
		return ListProductsFilters{}, invalidField("price_max", "price_max must be greater than or equal to zero")
	}
	if filters.PriceMin != nil && filters.PriceMax != nil && *filters.PriceMax < *filters.PriceMin {
		return ListProductsFilters{}, invalidField("price_max", "price_max must be greater than or equal to price_min")
	}
	return filters, nil
}

func normalizeSearchFilters(filters SearchFilters) (SearchFilters, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.City = strings.TrimSpace(filters.City)
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Type = strings.TrimSpace(filters.Type)
	if filters.Type == "" {
		filters.Type = SearchTypeAll
	}
	switch filters.Type {
	case SearchTypeAll, SearchTypeStores, SearchTypeProducts:
	default:
		return SearchFilters{}, invalidField("type", "type must be all, stores, or products")
	}
	filters.Limit = normalizeLimit(filters.Limit)
	return filters, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

func parseOptionalMoney(raw string, field string) (*int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, invalidField(field, field+" must be a valid integer")
	}
	return &value, nil
}

func nextStoreCursor(items []Store, hasMore bool) (*string, error) {
	if !hasMore || len(items) == 0 {
		return nil, nil
	}
	encoded, err := EncodeStoreCursor(items[len(items)-1])
	if err != nil {
		return nil, err
	}
	return &encoded, nil
}

func nextProductCursor(items []Product, hasMore bool) (*string, error) {
	if !hasMore || len(items) == 0 {
		return nil, nil
	}
	encoded, err := EncodeProductCursor(items[len(items)-1])
	if err != nil {
		return nil, err
	}
	return &encoded, nil
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{{"field": field, "message": message}})
}
