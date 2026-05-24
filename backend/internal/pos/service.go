package pos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

const maxSessionNoteLength = 500
const (
	defaultPOSListLimit        = 20
	maxPOSListLimit            = 100
	maxPOSTransactionBodyBytes = 1 << 20
	idempotencyLockTTL         = 5 * time.Minute
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type sessionStore interface {
	FindCurrentOpenByCashier(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*CashierSession, error)
	CreateSession(context.Context, db.Queryer, CreateSessionParams) (*CashierSession, error)
	LockSessionByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*CashierSession, error)
	SumCompletedCashTransactions(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (int64, error)
	CloseSession(context.Context, db.Queryer, CloseSessionParams) (*CashierSession, error)
	ListProducts(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ProductSearchFilters) ([]POSProduct, error)
	ListProductsByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, []uuid.UUID) ([]POSProduct, error)
	LockStockSnapshots(context.Context, db.Queryer, uuid.UUID, uuid.UUID, []uuid.UUID) ([]StockSnapshot, error)
	CreateTransaction(context.Context, db.Queryer, CreateTransactionParams) (*POSTransaction, error)
	CreateTransactionItem(context.Context, db.Queryer, CreateTransactionItemParams) error
	UpdateStockSnapshot(context.Context, db.Queryer, UpdateStockSnapshotParams) error
	CreateStockMovement(context.Context, db.Queryer, CreateStockMovementParams) error
	ListTransactions(context.Context, db.Queryer, uuid.UUID, uuid.UUID, TransactionListFilters) ([]POSTransaction, error)
	FindTransactionByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*POSTransaction, error)
	ListTransactionItems(context.Context, db.Queryer, uuid.UUID, uuid.UUID) ([]POSTransactionItem, error)
}

type auditStore interface {
	Create(context.Context, db.Queryer, audit.Entry) error
}

type outboxStore interface {
	Insert(context.Context, db.Queryer, outbox.InsertEventParams) (*outbox.Event, error)
}

type idempotencyStore interface {
	Begin(
		ctx context.Context,
		q db.Queryer,
		tenantID uuid.UUID,
		scope string,
		key string,
		requestHash string,
		lockedUntil time.Time,
	) (*idempotency.State, error)
	SaveCompletedResponse(
		ctx context.Context,
		q db.Queryer,
		tenantID uuid.UUID,
		scope string,
		key string,
		statusCode int,
		responseBody json.RawMessage,
	) error
}

type Service struct {
	db          database
	sessions    sessionStore
	auditLogs   auditStore
	idempotency idempotencyStore
	outbox      outboxStore
	now         func() time.Time
	newUUID     func() uuid.UUID
}

type OpenSessionInput struct {
	ActorUserID       uuid.UUID
	OpeningCashAmount int64
	Note              string
}

type CurrentSessionInput struct {
	ActorUserID uuid.UUID
}

type CloseSessionInput struct {
	ActorUserID       uuid.UUID
	Role              string
	ClosingCashAmount int64
	Note              string
}

type CreateTransactionCommand struct {
	TenantID       uuid.UUID
	StoreID        uuid.UUID
	ActorUserID    uuid.UUID
	Role           string
	IdempotencyKey string
	Method         string
	Path           string
	RawBody        []byte
	Request        CreateTransactionRequest
}

type CreateTransactionResult struct {
	Response   TransactionResponse
	StatusCode int
}

type normalizedTransactionItem struct {
	ProductID uuid.UUID
	Quantity  int
}

type normalizedTransactionRequest struct {
	SessionID     uuid.UUID
	Items         []normalizedTransactionItem
	PaymentMethod string
	AmountPaid    int64
	Note          string
}

func NewService(database database, sessions sessionStore, auditLogs auditStore, idempotencyRepo idempotencyStore, outboxRepo outboxStore) *Service {
	return &Service{
		db:          database,
		sessions:    sessions,
		auditLogs:   auditLogs,
		idempotency: idempotencyRepo,
		outbox:      outboxRepo,
		now:         time.Now,
		newUUID:     uuid.New,
	}
}

