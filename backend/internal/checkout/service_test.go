package checkout

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

var (
	testTenantID  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testStoreID   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testProductID = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testOrderID   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testZoneID    = uuid.MustParse("88888888-8888-8888-8888-888888888888")
)

func TestCheckoutSucceedsWithEnoughStockAndUpdatesSnapshot(t *testing.T) {
	service, repo, _, outboxRepo := newCheckoutTestService(t)

	result, err := service.Checkout(context.Background(), checkoutCommand(t, validCheckoutRequest(testProductID, 2), "idem-success"))
	if err != nil {
		t.Fatalf("checkout returned error: %v", err)
	}

	if result.StatusCode != http.StatusCreated {
		t.Fatalf("status code = %d, want %d", result.StatusCode, http.StatusCreated)
	}
	if result.Response.Totals.GrandTotal != 25000 {
		t.Fatalf("grand total = %d, want 25000", result.Response.Totals.GrandTotal)
	}
	if repo.ordersCreated != 1 {
		t.Fatalf("orders created = %d, want 1", repo.ordersCreated)
	}
	if len(repo.reservations) != 1 {
		t.Fatalf("reservations = %d, want 1", len(repo.reservations))
	}
	if len(repo.movements) != 1 || repo.movements[0].MovementType != "reserved" {
		t.Fatalf("reserved stock movement not created: %#v", repo.movements)
	}
	if len(outboxRepo.events) != 3 {
		t.Fatalf("outbox events = %d, want 3", len(outboxRepo.events))
	}

	snapshot := repo.snapshots[testProductID]
	if snapshot.QuantityReserved != 2 || snapshot.QuantityAvailable != 3 {
		t.Fatalf("snapshot reserved/available = %d/%d, want 2/3", snapshot.QuantityReserved, snapshot.QuantityAvailable)
	}
}

func TestCheckoutFailsWithInsufficientStock(t *testing.T) {
	service, repo, _, _ := newCheckoutTestService(t)
	snapshot := repo.snapshots[testProductID]
	snapshot.QuantityAvailable = 1
	repo.snapshots[testProductID] = snapshot

	_, err := service.Checkout(context.Background(), checkoutCommand(t, validCheckoutRequest(testProductID, 2), "idem-insufficient"))
	assertAppErrorCode(t, err, apperror.CodeInsufficientStock)
}

func TestCheckoutDoubleSubmitSameIdempotencyKeyReplaysOneOrder(t *testing.T) {
	service, repo, _, _ := newCheckoutTestService(t)
	cmd := checkoutCommand(t, validCheckoutRequest(testProductID, 1), "idem-replay")

	first, err := service.Checkout(context.Background(), cmd)
	if err != nil {
		t.Fatalf("first checkout returned error: %v", err)
	}
	second, err := service.Checkout(context.Background(), cmd)
	if err != nil {
		t.Fatalf("second checkout returned error: %v", err)
	}

	if repo.ordersCreated != 1 {
		t.Fatalf("orders created = %d, want 1", repo.ordersCreated)
	}
	if second.Response.OrderID != first.Response.OrderID ||
		second.Response.OrderNumber != first.Response.OrderNumber ||
		second.Response.Totals.GrandTotal != first.Response.Totals.GrandTotal {
		t.Fatalf("replayed response differs: first=%#v second=%#v", first.Response, second.Response)
	}
}

func TestCheckoutSameIdempotencyKeyDifferentPayloadReturnsConflict(t *testing.T) {
	service, repo, _, _ := newCheckoutTestService(t)

	_, err := service.Checkout(context.Background(), checkoutCommand(t, validCheckoutRequest(testProductID, 1), "idem-conflict"))
	if err != nil {
		t.Fatalf("first checkout returned error: %v", err)
	}

	_, err = service.Checkout(context.Background(), checkoutCommand(t, validCheckoutRequest(testProductID, 2), "idem-conflict"))
	assertAppErrorCode(t, err, apperror.CodeIdempotencyConflict)
	if repo.ordersCreated != 1 {
		t.Fatalf("orders created = %d, want 1", repo.ordersCreated)
	}
}

