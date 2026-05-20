package order

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type orderStore interface {
	List(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]Order, []int, error)
	FindByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*Order, error)
	LockByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*Order, error)
	ListItems(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID) ([]Item, error)
	ListStatusLogs(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID) ([]StatusLog, error)
	ListReservationSummary(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) ([]ReservationSummary, error)
	UpdateStatus(ctx context.Context, q db.Queryer, params UpdateStatusParams) (*Order, error)
	CreateStatusLog(ctx context.Context, q db.Queryer, params CreateStatusLogParams) (*StatusLog, error)
}

type outboxStore interface {
	Insert(ctx context.Context, q db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	db     database
	orders orderStore
	outbox outboxStore
}

type UpdateStatusInput struct {
	ActorUserID uuid.UUID
	Status      string
	Note        string
}

func NewService(database database, orders orderStore, outbox outboxStore) *Service {
	return &Service{
		db:     database,
		orders: orders,
		outbox: outbox,
	}
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]ListItemResponse, PaginationMeta, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, PaginationMeta{}, err
	}

	normalized, err := normalizeListFilters(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	orders, itemCounts, err := s.orders.List(ctx, s.db, tenantID, storeID, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(orders) > normalized.Limit
	if hasMore {
		orders = orders[:normalized.Limit]
		itemCounts = itemCounts[:normalized.Limit]
	}

	response := make([]ListItemResponse, 0, len(orders))
	for idx := range orders {
		response = append(response, NewListItemResponse(orders[idx], itemCounts[idx]))
	}

	var nextCursor *string
	if hasMore && len(orders) > 0 {
		encoded, err := EncodeCursor(orders[len(orders)-1])
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

func (s *Service) Detail(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (DetailResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return DetailResponse{}, err
	}
	if orderID == uuid.Nil {
		return DetailResponse{}, invalidField("order_id", "Order is required")
	}

	orderRecord, err := s.orders.FindByID(ctx, s.db, tenantID, storeID, orderID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			return DetailResponse{}, apperror.NotFound("Order not found")
		}
		return DetailResponse{}, apperror.Internal(err)
	}

	items, err := s.orders.ListItems(ctx, s.db, tenantID, orderID)
	if err != nil {
		return DetailResponse{}, apperror.Internal(err)
	}
	logs, err := s.orders.ListStatusLogs(ctx, s.db, tenantID, orderID)
	if err != nil {
		return DetailResponse{}, apperror.Internal(err)
	}
	reservations, err := s.orders.ListReservationSummary(ctx, s.db, tenantID, storeID, orderID)
	if err != nil {
		return DetailResponse{}, apperror.Internal(err)
	}

	return NewDetailResponse(*orderRecord, items, logs, reservations), nil
}

func (s *Service) UpdateStatus(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
	input UpdateStatusInput,
) (UpdateStatusResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return UpdateStatusResponse{}, err
	}
	if orderID == uuid.Nil {
		return UpdateStatusResponse{}, invalidField("order_id", "Order is required")
	}

	targetStatus := strings.TrimSpace(input.Status)
	if !knownStatus(targetStatus) {
		return UpdateStatusResponse{}, invalidStatus("Order status is not supported", "status", targetStatus)
	}
	note := strings.TrimSpace(input.Note)
	if len(note) > 500 {
		return UpdateStatusResponse{}, invalidField("note", "Note must be 500 characters or fewer")
	}

	var updated *Order
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.orders.LockByID(ctx, tx, tenantID, storeID, orderID)
		if err != nil {
			if errors.Is(err, ErrOrderNotFound) {
				return apperror.NotFound("Order not found")
			}
			return apperror.Internal(err)
		}

		if current.Status == targetStatus {
			updated = current
			return nil
		}

		if !canTransition(current.Status, targetStatus) {
			return invalidStatus("Invalid order status transition", "status", targetStatus)
		}

		updatedOrder, err := s.orders.UpdateStatus(ctx, tx, UpdateStatusParams{
			TenantID: tenantID,
			StoreID:  storeID,
			OrderID:  orderID,
			Status:   targetStatus,
		})
		if err != nil {
			if errors.Is(err, ErrOrderNotFound) {
				return apperror.NotFound("Order not found")
			}
			return apperror.Internal(err)
		}

		if _, err := s.orders.CreateStatusLog(ctx, tx, CreateStatusLogParams{
			TenantID:   tenantID,
			OrderID:    orderID,
			FromStatus: current.Status,
			ToStatus:   targetStatus,
			Note:       note,
			CreatedBy:  input.ActorUserID,
		}); err != nil {
			return apperror.Internal(err)
		}

		if err := s.insertStatusUpdatedEvent(ctx, tx, *updatedOrder, current.Status, input.ActorUserID, note); err != nil {
			return err
		}

		updated = updatedOrder
		return nil
	})
	if err != nil {
		return UpdateStatusResponse{}, err
	}
	if updated == nil {
		return UpdateStatusResponse{}, apperror.Internal(errors.New("updated order is nil"))
	}

	return NewUpdateStatusResponse(*updated), nil
}

