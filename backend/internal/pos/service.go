package pos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

const maxSessionNoteLength = 500

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
}

type auditStore interface {
	Create(context.Context, db.Queryer, audit.Entry) error
}

type outboxStore interface {
	Insert(context.Context, db.Queryer, outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	db        database
	sessions  sessionStore
	auditLogs auditStore
	outbox    outboxStore
	now       func() time.Time
	newUUID   func() uuid.UUID
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

func NewService(database database, sessions sessionStore, auditLogs auditStore, outboxRepo outboxStore) *Service {
	return &Service{
		db:        database,
		sessions:  sessions,
		auditLogs: auditLogs,
		outbox:    outboxRepo,
		now:       time.Now,
		newUUID:   uuid.New,
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

func (s *Service) generateSessionNumber() string {
	datePart := s.now().UTC().Format("20060102")
	randomPart := strings.ToUpper(strings.ReplaceAll(s.newUUID().String(), "-", ""))[:8]
	return fmt.Sprintf("CS-%s-%s", datePart, randomPart)
}

func canCloseAnySession(role string) bool {
	switch permission.Role(role) {
	case permission.RoleOwner, permission.RoleManager:
		return true
	default:
		return false
	}
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
