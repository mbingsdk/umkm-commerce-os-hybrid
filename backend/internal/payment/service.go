package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/idempotency"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

const idempotencyLockTTL = 5 * time.Minute

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type storeResolver interface {
	Resolve(ctx context.Context, slug string) (store.PublicContext, error)
}

type idempotencyStore interface {
	Begin(ctx context.Context, q db.Queryer, tenantID uuid.UUID, scope string, key string, requestHash string, lockedUntil time.Time) (*idempotency.State, error)
	SaveCompletedResponse(ctx context.Context, q db.Queryer, tenantID uuid.UUID, scope string, key string, statusCode int, responseBody json.RawMessage) error
}

type paymentStore interface {
	FindOrderByPublicReference(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderNumber string, customerPhone string) (*order.Order, error)
	LockOrderByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*order.Order, error)
	CreateConfirmation(ctx context.Context, q db.Queryer, params CreateConfirmationParams) (*Confirmation, error)
	ListConfirmations(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) ([]Confirmation, error)
	FindPendingConfirmation(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, confirmationID *uuid.UUID) (*Confirmation, error)
	MarkConfirmationReviewed(ctx context.Context, q db.Queryer, params ReviewConfirmationParams) (*Confirmation, error)
	UpdateOrderPaymentWaiting(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) error
	UpdateOrderPaid(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*order.Order, error)
	CreatePayment(ctx context.Context, q db.Queryer, params CreatePaymentParams) (*Payment, error)
	CreateOrderStatusLog(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID, fromStatus string, toStatus string, note string, actorID uuid.UUID) error
}

type outboxStore interface {
	Insert(ctx context.Context, q db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	db          database
	stores      storeResolver
	payments    paymentStore
	idempotency idempotencyStore
	outbox      outboxStore
	now         func() time.Time
}

type PublicConfirmationCommand struct {
	StoreSlug      string
	OrderNumber    string
	IdempotencyKey string
	Method         string
	Path           string
	RawBody        []byte
	Request        PublicConfirmationRequest
}

type ConfirmInput struct {
	ActorUserID    uuid.UUID
	ConfirmationID *uuid.UUID
	Note           string
}

type RejectInput struct {
	ActorUserID    uuid.UUID
	ConfirmationID *uuid.UUID
	Note           string
}

type normalizedPublicConfirmation struct {
	CustomerPhone  string
	PayerName      string
	BankName       string
	TransferAmount int64
	TransferDate   time.Time
	ProofURL       string
	Note           string
}

func NewService(database database, stores storeResolver, payments paymentStore, idempotency idempotencyStore, outbox outboxStore) *Service {
	return &Service{
		db:          database,
		stores:      stores,
		payments:    payments,
		idempotency: idempotency,
		outbox:      outbox,
		now:         time.Now,
	}
}

func (s *Service) PublicConfirm(ctx context.Context, cmd PublicConfirmationCommand) (PublicConfirmationResponse, int, error) {
	idempotencyKey := strings.TrimSpace(cmd.IdempotencyKey)
	if idempotencyKey == "" {
		return PublicConfirmationResponse{}, 0, apperror.Validation("Validation failed", []map[string]string{{"field": "Idempotency-Key", "message": "header is required"}})
	}

	currentStore, err := s.stores.Resolve(ctx, cmd.StoreSlug)
	if err != nil {
		return PublicConfirmationResponse{}, 0, err
	}

	requestHash, err := idempotency.RequestHash(cmd.Method, cmd.Path, cmd.RawBody)
	if err != nil {
		return PublicConfirmationResponse{}, 0, apperror.Validation("Invalid JSON payload", []map[string]string{{"field": "body", "message": err.Error()}})
	}

	normalized, err := normalizePublicConfirmation(cmd.Request)
	if err != nil {
		return PublicConfirmationResponse{}, 0, err
	}

	var response PublicConfirmationResponse
	statusCode := http.StatusCreated
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		state, err := s.idempotency.Begin(ctx, tx, currentStore.TenantID, idempotency.ScopePaymentConfirmation, idempotencyKey, requestHash, s.now().UTC().Add(idempotencyLockTTL))
		if err != nil {
			return err
		}
		if state.CanReplay {
			if err := json.Unmarshal(state.ResponseBody, &response); err != nil {
				return apperror.Internal(err)
			}
			statusCode = state.StatusCode
			if statusCode == 0 {
				statusCode = http.StatusCreated
			}
			return nil
		}
		if state.IsProcessing && !state.Created {
			return apperror.Conflict("Payment confirmation is still processing")
		}

		orderRecord, err := s.payments.FindOrderByPublicReference(ctx, tx, currentStore.TenantID, currentStore.StoreID, strings.TrimSpace(cmd.OrderNumber), normalized.CustomerPhone)
		if err != nil {
			if errors.Is(err, ErrOrderNotFound) {
				return apperror.NotFound("Order not found")
			}
			return apperror.Internal(err)
		}
		if orderRecord.Status == order.StatusCancelled {
			return apperror.Conflict("Order is already cancelled")
		}
		if orderRecord.PaymentStatus == order.PaymentStatusPaid {
			return apperror.Conflict("Order is already paid")
		}

		confirmation, err := s.payments.CreateConfirmation(ctx, tx, CreateConfirmationParams{
			TenantID:       currentStore.TenantID,
			StoreID:        currentStore.StoreID,
			OrderID:        orderRecord.ID,
			PayerName:      normalized.PayerName,
			BankName:       normalized.BankName,
			TransferAmount: normalized.TransferAmount,
			TransferDate:   normalized.TransferDate,
			ProofURL:       normalized.ProofURL,
			Note:           normalized.Note,
		})
		if err != nil {
			return apperror.Internal(err)
		}
		if err := s.payments.UpdateOrderPaymentWaiting(ctx, tx, currentStore.TenantID, currentStore.StoreID, orderRecord.ID); err != nil {
			return apperror.Internal(err)
		}

		response = PublicConfirmationResponse{
			ID:          confirmation.ID,
			OrderID:     orderRecord.ID,
			OrderNumber: orderRecord.OrderNumber,
			Status:      confirmation.Status,
			Message:     "Konfirmasi pembayaran diterima dan menunggu review penjual.",
		}

		responseBody, err := json.Marshal(response)
		if err != nil {
			return apperror.Internal(err)
		}
		return s.idempotency.SaveCompletedResponse(ctx, tx, currentStore.TenantID, idempotency.ScopePaymentConfirmation, idempotencyKey, statusCode, responseBody)
	})
	if err != nil {
		return PublicConfirmationResponse{}, 0, err
	}

	return response, statusCode, nil
}

