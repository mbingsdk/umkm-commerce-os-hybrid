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
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

var (
	posTenantA  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	posStoreA   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	posTenantB  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	posStoreB   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	posCashierA = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	posCashierB = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	posSessionA = uuid.MustParse("77777777-7777-7777-7777-777777777777")
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

func newPOSTestService() (*Service, *fakePOSRepository, *fakePOSAuditRepository, *fakePOSOutboxRepository) {
	repo := &fakePOSRepository{
		sessions:  make(map[uuid.UUID]CashierSession),
		cashSales: make(map[uuid.UUID]int64),
		now:       time.Date(2026, 5, 20, 9, 0, 0, 0, time.UTC),
	}
	auditRepo := &fakePOSAuditRepository{}
	outboxRepo := &fakePOSOutboxRepository{}
	service := NewService(fakePOSDB{}, repo, auditRepo, outboxRepo)
	service.now = func() time.Time { return repo.now }
	service.newUUID = func() uuid.UUID { return uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee") }
	return service, repo, auditRepo, outboxRepo
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
	sessions  map[uuid.UUID]CashierSession
	cashSales map[uuid.UUID]int64
	now       time.Time
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