func (s *Service) OpenSession(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, input OpenSessionInput) (SessionResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return SessionResponse{}, err
	}
	if input.ActorUserID == uuid.Nil {
		return SessionResponse{}, invalidField("actor_user_id", "Actor is required")
	}
	if input.OpeningCashAmount < 0 {
		return SessionResponse{}, invalidField("opening_cash_amount", "Opening cash amount must be zero or greater")
	}
	note := strings.TrimSpace(input.Note)
	if len(note) > maxSessionNoteLength {
		return SessionResponse{}, invalidField("note", "Note must be 500 characters or fewer")
	}

	var created *CashierSession
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		existing, err := s.sessions.FindCurrentOpenByCashier(ctx, tx, tenantID, storeID, input.ActorUserID)
		if err != nil && !errors.Is(err, ErrSessionNotFound) {
			return apperror.Internal(err)
		}
		if existing != nil {
			return apperror.Conflict("Cashier already has an open session")
		}

		session, err := s.sessions.CreateSession(ctx, tx, CreateSessionParams{
			TenantID:      tenantID,
			StoreID:       storeID,
			CashierID:     input.ActorUserID,
			SessionNumber: s.generateSessionNumber(),
			OpeningCash:   input.OpeningCashAmount,
		})
		if err != nil {
			if errors.Is(err, ErrOpenSessionExists) {
				return apperror.Conflict("Cashier already has an open session")
			}
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    tenantID,
			StoreID:     &storeID,
			ActorUserID: &input.ActorUserID,
			Action:      AuditActionSessionOpen,
			EntityType:  AggregateCashierSession,
			EntityID:    &session.ID,
			AfterData: map[string]any{
				"session_number": session.SessionNumber,
				"opening_cash":   session.OpeningCash,
				"status":         session.Status,
				"note":           note,
			},
		}); err != nil {
			return apperror.Internal(err)
		}

		if err := s.insertSessionEvent(ctx, tx, EventCashierSessionOpened, *session, input.ActorUserID, note); err != nil {
			return err
		}

		created = session
		return nil
	})
	if err != nil {
		return SessionResponse{}, err
	}
	if created == nil {
		return SessionResponse{}, apperror.Internal(errors.New("created session is nil"))
	}

	return NewSessionResponse(*created), nil
}

func (s *Service) CurrentSession(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, input CurrentSessionInput) (SessionResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return SessionResponse{}, err
	}
	if input.ActorUserID == uuid.Nil {
		return SessionResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	session, err := s.sessions.FindCurrentOpenByCashier(ctx, s.db, tenantID, storeID, input.ActorUserID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return SessionResponse{}, apperror.NotFound("Cashier session not found")
		}
		return SessionResponse{}, apperror.Internal(err)
	}
	return NewSessionResponse(*session), nil
}

func (s *Service) ListProducts(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, filters ProductSearchFilters) ([]POSProductResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, err
	}
	normalized := normalizeProductFilters(filters)
	items, err := s.sessions.ListProducts(ctx, s.db, tenantID, storeID, normalized)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	response := make([]POSProductResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewPOSProductResponse(item))
	}
	return response, nil
}

func (s *Service) CreateTransaction(ctx context.Context, cmd CreateTransactionCommand) (CreateTransactionResult, error) {
	if err := validateScope(cmd.TenantID, cmd.StoreID); err != nil {
		return CreateTransactionResult{}, err
	}
	if cmd.ActorUserID == uuid.Nil {
		return CreateTransactionResult{}, invalidField("actor_user_id", "Actor is required")
	}
	idempotencyKey := strings.TrimSpace(cmd.IdempotencyKey)
	if idempotencyKey == "" {
		return CreateTransactionResult{}, apperror.Validation("Validation failed", []map[string]string{
			{"field": "Idempotency-Key", "message": "header is required"},
		})
	}

	requestHash, err := idempotency.RequestHash(cmd.Method, cmd.Path, cmd.RawBody)
	if err != nil {
		return CreateTransactionResult{}, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": err.Error()},
		})
	}

	var result CreateTransactionResult
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		state, err := s.idempotency.Begin(
			ctx,
			tx,
			cmd.TenantID,
			idempotency.ScopePOS,
			idempotencyKey,
			requestHash,
			s.now().UTC().Add(idempotencyLockTTL),
		)
		if err != nil {
			return err
		}
		if state.CanReplay {
			var response TransactionResponse
			if err := json.Unmarshal(state.ResponseBody, &response); err != nil {
				return apperror.Internal(err)
			}
			result = CreateTransactionResult{Response: response, StatusCode: state.StatusCode}
			if result.StatusCode == 0 {
				result.StatusCode = http.StatusCreated
			}
			return nil
		}
		if state.IsProcessing && !state.Created {
			return apperror.Conflict("POS transaction request is still processing")
		}

		normalized, err := normalizeCreateTransactionRequest(cmd.Request)
		if err != nil {
			return err
		}

		response, err := s.createTransaction(ctx, tx, cmd.TenantID, cmd.StoreID, cmd.ActorUserID, cmd.Role, normalized)
		if err != nil {
			return err
		}

		responseBody, err := json.Marshal(response)
		if err != nil {
			return apperror.Internal(err)
		}
		if err := s.idempotency.SaveCompletedResponse(
			ctx,
			tx,
			cmd.TenantID,
			idempotency.ScopePOS,
			idempotencyKey,
			http.StatusCreated,
			responseBody,
		); err != nil {
			return err
		}

		result = CreateTransactionResult{Response: response, StatusCode: http.StatusCreated}
		return nil
	})
	if err != nil {
		return CreateTransactionResult{}, err
	}
	return result, nil
}

