package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	orderpkg "github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

var (
	testTenantA        = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	testStoreA         = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	testTenantB        = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	testStoreB         = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	testOrderA         = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	testOrderB         = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	testConfirmationID = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	testActorID        = uuid.MustParse("88888888-8888-8888-8888-888888888888")
)

func TestPublicConfirmationRejectsUnpublishedOrWrongStore(t *testing.T) {
	service, repo, _, _ := newPaymentTestService()
	service.stores = fakeStoreResolver{err: apperror.NotFound("Store not found")}

	_, _, err := service.PublicConfirm(context.Background(), publicConfirmationCommand(t, validPublicConfirmationRequest(), "idem-store"))
	assertAppErrorCode(t, err, apperror.CodeNotFound)
	if len(repo.confirmations) != 1 {
		t.Fatalf("confirmations = %d, want unchanged 1", len(repo.confirmations))
	}
}

func TestPublicConfirmationRejectsWrongPhone(t *testing.T) {
	service, repo, _, _ := newPaymentTestService()
	request := validPublicConfirmationRequest()
	request.CustomerPhone = "089999999999"

	_, _, err := service.PublicConfirm(context.Background(), publicConfirmationCommand(t, request, "idem-phone"))
	assertAppErrorCode(t, err, apperror.CodeNotFound)
	if len(repo.confirmations) != 1 {
		t.Fatalf("confirmations = %d, want unchanged 1", len(repo.confirmations))
	}
}

