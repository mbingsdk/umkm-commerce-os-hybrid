//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	checkoutpkg "github.com/sdkdev/umkm-commerce-os/backend/internal/checkout"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
	orderpkg "github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	paymentpkg "github.com/sdkdev/umkm-commerce-os/backend/internal/payment"
	dbpkg "github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	pospkg "github.com/sdkdev/umkm-commerce-os/backend/internal/pos"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	auditpkg "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	idempotencypkg "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	outboxpkg "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	storepkg "github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

var integrationDBMu sync.Mutex

type integrationEnv struct {
	db         *dbpkg.DB
	tenantID   uuid.UUID
	storeID    uuid.UUID
	productID  uuid.UUID
	categoryID uuid.UUID
	ownerID    uuid.UUID
	cashierID  uuid.UUID
	storeSlug  string
}

type outboxInserter interface {
	Insert(context.Context, dbpkg.Queryer, outboxpkg.InsertEventParams) (*outboxpkg.Event, error)
}

func newIntegrationEnv(t *testing.T, stockAvailable int) *integrationEnv {
	t.Helper()

	if os.Getenv("RUN_DB_INTEGRATION") != "1" {
		t.Skip("set RUN_DB_INTEGRATION=1 and TEST_DATABASE_URL to run PostgreSQL race/idempotency tests")
	}
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL race/idempotency tests")
	}

	integrationDBMu.Lock()
	t.Cleanup(integrationDBMu.Unlock)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)

	database, err := dbpkg.New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	t.Cleanup(database.Close)

	applyMigrations(t, ctx, database)
	truncatePublicData(t, ctx, database)

	env := &integrationEnv{
		db:         database,
		tenantID:   uuid.MustParse("aaaaaaaa-aaaa-4aaa-aaaa-aaaaaaaaaaaa"),
		storeID:    uuid.MustParse("bbbbbbbb-bbbb-4bbb-bbbb-bbbbbbbbbbbb"),
		productID:  uuid.MustParse("cccccccc-cccc-4ccc-cccc-cccccccccccc"),
		categoryID: uuid.MustParse("dddddddd-dddd-4ddd-dddd-dddddddddddd"),
		ownerID:    uuid.MustParse("eeeeeeee-eeee-4eee-eeee-eeeeeeeeeeee"),
		cashierID:  uuid.MustParse("ffffffff-ffff-4fff-8fff-ffffffffffff"),
		storeSlug:  "toko-integrasi",
	}
	env.seedBaseData(t, ctx, stockAvailable)

	return env
}

func (e *integrationEnv) checkoutService(outbox outboxInserter) *checkoutpkg.Service {
	return checkoutpkg.NewService(
		e.db,
		staticStoreResolver{ctx: e.publicStoreContext()},
		checkoutpkg.NewRepository(),
		idempotencypkg.NewRepository(),
		outbox,
	)
}

func (e *integrationEnv) posService(outbox outboxInserter) *pospkg.Service {
	return pospkg.NewService(
		e.db,
		pospkg.NewRepository(),
		auditpkg.NewRepository(),
		idempotencypkg.NewRepository(),
		outbox,
	)
}

func (e *integrationEnv) paymentService(outbox outboxInserter) *paymentpkg.Service {
	return paymentpkg.NewService(
		e.db,
		staticStoreResolver{ctx: e.publicStoreContext()},
		paymentpkg.NewRepository(),
		idempotencypkg.NewRepository(),
		outbox,
	)
}

func (e *integrationEnv) orderService(outbox outboxInserter) *orderpkg.Service {
	return orderpkg.NewService(e.db, orderpkg.NewRepository(), outbox)
}

func (e *integrationEnv) publicStoreContext() storepkg.PublicContext {
	return storepkg.PublicContext{
		TenantID: e.tenantID,
		StoreID:  e.storeID,
		Store: storepkg.Store{
			ID:             e.storeID,
			TenantID:       e.tenantID,
			Name:           "Toko Integrasi",
			Slug:           e.storeSlug,
			Status:         storepkg.StatusPublished,
			IsDiscoverable: true,
		},
	}
}