func (s *Service) ListConfirmations(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) ([]ConfirmationResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, err
	}
	if orderID == uuid.Nil {
		return nil, invalidField("order_id", "Order is required")
	}

	if _, err := s.payments.LockOrderByID(ctx, s.db, tenantID, storeID, orderID); err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			return nil, apperror.NotFound("Order not found")
		}
		return nil, apperror.Internal(err)
	}

	items, err := s.payments.ListConfirmations(ctx, s.db, tenantID, storeID, orderID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	response := make([]ConfirmationResponse, 0, len(items))
	for idx := range items {
		response = append(response, NewConfirmationResponse(items[idx]))
	}
	return response, nil
}

func (s *Service) ConfirmPayment(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, input ConfirmInput) (ConfirmPaymentResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ConfirmPaymentResponse{}, err
	}
	if orderID == uuid.Nil {
		return ConfirmPaymentResponse{}, invalidField("order_id", "Order is required")
	}
	if input.ActorUserID == uuid.Nil {
		return ConfirmPaymentResponse{}, invalidField("actor_user_id", "Actor is required")
	}
	note := strings.TrimSpace(input.Note)

	var response ConfirmPaymentResponse
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		orderRecord, err := s.payments.LockOrderByID(ctx, tx, tenantID, storeID, orderID)
		if err != nil {
			if errors.Is(err, ErrOrderNotFound) {
				return apperror.NotFound("Order not found")
			}
			return apperror.Internal(err)
		}
		if orderRecord.TenantID != tenantID || orderRecord.StoreID != storeID {
			return apperror.NotFound("Order not found")
		}
		if orderRecord.Status == order.StatusCancelled {
			return apperror.Conflict("Order is already cancelled")
		}
		if orderRecord.PaymentStatus == order.PaymentStatusPaid {
			return apperror.Conflict("Order is already paid")
		}

		confirmation, err := s.payments.FindPendingConfirmation(ctx, tx, tenantID, storeID, orderID, input.ConfirmationID)
		if err != nil {
			if errors.Is(err, ErrConfirmationNotFound) {
				return apperror.NotFound("Payment confirmation not found")
			}
			return apperror.Internal(err)
		}
		if confirmation.TransferAmount < orderRecord.GrandTotal {
			return apperror.Validation("Validation failed", []map[string]string{{"field": "transfer_amount", "message": "transfer amount is less than order total"}})
		}

		confirmation, err = s.payments.MarkConfirmationReviewed(ctx, tx, ReviewConfirmationParams{
			TenantID:       tenantID,
			StoreID:        storeID,
			OrderID:        orderID,
			ConfirmationID: confirmation.ID,
			Status:         ConfirmationStatusConfirmed,
			ReviewedBy:     input.ActorUserID,
			ReviewNote:     note,
		})
		if err != nil {
			return apperror.Internal(err)
		}

		updatedOrder, err := s.payments.UpdateOrderPaid(ctx, tx, tenantID, storeID, orderID)
		if err != nil {
			return apperror.Internal(err)
		}
		paymentRecord, err := s.payments.CreatePayment(ctx, tx, CreatePaymentParams{
			TenantID:              tenantID,
			StoreID:               storeID,
			OrderID:               orderID,
			PaymentConfirmationID: confirmation.ID,
			Method:                MethodManualTransfer,
			Status:                PaymentStatusPaid,
			Amount:                confirmation.TransferAmount,
			PayerName:             confirmation.PayerName,
			BankName:              confirmation.BankName,
			ProofURL:              confirmation.ProofURL,
			Note:                  note,
			ConfirmedBy:           input.ActorUserID,
		})
		if err != nil {
			return apperror.Internal(err)
		}

		if orderRecord.Status != updatedOrder.Status {
			if err := s.payments.CreateOrderStatusLog(ctx, tx, tenantID, orderID, orderRecord.Status, updatedOrder.Status, "Payment confirmed", input.ActorUserID); err != nil {
				return apperror.Internal(err)
			}
		}
		if err := s.insertConfirmOutboxEvents(ctx, tx, *updatedOrder, *paymentRecord); err != nil {
			return err
		}

		response = ConfirmPaymentResponse{
			OrderID:       updatedOrder.ID,
			OrderNumber:   updatedOrder.OrderNumber,
			OrderStatus:   updatedOrder.Status,
			PaymentStatus: updatedOrder.PaymentStatus,
			PaymentID:     paymentRecord.ID,
		}
		return nil
	})
	if err != nil {
		return ConfirmPaymentResponse{}, err
	}
	return response, nil
}

