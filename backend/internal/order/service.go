package order

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/inventory"
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
	LockActiveReservationsByOrder(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) ([]StockReservation, error)
	LockStockSnapshots(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productIDs []uuid.UUID) ([]StockSnapshot, error)
	UpdateStatus(ctx context.Context, q db.Queryer, params UpdateStatusParams) (*Order, error)
	UpdateStockSnapshot(ctx context.Context, q db.Queryer, params UpdateStockSnapshotParams) error
	ReleaseReservations(ctx context.Context, q db.Queryer, params ReleaseReservationsParams) error
	CreateStockMovement(ctx context.Context, q db.Queryer, params CreateStockMovementParams) error
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

type CancelInput struct {
	ActorUserID uuid.UUID
	Reason      string
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

func (s *Service) Cancel(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
	input CancelInput,
) (CancelResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return CancelResponse{}, err
	}
	if orderID == uuid.Nil {
		return CancelResponse{}, invalidField("order_id", "Order is required")
	}
	if input.ActorUserID == uuid.Nil {
		return CancelResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	reason := strings.TrimSpace(input.Reason)
	note := strings.TrimSpace(input.Note)
	if reason == "" {
		return CancelResponse{}, invalidField("reason", "Cancel reason is required")
	}
	if len(reason) > 200 {
		return CancelResponse{}, invalidField("reason", "Reason must be 200 characters or fewer")
	}
	if len(note) > 500 {
		return CancelResponse{}, invalidField("note", "Note must be 500 characters or fewer")
	}

	var response *CancelResponse
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.orders.LockByID(ctx, tx, tenantID, storeID, orderID)
		if err != nil {
			if errors.Is(err, ErrOrderNotFound) {
				return apperror.NotFound("Order not found")
			}
			return apperror.Internal(err)
		}

		if current.Status == StatusCancelled {
			cancelled := NewCancelResponse(*current, 0, 0)
			response = &cancelled
			return nil
		}
		if !canCancelStatus(current.Status) {
			return invalidStatus("Order cannot be cancelled in its current status", "status", current.Status)
		}

		reservations, err := s.orders.LockActiveReservationsByOrder(ctx, tx, tenantID, storeID, orderID)
		if err != nil {
			return apperror.Internal(err)
		}

		releasedQuantity, err := s.releaseStockReservations(ctx, tx, tenantID, storeID, orderID, input.ActorUserID, reservations)
		if err != nil {
			return err
		}

		updated, err := s.orders.UpdateStatus(ctx, tx, UpdateStatusParams{
			TenantID: tenantID,
			StoreID:  storeID,
			OrderID:  orderID,
			Status:   StatusCancelled,
		})
		if err != nil {
			if errors.Is(err, ErrOrderNotFound) {
				return apperror.NotFound("Order not found")
			}
			return apperror.Internal(err)
		}

		logNote := cancelLogNote(reason, note)
		if _, err := s.orders.CreateStatusLog(ctx, tx, CreateStatusLogParams{
			TenantID:   tenantID,
			OrderID:    orderID,
			FromStatus: current.Status,
			ToStatus:   StatusCancelled,
			Note:       logNote,
			CreatedBy:  input.ActorUserID,
		}); err != nil {
			return apperror.Internal(err)
		}

		if err := s.insertCancelEvents(ctx, tx, *updated, input.ActorUserID, reason, note, reservations, releasedQuantity); err != nil {
			return err
		}

		cancelled := NewCancelResponse(*updated, len(reservations), releasedQuantity)
		response = &cancelled
		return nil
	})
	if err != nil {
		return CancelResponse{}, err
	}
	if response == nil {
		return CancelResponse{}, apperror.Internal(errors.New("cancel response is nil"))
	}

	return *response, nil
}