type staticStoreResolver struct {
	ctx storepkg.PublicContext
}

func (r staticStoreResolver) Resolve(_ context.Context, slug string) (storepkg.PublicContext, error) {
	if slug != r.ctx.Store.Slug {
		return storepkg.PublicContext{}, apperror.NotFound("Store not found")
	}
	return r.ctx, nil
}

type failingOutbox struct {
	err error
}

func (f failingOutbox) Insert(context.Context, dbpkg.Queryer, outboxpkg.InsertEventParams) (*outboxpkg.Event, error) {
	if f.err != nil {
		return nil, f.err
	}
	return nil, errors.New("forced outbox failure")
}

func TestIntegrationCheckoutRaceLastStock(t *testing.T) {
	env := newIntegrationEnv(t, 1)
	service := env.checkoutService(outboxpkg.NewRepository())

	errs := runConcurrent(2, func(i int) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := env.checkoutCommand(t, fmt.Sprintf("checkout-race-%d", i), 1, fmt.Sprintf("62811100000%d", i))
		_, err := service.Checkout(ctx, cmd)
		return err
	})

	assertErrorCounts(t, errs, 1, map[apperror.Code]int{
		apperror.CodeInsufficientStock: 1,
	})
	assertSnapshot(t, env, stockSnapshot{OnHand: 1, Reserved: 1, Available: 0})
	assertCount(t, env, `SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND store_id = $2`, 1, env.tenantID, env.storeID)
	assertCount(t, env, `SELECT COUNT(*) FROM stock_reservations WHERE tenant_id = $1 AND store_id = $2 AND status = 'active'`, 1, env.tenantID, env.storeID)
}

func TestIntegrationCheckoutIdempotencyReplayAndConflict(t *testing.T) {
	env := newIntegrationEnv(t, 2)
	service := env.checkoutService(outboxpkg.NewRepository())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	firstCmd := env.checkoutCommand(t, "checkout-idempotent", 1, "628111100001")
	first, err := service.Checkout(ctx, firstCmd)
	if err != nil {
		t.Fatalf("first checkout failed: %v", err)
	}
	second, err := service.Checkout(ctx, firstCmd)
	if err != nil {
		t.Fatalf("second same-key checkout failed: %v", err)
	}
	if first.Response.OrderID != second.Response.OrderID || first.Response.OrderNumber != second.Response.OrderNumber {
		t.Fatalf("expected idempotent replay to return same order, got %#v then %#v", first.Response, second.Response)
	}
	assertCount(t, env, `SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND store_id = $2`, 1, env.tenantID, env.storeID)

	differentPayload := env.checkoutCommand(t, "checkout-idempotent", 2, "628111100001")
	if _, err := service.Checkout(ctx, differentPayload); appCode(err) != apperror.CodeIdempotencyConflict {
		t.Fatalf("expected IDEMPOTENCY_CONFLICT for same key with different payload, got %v (%v)", appCode(err), err)
	}
}

func TestIntegrationPOSTransactionRaceLastStock(t *testing.T) {
	env := newIntegrationEnv(t, 1)
	service := env.posService(outboxpkg.NewRepository())
	sessionID := env.openPOSSession(t, service)

	errs := runConcurrent(2, func(i int) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cmd := env.posTransactionCommand(t, sessionID, fmt.Sprintf("pos-race-%d", i), 1, 50_000)
		_, err := service.CreateTransaction(ctx, cmd)
		return err
	})

	assertErrorCounts(t, errs, 1, map[apperror.Code]int{
		apperror.CodeInsufficientStock: 1,
	})
	assertSnapshot(t, env, stockSnapshot{OnHand: 0, Reserved: 0, Available: 0})
	assertCount(t, env, `SELECT COUNT(*) FROM pos_transactions WHERE tenant_id = $1 AND store_id = $2`, 1, env.tenantID, env.storeID)
	assertCount(t, env, `SELECT COUNT(*) FROM stock_movements WHERE tenant_id = $1 AND store_id = $2 AND movement_type = 'pos_sale'`, 1, env.tenantID, env.storeID)
}