func TestCheckoutRejectsProductFromAnotherStore(t *testing.T) {
	service, repo, _, _ := newCheckoutTestService(t)
	productRecord := repo.products[testProductID]
	productRecord.StoreID = uuid.MustParse("99999999-9999-9999-9999-999999999999")
	repo.products[testProductID] = productRecord

	_, err := service.Checkout(context.Background(), checkoutCommand(t, validCheckoutRequest(testProductID, 1), "idem-other-store"))
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestCheckoutRejectsUnpublishedStore(t *testing.T) {
	service, repo, _, _ := newCheckoutTestService(t)
	service.stores = fakeStoreResolver{err: apperror.NotFound("Store not found")}

	_, err := service.Checkout(context.Background(), checkoutCommand(t, validCheckoutRequest(testProductID, 1), "idem-unpublished"))
	assertAppErrorCode(t, err, apperror.CodeNotFound)
	if repo.ordersCreated != 0 {
		t.Fatalf("orders created = %d, want 0", repo.ordersCreated)
	}
}

func TestCheckoutIgnoresClientPrice(t *testing.T) {
	service, _, _, _ := newCheckoutTestService(t)
	request := validCheckoutRequest(testProductID, 1)
	clientPrice := int64(1)
	request.Items[0].Price = &clientPrice

	result, err := service.Checkout(context.Background(), checkoutCommand(t, request, "idem-client-price"))
	if err != nil {
		t.Fatalf("checkout returned error: %v", err)
	}
	if result.Response.Totals.GrandTotal != 12500 {
		t.Fatalf("grand total = %d, want database price 12500", result.Response.Totals.GrandTotal)
	}
}

func TestCheckoutUsesActiveCourierZoneRateFromDatabase(t *testing.T) {
	service, _, _, _ := newCheckoutTestService(t)
	request := validCheckoutRequest(testProductID, 1)
	request.Shipping = &CheckoutShippingRequest{CourierZoneID: &testZoneID}

	result, err := service.Checkout(context.Background(), checkoutCommand(t, request, "idem-courier-zone"))
	if err != nil {
		t.Fatalf("checkout returned error: %v", err)
	}

	if result.Response.Totals.ShippingCost != 7000 {
		t.Fatalf("shipping cost = %d, want 7000", result.Response.Totals.ShippingCost)
	}
	if result.Response.Totals.GrandTotal != 19500 {
		t.Fatalf("grand total = %d, want 19500", result.Response.Totals.GrandTotal)
	}
}

func TestCheckoutRejectsInactiveOrUnknownCourierZone(t *testing.T) {
	service, repo, _, _ := newCheckoutTestService(t)
	delete(repo.courierZones, testZoneID)
	request := validCheckoutRequest(testProductID, 1)
	request.Shipping = &CheckoutShippingRequest{CourierZoneID: &testZoneID}

	_, err := service.Checkout(context.Background(), checkoutCommand(t, request, "idem-courier-zone-missing"))
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func newCheckoutTestService(t *testing.T) (*Service, *fakeCheckoutRepository, *fakeIdempotencyRepository, *fakeOutboxRepository) {
	t.Helper()

	repo := &fakeCheckoutRepository{
		products: map[uuid.UUID]ProductForCheckout{
			testProductID: {
				ID:             testProductID,
				TenantID:       testTenantID,
				StoreID:        testStoreID,
				Name:           "Kopi Susu",
				SKU:            "KOPI-001",
				Price:          12500,
				Status:         "active",
				TrackInventory: true,
			},
		},
		snapshots: map[uuid.UUID]StockSnapshot{
			testProductID: {
				ID:                uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				TenantID:          testTenantID,
				StoreID:           testStoreID,
				ProductID:         testProductID,
				QuantityOnHand:    5,
				QuantityReserved:  0,
				QuantityAvailable: 5,
			},
		},
		courierZones: map[uuid.UUID]CourierZoneForCheckout{
			testZoneID: {
				ID:       testZoneID,
				TenantID: testTenantID,
				StoreID:  testStoreID,
				Name:     "Dalam Kota",
				Rate:     7000,
			},
		},
		nextOrderID: testOrderID,
	}
	idempotencyRepo := newFakeIdempotencyRepository()
	outboxRepo := &fakeOutboxRepository{}

	service := NewService(
		fakeTxRunner{},
		fakeStoreResolver{context: store.PublicContext{
			TenantID: testTenantID,
			StoreID:  testStoreID,
			Store: store.Store{
				ID:       testStoreID,
				TenantID: testTenantID,
				Slug:     "warung-kopi",
				Status:   store.StatusPublished,
			},
		}},
		repo,
		idempotencyRepo,
		outboxRepo,
	)
	service.now = func() time.Time {
		return time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	}
	service.newUUID = func() uuid.UUID {
		return uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	}

	return service, repo, idempotencyRepo, outboxRepo
}

func validCheckoutRequest(productID uuid.UUID, quantity int) CheckoutRequest {
	return CheckoutRequest{
		Items: []CheckoutItemRequest{
			{ProductID: productID, Quantity: quantity},
		},
		Customer: CheckoutCustomerRequest{
			Name:  "Budi",
			Phone: "081234567890",
			Email: "budi@example.test",
		},
		ShippingAddress: CheckoutAddressRequest{
			Address:    "Jl. Mawar No. 1",
			City:       "Makassar",
			Province:   "Sulawesi Selatan",
			PostalCode: "90111",
		},
		PaymentMethod: PaymentMethodManualTransfer,
	}
}

func checkoutCommand(t *testing.T, request CheckoutRequest, key string) Command {
	t.Helper()
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return Command{
		StoreSlug:      "warung-kopi",
		IdempotencyKey: key,
		Method:         http.MethodPost,
		Path:           "/api/v1/public/stores/warung-kopi/checkout",
		RawBody:        body,
		Request:        request,
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

type fakeTxRunner struct{}

func (fakeTxRunner) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(nil)
}

type fakeStoreResolver struct {
	context store.PublicContext
	err     error
}

func (f fakeStoreResolver) Resolve(context.Context, string) (store.PublicContext, error) {
	if f.err != nil {
		return store.PublicContext{}, f.err
	}
	return f.context, nil
}

type fakeIdempotencyRepository struct {
	records map[string]fakeIdempotencyRecord
}

type fakeIdempotencyRecord struct {
	requestHash string
	response    json.RawMessage
	statusCode  int
}

func newFakeIdempotencyRepository() *fakeIdempotencyRepository {
	return &fakeIdempotencyRepository{
		records: make(map[string]fakeIdempotencyRecord),
	}
}

func (f *fakeIdempotencyRepository) Begin(
	_ context.Context,
	_ db.Queryer,
	tenantID uuid.UUID,
	scope string,
	key string,
	requestHash string,
	_ time.Time,
) (*idempotency.State, error) {
	recordKey := tenantID.String() + "|" + scope + "|" + key
	record, ok := f.records[recordKey]
	if !ok {
		f.records[recordKey] = fakeIdempotencyRecord{requestHash: requestHash}
		return &idempotency.State{Created: true, IsProcessing: true}, nil
	}
	if record.requestHash != requestHash {
		return nil, apperror.IdempotencyConflict("Idempotency key was already used with a different request")
	}
	if len(record.response) > 0 {
		return &idempotency.State{
			CanReplay:    true,
			ResponseBody: record.response,
			StatusCode:   record.statusCode,
		}, nil
	}
	return &idempotency.State{IsProcessing: true}, nil
}

func (f *fakeIdempotencyRepository) SaveCompletedResponse(
	_ context.Context,
	_ db.Queryer,
	tenantID uuid.UUID,
	scope string,
	key string,
	statusCode int,
	responseBody json.RawMessage,
) error {
	recordKey := tenantID.String() + "|" + scope + "|" + key
	record := f.records[recordKey]
	record.response = append(json.RawMessage(nil), responseBody...)
	record.statusCode = statusCode
	f.records[recordKey] = record
	return nil
}

type fakeCheckoutRepository struct {
	products      map[uuid.UUID]ProductForCheckout
	snapshots     map[uuid.UUID]StockSnapshot
	courierZones  map[uuid.UUID]CourierZoneForCheckout
	nextOrderID   uuid.UUID
	ordersCreated int
	reservations  []CreateReservationParams
	movements     []CreateStockMovementParams
}

func (f *fakeCheckoutRepository) ListProductsForCheckout(
	_ context.Context,
	_ db.Queryer,
	_ uuid.UUID,
	_ uuid.UUID,
	productIDs []uuid.UUID,
) ([]ProductForCheckout, error) {
	result := make([]ProductForCheckout, 0, len(productIDs))
	for _, productID := range productIDs {
		if item, ok := f.products[productID]; ok {
			result = append(result, item)
		}
	}
	return result, nil
}

func (f *fakeCheckoutRepository) LockStockSnapshots(
	_ context.Context,
	_ db.Queryer,
	_ uuid.UUID,
	_ uuid.UUID,
	productIDs []uuid.UUID,
) ([]StockSnapshot, error) {
	result := make([]StockSnapshot, 0, len(productIDs))
	for _, productID := range productIDs {
		if item, ok := f.snapshots[productID]; ok {
			result = append(result, item)
		}
	}
	return result, nil
}

func (f *fakeCheckoutRepository) FindActiveCourierZone(
	_ context.Context,
	_ db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	zoneID uuid.UUID,
) (*CourierZoneForCheckout, error) {
	item, ok := f.courierZones[zoneID]
	if !ok || item.TenantID != tenantID || item.StoreID != storeID {
		return nil, ErrCourierZoneNotFound
	}
	return &item, nil
}

func (f *fakeCheckoutRepository) FindOrCreateCustomer(
	context.Context,
	db.Queryer,
	FindOrCreateCustomerParams,
) (*CustomerRecord, error) {
	return &CustomerRecord{
		ID:       uuid.MustParse("66666666-6666-6666-6666-666666666666"),
		TenantID: testTenantID,
		StoreID:  testStoreID,
		Name:     "Budi",
		Phone:    "081234567890",
	}, nil
}

func (f *fakeCheckoutRepository) CreateCustomerAddress(
	context.Context,
	db.Queryer,
	CreateAddressParams,
) (*AddressRecord, error) {
	return &AddressRecord{
		ID:         uuid.MustParse("77777777-7777-7777-7777-777777777777"),
		TenantID:   testTenantID,
		CustomerID: uuid.MustParse("66666666-6666-6666-6666-666666666666"),
	}, nil
}

func (f *fakeCheckoutRepository) CreateOrder(_ context.Context, _ db.Queryer, params CreateOrderParams) (*OrderRecord, error) {
	f.ordersCreated++
	return &OrderRecord{
		ID:            f.nextOrderID,
		TenantID:      params.TenantID,
		StoreID:       params.StoreID,
		OrderNumber:   params.OrderNumber,
		Status:        params.Status,
		PaymentStatus: params.PaymentStatus,
		Subtotal:      params.Subtotal,
		DiscountTotal: params.DiscountTotal,
		ShippingCost:  params.ShippingCost,
		TaxTotal:      params.TaxTotal,
		GrandTotal:    params.GrandTotal,
		CreatedAt:     time.Now(),
	}, nil
}

func (f *fakeCheckoutRepository) CreateOrderItem(context.Context, db.Queryer, CreateOrderItemParams) error {
	return nil
}

func (f *fakeCheckoutRepository) CreateStockReservation(_ context.Context, _ db.Queryer, params CreateReservationParams) error {
	f.reservations = append(f.reservations, params)
	return nil
}

func (f *fakeCheckoutRepository) UpdateStockSnapshot(_ context.Context, _ db.Queryer, params UpdateSnapshotParams) error {
	snapshot := f.snapshots[params.ProductID]
	snapshot.QuantityReserved = params.QuantityReserved
	snapshot.QuantityAvailable = params.QuantityAvailable
	f.snapshots[params.ProductID] = snapshot
	return nil
}

func (f *fakeCheckoutRepository) CreateStockMovement(_ context.Context, _ db.Queryer, params CreateStockMovementParams) error {
	f.movements = append(f.movements, params)
	return nil
}

func (f *fakeCheckoutRepository) CreateOrderStatusLog(context.Context, db.Queryer, uuid.UUID, uuid.UUID, string, string) error {
	return nil
}

func (f *fakeCheckoutRepository) UpdateCustomerStats(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID, int64) error {
	return nil
}

type fakeOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakeOutboxRepository) Insert(
	_ context.Context,
	_ db.Queryer,
	params outbox.InsertEventParams,
) (*outbox.Event, error) {
	f.events = append(f.events, params)
	return &outbox.Event{}, nil
}