func (s *Service) ListTransactions(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, actorUserID uuid.UUID, role string, filters TransactionListFilters) ([]TransactionResponse, PaginationMeta, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, PaginationMeta{}, err
	}
	if actorUserID == uuid.Nil {
		return nil, PaginationMeta{}, invalidField("actor_user_id", "Actor is required")
	}
	normalized, err := normalizeTransactionFilters(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	if !canAccessAnyPOSSession(role) {
		normalized.CashierID = &actorUserID
	}
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.sessions.ListTransactions(ctx, s.db, tenantID, storeID, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}
	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	response := make([]TransactionResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewTransactionResponse(item))
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeTransactionCursor(items[len(items)-1])
		if err != nil {
			return nil, PaginationMeta{}, apperror.Internal(err)
		}
		nextCursor = &encoded
	}

	return response, PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) TransactionDetail(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, actorUserID uuid.UUID, role string, transactionID uuid.UUID) (TransactionDetailResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return TransactionDetailResponse{}, err
	}
	if actorUserID == uuid.Nil {
		return TransactionDetailResponse{}, invalidField("actor_user_id", "Actor is required")
	}
	if transactionID == uuid.Nil {
		return TransactionDetailResponse{}, invalidField("transaction_id", "Transaction is required")
	}

	transaction, err := s.sessions.FindTransactionByID(ctx, s.db, tenantID, storeID, transactionID)
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			return TransactionDetailResponse{}, apperror.NotFound("POS transaction not found")
		}
		return TransactionDetailResponse{}, apperror.Internal(err)
	}
	if !canAccessAnyPOSSession(role) && transaction.CashierID != actorUserID {
		return TransactionDetailResponse{}, apperror.NotFound("POS transaction not found")
	}

	items, err := s.sessions.ListTransactionItems(ctx, s.db, tenantID, transactionID)
	if err != nil {
		return TransactionDetailResponse{}, apperror.Internal(err)
	}
	return NewTransactionDetailResponse(*transaction, items), nil
}