func TestIntegrationPOSIdempotencyReplayAndConflict(t *testing.T) {
	env := newIntegrationEnv(t, 2)
	service := env.posService(outboxpkg.NewRepository())
	sessionID := env.openPOSSession(t, service)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	firstCmd := env.posTransactionCommand(t, sessionID, "pos-idempotent", 1, 50_000)
	first, err := service.CreateTransaction(ctx, firstCmd)
	if err != nil {
		t.Fatalf("first POS transaction failed: %v", err)
	}
	second, err := service.CreateTransaction(ctx, firstCmd)
	if err != nil {
		t.Fatalf("second same-key POS transaction failed: %v", err)
	}
	if first.Response.ID != second.Response.ID || first.Response.TransactionNumber != second.Response.TransactionNumber {
		t.Fatalf("expected idempotent POS replay to return same transaction, got %#v then %#v", first.Response, second.Response)
	}
	assertCount(t, env, `SELECT COUNT(*) FROM pos_transactions WHERE tenant_id = $1 AND store_id = $2`, 1, env.tenantID, env.storeID)

	differentPayload := env.posTransactionCommand(t, sessionID, "pos-idempotent", 2, 100_000)
	if _, err := service.CreateTransaction(ctx, differentPayload); appCode(err) != apperror.CodeIdempotencyConflict {
		t.Fatalf("expected IDEMPOTENCY_CONFLICT for same key with different POS payload, got %v (%v)", appCode(err), err)
	}
}

func TestIntegrationPaymentConfirmAndCancelRaceKeepsStockConsistent(t *testing.T) {
	env := newIntegrationEnv(t, 1)
	orderID, orderNumber, customerPhone := env.createCheckoutOrder(t, "payment-cancel-checkout", "628111200001")
	confirmationID := env.submitPaymentConfirmation(t, "payment-public-confirm", orderNumber, customerPhone)

	paymentService := env.paymentService(outboxpkg.NewRepository())
	orderService := env.orderService(outboxpkg.NewRepository())

	errs := runConcurrent(2, func(i int) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if i == 0 {
			_, err := paymentService.ConfirmPayment(ctx, env.tenantID, env.storeID, orderID, paymentpkg.ConfirmInput{
				ActorUserID:    env.ownerID,
				ConfirmationID: &confirmationID,
				Note:           "Konfirmasi race test",
			})
			return err
		}

		_, err := orderService.Cancel(ctx, env.tenantID, env.storeID, orderID, orderpkg.CancelInput{
			ActorUserID: env.ownerID,
			Reason:      "race_test",
			Note:        "Cancel race test",
		})
		return err
	})

	successes := countSuccess(errs)
	if successes < 1 || successes > 2 {
		t.Fatalf("expected at least one operation to win and no more than two safe operations, got successes=%d errors=%v", successes, errs)
	}
	finalStatus := queryString(t, env, `SELECT status FROM orders WHERE tenant_id = $1 AND store_id = $2 AND id = $3`, env.tenantID, env.storeID, orderID)
	if finalStatus != orderpkg.StatusConfirmed && finalStatus != orderpkg.StatusCancelled {
		t.Fatalf("unexpected final order status after payment/cancel race: %s", finalStatus)
	}

	snapshot := readSnapshot(t, env)
	if snapshot.OnHand != 1 || snapshot.Reserved < 0 || snapshot.Available < 0 || snapshot.Reserved+snapshot.Available != snapshot.OnHand {
		t.Fatalf("stock snapshot inconsistent after payment/cancel race: %#v", snapshot)
	}
	assertMaxCount(t, env, `SELECT COUNT(*) FROM stock_movements WHERE tenant_id = $1 AND store_id = $2 AND movement_type = 'released'`, 1, env.tenantID, env.storeID)
	assertMaxCount(t, env, `SELECT COUNT(*) FROM payments WHERE tenant_id = $1 AND store_id = $2 AND order_id = $3`, 1, env.tenantID, env.storeID, orderID)
}