func TestTenantCannotConfirmOtherTenantOrder(t *testing.T) {
	service, _, _, _ := newPaymentTestService()

	_, err := service.ConfirmPayment(context.Background(), testTenantA, testStoreA, testOrderB, ConfirmInput{
		ActorUserID: testActorID,
	})
	assertAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestConfirmPaymentUpdatesPaymentStatusOrderStatusLogsAndOutbox(t *testing.T) {
	service, repo, _, outboxRepo := newPaymentTestService()

	result, err := service.ConfirmPayment(context.Background(), testTenantA, testStoreA, testOrderA, ConfirmInput{
		ActorUserID:    testActorID,
		ConfirmationID: &testConfirmationID,
		Note:           "Transfer sudah dicek",
	})
	if err != nil {
		t.Fatalf("ConfirmPayment error = %v", err)
	}

	orderRecord := repo.orders[testOrderA]
	if result.PaymentStatus != orderpkg.PaymentStatusPaid || orderRecord.PaymentStatus != orderpkg.PaymentStatusPaid {
		t.Fatalf("payment status = response %s order %s, want paid", result.PaymentStatus, orderRecord.PaymentStatus)
	}
	if result.OrderStatus != orderpkg.StatusConfirmed || orderRecord.Status != orderpkg.StatusConfirmed {
		t.Fatalf("order status = response %s order %s, want confirmed", result.OrderStatus, orderRecord.Status)
	}
	if len(repo.payments) != 1 {
		t.Fatalf("payments = %d, want 1", len(repo.payments))
	}
	if len(repo.statusLogs) != 1 {
		t.Fatalf("status logs = %d, want 1", len(repo.statusLogs))
	}
	if repo.statusLogs[0].FromStatus != orderpkg.StatusPending || repo.statusLogs[0].ToStatus != orderpkg.StatusConfirmed {
		t.Fatalf("status log = %#v", repo.statusLogs[0])
	}
	if len(outboxRepo.events) != 3 {
		t.Fatalf("outbox events = %d, want 3", len(outboxRepo.events))
	}
	if outboxRepo.events[0].EventType != EventPaymentConfirmed || outboxRepo.events[1].EventType != EventOrderPaid || outboxRepo.events[2].EventType != EventNotificationRequested {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestRejectPaymentDoesNotMarkOrderPaid(t *testing.T) {
	service, repo, _, outboxRepo := newPaymentTestService()

	result, err := service.RejectPayment(context.Background(), testTenantA, testStoreA, testOrderA, RejectInput{
		ActorUserID:    testActorID,
		ConfirmationID: &testConfirmationID,
		Note:           "Nominal tidak cocok",
	})
	if err != nil {
		t.Fatalf("RejectPayment error = %v", err)
	}

	orderRecord := repo.orders[testOrderA]
	if orderRecord.PaymentStatus == orderpkg.PaymentStatusPaid {
		t.Fatal("order payment status = paid, want not paid")
	}
	if result.Status != ConfirmationStatusRejected {
		t.Fatalf("confirmation status = %s, want rejected", result.Status)
	}
	if len(repo.payments) != 0 {
		t.Fatalf("payments = %d, want 0", len(repo.payments))
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventNotificationRequested {
		t.Fatalf("outbox events = %#v, want one notification", outboxRepo.events)
	}
}

func newPaymentTestService() (*Service, *fakePaymentRepository, *fakeIdempotencyRepository, *fakeOutboxRepository) {
	now := time.Date(2026, 5, 20, 12, 0, 0, 0, time.UTC)
	repo := &fakePaymentRepository{
		orders: map[uuid.UUID]orderpkg.Order{
			testOrderA: testOrder(testOrderA, testTenantA, testStoreA, orderpkg.StatusPending, orderpkg.PaymentStatusWaitingConfirmation, now),
			testOrderB: testOrder(testOrderB, testTenantB, testStoreB, orderpkg.StatusPending, orderpkg.PaymentStatusWaitingConfirmation, now),
		},
		confirmations: map[uuid.UUID]Confirmation{
			testConfirmationID: testConfirmation(testConfirmationID, testTenantA, testStoreA, testOrderA, now),
		},
	}
	idempotencyRepo := newFakeIdempotencyRepository()
	outboxRepo := &fakeOutboxRepository{}
	service := NewService(fakePaymentDB{}, fakeStoreResolver{context: store.PublicContext{TenantID: testTenantA, StoreID: testStoreA, Store: store.Store{ID: testStoreA, TenantID: testTenantA, Slug: "toko-bunga", Status: store.StatusPublished}}}, repo, idempotencyRepo, outboxRepo)
	service.now = func() time.Time { return now }
	return service, repo, idempotencyRepo, outboxRepo
}

func testOrder(id uuid.UUID, tenantID uuid.UUID, storeID uuid.UUID, status string, paymentStatus string, createdAt time.Time) orderpkg.Order {
	return orderpkg.Order{
		ID:            id,
		TenantID:      tenantID,
		StoreID:       storeID,
		OrderNumber:   "ORD-20260520-000001",
		Source:        orderpkg.SourceStorefront,
		Status:        status,
		PaymentStatus: paymentStatus,
		GrandTotal:    100000,
		CustomerName:  "Budi",
		CustomerPhone: "081234567890",
		CreatedAt:     createdAt,
		UpdatedAt:     createdAt,
	}
}

func testConfirmation(id uuid.UUID, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, createdAt time.Time) Confirmation {
	return Confirmation{
		ID:             id,
		TenantID:       tenantID,
		StoreID:        storeID,
		OrderID:        orderID,
		PayerName:      "Budi",
		BankName:       "BCA",
		TransferAmount: 100000,
		TransferDate:   createdAt,
		Status:         ConfirmationStatusPending,
		CreatedAt:      createdAt,
		UpdatedAt:      createdAt,
	}
}

func validPublicConfirmationRequest() PublicConfirmationRequest {
	return PublicConfirmationRequest{
		CustomerPhone:  "081234567890",
		PayerName:      "Budi",
		BankName:       "BCA",
		TransferAmount: 100000,
		TransferDate:   "2026-05-20",
		ProofURL:       "/uploads/proof.jpg",
		Note:           "Sudah transfer",
	}
}

func publicConfirmationCommand(t *testing.T, request PublicConfirmationRequest, key string) PublicConfirmationCommand {
	t.Helper()
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	return PublicConfirmationCommand{
		StoreSlug:      "toko-bunga",
		OrderNumber:    "ORD-20260520-000001",
		IdempotencyKey: key,
		Method:         http.MethodPost,
		Path:           "/api/v1/public/stores/toko-bunga/orders/ORD-20260520-000001/payment-confirmation",
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

type fakePaymentDB struct{}

func (fakePaymentDB) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(fakePaymentDB{})
}

func (fakePaymentDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakePaymentDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakePaymentDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
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
	return &fakeIdempotencyRepository{records: make(map[string]fakeIdempotencyRecord)}
}

func (f *fakeIdempotencyRepository) Begin(_ context.Context, _ db.Queryer, tenantID uuid.UUID, scope string, key string, requestHash string, _ time.Time) (*idempotency.State, error) {
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
		return &idempotency.State{CanReplay: true, ResponseBody: record.response, StatusCode: record.statusCode}, nil
	}
	return &idempotency.State{IsProcessing: true}, nil
}

func (f *fakeIdempotencyRepository) SaveCompletedResponse(_ context.Context, _ db.Queryer, tenantID uuid.UUID, scope string, key string, statusCode int, responseBody json.RawMessage) error {
	recordKey := tenantID.String() + "|" + scope + "|" + key
	record := f.records[recordKey]
	record.response = append(json.RawMessage(nil), responseBody...)
	record.statusCode = statusCode
	f.records[recordKey] = record
	return nil
}

type fakePaymentRepository struct {
	orders        map[uuid.UUID]orderpkg.Order
	confirmations map[uuid.UUID]Confirmation
	payments      []Payment
	statusLogs    []struct {
		FromStatus string
		ToStatus   string
	}
}

func (f *fakePaymentRepository) FindOrderByPublicReference(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderNumber string, customerPhone string) (*orderpkg.Order, error) {
	for _, item := range f.orders {
		if item.TenantID == tenantID && item.StoreID == storeID && item.OrderNumber == orderNumber && item.CustomerPhone == customerPhone {
			copy := item
			return &copy, nil
		}
	}
	return nil, ErrOrderNotFound
}

func (f *fakePaymentRepository) LockOrderByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*orderpkg.Order, error) {
	item, ok := f.orders[orderID]
	if !ok || item.TenantID != tenantID || item.StoreID != storeID {
		return nil, ErrOrderNotFound
	}
	copy := item
	return &copy, nil
}

func (f *fakePaymentRepository) CreateConfirmation(_ context.Context, _ db.Queryer, params CreateConfirmationParams) (*Confirmation, error) {
	id := uuid.New()
	item := Confirmation{ID: id, TenantID: params.TenantID, StoreID: params.StoreID, OrderID: params.OrderID, PayerName: params.PayerName, BankName: params.BankName, TransferAmount: params.TransferAmount, TransferDate: params.TransferDate, ProofURL: params.ProofURL, Note: params.Note, Status: ConfirmationStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	f.confirmations[id] = item
	return &item, nil
}

func (f *fakePaymentRepository) ListConfirmations(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) ([]Confirmation, error) {
	return nil, nil
}

func (f *fakePaymentRepository) FindPendingConfirmation(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, confirmationID *uuid.UUID) (*Confirmation, error) {
	for _, item := range f.confirmations {
		if item.TenantID != tenantID || item.StoreID != storeID || item.OrderID != orderID || item.Status != ConfirmationStatusPending {
			continue
		}
		if confirmationID != nil && item.ID != *confirmationID {
			continue
		}
		copy := item
		return &copy, nil
	}
	return nil, ErrConfirmationNotFound
}

func (f *fakePaymentRepository) MarkConfirmationReviewed(_ context.Context, _ db.Queryer, params ReviewConfirmationParams) (*Confirmation, error) {
	item, ok := f.confirmations[params.ConfirmationID]
	if !ok || item.TenantID != params.TenantID || item.StoreID != params.StoreID || item.OrderID != params.OrderID || item.Status != ConfirmationStatusPending {
		return nil, ErrConfirmationNotFound
	}
	item.Status = params.Status
	item.ReviewedBy = &params.ReviewedBy
	now := time.Now()
	item.ReviewedAt = &now
	item.ReviewNote = params.ReviewNote
	f.confirmations[item.ID] = item
	return &item, nil
}

func (f *fakePaymentRepository) UpdateOrderPaymentWaiting(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) error {
	item := f.orders[orderID]
	if item.TenantID == tenantID && item.StoreID == storeID && item.PaymentStatus == orderpkg.PaymentStatusUnpaid {
		item.PaymentStatus = orderpkg.PaymentStatusWaitingConfirmation
		f.orders[orderID] = item
	}
	return nil
}

func (f *fakePaymentRepository) UpdateOrderPaid(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*orderpkg.Order, error) {
	item, ok := f.orders[orderID]
	if !ok || item.TenantID != tenantID || item.StoreID != storeID {
		return nil, ErrOrderNotFound
	}
	item.PaymentStatus = orderpkg.PaymentStatusPaid
	if item.Status == orderpkg.StatusPending {
		item.Status = orderpkg.StatusConfirmed
	}
	f.orders[orderID] = item
	return &item, nil
}

func (f *fakePaymentRepository) CreatePayment(_ context.Context, _ db.Queryer, params CreatePaymentParams) (*Payment, error) {
	item := Payment{ID: uuid.New(), TenantID: params.TenantID, StoreID: params.StoreID, OrderID: params.OrderID, PaymentConfirmationID: &params.PaymentConfirmationID, Method: params.Method, Status: params.Status, Amount: params.Amount, PayerName: params.PayerName, BankName: params.BankName, ProofURL: params.ProofURL, Note: params.Note, ConfirmedBy: &params.ConfirmedBy, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	f.payments = append(f.payments, item)
	return &item, nil
}

func (f *fakePaymentRepository) CreateOrderStatusLog(_ context.Context, _ db.Queryer, _ uuid.UUID, _ uuid.UUID, fromStatus string, toStatus string, _ string, _ uuid.UUID) error {
	f.statusLogs = append(f.statusLogs, struct {
		FromStatus string
		ToStatus   string
	}{FromStatus: fromStatus, ToStatus: toStatus})
	return nil
}

type fakeOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakeOutboxRepository) Insert(_ context.Context, _ db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error) {
	if !json.Valid(params.Payload) {
		return nil, errors.New("invalid payload")
	}
	f.events = append(f.events, params)
	return &outbox.Event{}, nil
}