func (s *Service) insertStatusUpdatedEvent(
	ctx context.Context,
	tx db.Tx,
	orderRecord Order,
	fromStatus string,
	actorUserID uuid.UUID,
	note string,
) error {
	payload, err := json.Marshal(map[string]any{
		"tenant_id":     orderRecord.TenantID.String(),
		"store_id":      orderRecord.StoreID.String(),
		"order_id":      orderRecord.ID.String(),
		"order_number":  orderRecord.OrderNumber,
		"from_status":   fromStatus,
		"to_status":     orderRecord.Status,
		"actor_user_id": actorUserID.String(),
		"note":          note,
	})
	if err != nil {
		return apperror.Internal(err)
	}

	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      orderRecord.TenantID,
		EventType:     EventOrderStatusUpdated,
		AggregateType: AggregateOrder,
		AggregateID:   orderRecord.ID,
		Payload:       payload,
	}); err != nil {
		return apperror.Internal(err)
	}

	return nil
}

func normalizeListFilters(filters ListFilters) (ListFilters, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	if filters.Status != nil {
		status := strings.TrimSpace(*filters.Status)
		if !knownStatus(status) {
			return ListFilters{}, invalidStatus("Order status is not supported", "status", status)
		}
		filters.Status = &status
	}
	if filters.PaymentStatus != nil {
		paymentStatus := strings.TrimSpace(*filters.PaymentStatus)
		if !knownPaymentStatus(paymentStatus) {
			return ListFilters{}, apperror.Validation("Validation failed", []map[string]string{{"field": "payment_status", "message": "payment_status is not supported"}})
		}
		filters.PaymentStatus = &paymentStatus
	}
	if filters.Source != nil {
		source := strings.TrimSpace(*filters.Source)
		if source == "" {
			filters.Source = nil
		} else {
			filters.Source = &source
		}
	}
	if filters.Limit <= 0 {
		filters.Limit = defaultListLimit
	}
	if filters.Limit > maxListLimit {
		filters.Limit = maxListLimit
	}
	return filters, nil
}

func canTransition(from string, to string) bool {
	return allowedTransitions[from][to]
}

func knownStatus(status string) bool {
	switch status {
	case StatusPending, StatusConfirmed, StatusProcessing, StatusReadyToShip, StatusShipped, StatusCompleted, StatusCancelled:
		return true
	default:
		return false
	}
}

func knownPaymentStatus(status string) bool {
	switch status {
	case PaymentStatusUnpaid, PaymentStatusWaitingConfirmation, PaymentStatusPaid, PaymentStatusFailed, PaymentStatusRefunded:
		return true
	default:
		return false
	}
}

var allowedTransitions = map[string]map[string]bool{
	StatusPending: {
		StatusConfirmed: true,
		StatusCancelled: true,
	},
	StatusConfirmed: {
		StatusProcessing: true,
		StatusCancelled:  true,
	},
	StatusProcessing: {
		StatusReadyToShip: true,
		StatusCancelled:   true,
	},
	StatusReadyToShip: {
		StatusShipped: true,
	},
	StatusShipped: {
		StatusCompleted: true,
	},
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

func invalidStatus(message string, field string, status string) error {
	return apperror.InvalidOrderStatus(message, []map[string]string{{"field": field, "message": "unsupported or invalid status: " + status}})
}