func (s *Service) CloseSession(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, sessionID uuid.UUID, input CloseSessionInput) (SessionResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return SessionResponse{}, err
	}
	if sessionID == uuid.Nil {
		return SessionResponse{}, invalidField("session_id", "Session is required")
	}
	if input.ActorUserID == uuid.Nil {
		return SessionResponse{}, invalidField("actor_user_id", "Actor is required")
	}
	if input.ClosingCashAmount < 0 {
		return SessionResponse{}, invalidField("closing_cash_amount", "Closing cash amount must be zero or greater")
	}
	note := strings.TrimSpace(input.Note)
	if len(note) > maxSessionNoteLength {
		return SessionResponse{}, invalidField("note", "Note must be 500 characters or fewer")
	}

	var closed *CashierSession
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.sessions.LockSessionByID(ctx, tx, tenantID, storeID, sessionID)
		if err != nil {
			if errors.Is(err, ErrSessionNotFound) {
				return apperror.NotFound("Cashier session not found")
			}
			return apperror.Internal(err)
		}
		if current.Status != SessionStatusOpen {
			return apperror.Conflict("Cashier session is already closed")
		}
		if !canCloseAnySession(input.Role) && current.CashierID != input.ActorUserID {
			return apperror.Forbidden("Cannot close another cashier's session")
		}

		cashSales, err := s.sessions.SumCompletedCashTransactions(ctx, tx, tenantID, storeID, sessionID)
		if err != nil {
			return apperror.Internal(err)
		}
		expectedCash := current.OpeningCash + cashSales
		difference := input.ClosingCashAmount - expectedCash

		updated, err := s.sessions.CloseSession(ctx, tx, CloseSessionParams{
			TenantID:     tenantID,
			StoreID:      storeID,
			SessionID:    sessionID,
			ClosingCash:  input.ClosingCashAmount,
			ExpectedCash: expectedCash,
			Difference:   difference,
		})
		if err != nil {
			if errors.Is(err, ErrSessionAlreadyDone) {
				return apperror.Conflict("Cashier session is already closed")
			}
			if errors.Is(err, ErrSessionNotFound) {
				return apperror.NotFound("Cashier session not found")
			}
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    tenantID,
			StoreID:     &storeID,
			ActorUserID: &input.ActorUserID,
			Action:      AuditActionSessionClose,
			EntityType:  AggregateCashierSession,
			EntityID:    &sessionID,
			BeforeData: map[string]any{
				"status":       current.Status,
				"opening_cash": current.OpeningCash,
			},
			AfterData: map[string]any{
				"status":        updated.Status,
				"closing_cash":  updated.ClosingCash,
				"expected_cash": updated.ExpectedCash,
				"difference":    updated.Difference,
				"note":          note,
			},
		}); err != nil {
			return apperror.Internal(err)
		}

		if err := s.insertSessionEvent(ctx, tx, EventCashierSessionClosed, *updated, input.ActorUserID, note); err != nil {
			return err
		}

		closed = updated
		return nil
	})
	if err != nil {
		return SessionResponse{}, err
	}
	if closed == nil {
		return SessionResponse{}, apperror.Internal(errors.New("closed session is nil"))
	}

	return NewSessionResponse(*closed), nil
}

