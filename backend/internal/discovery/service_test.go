package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	discoveryStoreA   = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	discoveryStoreB   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	discoveryStoreC   = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	discoveryProductA = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	discoveryProductB = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	discoveryProductC = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

func TestUnpublishedStoreHidden(t *testing.T) {
	service := newDiscoveryTestService()

	stores, _, err := service.ListStores(context.Background(), ListStoresFilters{})
	if err != nil {
		t.Fatalf("ListStores error = %v", err)
	}
	for _, store := range stores {
		if store.ID == discoveryStoreB.String() {
			t.Fatalf("unpublished store leaked: %#v", stores)
		}
	}
}

func TestInactiveTenantHidden(t *testing.T) {
	service := newDiscoveryTestService()

	stores, _, err := service.ListStores(context.Background(), ListStoresFilters{})
	if err != nil {
		t.Fatalf("ListStores error = %v", err)
	}
	for _, store := range stores {
		if store.ID == discoveryStoreC.String() {
			t.Fatalf("inactive tenant store leaked: %#v", stores)
		}
	}
}

func TestNonDiscoverableAndInactiveProductsHidden(t *testing.T) {
	service := newDiscoveryTestService()

	products, _, err := service.ListProducts(context.Background(), ListProductsFilters{})
	if err != nil {
		t.Fatalf("ListProducts error = %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("products len = %d, want only eligible pagination window 2 including extra candidate handling got %#v", len(products), products)
	}
	for _, product := range products {
		if product.ID == discoveryProductB.String() || product.ID == discoveryProductC.String() {
			t.Fatalf("hidden product leaked: %#v", products)
		}
	}
}

func TestCostPriceNeverExposed(t *testing.T) {
	service := newDiscoveryTestService()

	result, err := service.Search(context.Background(), SearchFilters{Query: "Bouquet", Type: SearchTypeProducts, Limit: 10})
	if err != nil {
		t.Fatalf("Search error = %v", err)
	}
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal search result: %v", err)
	}
	payload := string(raw)
	if strings.Contains(payload, "cost_price") || strings.Contains(payload, "999999") {
		t.Fatalf("public discovery leaked cost price: %s", payload)
	}
}

func TestSearchScopedToPublicEligibleRecordsOnly(t *testing.T) {
	service := newDiscoveryTestService()

	result, err := service.Search(context.Background(), SearchFilters{Query: "Rahasia", Type: SearchTypeAll, Limit: 10})
	if err != nil {
		t.Fatalf("Search error = %v", err)
	}
	if len(result.Stores) != 0 || len(result.Products) != 0 {
		t.Fatalf("search returned non-public records: %#v", result)
	}
}