func (s *Service) RejectPayment(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, input RejectInput) (RejectPaymentResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return RejectPaymentResponse{}, err
	}
	if orderID == uuid.Nil {
		return RejectPaymentResponse{}, invalidField("order_id", "Order is required")
	}
	if input.ActorUserID == uuid.Nil {
		return RejectPaymentResponse{}, invalidField("actor_user_id", "Actor is required")
	}
	note := strings.TrimSpace(input.Note)

	var response RejectPaymentResponse
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		orderRecord, err := s.payments.LockOrderByID(ctx, tx, tenantID, storeID, orderID)
		if err != nil {
			if errors.Is(err, ErrOrderNotFound) {
				return apperror.NotFound("Order not found")
			}
			return apperror.Internal(err)
		}
		confirmation, err := s.payments.FindPendingConfirmation(ctx, tx, tenantID, storeID, orderID, input.ConfirmationID)
		if err != nil {
			if errors.Is(err, ErrConfirmationNotFound) {
				return apperror.NotFound("Payment confirmation not found")
			}
			return apperror.Internal(err)
		}
		confirmation, err = s.payments.MarkConfirmationReviewed(ctx, tx, ReviewConfirmationParams{
			TenantID:       tenantID,
			StoreID:        storeID,
			OrderID:        orderID,
			ConfirmationID: confirmation.ID,
			Status:         ConfirmationStatusRejected,
			ReviewedBy:     input.ActorUserID,
			ReviewNote:     note,
		})
		if err != nil {
			return apperror.Internal(err)
		}
		if err := s.insertNotificationEvent(ctx, tx, *orderRecord, "payment_rejected"); err != nil {
			return err
		}
		response = RejectPaymentResponse{
			OrderID:        orderRecord.ID,
			OrderNumber:    orderRecord.OrderNumber,
			PaymentStatus:  orderRecord.PaymentStatus,
			ConfirmationID: confirmation.ID,
			Status:         confirmation.Status,
		}
		return nil
	})
	if err != nil {
		return RejectPaymentResponse{}, err
	}
	return response, nil
}

