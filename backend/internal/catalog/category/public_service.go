package category

import (
	"context"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

type publicStoreResolver interface {
	Resolve(ctx context.Context, slug string) (store.PublicContext, error)
}

type publicCategoryStore interface {
	ListPublic(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) ([]PublicCategory, error)
}

type PublicService struct {
	db         db.Queryer
	categories publicCategoryStore
	stores     publicStoreResolver
}

func NewPublicService(database db.Queryer, categories publicCategoryStore, stores publicStoreResolver) *PublicService {
	return &PublicService{
		db:         database,
		categories: categories,
		stores:     stores,
	}
}

func (s *PublicService) List(ctx context.Context, storeSlug string) ([]PublicResponse, error) {
	currentStore, err := s.stores.Resolve(ctx, storeSlug)
	if err != nil {
		return nil, err
	}

	items, err := s.categories.ListPublic(ctx, s.db, currentStore.TenantID, currentStore.StoreID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	response := make([]PublicResponse, 0, len(items))
	for idx := range items {
		response = append(response, NewPublicResponse(&items[idx]))
	}

	return response, nil
}
