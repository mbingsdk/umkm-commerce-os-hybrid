package order

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
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
)

var (
	tenantA  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	storeA   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	tenantB  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	storeB   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	orderA   = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	orderB   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	actorID  = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	productA = uuid.MustParse("88888888-8888-8888-8888-888888888888")
	productB = uuid.MustParse("99999999-9999-9999-9999-999999999999")
	reserveA = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
)

func TestDetailRejectsOtherTenantOrder(t *testing.T) {
	service, _, _ := newOrderTestService()

	_, err := service.Detail(context.Background(), tenantA, storeA, orderB)
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestListOrdersIsTenantScoped(t *testing.T) {
	service, _, _ := newOrderTestService()

	result, _, err := service.List(context.Background(), tenantA, storeA, ListFilters{})
	if err != nil {
		t.Fatalf("List error = %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("orders len = %d, want 1", len(result))
	}
	if result[0].ID != orderA {
		t.Fatalf("order id = %s, want %s", result[0].ID, orderA)
	}
}

func TestInvalidStatusTransitionRejected(t *testing.T) {
	service, repo, outboxRepo := newOrderTestService()

	_, err := service.UpdateStatus(context.Background(), tenantA, storeA, orderA, UpdateStatusInput{
		ActorUserID: actorID,
		Status:      StatusShipped,
	})
	assertAppErrorCode(t, err, apperror.CodeInvalidOrderStatus)
	if len(repo.logs) != 0 {
		t.Fatalf("status logs = %d, want 0", len(repo.logs))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestValidStatusTransitionCreatesStatusLogAndOutboxEvent(t *testing.T) {
	service, repo, outboxRepo := newOrderTestService()

	result, err := service.UpdateStatus(context.Background(), tenantA, storeA, orderA, UpdateStatusInput{
		ActorUserID: actorID,
		Status:      StatusConfirmed,
		Note:        "Pesanan dikonfirmasi",
	})
	if err != nil {
		t.Fatalf("UpdateStatus error = %v", err)
	}
	if result.Status != StatusConfirmed {
		t.Fatalf("status = %s, want %s", result.Status, StatusConfirmed)
	}
	if len(repo.logs) != 1 {
		t.Fatalf("status logs = %d, want 1", len(repo.logs))
	}
	if repo.logs[0].FromStatus != StatusPending || repo.logs[0].ToStatus != StatusConfirmed {
		t.Fatalf("status log = %#v", repo.logs[0])
	}
	if len(outboxRepo.events) != 1 {
		t.Fatalf("outbox events = %d, want 1", len(outboxRepo.events))
	}
	if outboxRepo.events[0].EventType != EventOrderStatusUpdated {
		t.Fatalf("event type = %s, want %s", outboxRepo.events[0].EventType, EventOrderStatusUpdated)
	}
	if outboxRepo.events[0].TenantID != tenantA || outboxRepo.events[0].AggregateID != orderA {
		t.Fatalf("outbox event scope = %#v", outboxRepo.events[0])
	}
}

func TestCancelPendingOrderReleasesReservedStock(t *testing.T) {
	service, repo, outboxRepo := newOrderTestService()

	result, err := service.Cancel(context.Background(), tenantA, storeA, orderA, CancelInput{
		ActorUserID: actorID,
		Reason:      "Customer requested cancellation",
		Note:        "Belum dikirim",
	})
	if err != nil {
		t.Fatalf("Cancel error = %v", err)
	}

	if result.Status != StatusCancelled {
		t.Fatalf("status = %s, want %s", result.Status, StatusCancelled)
	}
	if result.ReleasedReservations != 1 || result.ReleasedQuantity != 2 {
		t.Fatalf("release result = %#v, want 1 reservation and quantity 2", result)
	}

	snapshot := repo.snapshots[productA]
	if snapshot.QuantityReserved != 0 || snapshot.QuantityAvailable != 10 {
		t.Fatalf("snapshot = %#v, want reserved 0 available 10", snapshot)
	}
	if repo.reservations[reserveA].Status != ReservationStatusReleased {
		t.Fatalf("reservation status = %s, want released", repo.reservations[reserveA].Status)
	}
	if len(repo.movements) != 1 {
		t.Fatalf("movements = %d, want 1", len(repo.movements))
	}
	if repo.movements[0].MovementType != "released" || repo.movements[0].Quantity != 2 || repo.movements[0].BalanceAfter != 10 {
		t.Fatalf("movement = %#v", repo.movements[0])
	}
	if len(repo.logs) != 1 || repo.logs[0].FromStatus != StatusPending || repo.logs[0].ToStatus != StatusCancelled {
		t.Fatalf("status logs = %#v", repo.logs)
	}
	if len(outboxRepo.events) != 3 {
		t.Fatalf("outbox events = %d, want 3", len(outboxRepo.events))
	}
	eventTypes := []string{outboxRepo.events[0].EventType, outboxRepo.events[1].EventType, outboxRepo.events[2].EventType}
	if eventTypes[0] != EventOrderCancelled || eventTypes[1] != EventStockReservationReleased || eventTypes[2] != EventNotificationRequested {
		t.Fatalf("event types = %#v", eventTypes)
	}
}

func TestCancelPaidOrderReleasesReservedStock(t *testing.T) {
	service, repo, _ := newOrderTestService()
	orderRecord := repo.orders[orderA]
	orderRecord.Status = StatusConfirmed
	orderRecord.PaymentStatus = PaymentStatusPaid
	repo.orders[orderA] = orderRecord

	result, err := service.Cancel(context.Background(), tenantA, storeA, orderA, CancelInput{
		ActorUserID: actorID,
		Reason:      "Pesanan dibatalkan sebelum dikirim",
	})
	if err != nil {
		t.Fatalf("Cancel error = %v", err)
	}

	if result.Status != StatusCancelled || result.ReleasedQuantity != 2 {
		t.Fatalf("cancel result = %#v", result)
	}
	if repo.snapshots[productA].QuantityReserved != 0 || repo.snapshots[productA].QuantityAvailable != 10 {
		t.Fatalf("snapshot = %#v", repo.snapshots[productA])
	}
}

func TestCancelShippedOrCompletedRejected(t *testing.T) {
	tests := []string{StatusShipped, StatusCompleted}

	for _, status := range tests {
		status := status
		t.Run(status, func(t *testing.T) {
			service, repo, outboxRepo := newOrderTestService()
			orderRecord := repo.orders[orderA]
			orderRecord.Status = status
			repo.orders[orderA] = orderRecord

			_, err := service.Cancel(context.Background(), tenantA, storeA, orderA, CancelInput{
				ActorUserID: actorID,
				Reason:      "Tidak bisa dibatalkan",
			})
			assertAppErrorCode(t, err, apperror.CodeInvalidOrderStatus)
			if len(repo.movements) != 0 {
				t.Fatalf("movements = %d, want 0", len(repo.movements))
			}
			if len(outboxRepo.events) != 0 {
				t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
			}
		})
	}
}

func TestDoubleCancelDoesNotDoubleReleaseStock(t *testing.T) {
	service, repo, outboxRepo := newOrderTestService()

	first, err := service.Cancel(context.Background(), tenantA, storeA, orderA, CancelInput{
		ActorUserID: actorID,
		Reason:      "Customer requested cancellation",
	})
	if err != nil {
		t.Fatalf("first Cancel error = %v", err)
	}
	second, err := service.Cancel(context.Background(), tenantA, storeA, orderA, CancelInput{
		ActorUserID: actorID,
		Reason:      "Customer requested cancellation",
	})
	if err != nil {
		t.Fatalf("second Cancel error = %v", err)
	}

	if first.ReleasedQuantity != 2 || second.ReleasedQuantity != 0 {
		t.Fatalf("release quantities first=%d second=%d, want 2 then 0", first.ReleasedQuantity, second.ReleasedQuantity)
	}
	if len(repo.movements) != 1 {
		t.Fatalf("movements = %d, want 1", len(repo.movements))
	}
	if repo.snapshots[productA].QuantityReserved != 0 || repo.snapshots[productA].QuantityAvailable != 10 {
		t.Fatalf("snapshot = %#v", repo.snapshots[productA])
	}
	if len(outboxRepo.events) != 3 {
		t.Fatalf("outbox events = %d, want only first cancel events", len(outboxRepo.events))
	}
}

func TestCancelRejectsOtherTenantOrder(t *testing.T) {
	service, repo, outboxRepo := newOrderTestService()

	_, err := service.Cancel(context.Background(), tenantA, storeA, orderB, CancelInput{
		ActorUserID: actorID,
		Reason:      "Wrong tenant",
	})
	assertAppErrorCode(t, err, apperror.CodeNotFound)
	if len(repo.movements) != 0 {
		t.Fatalf("movements = %d, want 0", len(repo.movements))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func newOrderTestService() (*Service, *fakeOrderRepository, *fakeOutboxRepository) {
	now := time.Date(2026, 5, 20, 10, 0, 0, 0, time.UTC)
	repo := &fakeOrderRepository{
		orders: map[uuid.UUID]Order{
			orderA: testOrder(orderA, tenantA, storeA, StatusPending, now),
			orderB: testOrder(orderB, tenantB, storeB, StatusPending, now.Add(-time.Hour)),
		},
		reservations: map[uuid.UUID]StockReservation{
			reserveA: {
				ID:        reserveA,
				TenantID:  tenantA,
				StoreID:   storeA,
				ProductID: productA,
				OrderID:   orderA,
				Quantity:  2,
				Status:    ReservationStatusActive,
				CreatedAt: now,
			},
		},
		snapshots: map[uuid.UUID]StockSnapshot{
			productA: {
				ID:                uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
				TenantID:          tenantA,
				StoreID:           storeA,
				ProductID:         productA,
				QuantityOnHand:    10,
				QuantityReserved:  2,
				QuantityAvailable: 8,
			},
			productB: {
				ID:                uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
				TenantID:          tenantB,
				StoreID:           storeB,
				ProductID:         productB,
				QuantityOnHand:    5,
				QuantityReserved:  0,
				QuantityAvailable: 5,
			},
		},
	}
	outboxRepo := &fakeOutboxRepository{}
	service := NewService(fakeOrderDB{}, repo, outboxRepo)
	return service, repo, outboxRepo
}

func testOrder(id uuid.UUID, tenantID uuid.UUID, storeID uuid.UUID, status string, createdAt time.Time) Order {
	return Order{
		ID:            id,
		TenantID:      tenantID,
		StoreID:       storeID,
		OrderNumber:   "ORD-20260520-" + id.String()[:8],
		Source:        SourceStorefront,
		Status:        status,
		PaymentStatus: PaymentStatusUnpaid,
		GrandTotal:    12500,
		CustomerName:  "Budi",
		CustomerPhone: "081234567890",
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}
}

func assertAppErrorCode(t *testing.T, err error, code apperror.Code) {
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

type fakeOrderDB struct{}

func (fakeOrderDB) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(fakeOrderDB{})
}

func (fakeOrderDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakeOrderDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakeOrderDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakeOrderRepository struct {
	orders       map[uuid.UUID]Order
	reservations map[uuid.UUID]StockReservation
	snapshots    map[uuid.UUID]StockSnapshot
	logs         []CreateStatusLogParams
	movements    []CreateStockMovementParams
}

func (f *fakeOrderRepository) List(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]Order, []int, error) {
	items := make([]Order, 0)
	counts := make([]int, 0)
	for _, item := range f.orders {
		if item.TenantID != tenantID || item.StoreID != storeID {
			continue
		}
		if filters.Status != nil && item.Status != *filters.Status {
			continue
		}
		items = append(items, item)
		counts = append(counts, 1)
	}
	return items, counts, nil
}

func (f *fakeOrderRepository) FindByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*Order, error) {
	item, ok := f.orders[orderID]
	if !ok || item.TenantID != tenantID || item.StoreID != storeID {
		return nil, ErrOrderNotFound
	}
	return &item, nil
}

func (f *fakeOrderRepository) LockByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*Order, error) {
	return f.FindByID(ctx, q, tenantID, storeID, orderID)
}

func (f *fakeOrderRepository) ListItems(context.Context, db.Queryer, uuid.UUID, uuid.UUID) ([]Item, error) {
	return nil, nil
}

func (f *fakeOrderRepository) ListStatusLogs(context.Context, db.Queryer, uuid.UUID, uuid.UUID) ([]StatusLog, error) {
	return nil, nil
}

func (f *fakeOrderRepository) ListReservationSummary(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) ([]ReservationSummary, error) {
	return nil, nil
}

func (f *fakeOrderRepository) LockActiveReservationsByOrder(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) ([]StockReservation, error) {
	items := make([]StockReservation, 0)
	for _, item := range f.reservations {
		if item.TenantID != tenantID || item.StoreID != storeID || item.OrderID != orderID {
			continue
		}
		if item.Status != ReservationStatusActive && item.Status != ReservationStatusConfirmed {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (f *fakeOrderRepository) LockStockSnapshots(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productIDs []uuid.UUID) ([]StockSnapshot, error) {
	items := make([]StockSnapshot, 0, len(productIDs))
	for _, productID := range productIDs {
		item, ok := f.snapshots[productID]
		if !ok || item.TenantID != tenantID || item.StoreID != storeID {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (f *fakeOrderRepository) UpdateStatus(_ context.Context, _ db.Queryer, params UpdateStatusParams) (*Order, error) {
	item, ok := f.orders[params.OrderID]
	if !ok || item.TenantID != params.TenantID || item.StoreID != params.StoreID {
		return nil, ErrOrderNotFound
	}
	item.Status = params.Status
	item.UpdatedAt = item.UpdatedAt.Add(time.Minute)
	f.orders[params.OrderID] = item
	return &item, nil
}

func (f *fakeOrderRepository) UpdateStockSnapshot(_ context.Context, _ db.Queryer, params UpdateStockSnapshotParams) error {
	item, ok := f.snapshots[params.ProductID]
	if !ok || item.TenantID != params.TenantID || item.StoreID != params.StoreID {
		return ErrStockSnapshotNotFound
	}
	item.QuantityReserved = params.QuantityReserved
	item.QuantityAvailable = params.QuantityAvailable
	f.snapshots[params.ProductID] = item
	return nil
}

func (f *fakeOrderRepository) ReleaseReservations(_ context.Context, _ db.Queryer, params ReleaseReservationsParams) error {
	for _, reservationID := range params.ReservationIDs {
		item, ok := f.reservations[reservationID]
		if !ok || item.TenantID != params.TenantID || item.StoreID != params.StoreID {
			continue
		}
		if item.Status != ReservationStatusActive && item.Status != ReservationStatusConfirmed {
			continue
		}
		item.Status = params.Status
		f.reservations[reservationID] = item
	}
	return nil
}

func (f *fakeOrderRepository) CreateStockMovement(_ context.Context, _ db.Queryer, params CreateStockMovementParams) error {
	f.movements = append(f.movements, params)
	return nil
}

func (f *fakeOrderRepository) CreateStatusLog(_ context.Context, _ db.Queryer, params CreateStatusLogParams) (*StatusLog, error) {
	f.logs = append(f.logs, params)
	return &StatusLog{
		ID:         uuid.New(),
		TenantID:   params.TenantID,
		OrderID:    params.OrderID,
		FromStatus: params.FromStatus,
		ToStatus:   params.ToStatus,
		Note:       params.Note,
		CreatedBy:  &params.CreatedBy,
		CreatedAt:  time.Now(),
	}, nil
}

type fakeOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakeOutboxRepository) Insert(_ context.Context, _ db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error) {
	if !json.Valid(params.Payload) {
		return nil, errors.New("invalid json payload")
	}
	f.events = append(f.events, params)
	return &outbox.Event{}, nil
}
