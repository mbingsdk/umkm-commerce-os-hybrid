package product

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/catalog/category"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/storage"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	plans "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/plan"
)

func TestValidateCreate(t *testing.T) {
	t.Parallel()

	negativeCost := int64(-1)
	discountTooLow := int64(100)
	tests := []struct {
		name  string
		input CreateInput
	}{
		{
			name: "negative price",
			input: CreateInput{
				Name:   "Bouquet Mawar",
				Slug:   "bouquet-mawar",
				Price:  -1,
				Status: StatusDraft,
			},
		},
		{
			name: "compare at price below price",
			input: CreateInput{
				Name:           "Bouquet Mawar",
				Slug:           "bouquet-mawar",
				Price:          120,
				CompareAtPrice: &discountTooLow,
				Status:         StatusDraft,
			},
		},
		{
			name: "negative cost price",
			input: CreateInput{
				Name:      "Bouquet Mawar",
				Slug:      "bouquet-mawar",
				Price:     120,
				CostPrice: &negativeCost,
				Status:    StatusDraft,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := validateCreate(tt.input)
			appErr, ok := err.(*apperror.AppError)
			if !ok {
				t.Fatalf("validateCreate error type = %T, want *apperror.AppError", err)
			}
			if appErr.Code != apperror.CodeValidation {
				t.Fatalf("validateCreate code = %s, want %s", appErr.Code, apperror.CodeValidation)
			}
		})
	}
}

func TestCreateRejectsDuplicateSlug(t *testing.T) {
	t.Parallel()

	service := NewService(
		fakeDatabase{},
		&fakeProductStore{
			create: func(context.Context, db.Queryer, CreateParams) (*Product, error) {
				return nil, ErrProductSlugAlreadyInUse
			},
		},
		&fakeCategoryReader{},
		&fakeStockWriter{},
		&fakeImageStore{},
		&fakeAssetStore{},
	)

	_, err := service.Create(context.Background(), uuid.New(), uuid.New(), uuid.New(), validCreateInput())
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Create error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeValidation {
		t.Fatalf("Create code = %s, want %s", appErr.Code, apperror.CodeValidation)
	}
}

func TestCreateInitialStockCreatesSnapshotAndMovement(t *testing.T) {
	t.Parallel()

	productID := uuid.New()
	stockWriter := &fakeStockWriter{}
	service := NewService(
		fakeDatabase{},
		&fakeProductStore{
			create: func(_ context.Context, _ db.Queryer, params CreateParams) (*Product, error) {
				return &Product{
					ID:             productID,
					TenantID:       params.TenantID,
					StoreID:        params.StoreID,
					Name:           params.Name,
					Slug:           params.Slug,
					Status:         params.Status,
					TrackInventory: params.TrackInventory,
				}, nil
			},
		},
		&fakeCategoryReader{},
		stockWriter,
		&fakeImageStore{},
		&fakeAssetStore{},
	)

	input := validCreateInput()
	input.InitialStock = 7
	result, err := service.Create(context.Background(), uuid.New(), uuid.New(), uuid.New(), input)
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if result.Stock.QuantityOnHand != 7 || result.Stock.QuantityAvailable != 7 {
		t.Fatalf("Create stock = %+v, want initial stock quantities of 7", result.Stock)
	}
	if len(stockWriter.snapshots) != 1 {
		t.Fatalf("snapshot calls = %d, want 1", len(stockWriter.snapshots))
	}
	if len(stockWriter.movements) != 1 {
		t.Fatalf("movement calls = %d, want 1", len(stockWriter.movements))
	}
	if stockWriter.movements[0].MovementType != inventory.MovementTypeInitial {
		t.Fatalf("movement type = %s, want %s", stockWriter.movements[0].MovementType, inventory.MovementTypeInitial)
	}
}

func TestCreateRejectsProductPlanLimit(t *testing.T) {
	t.Parallel()

	service := NewService(
		fakeDatabase{},
		&fakeProductStore{},
		&fakeCategoryReader{},
		&fakeStockWriter{},
		&fakeImageStore{},
		&fakeAssetStore{},
		&fakePlanChecker{productLimitErr: apperror.PlanLimitExceeded("Product limit exceeded", nil)},
	)

	_, err := service.Create(context.Background(), uuid.New(), uuid.New(), uuid.New(), validCreateInput())
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Create error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodePlanLimitExceeded {
		t.Fatalf("Create code = %s, want %s", appErr.Code, apperror.CodePlanLimitExceeded)
	}
}

func TestUpdateRejectsDiscoveryWhenPlanDisallowsIt(t *testing.T) {
	t.Parallel()

	productID := uuid.New()
	discoverable := true
	service := NewService(
		fakeDatabase{},
		&fakeProductStore{
			findByID: func(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, id uuid.UUID) (*Product, error) {
				return &Product{
					ID:             id,
					TenantID:       tenantID,
					StoreID:        storeID,
					Name:           "Bouquet Mawar",
					Slug:           "bouquet-mawar",
					Price:          150000,
					Status:         StatusActive,
					TrackInventory: true,
				}, nil
			},
		},
		&fakeCategoryReader{},
		&fakeStockWriter{},
		&fakeImageStore{},
		&fakeAssetStore{},
		&fakePlanChecker{featureErr: apperror.Forbidden("Feature is not available on current plan")},
	)

	_, err := service.Update(context.Background(), uuid.New(), uuid.New(), productID, UpdateInput{IsDiscoverable: &discoverable})
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Update error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeForbidden {
		t.Fatalf("Update code = %s, want %s", appErr.Code, apperror.CodeForbidden)
	}
}

