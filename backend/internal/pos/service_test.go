package pos

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
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

var (
	posTenantA      = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	posStoreA       = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	posTenantB      = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	posStoreB       = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	posCashierA     = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	posCashierB     = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	posSessionA     = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	posProductA     = uuid.MustParse("88888888-8888-8888-8888-888888888888")
	posProductB     = uuid.MustParse("99999999-9999-9999-9999-999999999999")
	posTransactionA = uuid.MustParse("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
)

func TestOpenSessionSucceeds(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newPOSTestService()

	result, err := service.OpenSession(context.Background(), posTenantA, posStoreA, OpenSessionInput{
		ActorUserID:       posCashierA,
		OpeningCashAmount: 200000,
		Note:              "Shift pagi",
	})
	if err != nil {
		t.Fatalf("OpenSession error = %v", err)
	}
	if result.Status != SessionStatusOpen || result.OpeningCash != 200000 {
		t.Fatalf("session response = %#v", result)
	}
	if len(repo.sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(repo.sessions))
	}
	if len(auditRepo.entries) != 1 || auditRepo.entries[0].Action != AuditActionSessionOpen {
		t.Fatalf("audit entries = %#v", auditRepo.entries)
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventCashierSessionOpened {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestCannotOpenSecondSessionForSameUserTenantStore(t *testing.T) {
	service, _, auditRepo, outboxRepo := newPOSTestService()

	_, err := service.OpenSession(context.Background(), posTenantA, posStoreA, OpenSessionInput{
		ActorUserID:       posCashierA,
		OpeningCashAmount: 100000,
	})
	if err != nil {
		t.Fatalf("first OpenSession error = %v", err)
	}
	_, err = service.OpenSession(context.Background(), posTenantA, posStoreA, OpenSessionInput{
		ActorUserID:       posCashierA,
		OpeningCashAmount: 50000,
	})
	assertPOSAppErrorCode(t, err, apperror.CodeConflict)
	if len(auditRepo.entries) != 1 {
		t.Fatalf("audit entries = %d, want only first open", len(auditRepo.entries))
	}
	if len(outboxRepo.events) != 1 {
		t.Fatalf("outbox events = %d, want only first open", len(outboxRepo.events))
	}
}

func TestCurrentSessionIsTenantScoped(t *testing.T) {
	service, repo, _, _ := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)

	result, err := service.CurrentSession(context.Background(), posTenantA, posStoreA, CurrentSessionInput{ActorUserID: posCashierA})
	if err != nil {
		t.Fatalf("CurrentSession error = %v", err)
	}
	if result.ID != posSessionA.String() {
		t.Fatalf("session id = %s, want %s", result.ID, posSessionA)
	}

	_, err = service.CurrentSession(context.Background(), posTenantB, posStoreB, CurrentSessionInput{ActorUserID: posCashierA})
	assertPOSAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestCashierCannotCloseAnotherUsersSession(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)

	_, err := service.CloseSession(context.Background(), posTenantA, posStoreA, posSessionA, CloseSessionInput{
		ActorUserID:       posCashierB,
		Role:              string(permission.RoleCashier),
		ClosingCashAmount: 250000,
	})
	assertPOSAppErrorCode(t, err, apperror.CodeForbidden)
	if repo.sessions[posSessionA].Status != SessionStatusOpen {
		t.Fatalf("session status = %s, want open", repo.sessions[posSessionA].Status)
	}
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestOwnerCanCloseAnotherUsersSessionAndCalculatesCash(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)
	repo.cashSales[posSessionA] = 150000

	result, err := service.CloseSession(context.Background(), posTenantA, posStoreA, posSessionA, CloseSessionInput{
		ActorUserID:       posCashierB,
		Role:              string(permission.RoleOwner),
		ClosingCashAmount: 360000,
		Note:              "Kas lebih sepuluh ribu",
	})
	if err != nil {
		t.Fatalf("CloseSession error = %v", err)
	}
	if result.Status != SessionStatusClosed {
		t.Fatalf("status = %s, want closed", result.Status)
	}
	if result.ExpectedCash == nil || *result.ExpectedCash != 350000 {
		t.Fatalf("expected cash = %#v, want 350000", result.ExpectedCash)
	}
	if result.Difference == nil || *result.Difference != 10000 {
		t.Fatalf("difference = %#v, want 10000", result.Difference)
	}
	if len(auditRepo.entries) != 1 || auditRepo.entries[0].Action != AuditActionSessionClose {
		t.Fatalf("audit entries = %#v", auditRepo.entries)
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventCashierSessionClosed {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestDoubleCloseRejected(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newPOSTestService()
	session := testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)
	session.Status = SessionStatusClosed
	now := time.Now().UTC()
	session.ClosedAt = &now
	repo.sessions[posSessionA] = session

	_, err := service.CloseSession(context.Background(), posTenantA, posStoreA, posSessionA, CloseSessionInput{
		ActorUserID:       posCashierA,
		Role:              string(permission.RoleCashier),
		ClosingCashAmount: 200000,
	})
	assertPOSAppErrorCode(t, err, apperror.CodeConflict)
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestPOSTransactionSucceedsWithEnoughStock(t *testing.T) {
	service, repo, _, outboxRepo := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)

	result, err := service.CreateTransaction(context.Background(), posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductA, 2), "pos-success"))
	if err != nil {
		t.Fatalf("CreateTransaction error = %v", err)
	}
	if result.StatusCode != 201 {
		t.Fatalf("status code = %d, want 201", result.StatusCode)
	}
	if result.Response.GrandTotal != 50000 || result.Response.ChangeAmount != 10000 {
		t.Fatalf("response = %#v, want total 50000 change 10000", result.Response)
	}
	if len(repo.transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(repo.transactions))
	}
	if len(repo.transactionItems) != 1 {
		t.Fatalf("transaction items = %d, want 1", len(repo.transactionItems))
	}
	snapshot := repo.snapshots[posProductA]
	if snapshot.QuantityOnHand != 8 || snapshot.QuantityReserved != 1 || snapshot.QuantityAvailable != 7 {
		t.Fatalf("snapshot = %#v, want on hand 8 reserved 1 available 7", snapshot)
	}
	if len(repo.stockMovements) != 1 {
		t.Fatalf("stock movements = %d, want 1", len(repo.stockMovements))
	}
	if repo.stockMovements[0].MovementType != "pos_sale" || repo.stockMovements[0].Quantity != -2 {
		t.Fatalf("stock movement = %#v", repo.stockMovements[0])
	}
	if len(outboxRepo.events) != 3 {
		t.Fatalf("outbox events = %d, want 3", len(outboxRepo.events))
	}
}

func TestPOSTransactionRejectsInsufficientStock(t *testing.T) {
	service, repo, _, outboxRepo := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)

	_, err := service.CreateTransaction(context.Background(), posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductA, 20), "pos-insufficient"))
	assertPOSAppErrorCode(t, err, apperror.CodeInsufficientStock)
	if len(repo.transactions) != 0 {
		t.Fatalf("transactions = %d, want 0", len(repo.transactions))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestPOSTransactionRejectsAmountPaidLessThanTotal(t *testing.T) {
	service, repo, _, _ := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)
	request := validPOSTransactionRequest(posSessionA, posProductA, 2)
	amountPaid := int64(49999)
	request.AmountPaid = &amountPaid

	_, err := service.CreateTransaction(context.Background(), posTransactionCommand(t, request, "pos-underpaid"))
	assertPOSAppErrorCode(t, err, apperror.CodeValidation)
}