func (s *Service) createTransaction(
	ctx context.Context,
	tx db.Tx,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	actorUserID uuid.UUID,
	role string,
	request normalizedTransactionRequest,
) (TransactionResponse, error) {
	session, err := s.sessions.LockSessionByID(ctx, tx, tenantID, storeID, request.SessionID)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return TransactionResponse{}, apperror.NotFound("Cashier session not found")
		}
		return TransactionResponse{}, apperror.Internal(err)
	}
	if session.Status != SessionStatusOpen {
		return TransactionResponse{}, apperror.Conflict("Cashier session is not open")
	}
	if !canAccessAnyPOSSession(role) && session.CashierID != actorUserID {
		return TransactionResponse{}, apperror.Forbidden("Cannot use another cashier's session")
	}

	productIDs := productIDsFromTransactionItems(request.Items)
	products, err := s.sessions.ListProductsByID(ctx, tx, tenantID, storeID, productIDs)
	if err != nil {
		return TransactionResponse{}, apperror.Internal(err)
	}
	productByID := make(map[uuid.UUID]POSProduct, len(products))
	for _, item := range products {
		productByID[item.ID] = item
	}

	snapshots, err := s.sessions.LockStockSnapshots(ctx, tx, tenantID, storeID, productIDs)
	if err != nil {
		return TransactionResponse{}, apperror.Internal(err)
	}
	snapshotByProduct := make(map[uuid.UUID]StockSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotByProduct[snapshot.ProductID] = snapshot
	}

	var subtotal int64
	for _, item := range request.Items {
		productRecord, ok := productByID[item.ProductID]
		if !ok || productRecord.TenantID != tenantID || productRecord.StoreID != storeID || productRecord.Status != "active" {
			return TransactionResponse{}, apperror.NotFound("Product not found")
		}
		lineSubtotal, err := multiplyMoney(productRecord.Price, item.Quantity)
		if err != nil {
			return TransactionResponse{}, err
		}
		if subtotal > math.MaxInt64-lineSubtotal {
			return TransactionResponse{}, invalidField("items", "Transaction total is too large")
		}
		subtotal += lineSubtotal

		snapshot, ok := snapshotByProduct[item.ProductID]
		if !ok || snapshot.TenantID != tenantID || snapshot.StoreID != storeID {
			return TransactionResponse{}, apperror.InsufficientStock("Insufficient stock", []map[string]any{
				{"product_id": item.ProductID.String(), "available": 0, "requested": item.Quantity},
			})
		}
		if snapshot.QuantityAvailable < item.Quantity {
			return TransactionResponse{}, apperror.InsufficientStock("Insufficient stock", []map[string]any{
				{"product_id": item.ProductID.String(), "available": snapshot.QuantityAvailable, "requested": item.Quantity},
			})
		}
	}

	discountTotal := int64(0)
	taxTotal := int64(0)
	grandTotal := subtotal - discountTotal + taxTotal
	if request.AmountPaid < grandTotal {
		return TransactionResponse{}, invalidField("amount_paid", "Amount paid must be greater than or equal to total")
	}
	changeAmount := int64(0)
	if request.PaymentMethod == PaymentMethodCash {
		changeAmount = request.AmountPaid - grandTotal
	} else if request.AmountPaid != grandTotal {
		return TransactionResponse{}, invalidField("amount_paid", "QRIS manual amount must equal total")
	}

	transaction, err := s.sessions.CreateTransaction(ctx, tx, CreateTransactionParams{
		TenantID:          tenantID,
		StoreID:           storeID,
		CashierSessionID:  session.ID,
		CashierID:         session.CashierID,
		TransactionNumber: s.generateTransactionNumber(),
		Subtotal:          subtotal,
		DiscountTotal:     discountTotal,
		TaxTotal:          taxTotal,
		GrandTotal:        grandTotal,
		PaymentMethod:     request.PaymentMethod,
		PaymentAmount:     request.AmountPaid,
		ChangeAmount:      changeAmount,
	})
	if err != nil {
		return TransactionResponse{}, apperror.Internal(err)
	}

	stockEventItems := make([]map[string]any, 0, len(request.Items))
	for _, item := range request.Items {
		productRecord := productByID[item.ProductID]
		snapshot := snapshotByProduct[item.ProductID]
		lineSubtotal, err := multiplyMoney(productRecord.Price, item.Quantity)
		if err != nil {
			return TransactionResponse{}, err
		}
		if err := s.sessions.CreateTransactionItem(ctx, tx, CreateTransactionItemParams{
			TenantID:         tenantID,
			POSTransactionID: transaction.ID,
			ProductID:        productRecord.ID,
			ProductName:      productRecord.Name,
			SKU:              productRecord.SKU,
			Quantity:         item.Quantity,
			UnitPrice:        productRecord.Price,
			DiscountTotal:    0,
			Subtotal:         lineSubtotal,
		}); err != nil {
			return TransactionResponse{}, apperror.Internal(err)
		}

		newOnHand := snapshot.QuantityOnHand - item.Quantity
		newAvailable := newOnHand - snapshot.QuantityReserved
		if newOnHand < 0 || newAvailable < 0 {
			return TransactionResponse{}, apperror.InsufficientStock("Insufficient stock", []map[string]any{
				{"product_id": item.ProductID.String(), "available": snapshot.QuantityAvailable, "requested": item.Quantity},
			})
		}
		if err := s.sessions.UpdateStockSnapshot(ctx, tx, UpdateStockSnapshotParams{
			TenantID:          tenantID,
			StoreID:           storeID,
			ProductID:         productRecord.ID,
			QuantityOnHand:    newOnHand,
			QuantityReserved:  snapshot.QuantityReserved,
			QuantityAvailable: newAvailable,
		}); err != nil {
			return TransactionResponse{}, apperror.Internal(err)
		}
		if err := s.sessions.CreateStockMovement(ctx, tx, CreateStockMovementParams{
			TenantID:      tenantID,
			StoreID:       storeID,
			ProductID:     productRecord.ID,
			MovementType:  inventory.MovementTypePOSSale,
			Quantity:      -item.Quantity,
			BalanceAfter:  newAvailable,
			ReferenceType: AggregatePOSTransaction,
			ReferenceID:   transaction.ID,
			Note:          "Stock reduced by POS transaction",
			CreatedBy:     actorUserID,
		}); err != nil {
			return TransactionResponse{}, apperror.Internal(err)
		}

		stockEventItems = append(stockEventItems, map[string]any{
			"product_id": productRecord.ID.String(),
			"quantity":   item.Quantity,
		})
	}

	if err := s.insertPOSTransactionEvents(ctx, tx, *transaction, stockEventItems, request.Note); err != nil {
		return TransactionResponse{}, err
	}

	return NewTransactionResponse(*transaction), nil
}