func TestIntegrationDoublePaymentConfirmCreatesOnePayment(t *testing.T) {
	env := newIntegrationEnv(t, 1)
	orderID, orderNumber, customerPhone := env.createCheckoutOrder(t, "double-payment-checkout", "628111300001")
	confirmationID := env.submitPaymentConfirmation(t, "double-payment-public-confirm", orderNumber, customerPhone)
	service := env.paymentService(outboxpkg.NewRepository())

	errs := runConcurrent(2, func(int) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := service.ConfirmPayment(ctx, env.tenantID, env.storeID, orderID, paymentpkg.ConfirmInput{
			ActorUserID:    env.ownerID,
			ConfirmationID: &confirmationID,
			Note:           "Double confirm test",
		})
		return err
	})

	assertErrorCounts(t, errs, 1, map[apperror.Code]int{
		apperror.CodeConflict: 1,
	})
	assertCount(t, env, `SELECT COUNT(*) FROM payments WHERE tenant_id = $1 AND store_id = $2 AND order_id = $3`, 1, env.tenantID, env.storeID, orderID)
	assertCount(t, env, `SELECT COUNT(*) FROM payment_confirmations WHERE tenant_id = $1 AND store_id = $2 AND order_id = $3 AND status = 'confirmed'`, 1, env.tenantID, env.storeID, orderID)
	assertMaxCount(t, env, `SELECT COUNT(*) FROM outbox_events WHERE tenant_id = $1 AND event_type = 'PaymentConfirmed'`, 1, env.tenantID)
	assertSnapshot(t, env, stockSnapshot{OnHand: 1, Reserved: 1, Available: 0})
}

func TestIntegrationOutboxFetchPendingUsesSkipLocked(t *testing.T) {
	env := newIntegrationEnv(t, 1)
	repo := outboxpkg.NewRepository()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inserted, err := repo.Insert(ctx, env.db, outboxpkg.InsertEventParams{
		TenantID:      env.tenantID,
		EventType:     "RaceTestEvent",
		AggregateType: "test",
		AggregateID:   env.productID,
		Payload:       json.RawMessage(`{"safe":true}`),
	})
	if err != nil {
		t.Fatalf("insert outbox event: %v", err)
	}

	firstFetched := make(chan []outboxpkg.Event, 1)
	releaseFirstTx := make(chan struct{})
	firstDone := make(chan error, 1)

	go func() {
		firstDone <- env.db.WithTx(ctx, func(tx dbpkg.Tx) error {
			events, err := repo.FetchPending(ctx, tx, 1, 5)
			if err != nil {
				return err
			}
			firstFetched <- events
			<-releaseFirstTx
			if len(events) == 0 {
				return nil
			}
			_, err = repo.MarkSucceeded(ctx, tx, events[0].ID)
			return err
		})
	}()

	events := <-firstFetched
	if len(events) != 1 || events[0].ID != inserted.ID {
		close(releaseFirstTx)
		t.Fatalf("expected first worker to fetch inserted event, got %#v", events)
	}

	var secondEvents []outboxpkg.Event
	if err := env.db.WithTx(ctx, func(tx dbpkg.Tx) error {
		var err error
		secondEvents, err = repo.FetchPending(ctx, tx, 1, 5)
		return err
	}); err != nil {
		close(releaseFirstTx)
		t.Fatalf("second fetch pending: %v", err)
	}
	if len(secondEvents) != 0 {
		close(releaseFirstTx)
		t.Fatalf("expected SKIP LOCKED to prevent second worker from fetching locked event, got %#v", secondEvents)
	}

	close(releaseFirstTx)
	if err := <-firstDone; err != nil {
		t.Fatalf("first worker finish: %v", err)
	}
	assertCount(t, env, `SELECT COUNT(*) FROM outbox_events WHERE id = $1 AND status = 'processed'`, 1, inserted.ID)
}