func (s *Service) releaseStockReservations(
	ctx context.Context,
	tx db.Tx,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
	actorUserID uuid.UUID,
	reservations []StockReservation,
) (int, error) {
	if len(reservations) == 0 {
		return 0, nil
	}

	releaseByProduct := make(map[uuid.UUID]int)
	reservationIDs := make([]uuid.UUID, 0, len(reservations))
	for _, reservation := range reservations {
		if reservation.TenantID != tenantID || reservation.StoreID != storeID || reservation.OrderID != orderID {
			return 0, apperror.NotFound("Order not found")
		}
		releaseByProduct[reservation.ProductID] += reservation.Quantity
		reservationIDs = append(reservationIDs, reservation.ID)
	}

	productIDs := make([]uuid.UUID, 0, len(releaseByProduct))
	for productID := range releaseByProduct {
		productIDs = append(productIDs, productID)
	}
	sort.Slice(productIDs, func(i, j int) bool {
		return strings.Compare(productIDs[i].String(), productIDs[j].String()) < 0
	})

	snapshots, err := s.orders.LockStockSnapshots(ctx, tx, tenantID, storeID, productIDs)
	if err != nil {
		return 0, apperror.Internal(err)
	}
	snapshotByProduct := make(map[uuid.UUID]StockSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotByProduct[snapshot.ProductID] = snapshot
	}

	totalReleased := 0
	for _, productID := range productIDs {
		quantity := releaseByProduct[productID]
		snapshot, ok := snapshotByProduct[productID]
		if !ok || snapshot.TenantID != tenantID || snapshot.StoreID != storeID {
			return 0, apperror.Internal(ErrStockSnapshotNotFound)
		}
		if quantity > snapshot.QuantityReserved {
			return 0, apperror.Conflict("Stock reservation exceeds reserved stock")
		}

		newReserved := snapshot.QuantityReserved - quantity
		newAvailable := snapshot.QuantityAvailable + quantity
		if err := s.orders.UpdateStockSnapshot(ctx, tx, UpdateStockSnapshotParams{
			TenantID:          tenantID,
			StoreID:           storeID,
			ProductID:         productID,
			QuantityReserved:  newReserved,
			QuantityAvailable: newAvailable,
		}); err != nil {
			if errors.Is(err, ErrStockSnapshotNotFound) {
				return 0, apperror.Internal(err)
			}
			return 0, apperror.Internal(err)
		}

		if err := s.orders.CreateStockMovement(ctx, tx, CreateStockMovementParams{
			TenantID:      tenantID,
			StoreID:       storeID,
			ProductID:     productID,
			MovementType:  inventory.MovementTypeReleased,
			Quantity:      quantity,
			BalanceAfter:  newAvailable,
			ReferenceType: AggregateOrder,
			ReferenceID:   orderID,
			Note:          "Stock reservation released because order was cancelled",
			CreatedBy:     actorUserID,
		}); err != nil {
			return 0, apperror.Internal(err)
		}
		totalReleased += quantity
	}

	if err := s.orders.ReleaseReservations(ctx, tx, ReleaseReservationsParams{
		TenantID:       tenantID,
		StoreID:        storeID,
		ReservationIDs: reservationIDs,
		Status:         ReservationStatusReleased,
	}); err != nil {
		return 0, apperror.Internal(err)
	}

	return totalReleased, nil
}

func (s *Service) insertCancelEvents(
	ctx context.Context,
	tx db.Tx,
	orderRecord Order,
	actorUserID uuid.UUID,
	reason string,
	note string,
	reservations []StockReservation,
	releasedQuantity int,
) error {
	stockItems := make([]map[string]any, 0, len(reservations))
	for _, reservation := range reservations {
		stockItems = append(stockItems, map[string]any{
			"reservation_id": reservation.ID.String(),
			"product_id":     reservation.ProductID.String(),
			"quantity":       reservation.Quantity,
		})
	}

	orderPayload, err := json.Marshal(map[string]any{
		"tenant_id":     orderRecord.TenantID.String(),
		"store_id":      orderRecord.StoreID.String(),
		"order_id":      orderRecord.ID.String(),
		"order_number":  orderRecord.OrderNumber,
		"actor_user_id": actorUserID.String(),
		"reason":        reason,
		"note":          note,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	stockPayload, err := json.Marshal(map[string]any{
		"tenant_id":         orderRecord.TenantID.String(),
		"store_id":          orderRecord.StoreID.String(),
		"order_id":          orderRecord.ID.String(),
		"released_quantity": releasedQuantity,
		"items":             stockItems,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	notificationPayload, err := json.Marshal(map[string]any{
		"tenant_id": orderRecord.TenantID.String(),
		"store_id":  orderRecord.StoreID.String(),
		"order_id":  orderRecord.ID.String(),
		"type":      "order_cancelled",
	})
	if err != nil {
		return apperror.Internal(err)
	}

	for _, event := range []outbox.InsertEventParams{
		{TenantID: orderRecord.TenantID, EventType: EventOrderCancelled, AggregateType: AggregateOrder, AggregateID: orderRecord.ID, Payload: orderPayload},
		{TenantID: orderRecord.TenantID, EventType: EventStockReservationReleased, AggregateType: AggregateOrder, AggregateID: orderRecord.ID, Payload: stockPayload},
		{TenantID: orderRecord.TenantID, EventType: EventNotificationRequested, AggregateType: AggregateOrder, AggregateID: orderRecord.ID, Payload: notificationPayload},
	} {
		if _, err := s.outbox.Insert(ctx, tx, event); err != nil {
			return apperror.Internal(err)
		}
	}

	return nil
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

func canCancelStatus(status string) bool {
	switch status {
	case StatusPending, StatusConfirmed, StatusProcessing, StatusReadyToShip:
		return true
	default:
		return false
	}
}

func cancelLogNote(reason string, note string) string {
	if note == "" {
		return "Cancel reason: " + reason
	}
	return "Cancel reason: " + reason + ". " + note
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