func (s *Service) insertSessionEvent(ctx context.Context, tx db.Tx, eventType string, session CashierSession, actorUserID uuid.UUID, note string) error {
	payload, err := json.Marshal(map[string]any{
		"tenant_id":      session.TenantID.String(),
		"store_id":       session.StoreID.String(),
		"session_id":     session.ID.String(),
		"session_number": session.SessionNumber,
		"cashier_id":     session.CashierID.String(),
		"actor_user_id":  actorUserID.String(),
		"status":         session.Status,
		"note":           note,
	})
	if err != nil {
		return apperror.Internal(err)
	}

	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      session.TenantID,
		EventType:     eventType,
		AggregateType: AggregateCashierSession,
		AggregateID:   session.ID,
		Payload:       payload,
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) insertPOSTransactionEvents(ctx context.Context, tx db.Tx, transaction POSTransaction, stockItems []map[string]any, note string) error {
	transactionPayload, err := json.Marshal(map[string]any{
		"tenant_id":          transaction.TenantID.String(),
		"store_id":           transaction.StoreID.String(),
		"transaction_id":     transaction.ID.String(),
		"transaction_number": transaction.TransactionNumber,
		"session_id":         transaction.CashierSessionID.String(),
		"cashier_id":         transaction.CashierID.String(),
		"grand_total":        transaction.GrandTotal,
		"payment_method":     transaction.PaymentMethod,
		"note":               note,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	stockPayload, err := json.Marshal(map[string]any{
		"tenant_id":      transaction.TenantID.String(),
		"store_id":       transaction.StoreID.String(),
		"transaction_id": transaction.ID.String(),
		"items":          stockItems,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	notificationPayload, err := json.Marshal(map[string]any{
		"tenant_id":      transaction.TenantID.String(),
		"store_id":       transaction.StoreID.String(),
		"transaction_id": transaction.ID.String(),
		"type":           "pos_transaction_created",
	})
	if err != nil {
		return apperror.Internal(err)
	}

	for _, event := range []outbox.InsertEventParams{
		{TenantID: transaction.TenantID, EventType: EventPOSTransactionCreated, AggregateType: AggregatePOSTransaction, AggregateID: transaction.ID, Payload: transactionPayload},
		{TenantID: transaction.TenantID, EventType: EventStockReduced, AggregateType: AggregatePOSTransaction, AggregateID: transaction.ID, Payload: stockPayload},
		{TenantID: transaction.TenantID, EventType: EventNotificationRequested, AggregateType: AggregatePOSTransaction, AggregateID: transaction.ID, Payload: notificationPayload},
	} {
		if _, err := s.outbox.Insert(ctx, tx, event); err != nil {
			return apperror.Internal(err)
		}
	}
	return nil
}

func (s *Service) generateSessionNumber() string {
	datePart := s.now().UTC().Format("20060102")
	randomPart := strings.ToUpper(strings.ReplaceAll(s.newUUID().String(), "-", ""))[:8]
	return fmt.Sprintf("CS-%s-%s", datePart, randomPart)
}

func (s *Service) generateTransactionNumber() string {
	datePart := s.now().UTC().Format("20060102")
	randomPart := strings.ToUpper(strings.ReplaceAll(s.newUUID().String(), "-", ""))[:8]
	return fmt.Sprintf("POS-%s-%s", datePart, randomPart)
}

func canCloseAnySession(role string) bool {
	return canAccessAnyPOSSession(role)
}

func canAccessAnyPOSSession(role string) bool {
	switch permission.Role(role) {
	case permission.RoleOwner, permission.RoleManager:
		return true
	default:
		return false
	}
}

func normalizeProductFilters(filters ProductSearchFilters) ProductSearchFilters {
	filters.Query = querytext.NormalizeSearch(filters.Query)
	filters.Barcode = querytext.NormalizeSearch(filters.Barcode)
	if filters.Limit <= 0 {
		filters.Limit = defaultPOSListLimit
	}
	if filters.Limit > maxPOSListLimit {
		filters.Limit = maxPOSListLimit
	}
	return filters
}

func normalizeTransactionFilters(filters TransactionListFilters) (TransactionListFilters, error) {
	if filters.PaymentMethod != nil {
		paymentMethod := strings.TrimSpace(*filters.PaymentMethod)
		if paymentMethod == "" {
			filters.PaymentMethod = nil
		} else if paymentMethod != PaymentMethodCash && paymentMethod != PaymentMethodQRISManual {
			return TransactionListFilters{}, invalidField("payment_method", "payment_method is not supported")
		} else {
			filters.PaymentMethod = &paymentMethod
		}
	}
	if filters.Limit <= 0 {
		filters.Limit = defaultPOSListLimit
	}
	if filters.Limit > maxPOSListLimit {
		filters.Limit = maxPOSListLimit
	}
	return filters, nil
}

func normalizeCreateTransactionRequest(request CreateTransactionRequest) (normalizedTransactionRequest, error) {
	details := make([]map[string]string, 0)
	sessionID := request.SessionID
	if sessionID == uuid.Nil {
		sessionID = request.CashierSessionID
	}
	if sessionID == uuid.Nil {
		details = append(details, map[string]string{"field": "session_id", "message": "session_id is required"})
	}
	if len(request.Items) == 0 {
		details = append(details, map[string]string{"field": "items", "message": "items must not be empty"})
	}

	quantityByProduct := make(map[uuid.UUID]int)
	for idx, item := range request.Items {
		fieldPrefix := fmt.Sprintf("items[%d]", idx)
		if item.ProductID == uuid.Nil {
			details = append(details, map[string]string{"field": fieldPrefix + ".product_id", "message": "product_id is required"})
			continue
		}
		if item.Quantity <= 0 {
			details = append(details, map[string]string{"field": fieldPrefix + ".quantity", "message": "quantity must be greater than zero"})
			continue
		}
		quantityByProduct[item.ProductID] += item.Quantity
	}

	paymentMethod := strings.TrimSpace(request.PaymentMethod)
	if paymentMethod != PaymentMethodCash && paymentMethod != PaymentMethodQRISManual {
		details = append(details, map[string]string{"field": "payment_method", "message": "payment_method must be cash or qris_manual"})
	}
	amountPaidPtr := request.AmountPaid
	if amountPaidPtr == nil {
		amountPaidPtr = request.PaymentAmount
	}
	if amountPaidPtr == nil {
		details = append(details, map[string]string{"field": "amount_paid", "message": "amount_paid is required"})
	} else if *amountPaidPtr < 0 {
		details = append(details, map[string]string{"field": "amount_paid", "message": "amount_paid must be zero or greater"})
	}
	note := strings.TrimSpace(request.Note)
	if len(note) > maxSessionNoteLength {
		details = append(details, map[string]string{"field": "note", "message": "Note must be 500 characters or fewer"})
	}

	if len(details) > 0 {
		return normalizedTransactionRequest{}, apperror.Validation("Validation failed", details)
	}

	items := make([]normalizedTransactionItem, 0, len(quantityByProduct))
	for productID, quantity := range quantityByProduct {
		items = append(items, normalizedTransactionItem{ProductID: productID, Quantity: quantity})
	}
	sort.Slice(items, func(i, j int) bool {
		return strings.Compare(items[i].ProductID.String(), items[j].ProductID.String()) < 0
	})

	return normalizedTransactionRequest{
		SessionID:     sessionID,
		Items:         items,
		PaymentMethod: paymentMethod,
		AmountPaid:    *amountPaidPtr,
		Note:          note,
	}, nil
}

func productIDsFromTransactionItems(items []normalizedTransactionItem) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ProductID)
	}
	return ids
}

func multiplyMoney(unitPrice int64, quantity int) (int64, error) {
	if unitPrice < 0 {
		return 0, invalidField("items", "product price is invalid")
	}
	if quantity <= 0 {
		return 0, invalidField("items.quantity", "quantity must be greater than zero")
	}
	if unitPrice > math.MaxInt64/int64(quantity) {
		return 0, invalidField("items", "line total is too large")
	}
	return unitPrice * int64(quantity), nil
}

func validateScope(tenantID uuid.UUID, storeID uuid.UUID) error {
	var details []map[string]string
	if tenantID == uuid.Nil {
		details = append(details, map[string]string{"field": "tenant_id", "message": "Tenant is required"})
	}
	if storeID == uuid.Nil {
		details = append(details, map[string]string{"field": "store_id", "message": "Store is required"})
	}
	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}
	return nil
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{{"field": field, "message": message}})
}