func TestIntegrationCheckoutRollbackOnOutboxFailure(t *testing.T) {
	env := newIntegrationEnv(t, 1)
	service := env.checkoutService(failingOutbox{err: errors.New("forced checkout outbox failure")})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := service.Checkout(ctx, env.checkoutCommand(t, "checkout-rollback", 1, "628111400001"))
	if err == nil {
		t.Fatal("expected checkout to fail when outbox insert fails")
	}
	assertSnapshot(t, env, stockSnapshot{OnHand: 1, Reserved: 0, Available: 1})
	assertCount(t, env, `SELECT COUNT(*) FROM orders WHERE tenant_id = $1 AND store_id = $2`, 0, env.tenantID, env.storeID)
	assertCount(t, env, `SELECT COUNT(*) FROM stock_reservations WHERE tenant_id = $1 AND store_id = $2`, 0, env.tenantID, env.storeID)
	assertCount(t, env, `SELECT COUNT(*) FROM idempotency_keys WHERE tenant_id = $1`, 0, env.tenantID)
}

func TestIntegrationPOSTransactionRollbackOnOutboxFailure(t *testing.T) {
	env := newIntegrationEnv(t, 1)
	normalService := env.posService(outboxpkg.NewRepository())
	sessionID := env.openPOSSession(t, normalService)
	service := env.posService(failingOutbox{err: errors.New("forced POS outbox failure")})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := service.CreateTransaction(ctx, env.posTransactionCommand(t, sessionID, "pos-rollback", 1, 50_000))
	if err == nil {
		t.Fatal("expected POS transaction to fail when outbox insert fails")
	}
	assertSnapshot(t, env, stockSnapshot{OnHand: 1, Reserved: 0, Available: 1})
	assertCount(t, env, `SELECT COUNT(*) FROM pos_transactions WHERE tenant_id = $1 AND store_id = $2`, 0, env.tenantID, env.storeID)
	assertCount(t, env, `SELECT COUNT(*) FROM stock_movements WHERE tenant_id = $1 AND store_id = $2 AND movement_type = 'pos_sale'`, 0, env.tenantID, env.storeID)
	assertCount(t, env, `SELECT COUNT(*) FROM idempotency_keys WHERE tenant_id = $1 AND scope = $2`, 0, env.tenantID, idempotencypkg.ScopePOS)
}

func (e *integrationEnv) checkoutCommand(t *testing.T, key string, quantity int, customerPhone string) checkoutpkg.Command {
	t.Helper()

	request := checkoutpkg.CheckoutRequest{
		Items: []checkoutpkg.CheckoutItemRequest{
			{ProductID: e.productID, Quantity: quantity},
		},
		Customer: checkoutpkg.CheckoutCustomerRequest{
			Name:  "Customer Integrasi",
			Phone: customerPhone,
			Email: "customer-integrasi@example.test",
		},
		ShippingAddress: checkoutpkg.CheckoutAddressRequest{
			RecipientName:  "Customer Integrasi",
			RecipientPhone: customerPhone,
			Address:        "Jl. Integrasi No. 1",
			City:           "Makassar",
			Province:       "Sulawesi Selatan",
			PostalCode:     "90111",
		},
		PaymentMethod: checkoutpkg.PaymentMethodManualTransfer,
	}
	rawBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal checkout request: %v", err)
	}

	return checkoutpkg.Command{
		StoreSlug:      e.storeSlug,
		IdempotencyKey: key,
		Method:         http.MethodPost,
		Path:           "/api/v1/public/stores/" + e.storeSlug + "/checkout",
		RawBody:        rawBody,
		Request:        request,
	}
}