func TestPOSTransactionDoubleSubmitSameKeyCreatesOneTransaction(t *testing.T) {
	service, repo, _, _ := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)
	cmd := posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductA, 1), "pos-replay")

	first, err := service.CreateTransaction(context.Background(), cmd)
	if err != nil {
		t.Fatalf("first CreateTransaction error = %v", err)
	}
	second, err := service.CreateTransaction(context.Background(), cmd)
	if err != nil {
		t.Fatalf("second CreateTransaction error = %v", err)
	}
	if len(repo.transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(repo.transactions))
	}
	if second.Response.ID != first.Response.ID || second.Response.GrandTotal != first.Response.GrandTotal {
		t.Fatalf("replayed response differs: first=%#v second=%#v", first.Response, second.Response)
	}
}

func TestPOSTransactionSameKeyDifferentPayloadReturnsConflict(t *testing.T) {
	service, repo, _, _ := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)

	if _, err := service.CreateTransaction(context.Background(), posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductA, 1), "pos-conflict")); err != nil {
		t.Fatalf("first CreateTransaction error = %v", err)
	}
	_, err := service.CreateTransaction(context.Background(), posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductA, 2), "pos-conflict"))
	assertPOSAppErrorCode(t, err, apperror.CodeIdempotencyConflict)
	if len(repo.transactions) != 1 {
		t.Fatalf("transactions = %d, want 1", len(repo.transactions))
	}
}

