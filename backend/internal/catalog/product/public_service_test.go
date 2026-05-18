package product

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

func TestPublicListHidesInactiveAndCrossStoreProducts(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	storeID := uuid.New()
	service := NewPublicService(
		fakeDatabase{},
		&fakePublicProductStore{
			listPublic: func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, PublicListFilters) ([]PublicListItem, error) {
				return []PublicListItem{
					{
						ID:        uuid.New(),
						TenantID:  tenantID,
						StoreID:   storeID,
						Name:      "Bouquet Mawar",
						Slug:      "bouquet-mawar",
						Status:    StatusActive,
						CreatedAt: time.Now(),
					},
					{
						ID:        uuid.New(),
						TenantID:  tenantID,
						StoreID:   storeID,
						Name:      "Draft",
						Slug:      "draft",
						Status:    StatusDraft,
						CreatedAt: time.Now(),
					},
					{
						ID:        uuid.New(),
						TenantID:  tenantID,
						StoreID:   uuid.New(),
						Name:      "Produk Tenant Lain",
						Slug:      "produk-tenant-lain",
						Status:    StatusActive,
						CreatedAt: time.Now(),
					},
				}, nil
			},
		},
		fakePublicStoreResolver{
			currentStore: publishedPublicStore(tenantID, storeID),
		},
	)

	result, _, err := service.List(context.Background(), "toko-bunga-ayu", PublicListFilters{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("List len = %d, want 1", len(result))
	}
	if result[0].Slug != "bouquet-mawar" {
		t.Fatalf("List slug = %s, want bouquet-mawar", result[0].Slug)
	}
}

func TestPublicGetHidesProductFromAnotherStore(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	storeID := uuid.New()
	service := NewPublicService(
		fakeDatabase{},
		&fakePublicProductStore{
			findPublicBySlug: func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, string) (*PublicProduct, error) {
				return &PublicProduct{
					ID:       uuid.New(),
					TenantID: tenantID,
					StoreID:  uuid.New(),
					Name:     "Produk Tenant Lain",
					Slug:     "produk-tenant-lain",
					Status:   StatusActive,
				}, nil
			},
		},
		fakePublicStoreResolver{
			currentStore: publishedPublicStore(tenantID, storeID),
		},
	)

	_, err := service.Get(context.Background(), "toko-bunga-ayu", "produk-tenant-lain")
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Get error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeNotFound {
		t.Fatalf("Get code = %s, want %s", appErr.Code, apperror.CodeNotFound)
	}
}

func TestPublicDetailResponseNeverExposesCostPrice(t *testing.T) {
	t.Parallel()

	payload, err := json.Marshal(NewPublicDetailResponse(&PublicProduct{
		ID:          uuid.New(),
		Name:        "Bouquet Mawar",
		Slug:        "bouquet-mawar",
		Description: "Bouquet mawar merah",
		Price:       150000,
		Status:      StatusActive,
		Stock: Stock{
			QuantityAvailable: 8,
		},
	}, publishedPublicStore(uuid.New(), uuid.New())))
	if err != nil {
		t.Fatalf("Marshal error = %v", err)
	}
	if strings.Contains(string(payload), "cost_price") {
		t.Fatalf("public response leaked cost_price: %s", payload)
	}
}

type fakePublicStoreResolver struct {
	currentStore store.PublicContext
	err          error
}

func (f fakePublicStoreResolver) Resolve(context.Context, string) (store.PublicContext, error) {
	if f.err != nil {
		return store.PublicContext{}, f.err
	}
	return f.currentStore, nil
}

type fakePublicProductStore struct {
	listPublic       func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, PublicListFilters) ([]PublicListItem, error)
	findPublicBySlug func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, string) (*PublicProduct, error)
}

func (f *fakePublicProductStore) ListPublic(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters PublicListFilters,
) ([]PublicListItem, error) {
	if f.listPublic == nil {
		return nil, nil
	}
	return f.listPublic(ctx, q, tenantID, storeID, filters)
}

func (f *fakePublicProductStore) FindPublicBySlug(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	slug string,
) (*PublicProduct, error) {
	if f.findPublicBySlug == nil {
		return nil, ErrProductNotFound
	}
	return f.findPublicBySlug(ctx, q, tenantID, storeID, slug)
}

func (*fakePublicProductStore) ListImages(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) ([]Image, error) {
	return nil, nil
}

func publishedPublicStore(tenantID uuid.UUID, storeID uuid.UUID) store.PublicContext {
	return store.PublicContext{
		TenantID: tenantID,
		StoreID:  storeID,
		Store: store.Store{
			ID:     storeID,
			Name:   "Toko Bunga Ayu",
			Slug:   "toko-bunga-ayu",
			City:   "Makassar",
			Status: store.StatusPublished,
		},
	}
}