func (e *integrationEnv) posTransactionCommand(t *testing.T, sessionID uuid.UUID, key string, quantity int, amountPaid int64) pospkg.CreateTransactionCommand {
	t.Helper()

	request := pospkg.CreateTransactionRequest{
		SessionID: sessionID,
		Items: []pospkg.CreateTransactionItemRequest{
			{ProductID: e.productID, Quantity: quantity},
		},
		PaymentMethod: pospkg.PaymentMethodCash,
		AmountPaid:    &amountPaid,
	}
	rawBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal POS request: %v", err)
	}

	return pospkg.CreateTransactionCommand{
		TenantID:       e.tenantID,
		StoreID:        e.storeID,
		ActorUserID:    e.cashierID,
		Role:           string(permission.RoleCashier),
		IdempotencyKey: key,
		Method:         http.MethodPost,
		Path:           "/api/v1/pos/transactions",
		RawBody:        rawBody,
		Request:        request,
	}
}

func (e *integrationEnv) createCheckoutOrder(t *testing.T, key string, customerPhone string) (uuid.UUID, string, string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := e.checkoutService(outboxpkg.NewRepository()).Checkout(ctx, e.checkoutCommand(t, key, 1, customerPhone))
	if err != nil {
		t.Fatalf("create checkout order: %v", err)
	}
	return result.Response.OrderID, result.Response.OrderNumber, customerPhone
}

func (e *integrationEnv) submitPaymentConfirmation(t *testing.T, key string, orderNumber string, customerPhone string) uuid.UUID {
	t.Helper()

	request := paymentpkg.PublicConfirmationRequest{
		CustomerPhone:  customerPhone,
		PayerName:      "Customer Integrasi",
		BankName:       "BCA",
		TransferAmount: 50_000,
		TransferDate:   "2026-05-24",
		Note:           "Konfirmasi integration test",
	}
	rawBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal payment confirmation: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	response, _, err := e.paymentService(outboxpkg.NewRepository()).PublicConfirm(ctx, paymentpkg.PublicConfirmationCommand{
		StoreSlug:      e.storeSlug,
		OrderNumber:    orderNumber,
		IdempotencyKey: key,
		Method:         http.MethodPost,
		Path:           "/api/v1/public/stores/" + e.storeSlug + "/orders/" + orderNumber + "/payment-confirmation",
		RawBody:        rawBody,
		Request:        request,
	})
	if err != nil {
		t.Fatalf("submit public payment confirmation: %v", err)
	}
	return response.ID
}

func (e *integrationEnv) openPOSSession(t *testing.T, service *pospkg.Service) uuid.UUID {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	session, err := service.OpenSession(ctx, e.tenantID, e.storeID, pospkg.OpenSessionInput{
		ActorUserID:       e.cashierID,
		OpeningCashAmount: 100_000,
		Note:              "Session integration test",
	})
	if err != nil {
		t.Fatalf("open POS session: %v", err)
	}
	sessionID, err := uuid.Parse(session.ID)
	if err != nil {
		t.Fatalf("parse POS session id: %v", err)
	}
	return sessionID
}

func runConcurrent(n int, fn func(int) error) []error {
	start := make(chan struct{})
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			<-start
			errs[idx] = fn(idx)
		}(i)
	}
	close(start)
	wg.Wait()
	return errs
}

func assertErrorCounts(t *testing.T, errs []error, wantSuccess int, wantCodes map[apperror.Code]int) {
	t.Helper()

	if got := countSuccess(errs); got != wantSuccess {
		t.Fatalf("success count mismatch: got %d want %d; errors=%v", got, wantSuccess, errs)
	}
	gotCodes := make(map[apperror.Code]int)
	for _, err := range errs {
		if err != nil {
			gotCodes[appCode(err)]++
		}
	}
	for code, want := range wantCodes {
		if gotCodes[code] != want {
			t.Fatalf("error code count mismatch for %s: got %d want %d; all=%v errors=%v", code, gotCodes[code], want, gotCodes, errs)
		}
	}
}

func countSuccess(errs []error) int {
	count := 0
	for _, err := range errs {
		if err == nil {
			count++
		}
	}
	return count
}