func TestCreateZeroInitialStockStillCreatesSnapshot(t *testing.T) {
	t.Parallel()

	stockWriter := &fakeStockWriter{}
	service := NewService(
		fakeDatabase{},
		&fakeProductStore{
			create: func(_ context.Context, _ db.Queryer, params CreateParams) (*Product, error) {
				return &Product{
					ID:             uuid.New(),
					TenantID:       params.TenantID,
					StoreID:        params.StoreID,
					Name:           params.Name,
					Slug:           params.Slug,
					Status:         params.Status,
					TrackInventory: params.TrackInventory,
				}, nil
			},
		},
		&fakeCategoryReader{},
		stockWriter,
		&fakeImageStore{},
		&fakeAssetStore{},
	)

	if _, err := service.Create(context.Background(), uuid.New(), uuid.New(), uuid.New(), validCreateInput()); err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if len(stockWriter.snapshots) != 1 {
		t.Fatalf("snapshot calls = %d, want 1", len(stockWriter.snapshots))
	}
	if len(stockWriter.movements) != 0 {
		t.Fatalf("movement calls = %d, want 0", len(stockWriter.movements))
	}
}

func validCreateInput() CreateInput {
	return CreateInput{
		Name:           "Bouquet Mawar",
		Slug:           "bouquet-mawar",
		Price:          150000,
		Status:         StatusDraft,
		TrackInventory: true,
	}
}

type fakeDatabase struct{}

func (fakeDatabase) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func (fakeDatabase) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (fakeDatabase) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}

func (fakeDatabase) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(noopTx{})
}

type noopTx struct{}

func (noopTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func (noopTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (noopTx) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}

type fakeProductStore struct {
	create   func(context.Context, db.Queryer, CreateParams) (*Product, error)
	findByID func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Product, error)
}

func (f *fakeProductStore) List(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ListFilters) ([]Product, error) {
	return nil, nil
}

func (f *fakeProductStore) Create(ctx context.Context, q db.Queryer, params CreateParams) (*Product, error) {
	if f.create != nil {
		return f.create(ctx, q, params)
	}
	return nil, errors.New("unexpected Create")
}

func (f *fakeProductStore) FindByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) (*Product, error) {
	if f.findByID != nil {
		return f.findByID(ctx, q, tenantID, storeID, productID)
	}
	return nil, ErrProductNotFound
}

func (f *fakeProductStore) ListImages(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) ([]Image, error) {
	return nil, nil
}

func (f *fakeProductStore) Update(context.Context, db.Queryer, UpdateParams) error {
	return nil
}

func (f *fakeProductStore) SoftDelete(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

type fakeCategoryReader struct{}

func (fakeCategoryReader) FindByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*category.Category, error) {
	return nil, category.ErrCategoryNotFound
}

type fakeStockWriter struct {
	snapshots []inventory.CreateSnapshotParams
	movements []inventory.CreateMovementParams
}

func (f *fakeStockWriter) CreateSnapshot(_ context.Context, _ db.Queryer, params inventory.CreateSnapshotParams) (*inventory.StockSnapshot, error) {
	f.snapshots = append(f.snapshots, params)
	return &inventory.StockSnapshot{
		QuantityOnHand:    params.QuantityOnHand,
		QuantityReserved:  params.QuantityReserved,
		QuantityAvailable: params.QuantityAvailable,
		LowStockThreshold: params.LowStockThreshold,
	}, nil
}

func (f *fakeStockWriter) CreateMovement(_ context.Context, _ db.Queryer, params inventory.CreateMovementParams) (*inventory.StockMovement, error) {
	f.movements = append(f.movements, params)
	return &inventory.StockMovement{}, nil
}

type fakeImageStore struct{}

func (fakeImageStore) Create(context.Context, db.Queryer, CreateImageParams) (*Image, error) {
	return nil, errors.New("unexpected Create")
}

func (fakeImageStore) Delete(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return errors.New("unexpected Delete")
}

type fakeAssetStore struct{}

func (fakeAssetStore) Store(context.Context, storage.StoreInput) (storage.Asset, error) {
	return storage.Asset{}, errors.New("unexpected Store")
}

func (fakeAssetStore) Delete(context.Context, string) error {
	return errors.New("unexpected Delete")
}

type fakePlanChecker struct {
	productLimitErr error
	featureErr      error
}

func (f *fakePlanChecker) CheckProductLimit(context.Context, db.Queryer, uuid.UUID, uuid.UUID) error {
	return f.productLimitErr
}

func (f *fakePlanChecker) RequireFeature(context.Context, db.Queryer, uuid.UUID, plans.Feature) error {
	return f.featureErr
}