func (s *Service) insertConfirmOutboxEvents(ctx context.Context, tx db.Tx, orderRecord order.Order, paymentRecord Payment) error {
	paymentPayload, err := json.Marshal(map[string]any{
		"tenant_id":  orderRecord.TenantID.String(),
		"store_id":   orderRecord.StoreID.String(),
		"order_id":   orderRecord.ID.String(),
		"payment_id": paymentRecord.ID.String(),
		"amount":     paymentRecord.Amount,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	orderPayload, err := json.Marshal(map[string]any{
		"tenant_id":    orderRecord.TenantID.String(),
		"store_id":     orderRecord.StoreID.String(),
		"order_id":     orderRecord.ID.String(),
		"order_number": orderRecord.OrderNumber,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	notificationPayload, err := json.Marshal(map[string]any{
		"tenant_id": orderRecord.TenantID.String(),
		"store_id":  orderRecord.StoreID.String(),
		"order_id":  orderRecord.ID.String(),
		"type":      "payment_confirmed",
	})
	if err != nil {
		return apperror.Internal(err)
	}

	for _, event := range []outbox.InsertEventParams{
		{TenantID: orderRecord.TenantID, EventType: EventPaymentConfirmed, AggregateType: AggregatePayment, AggregateID: paymentRecord.ID, Payload: paymentPayload},
		{TenantID: orderRecord.TenantID, EventType: EventOrderPaid, AggregateType: AggregateOrder, AggregateID: orderRecord.ID, Payload: orderPayload},
		{TenantID: orderRecord.TenantID, EventType: EventNotificationRequested, AggregateType: AggregateOrder, AggregateID: orderRecord.ID, Payload: notificationPayload},
	} {
		if _, err := s.outbox.Insert(ctx, tx, event); err != nil {
			return apperror.Internal(err)
		}
	}
	return nil
}

func (s *Service) insertNotificationEvent(ctx context.Context, tx db.Tx, orderRecord order.Order, notificationType string) error {
	payload, err := json.Marshal(map[string]any{
		"tenant_id": orderRecord.TenantID.String(),
		"store_id":  orderRecord.StoreID.String(),
		"order_id":  orderRecord.ID.String(),
		"type":      notificationType,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{TenantID: orderRecord.TenantID, EventType: EventNotificationRequested, AggregateType: AggregateOrder, AggregateID: orderRecord.ID, Payload: payload}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func normalizePublicConfirmation(request PublicConfirmationRequest) (normalizedPublicConfirmation, error) {
	result := normalizedPublicConfirmation{
		CustomerPhone:  strings.TrimSpace(request.CustomerPhone),
		PayerName:      strings.TrimSpace(request.PayerName),
		BankName:       strings.TrimSpace(request.BankName),
		TransferAmount: request.TransferAmount,
		ProofURL:       strings.TrimSpace(request.ProofURL),
		Note:           strings.TrimSpace(request.Note),
	}

	var details []map[string]string
	if result.CustomerPhone == "" {
		details = append(details, map[string]string{"field": "customer_phone", "message": "customer_phone is required"})
	}
	if result.PayerName == "" {
		details = append(details, map[string]string{"field": "payer_name", "message": "payer_name is required"})
	}
	if result.BankName == "" {
		details = append(details, map[string]string{"field": "bank_name", "message": "bank_name is required"})
	}
	if result.TransferAmount <= 0 {
		details = append(details, map[string]string{"field": "transfer_amount", "message": "transfer_amount must be greater than zero"})
	}
	transferDate, err := parseTransferDate(strings.TrimSpace(request.TransferDate))
	if err != nil {
		details = append(details, map[string]string{"field": "transfer_date", "message": "transfer_date must be YYYY-MM-DD or RFC3339"})
	} else {
		result.TransferDate = transferDate
	}
	if len(result.ProofURL) > 500 {
		details = append(details, map[string]string{"field": "proof_url", "message": "proof_url is too long"})
	}
	if len(result.Note) > 500 {
		details = append(details, map[string]string{"field": "note", "message": "note must be 500 characters or fewer"})
	}
	if len(details) > 0 {
		return normalizedPublicConfirmation{}, apperror.Validation("Validation failed", details)
	}
	return result, nil
}

func parseTransferDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, errors.New("transfer date is required")
	}
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", raw)
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