func appCode(err error) apperror.Code {
	if err == nil {
		return ""
	}
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

type stockSnapshot struct {
	OnHand    int
	Reserved  int
	Available int
}

func assertSnapshot(t *testing.T, env *integrationEnv, want stockSnapshot) {
	t.Helper()

	got := readSnapshot(t, env)
	if got != want {
		t.Fatalf("stock snapshot mismatch: got %#v want %#v", got, want)
	}
}

func readSnapshot(t *testing.T, env *integrationEnv) stockSnapshot {
	t.Helper()

	var snapshot stockSnapshot
	err := env.db.QueryRow(
		context.Background(),
		`
			SELECT quantity_on_hand, quantity_reserved, quantity_available
			FROM product_stock_snapshots
			WHERE tenant_id = $1 AND store_id = $2 AND product_id = $3
		`,
		env.tenantID,
		env.storeID,
		env.productID,
	).Scan(&snapshot.OnHand, &snapshot.Reserved, &snapshot.Available)
	if err != nil {
		t.Fatalf("read stock snapshot: %v", err)
	}
	return snapshot
}

func assertCount(t *testing.T, env *integrationEnv, query string, want int, args ...any) {
	t.Helper()

	got := queryInt(t, env, query, args...)
	if got != want {
		t.Fatalf("count mismatch: got %d want %d for query %q", got, want, query)
	}
}

func assertMaxCount(t *testing.T, env *integrationEnv, query string, max int, args ...any) {
	t.Helper()

	got := queryInt(t, env, query, args...)
	if got > max {
		t.Fatalf("count too high: got %d max %d for query %q", got, max, query)
	}
}

func queryInt(t *testing.T, env *integrationEnv, query string, args ...any) int {
	t.Helper()

	var count int
	if err := env.db.QueryRow(context.Background(), query, args...).Scan(&count); err != nil {
		t.Fatalf("query int: %v", err)
	}
	return count
}

func queryString(t *testing.T, env *integrationEnv, query string, args ...any) string {
	t.Helper()

	var value string
	if err := env.db.QueryRow(context.Background(), query, args...).Scan(&value); err != nil {
		t.Fatalf("query string: %v", err)
	}
	return value
}

func applyMigrations(t *testing.T, ctx context.Context, database *dbpkg.DB) {
	t.Helper()

	if _, err := database.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`); err != nil {
		t.Fatalf("ensure schema_migrations: %v", err)
	}

	applied := map[string]bool{}
	rows, err := database.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		t.Fatalf("load applied migrations: %v", err)
	}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			rows.Close()
			t.Fatalf("scan applied migration: %v", err)
		}
		applied[version] = true
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatalf("iterate applied migrations: %v", err)
	}
	rows.Close()

	migrationsDir := findMigrationsDir(t)
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		if applied[name] {
			continue
		}
		path := filepath.Join(migrationsDir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read migration %s: %v", name, err)
		}
		sqlText := strings.TrimPrefix(string(sqlBytes), "\ufeff")
		if err := database.WithTx(ctx, func(tx dbpkg.Tx) error {
			if _, err := tx.Exec(ctx, sqlText); err != nil {
				return fmt.Errorf("execute migration %s: %w", name, err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, name, name); err != nil {
				return fmt.Errorf("record migration %s: %w", name, err)
			}
			return nil
		}); err != nil {
			t.Fatalf("apply migration %s: %v", name, err)
		}
	}
}

func findMigrationsDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot resolve current test file path")
	}
	dir := filepath.Dir(file)
	for {
		candidate := filepath.Join(dir, "migrations")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("migrations directory not found")
		}
		dir = parent
	}
}

func truncatePublicData(t *testing.T, ctx context.Context, database *dbpkg.DB) {
	t.Helper()

	rows, err := database.Query(ctx, `
		SELECT tablename
		FROM pg_tables
		WHERE schemaname = 'public'
		  AND tablename <> 'schema_migrations'
	`)
	if err != nil {
		t.Fatalf("list public tables: %v", err)
	}
	defer rows.Close()

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		tables = append(tables, quoteIdent(table))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate public tables: %v", err)
	}
	if len(tables) == 0 {
		return
	}
	sort.Strings(tables)
	if _, err := database.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(tables, ", "))); err != nil {
		t.Fatalf("truncate public tables: %v", err)
	}
}

func quoteIdent(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

func (e *integrationEnv) seedBaseData(t *testing.T, ctx context.Context, stockAvailable int) {
	t.Helper()

	planID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	if err := e.db.WithTx(ctx, func(tx dbpkg.Tx) error {
		statements := []struct {
			sql  string
			args []any
		}{
			{
				sql: `
					INSERT INTO plans (
						id, code, name, description, price_monthly,
						product_limit, staff_limit, can_use_pos, can_use_discovery, can_use_courier, is_active
					)
					VALUES ($1, 'integration', 'Integration', 'Integration test plan', 0, 100, 10, true, true, true, true)
				`,
				args: []any{planID},
			},
			{
				sql: `
					INSERT INTO users (id, name, email, phone, password_hash, platform_role, status)
					VALUES
						($1, 'Owner Integrasi', 'owner-integrasi@example.test', '628111900001', 'hash', 'user', 'active'),
						($2, 'Kasir Integrasi', 'kasir-integrasi@example.test', '628111900002', 'hash', 'user', 'active')
				`,
				args: []any{e.ownerID, e.cashierID},
			},
			{
				sql: `
					INSERT INTO tenants (id, plan_id, name, slug, status)
					VALUES ($1, $2, 'Tenant Integrasi', 'tenant-integrasi', 'active')
				`,
				args: []any{e.tenantID, planID},
			},
			{
				sql: `
					INSERT INTO user_tenants (user_id, tenant_id, role, status, joined_at)
					VALUES
						($1, $3, 'owner', 'active', now()),
						($2, $3, 'cashier', 'active', now())
				`,
				args: []any{e.ownerID, e.cashierID, e.tenantID},
			},
			{
				sql: `
					INSERT INTO stores (
						id, tenant_id, name, slug, description, phone, whatsapp,
						address, city, province, status, is_discoverable, published_at
					)
					VALUES (
						$1, $2, 'Toko Integrasi', $3, 'Store untuk integration test',
						'628111900001', '628111900001', 'Jl. Integrasi No. 1',
						'Makassar', 'Sulawesi Selatan', 'published', true, now()
					)
				`,
				args: []any{e.storeID, e.tenantID, e.storeSlug},
			},
			{
				sql: `
					INSERT INTO categories (id, tenant_id, store_id, name, slug, is_active)
					VALUES ($1, $2, $3, 'Bouquet', 'bouquet', true)
				`,
				args: []any{e.categoryID, e.tenantID, e.storeID},
			},
			{
				sql: `
					INSERT INTO products (
						id, tenant_id, store_id, category_id, name, slug, sku,
						price, status, is_discoverable, track_inventory, allow_backorder
					)
					VALUES (
						$1, $2, $3, $4, 'Bouquet Integrasi', 'bouquet-integrasi',
						'BQT-INT-001', 50000, 'active', true, true, false
					)
				`,
				args: []any{e.productID, e.tenantID, e.storeID, e.categoryID},
			},
			{
				sql: `
					INSERT INTO product_stock_snapshots (
						tenant_id, store_id, product_id, quantity_on_hand, quantity_reserved, quantity_available, low_stock_threshold
					)
					VALUES ($1, $2, $3, $4, 0, $4, 1)
				`,
				args: []any{e.tenantID, e.storeID, e.productID, stockAvailable},
			},
			{
				sql: `
					INSERT INTO stock_movements (
						tenant_id, store_id, product_id, movement_type, quantity, balance_after, reference_type, note
					)
					VALUES ($1, $2, $3, $4, $5, $5, 'seed', 'Initial integration stock')
				`,
				args: []any{e.tenantID, e.storeID, e.productID, inventory.MovementTypeInitial, stockAvailable},
			},
		}
		for _, statement := range statements {
			if _, err := tx.Exec(ctx, statement.sql, statement.args...); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed integration data: %v", err)
	}
}
