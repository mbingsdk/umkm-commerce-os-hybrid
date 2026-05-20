package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
)

var (
	invTenantA  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	invStoreA   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	invTenantB  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	invStoreB   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	invProductA = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	invProductB = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	invActorID  = uuid.MustParse("77777777-7777-7777-7777-777777777777")
)

func TestStockListIsTenantScoped(t *testing.T) {
	service, _, _, _ := newInventoryTestService()

	items, _, err := service.ListStocks(context.Background(), invTenantA, invStoreA, ListStockFilters{})
	if err != nil {
		t.Fatalf("ListStocks error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("stocks len = %d, want 1", len(items))
	}
	if items[0].ProductID != invProductA.String() {
		t.Fatalf("product id = %s, want %s", items[0].ProductID, invProductA)
	}
}

func TestMovementListIsTenantScoped(t *testing.T) {
	service, _, _, _ := newInventoryTestService()

	_, _, err := service.ListMovements(context.Background(), invTenantA, invStoreA, invProductB, ListMovementFilters{})
	assertInventoryAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestAdjustmentInIncreasesStockAndCreatesSideEffects(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newInventoryTestService()

	result, err := service.AdjustStock(context.Background(), invTenantA, invStoreA, invProductA, AdjustStockInput{
		ActorUserID:    invActorID,
		AdjustmentType: AdjustmentTypeIn,
		Quantity:       3,
		Reason:         "restock_supplier",
		Note:           "Restock dari supplier lokal",
	})
	if err != nil {
		t.Fatalf("AdjustStock error = %v", err)
	}
	if result.QuantityOnHand != 13 || result.QuantityAvailable != 11 || result.QuantityReserved != 2 {
		t.Fatalf("adjust result = %#v, want on hand 13 available 11 reserved 2", result)
	}
	if len(repo.movements) != 1 {
		t.Fatalf("movements = %d, want 1", len(repo.movements))
	}
	if repo.movements[0].MovementType != MovementTypeAdjustmentIn || repo.movements[0].Quantity != 3 {
		t.Fatalf("movement = %#v", repo.movements[0])
	}
	if len(auditRepo.entries) != 1 || auditRepo.entries[0].Action != AuditActionStockAdjusted {
		t.Fatalf("audit entries = %#v", auditRepo.entries)
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventStockAdjusted {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestAdjustmentOutDecreasesStock(t *testing.T) {
	service, repo, _, _ := newInventoryTestService()

	result, err := service.AdjustStock(context.Background(), invTenantA, invStoreA, invProductA, AdjustStockInput{
		ActorUserID:    invActorID,
		AdjustmentType: AdjustmentTypeOut,
		Quantity:       3,
		Reason:         "damaged_stock",
	})
	if err != nil {
		t.Fatalf("AdjustStock error = %v", err)
	}
	if result.QuantityOnHand != 7 || result.QuantityAvailable != 5 || result.QuantityReserved != 2 {
		t.Fatalf("adjust result = %#v, want on hand 7 available 5 reserved 2", result)
	}
	if len(repo.movements) != 1 {
		t.Fatalf("movements = %d, want 1", len(repo.movements))
	}
	if repo.movements[0].MovementType != MovementTypeAdjustmentOut || repo.movements[0].Quantity != -3 {
		t.Fatalf("movement = %#v", repo.movements[0])
	}
}

func TestAdjustmentOutCannotMakeAvailableStockNegative(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newInventoryTestService()

	_, err := service.AdjustStock(context.Background(), invTenantA, invStoreA, invProductA, AdjustStockInput{
		ActorUserID:    invActorID,
		AdjustmentType: AdjustmentTypeOut,
		Quantity:       9,
		Reason:         "stock_count_correction",
	})
	assertInventoryAppErrorCode(t, err, apperror.CodeInsufficientStock)
	if len(repo.movements) != 0 {
		t.Fatalf("movements = %d, want 0", len(repo.movements))
	}
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestAdjustmentRejectsOtherTenantProduct(t *testing.T) {
	service, repo, _, _ := newInventoryTestService()

	_, err := service.AdjustStock(context.Background(), invTenantA, invStoreA, invProductB, AdjustStockInput{
		ActorUserID:    invActorID,
		AdjustmentType: AdjustmentTypeIn,
		Quantity:       1,
		Reason:         "wrong_tenant",
	})
	assertInventoryAppErrorCode(t, err, apperror.CodeNotFound)
	if len(repo.movements) != 0 {
		t.Fatalf("movements = %d, want 0", len(repo.movements))
	}
}

func TestThresholdUpdateCreatesAuditLog(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newInventoryTestService()

	result, err := service.UpdateThreshold(context.Background(), invTenantA, invStoreA, invProductA, UpdateThresholdInput{
		ActorUserID:       invActorID,
		LowStockThreshold: 3,
	})
	if err != nil {
		t.Fatalf("UpdateThreshold error = %v", err)
	}
	if result.LowStockThreshold != 3 {
		t.Fatalf("threshold = %d, want 3", result.LowStockThreshold)
	}
	if repo.snapshots[invProductA].LowStockThreshold != 3 {
		t.Fatalf("stored threshold = %d, want 3", repo.snapshots[invProductA].LowStockThreshold)
	}
	if len(auditRepo.entries) != 1 || auditRepo.entries[0].Action != AuditActionThresholdUpdated {
		t.Fatalf("audit entries = %#v", auditRepo.entries)
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func newInventoryTestService() (*Service, *fakeInventoryRepository, *fakeAuditRepository, *fakeInventoryOutboxRepository) {
	now := time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
	repo := &fakeInventoryRepository{
		products: map[uuid.UUID]ProductRef{
			invProductA: {
				ID:       invProductA,
				TenantID: invTenantA,
				StoreID:  invStoreA,
				Name:     "Bouquet Mawar",
				SKU:      "BQT-001",
			},
			invProductB: {
				ID:       invProductB,
				TenantID: invTenantB,
				StoreID:  invStoreB,
				Name:     "Produk Tenant B",
				SKU:      "B-001",
			},
		},
		snapshots: map[uuid.UUID]StockSnapshot{
			invProductA: {
				ID:                uuid.MustParse("88888888-8888-8888-8888-888888888888"),
				TenantID:          invTenantA,
				StoreID:           invStoreA,
				ProductID:         invProductA,
				QuantityOnHand:    10,
				QuantityReserved:  2,
				QuantityAvailable: 8,
				LowStockThreshold: 5,
				UpdatedAt:         now,
			},
			invProductB: {
				ID:                uuid.MustParse("99999999-9999-9999-9999-999999999999"),
				TenantID:          invTenantB,
				StoreID:           invStoreB,
				ProductID:         invProductB,
				QuantityOnHand:    5,
				QuantityReserved:  0,
				QuantityAvailable: 5,
				LowStockThreshold: 2,
				UpdatedAt:         now,
			},
		},
		now: now,
	}
	auditRepo := &fakeAuditRepository{}
	outboxRepo := &fakeInventoryOutboxRepository{}
	service := NewService(fakeInventoryDB{}, repo, auditRepo, outboxRepo)
	return service, repo, auditRepo, outboxRepo
}

func assertInventoryAppErrorCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error %s, got %T: %v", code, err, err)
	}
	if appErr.Code != code {
		t.Fatalf("error code = %s, want %s", appErr.Code, code)
	}
}

type fakeInventoryDB struct{}

func (fakeInventoryDB) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(fakeInventoryDB{})
}

func (fakeInventoryDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakeInventoryDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakeInventoryDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakeInventoryRepository struct {
	products  map[uuid.UUID]ProductRef
	snapshots map[uuid.UUID]StockSnapshot
	movements []StockMovement
	now       time.Time
}

func (f *fakeInventoryRepository) ListStocks(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListStockFilters) ([]StockListItem, error) {
	items := make([]StockListItem, 0)
	for productID, snapshot := range f.snapshots {
		product := f.products[productID]
		if snapshot.TenantID != tenantID || snapshot.StoreID != storeID || product.TenantID != tenantID || product.StoreID != storeID {
			continue
		}
		items = append(items, StockListItem{
			ProductID:         product.ID,
			TenantID:          tenantID,
			StoreID:           storeID,
			ProductName:       product.Name,
			SKU:               product.SKU,
			QuantityOnHand:    snapshot.QuantityOnHand,
			QuantityReserved:  snapshot.QuantityReserved,
			QuantityAvailable: snapshot.QuantityAvailable,
			LowStockThreshold: snapshot.LowStockThreshold,
			UpdatedAt:         snapshot.UpdatedAt,
		})
	}
	return items, nil
}

func (f *fakeInventoryRepository) FindProduct(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) (*ProductRef, error) {
	product, ok := f.products[productID]
	if !ok || product.TenantID != tenantID || product.StoreID != storeID {
		return nil, ErrProductNotFound
	}
	return &product, nil
}

func (f *fakeInventoryRepository) LockProduct(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) (*ProductRef, error) {
	return f.FindProduct(ctx, q, tenantID, storeID, productID)
}

func (f *fakeInventoryRepository) ListMovements(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID, filters ListMovementFilters) ([]StockMovement, error) {
	items := make([]StockMovement, 0)
	for _, movement := range f.movements {
		if movement.TenantID == tenantID && movement.StoreID == storeID && movement.ProductID == productID {
			items = append(items, movement)
		}
	}
	return items, nil
}

func (f *fakeInventoryRepository) LockStockSnapshot(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) (*StockSnapshot, error) {
	snapshot, ok := f.snapshots[productID]
	if !ok || snapshot.TenantID != tenantID || snapshot.StoreID != storeID {
		return nil, ErrStockSnapshotNotFound
	}
	return &snapshot, nil
}

func (f *fakeInventoryRepository) UpdateSnapshot(_ context.Context, _ db.Queryer, params UpdateSnapshotParams) (*StockSnapshot, error) {
	snapshot, ok := f.snapshots[params.ProductID]
	if !ok || snapshot.TenantID != params.TenantID || snapshot.StoreID != params.StoreID {
		return nil, ErrStockSnapshotNotFound
	}
	snapshot.QuantityOnHand = params.QuantityOnHand
	snapshot.QuantityReserved = params.QuantityReserved
	snapshot.QuantityAvailable = params.QuantityAvailable
	snapshot.UpdatedAt = f.now.Add(time.Minute)
	f.snapshots[params.ProductID] = snapshot
	return &snapshot, nil
}

func (f *fakeInventoryRepository) UpdateThreshold(_ context.Context, _ db.Queryer, params UpdateThresholdParams) (*StockSnapshot, error) {
	snapshot, ok := f.snapshots[params.ProductID]
	if !ok || snapshot.TenantID != params.TenantID || snapshot.StoreID != params.StoreID {
		return nil, ErrStockSnapshotNotFound
	}
	snapshot.LowStockThreshold = params.LowStockThreshold
	snapshot.UpdatedAt = f.now.Add(time.Minute)
	f.snapshots[params.ProductID] = snapshot
	return &snapshot, nil
}

func (f *fakeInventoryRepository) CreateMovement(_ context.Context, _ db.Queryer, params CreateMovementParams) (*StockMovement, error) {
	movement := StockMovement{
		ID:            uuid.New(),
		TenantID:      params.TenantID,
		StoreID:       params.StoreID,
		ProductID:     params.ProductID,
		MovementType:  params.MovementType,
		Quantity:      params.Quantity,
		BalanceAfter:  params.BalanceAfter,
		ReferenceType: params.ReferenceType,
		ReferenceID:   params.ReferenceID,
		Reason:        params.Reason,
		Note:          params.Note,
		CreatedBy:     params.CreatedBy,
		CreatedAt:     f.now,
	}
	f.movements = append(f.movements, movement)
	return &movement, nil
}

type fakeAuditRepository struct {
	entries []audit.Entry
}

func (f *fakeAuditRepository) Create(_ context.Context, _ db.Queryer, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type fakeInventoryOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakeInventoryOutboxRepository) Insert(_ context.Context, _ db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error) {
	if !json.Valid(params.Payload) {
		return nil, errors.New("invalid json payload")
	}
	f.events = append(f.events, params)
	return &outbox.Event{}, nil
}