func TestPOSTransactionWithoutOpenSessionRejected(t *testing.T) {
	service, _, _, _ := newPOSTestService()

	_, err := service.CreateTransaction(context.Background(), posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductA, 1), "pos-no-session"))
	assertPOSAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestCashierCannotUseAnotherCashierSession(t *testing.T) {
	service, repo, _, _ := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)
	cmd := posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductA, 1), "pos-other-session")
	cmd.ActorUserID = posCashierB
	cmd.Role = string(permission.RoleCashier)

	_, err := service.CreateTransaction(context.Background(), cmd)
	assertPOSAppErrorCode(t, err, apperror.CodeForbidden)
}

func TestPOSTransactionRejectsProductFromAnotherTenantStore(t *testing.T) {
	service, repo, _, _ := newPOSTestService()
	repo.sessions[posSessionA] = testOpenSession(posSessionA, posTenantA, posStoreA, posCashierA)

	_, err := service.CreateTransaction(context.Background(), posTransactionCommand(t, validPOSTransactionRequest(posSessionA, posProductB, 1), "pos-other-product"))
	assertPOSAppErrorCode(t, err, apperror.CodeNotFound)
}

func TestPOSProductSearchIsTenantScoped(t *testing.T) {
	service, _, _, _ := newPOSTestService()

	items, err := service.ListProducts(context.Background(), posTenantA, posStoreA, ProductSearchFilters{Query: "Kopi"})
	if err != nil {
		t.Fatalf("ListProducts error = %v", err)
	}
	if len(items) != 1 || items[0].ProductID != posProductA.String() {
		t.Fatalf("items = %#v, want only tenant A product", items)
	}

	items, err = service.ListProducts(context.Background(), posTenantA, posStoreA, ProductSearchFilters{Query: "Produk tenant B"})
	if err != nil {
		t.Fatalf("ListProducts tenant B query error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("items = %#v, want no tenant B product", items)
	}
}

func newPOSTestService() (*Service, *fakePOSRepository, *fakePOSAuditRepository, *fakePOSOutboxRepository) {
	repo := &fakePOSRepository{
		sessions:          make(map[uuid.UUID]CashierSession),
		cashSales:         make(map[uuid.UUID]int64),
		products:          testPOSProducts(),
		snapshots:         testPOSStockSnapshots(),
		transactions:      make(map[uuid.UUID]POSTransaction),
		transactionItems:  make([]POSTransactionItem, 0),
		itemsByTx:         make(map[uuid.UUID][]POSTransactionItem),
		stockMovements:    make([]CreateStockMovementParams, 0),
		transactionNumber: "POS-20260520-TEST0001",
		now:               time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
	}
	auditRepo := &fakePOSAuditRepository{}
	idempotencyRepo := &fakePOSIdempotencyRepository{records: make(map[string]idempotency.Record)}
	outboxRepo := &fakePOSOutboxRepository{}
	service := NewService(fakePOSDB{}, repo, auditRepo, idempotencyRepo, outboxRepo)
	service.now = func() time.Time { return repo.now }
	service.newUUID = func() uuid.UUID { return uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee") }
	return service, repo, auditRepo, outboxRepo
}

func testPOSProducts() map[uuid.UUID]POSProduct {
	return map[uuid.UUID]POSProduct{
		posProductA: {
			ID:             posProductA,
			TenantID:       posTenantA,
			StoreID:        posStoreA,
			Name:           "Kopi Susu",
			SKU:            "KOPI-SUSU",
			Price:          25000,
			Status:         "active",
			StockAvailable: 9,
		},
		posProductB: {
			ID:             posProductB,
			TenantID:       posTenantB,
			StoreID:        posStoreB,
			Name:           "Produk tenant B",
			SKU:            "TENANT-B",
			Price:          10000,
			Status:         "active",
			StockAvailable: 5,
		},
	}
}

func testPOSStockSnapshots() map[uuid.UUID]StockSnapshot {
	return map[uuid.UUID]StockSnapshot{
		posProductA: {
			ID:                uuid.MustParse("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
			TenantID:          posTenantA,
			StoreID:           posStoreA,
			ProductID:         posProductA,
			QuantityOnHand:    10,
			QuantityReserved:  1,
			QuantityAvailable: 9,
		},
		posProductB: {
			ID:                uuid.MustParse("cccccccc-cccc-cccc-cccc-cccccccccccc"),
			TenantID:          posTenantB,
			StoreID:           posStoreB,
			ProductID:         posProductB,
			QuantityOnHand:    5,
			QuantityReserved:  0,
			QuantityAvailable: 5,
		},
	}
}

func validPOSTransactionRequest(sessionID uuid.UUID, productID uuid.UUID, quantity int) CreateTransactionRequest {
	amountPaid := int64(60000)
	return CreateTransactionRequest{
		SessionID:     sessionID,
		PaymentMethod: PaymentMethodCash,
		AmountPaid:    &amountPaid,
		Items: []CreateTransactionItemRequest{
			{ProductID: productID, Quantity: quantity},
		},
	}
}

func posTransactionCommand(t *testing.T, request CreateTransactionRequest, idempotencyKey string) CreateTransactionCommand {
	t.Helper()

	rawBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	return CreateTransactionCommand{
		TenantID:       posTenantA,
		StoreID:        posStoreA,
		ActorUserID:    posCashierA,
		Role:           string(permission.RoleCashier),
		IdempotencyKey: idempotencyKey,
		Method:         "POST",
		Path:           "/api/v1/pos/transactions",
		RawBody:        rawBody,
		Request:        request,
	}
}

func testOpenSession(id uuid.UUID, tenantID uuid.UUID, storeID uuid.UUID, cashierID uuid.UUID) CashierSession {
	now := time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
	return CashierSession{
		ID:            id,
		TenantID:      tenantID,
		StoreID:       storeID,
		CashierID:     cashierID,
		SessionNumber: "CS-20260520-TEST0001",
		OpeningCash:   200000,
		Status:        SessionStatusOpen,
		OpenedAt:      now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func assertPOSAppErrorCode(t *testing.T, err error, code apperror.Code) {
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

type fakePOSDB struct{}

func (fakePOSDB) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(fakePOSDB{})
}

func (fakePOSDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakePOSDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakePOSDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakePOSRepository struct {
	sessions          map[uuid.UUID]CashierSession
	cashSales         map[uuid.UUID]int64
	products          map[uuid.UUID]POSProduct
	snapshots         map[uuid.UUID]StockSnapshot
	transactions      map[uuid.UUID]POSTransaction
	transactionItems  []POSTransactionItem
	itemsByTx         map[uuid.UUID][]POSTransactionItem
	stockMovements    []CreateStockMovementParams
	transactionNumber string
	now               time.Time
}

func (f *fakePOSRepository) FindCurrentOpenByCashier(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, cashierID uuid.UUID) (*CashierSession, error) {
	for _, session := range f.sessions {
		if session.TenantID == tenantID && session.StoreID == storeID && session.CashierID == cashierID && session.Status == SessionStatusOpen {
			copy := session
			return &copy, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (f *fakePOSRepository) CreateSession(_ context.Context, _ db.Queryer, params CreateSessionParams) (*CashierSession, error) {
	if existing, err := f.FindCurrentOpenByCashier(context.Background(), nil, params.TenantID, params.StoreID, params.CashierID); err == nil && existing != nil {
		return nil, ErrOpenSessionExists
	}
	id := posSessionA
	if _, exists := f.sessions[id]; exists {
		id = uuid.New()
	}
	session := CashierSession{
		ID:            id,
		TenantID:      params.TenantID,
		StoreID:       params.StoreID,
		CashierID:     params.CashierID,
		SessionNumber: params.SessionNumber,
		OpeningCash:   params.OpeningCash,
		Status:        SessionStatusOpen,
		OpenedAt:      f.now,
		CreatedAt:     f.now,
		UpdatedAt:     f.now,
	}
	f.sessions[id] = session
	return &session, nil
}

func (f *fakePOSRepository) LockSessionByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, sessionID uuid.UUID) (*CashierSession, error) {
	session, ok := f.sessions[sessionID]
	if !ok || session.TenantID != tenantID || session.StoreID != storeID {
		return nil, ErrSessionNotFound
	}
	copy := session
	return &copy, nil
}

func (f *fakePOSRepository) SumCompletedCashTransactions(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, sessionID uuid.UUID) (int64, error) {
	session, ok := f.sessions[sessionID]
	if !ok || session.TenantID != tenantID || session.StoreID != storeID {
		return 0, ErrSessionNotFound
	}
	return f.cashSales[sessionID], nil
}

func (f *fakePOSRepository) CloseSession(_ context.Context, _ db.Queryer, params CloseSessionParams) (*CashierSession, error) {
	session, ok := f.sessions[params.SessionID]
	if !ok || session.TenantID != params.TenantID || session.StoreID != params.StoreID {
		return nil, ErrSessionNotFound
	}
	if session.Status != SessionStatusOpen {
		return nil, ErrSessionAlreadyDone
	}
	now := f.now.Add(8 * time.Hour)
	session.Status = SessionStatusClosed
	session.ClosingCash = &params.ClosingCash
	session.ExpectedCash = &params.ExpectedCash
	session.Difference = &params.Difference
	session.ClosedAt = &now
	session.UpdatedAt = now
	f.sessions[params.SessionID] = session
	return &session, nil
}

func (f *fakePOSRepository) ListProducts(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ProductSearchFilters) ([]POSProduct, error) {
	items := make([]POSProduct, 0)
	for _, product := range f.products {
		if product.TenantID != tenantID || product.StoreID != storeID || product.Status != "active" {
			continue
		}
		if filters.Query != "" && product.Name != filters.Query && product.SKU != filters.Query {
			if !(containsFold(product.Name, filters.Query) || containsFold(product.SKU, filters.Query)) {
				continue
			}
		}
		if filters.Barcode != "" && product.Barcode != filters.Barcode {
			continue
		}
		if snapshot, ok := f.snapshots[product.ID]; ok {
			product.StockAvailable = snapshot.QuantityAvailable
		}
		items = append(items, product)
	}
	if filters.Limit > 0 && len(items) > filters.Limit {
		items = items[:filters.Limit]
	}
	return items, nil
}

func (f *fakePOSRepository) ListProductsByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productIDs []uuid.UUID) ([]POSProduct, error) {
	items := make([]POSProduct, 0, len(productIDs))
	for _, productID := range productIDs {
		product, ok := f.products[productID]
		if !ok || product.TenantID != tenantID || product.StoreID != storeID || product.Status != "active" {
			continue
		}
		if snapshot, ok := f.snapshots[product.ID]; ok {
			product.StockAvailable = snapshot.QuantityAvailable
		}
		items = append(items, product)
	}
	return items, nil
}

func (f *fakePOSRepository) LockStockSnapshots(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productIDs []uuid.UUID) ([]StockSnapshot, error) {
	items := make([]StockSnapshot, 0, len(productIDs))
	for _, productID := range productIDs {
		snapshot, ok := f.snapshots[productID]
		if !ok || snapshot.TenantID != tenantID || snapshot.StoreID != storeID {
			continue
		}
		items = append(items, snapshot)
	}
	return items, nil
}

func (f *fakePOSRepository) CreateTransaction(_ context.Context, _ db.Queryer, params CreateTransactionParams) (*POSTransaction, error) {
	id := posTransactionA
	if _, exists := f.transactions[id]; exists {
		id = uuid.New()
	}
	transaction := POSTransaction{
		ID:                id,
		TenantID:          params.TenantID,
		StoreID:           params.StoreID,
		CashierSessionID:  params.CashierSessionID,
		CashierID:         params.CashierID,
		TransactionNumber: params.TransactionNumber,
		Subtotal:          params.Subtotal,
		DiscountTotal:     params.DiscountTotal,
		TaxTotal:          params.TaxTotal,
		GrandTotal:        params.GrandTotal,
		PaymentMethod:     params.PaymentMethod,
		PaymentAmount:     params.PaymentAmount,
		ChangeAmount:      params.ChangeAmount,
		Status:            TransactionStatusDone,
		CreatedAt:         f.now,
		UpdatedAt:         f.now,
	}
	if transaction.TransactionNumber == "" {
		transaction.TransactionNumber = f.transactionNumber
	}
	f.transactions[id] = transaction
	return &transaction, nil
}

func (f *fakePOSRepository) CreateTransactionItem(_ context.Context, _ db.Queryer, params CreateTransactionItemParams) error {
	id := uuid.New()
	productID := params.ProductID
	item := POSTransactionItem{
		ID:               id,
		TenantID:         params.TenantID,
		POSTransactionID: params.POSTransactionID,
		ProductID:        &productID,
		ProductName:      params.ProductName,
		SKU:              params.SKU,
		Quantity:         params.Quantity,
		UnitPrice:        params.UnitPrice,
		DiscountTotal:    params.DiscountTotal,
		Subtotal:         params.Subtotal,
		CreatedAt:        f.now,
	}
	f.transactionItems = append(f.transactionItems, item)
	f.itemsByTx[params.POSTransactionID] = append(f.itemsByTx[params.POSTransactionID], item)
	return nil
}

func (f *fakePOSRepository) UpdateStockSnapshot(_ context.Context, _ db.Queryer, params UpdateStockSnapshotParams) error {
	snapshot, ok := f.snapshots[params.ProductID]
	if !ok || snapshot.TenantID != params.TenantID || snapshot.StoreID != params.StoreID {
		return ErrStockSnapshotNotFound
	}
	snapshot.QuantityOnHand = params.QuantityOnHand
	snapshot.QuantityReserved = params.QuantityReserved
	snapshot.QuantityAvailable = params.QuantityAvailable
	f.snapshots[params.ProductID] = snapshot
	return nil
}

func (f *fakePOSRepository) CreateStockMovement(_ context.Context, _ db.Queryer, params CreateStockMovementParams) error {
	f.stockMovements = append(f.stockMovements, params)
	return nil
}

func (f *fakePOSRepository) ListTransactions(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters TransactionListFilters) ([]POSTransaction, error) {
	items := make([]POSTransaction, 0)
	for _, transaction := range f.transactions {
		if transaction.TenantID != tenantID || transaction.StoreID != storeID {
			continue
		}
		if filters.CashierID != nil && transaction.CashierID != *filters.CashierID {
			continue
		}
		if filters.PaymentMethod != nil && transaction.PaymentMethod != *filters.PaymentMethod {
			continue
		}
		items = append(items, transaction)
	}
	if filters.Limit > 0 && len(items) > filters.Limit {
		items = items[:filters.Limit]
	}
	return items, nil
}

func (f *fakePOSRepository) FindTransactionByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, transactionID uuid.UUID) (*POSTransaction, error) {
	transaction, ok := f.transactions[transactionID]
	if !ok || transaction.TenantID != tenantID || transaction.StoreID != storeID {
		return nil, ErrTransactionNotFound
	}
	copy := transaction
	return &copy, nil
}

func (f *fakePOSRepository) ListTransactionItems(_ context.Context, _ db.Queryer, tenantID uuid.UUID, transactionID uuid.UUID) ([]POSTransactionItem, error) {
	items := make([]POSTransactionItem, 0)
	for _, item := range f.itemsByTx[transactionID] {
		if item.TenantID == tenantID {
			items = append(items, item)
		}
	}
	return items, nil
}

func containsFold(value string, needle string) bool {
	if needle == "" {
		return true
	}
	valueRunes := []rune(value)
	needleRunes := []rune(needle)
	if len(needleRunes) > len(valueRunes) {
		return false
	}
	for i := 0; i <= len(valueRunes)-len(needleRunes); i++ {
		if stringEqualFold(string(valueRunes[i:i+len(needleRunes)]), needle) {
			return true
		}
	}
	return false
}

func stringEqualFold(a string, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

type fakePOSIdempotencyRepository struct {
	records map[string]idempotency.Record
}

func (f *fakePOSIdempotencyRepository) Begin(_ context.Context, _ db.Queryer, tenantID uuid.UUID, scope string, key string, requestHash string, lockedUntil time.Time) (*idempotency.State, error) {
	mapKey := idempotencyMapKey(tenantID, scope, key)
	record, exists := f.records[mapKey]
	if !exists {
		record = idempotency.Record{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Scope:       scope,
			Key:         key,
			RequestHash: requestHash,
			LockedUntil: &lockedUntil,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
		}
		f.records[mapKey] = record
		return &idempotency.State{
			Record:       &record,
			Created:      true,
			IsProcessing: true,
			LockedUntil:  record.LockedUntil,
		}, nil
	}
	return idempotency.ResolveExisting(record, requestHash)
}

func (f *fakePOSIdempotencyRepository) SaveCompletedResponse(_ context.Context, _ db.Queryer, tenantID uuid.UUID, scope string, key string, statusCode int, responseBody json.RawMessage) error {
	mapKey := idempotencyMapKey(tenantID, scope, key)
	record, exists := f.records[mapKey]
	if !exists {
		return idempotency.ErrKeyNotFound
	}
	record.StatusCode = &statusCode
	record.ResponseBody = append(json.RawMessage(nil), responseBody...)
	record.LockedUntil = nil
	record.UpdatedAt = time.Now().UTC()
	f.records[mapKey] = record
	return nil
}

func idempotencyMapKey(tenantID uuid.UUID, scope string, key string) string {
	return tenantID.String() + "|" + scope + "|" + key
}

type fakePOSAuditRepository struct {
	entries []audit.Entry
}

func (f *fakePOSAuditRepository) Create(_ context.Context, _ db.Queryer, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type fakePOSOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakePOSOutboxRepository) Insert(_ context.Context, _ db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error) {
	if !json.Valid(params.Payload) {
		return nil, errors.New("invalid json payload")
	}
	f.events = append(f.events, params)
	return &outbox.Event{}, nil
}