func TestPaginationWorks(t *testing.T) {
	service := newDiscoveryTestService()

	products, meta, err := service.ListProducts(context.Background(), ListProductsFilters{Limit: 1})
	if err != nil {
		t.Fatalf("ListProducts error = %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("products len = %d, want 1", len(products))
	}
	if !meta.Pagination.HasMore || meta.Pagination.NextCursor == nil {
		t.Fatalf("pagination meta = %#v, want has_more with cursor", meta)
	}
}

func TestDiscoveryEndpointDoesNotRequireAuth(t *testing.T) {
	service := newDiscoveryTestService()
	handler := NewHandler(service, slog.New(slog.NewTextHandler(httptest.NewRecorder(), nil)))
	r := chi.NewRouter()
	RegisterPublicRoutes(r, handler)

	req := httptest.NewRequest(http.MethodGet, "/public/discovery/home", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func newDiscoveryTestService() *Service {
	now := time.Date(2026, 5, 21, 8, 0, 0, 0, time.UTC)
	repo := &fakeDiscoveryRepository{
		stores: []fakeDiscoveryStore{
			{
				Store: Store{
					ID:          discoveryStoreA,
					Name:        "Toko Bunga Ayu",
					Slug:        "toko-bunga-ayu",
					Description: "Bouquet Makassar",
					City:        "Makassar",
					Province:    "Sulawesi Selatan",
					CreatedAt:   now,
				},
				TenantStatus: tenantStatusActive,
				StoreStatus:  storeStatusPublished,
				Discoverable: true,
			},
			{
				Store: Store{
					ID:          discoveryStoreB,
					Name:        "Toko Rahasia",
					Slug:        "toko-rahasia",
					Description: "Tidak publik",
					City:        "Makassar",
					CreatedAt:   now.Add(-time.Hour),
				},
				TenantStatus: tenantStatusActive,
				StoreStatus:  "draft",
				Discoverable: true,
			},
			{
				Store: Store{
					ID:        discoveryStoreC,
					Name:      "Tenant Suspended",
					Slug:      "tenant-suspended",
					City:      "Makassar",
					CreatedAt: now.Add(-2 * time.Hour),
				},
				TenantStatus: "suspended",
				StoreStatus:  storeStatusPublished,
				Discoverable: true,
			},
		},
		products: []fakeDiscoveryProduct{
			{
				Product: Product{
					ID:              discoveryProductA,
					Name:            "Bouquet Mawar",
					Slug:            "bouquet-mawar",
					Description:     "Produk publik",
					Price:           50000,
					PrimaryImageURL: "/uploads/mawar.webp",
					CategoryName:    "Bouquet",
					CategorySlug:    "bouquet",
					StoreID:         discoveryStoreA,
					StoreName:       "Toko Bunga Ayu",
					StoreSlug:       "toko-bunga-ayu",
					StoreCity:       "Makassar",
					CreatedAt:       now,
				},
				TenantStatus:        tenantStatusActive,
				StoreStatus:         storeStatusPublished,
				StoreDiscoverable:   true,
				ProductStatus:       productStatusActive,
				ProductDiscoverable: true,
				CostPrice:           999999,
			},
			{
				Product: Product{
					ID:           uuid.New(),
					Name:         "Bouquet Melati",
					Slug:         "bouquet-melati",
					Price:        60000,
					CategoryName: "Bouquet",
					CategorySlug: "bouquet",
					StoreID:      discoveryStoreA,
					StoreName:    "Toko Bunga Ayu",
					StoreSlug:    "toko-bunga-ayu",
					StoreCity:    "Makassar",
					CreatedAt:    now.Add(-time.Minute),
				},
				TenantStatus:        tenantStatusTrialing,
				StoreStatus:         storeStatusPublished,
				StoreDiscoverable:   true,
				ProductStatus:       productStatusActive,
				ProductDiscoverable: true,
			},
			{
				Product: Product{
					ID:           discoveryProductB,
					Name:         "Produk Rahasia",
					Slug:         "produk-rahasia",
					Price:        100000,
					CategorySlug: "secret",
					StoreID:      discoveryStoreA,
					StoreName:    "Toko Bunga Ayu",
					StoreSlug:    "toko-bunga-ayu",
					StoreCity:    "Makassar",
					CreatedAt:    now.Add(-2 * time.Minute),
				},
				TenantStatus:        tenantStatusActive,
				StoreStatus:         storeStatusPublished,
				StoreDiscoverable:   true,
				ProductStatus:       productStatusActive,
				ProductDiscoverable: false,
			},
			{
				Product: Product{
					ID:        discoveryProductC,
					Name:      "Produk Inactive",
					Slug:      "produk-inactive",
					Price:     100000,
					StoreID:   discoveryStoreA,
					StoreName: "Toko Bunga Ayu",
					StoreSlug: "toko-bunga-ayu",
					CreatedAt: now.Add(-3 * time.Minute),
				},
				TenantStatus:        tenantStatusActive,
				StoreStatus:         storeStatusPublished,
				StoreDiscoverable:   true,
				ProductStatus:       "draft",
				ProductDiscoverable: true,
			},
		},
	}
	return NewService(fakeDiscoveryDB{}, repo)
}

type fakeDiscoveryDB struct{}

func (fakeDiscoveryDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakeDiscoveryDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakeDiscoveryDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakeDiscoveryRepository struct {
	stores   []fakeDiscoveryStore
	products []fakeDiscoveryProduct
}

type fakeDiscoveryStore struct {
	Store
	TenantStatus string
	StoreStatus  string
	Discoverable bool
	Deleted      bool
}

type fakeDiscoveryProduct struct {
	Product
	TenantStatus        string
	StoreStatus         string
	StoreDiscoverable   bool
	StoreDeleted        bool
	ProductStatus       string
	ProductDiscoverable bool
	ProductDeleted      bool
	CostPrice           int64
}

func (f *fakeDiscoveryRepository) ListFeaturedStores(ctx context.Context, q db.Queryer, limit int) ([]Store, error) {
	return f.ListStores(ctx, q, ListStoresFilters{Limit: limit})
}

func (f *fakeDiscoveryRepository) ListFeaturedProducts(ctx context.Context, q db.Queryer, limit int) ([]Product, error) {
	return f.ListProducts(ctx, q, ListProductsFilters{Limit: limit})
}

func (f *fakeDiscoveryRepository) ListStores(_ context.Context, _ db.Queryer, filters ListStoresFilters) ([]Store, error) {
	items := make([]Store, 0)
	for _, item := range f.stores {
		if !eligibleFakeStore(item) {
			continue
		}
		if filters.Query != "" && !containsAny(filters.Query, item.Name, item.Description, item.City, item.Province) {
			continue
		}
		if filters.City != "" && !strings.EqualFold(filters.City, item.City) {
			continue
		}
		items = append(items, item.Store)
	}
	return capStores(items, filters.Limit), nil
}

func (f *fakeDiscoveryRepository) ListProducts(_ context.Context, _ db.Queryer, filters ListProductsFilters) ([]Product, error) {
	items := make([]Product, 0)
	for _, item := range f.products {
		if !eligibleFakeProduct(item) {
			continue
		}
		if filters.Query != "" && !containsAny(filters.Query, item.Name, item.Description, item.StoreName, item.CategoryName, item.CategorySlug) {
			continue
		}
		if filters.City != "" && !strings.EqualFold(filters.City, item.StoreCity) {
			continue
		}
		if filters.Category != "" && !strings.EqualFold(filters.Category, item.CategorySlug) && !strings.Contains(strings.ToLower(item.CategoryName), strings.ToLower(filters.Category)) {
			continue
		}
		if filters.PriceMin != nil && item.Price < *filters.PriceMin {
			continue
		}
		if filters.PriceMax != nil && item.Price > *filters.PriceMax {
			continue
		}
		items = append(items, item.Product)
	}
	return capProducts(items, filters.Limit), nil
}

func (f *fakeDiscoveryRepository) PopularCategories(_ context.Context, _ db.Queryer, limit int) ([]CategoryAggregate, error) {
	items := []CategoryAggregate{{Name: "Bouquet", Slug: "bouquet", Count: 2}}
	if limit < len(items) {
		return items[:limit], nil
	}
	return items, nil
}

func (f *fakeDiscoveryRepository) PopularCities(_ context.Context, _ db.Queryer, limit int) ([]CityAggregate, error) {
	items := []CityAggregate{{City: "Makassar", Count: 1}}
	if limit < len(items) {
		return items[:limit], nil
	}
	return items, nil
}

func eligibleFakeStore(item fakeDiscoveryStore) bool {
	return !item.Deleted &&
		(item.TenantStatus == tenantStatusActive || item.TenantStatus == tenantStatusTrialing) &&
		item.StoreStatus == storeStatusPublished &&
		item.Discoverable
}

func eligibleFakeProduct(item fakeDiscoveryProduct) bool {
	return !item.StoreDeleted &&
		!item.ProductDeleted &&
		(item.TenantStatus == tenantStatusActive || item.TenantStatus == tenantStatusTrialing) &&
		item.StoreStatus == storeStatusPublished &&
		item.StoreDiscoverable &&
		item.ProductStatus == productStatusActive &&
		item.ProductDiscoverable
}

func containsAny(query string, values ...string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	for _, value := range values {
		if strings.Contains(strings.ToLower(value), query) {
			return true
		}
	}
	return false
}

func capStores(items []Store, limit int) []Store {
	if limit <= 0 || limit > len(items) {
		return items
	}
	return items[:limit]
}

func capProducts(items []Product, limit int) []Product {
	if limit <= 0 || limit > len(items) {
		return items
	}
	return items[:limit]
}
