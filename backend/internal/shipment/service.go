package shipment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100

	maxCourierNameLength     = 120
	maxTrackingNumberLength  = 120
	maxAssignedNameLength    = 120
	maxAssignedPhoneLength   = 40
	maxShipmentNoteLength    = 500
	defaultShipmentLogNote   = "Shipment created"
	deliveredShipmentLogNote = "Shipment delivered"
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type shipmentStore interface {
	List(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ListFilters) ([]Shipment, error)
	FindByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Shipment, error)
	LockByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Shipment, error)
	FindLatestByOrder(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Shipment, error)
	Create(context.Context, db.Queryer, CreateShipmentParams) (*Shipment, error)
	UpdateStatus(context.Context, db.Queryer, UpdateShipmentStatusParams) (*Shipment, error)
	CreateStatusLog(context.Context, db.Queryer, CreateStatusLogParams) (*StatusLog, error)
	ListStatusLogs(context.Context, db.Queryer, uuid.UUID, uuid.UUID) ([]StatusLog, error)
	LockOrderByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*order.Order, error)
	FindPublicOrderByNumber(context.Context, db.Queryer, uuid.UUID, uuid.UUID, string) (*order.Order, error)
	ListOrderItems(context.Context, db.Queryer, uuid.UUID, uuid.UUID) ([]order.Item, error)
	UpdateOrderShipmentStatus(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID, string) error
	UpdateOrderStatus(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID, string) (*order.Order, error)
	CreateOrderStatusLog(context.Context, db.Queryer, uuid.UUID, uuid.UUID, string, string, string, uuid.UUID) error
}

type publicStoreResolver interface {
	Resolve(context.Context, string) (store.PublicContext, error)
}

type outboxStore interface {
	Insert(context.Context, db.Queryer, outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	db           database
	shipments    shipmentStore
	publicStores publicStoreResolver
	outbox       outboxStore
	now          func() time.Time
}

type CreateInput struct {
	ActorUserID     uuid.UUID
	CourierType     string
	CourierName     string
	TrackingNumber  string
	ShippingCost    int64
	AssignedToName  string
	AssignedToPhone string
	Note            string
}

type UpdateStatusInput struct {
	ActorUserID uuid.UUID
	Status      string
	Note        string
}

type normalizedCreateInput struct {
	CourierType     string
	CourierName     string
	TrackingNumber  string
	ShippingCost    int64
	AssignedToName  string
	AssignedToPhone string
	Note            string
}

func NewService(database database, shipments shipmentStore, publicStores publicStoreResolver, outbox outboxStore) *Service {
	return &Service{
		db:           database,
		shipments:    shipments,
		publicStores: publicStores,
		outbox:       outbox,
		now:          time.Now,
	}
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]ShipmentResponse, PaginationMeta, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, PaginationMeta{}, err
	}

	normalized, err := normalizeListFilters(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.shipments.List(ctx, s.db, tenantID, storeID, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeCursor(items[len(items)-1])
		if err != nil {
			return nil, PaginationMeta{}, apperror.Internal(err)
		}
		nextCursor = &encoded
	}

	return NewShipmentResponses(items), PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) Detail(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, shipmentID uuid.UUID) (DetailResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return DetailResponse{}, err
	}
	if shipmentID == uuid.Nil {
		return DetailResponse{}, invalidField("shipment_id", "Shipment is required")
	}

	item, err := s.shipments.FindByID(ctx, s.db, tenantID, storeID, shipmentID)
	if err != nil {
		return DetailResponse{}, shipmentReadError(err)
	}

	logs, err := s.shipments.ListStatusLogs(ctx, s.db, tenantID, shipmentID)
	if err != nil {
		return DetailResponse{}, apperror.Internal(err)
	}

	return NewDetailResponse(*item, logs), nil
}

func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, input CreateInput) (CreateShipmentResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return CreateShipmentResponse{}, err
	}
	if orderID == uuid.Nil {
		return CreateShipmentResponse{}, invalidField("order_id", "Order is required")
	}
	if input.ActorUserID == uuid.Nil {
		return CreateShipmentResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	normalized, err := normalizeCreateInput(input)
	if err != nil {
		return CreateShipmentResponse{}, err
	}
	if normalized.TrackingNumber == "" {
		normalized.TrackingNumber = s.generateTrackingNumber()
	}

	var created *Shipment
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		orderRecord, err := s.shipments.LockOrderByID(ctx, tx, tenantID, storeID, orderID)
		if err != nil {
			return orderReadError(err)
		}
		if err := validateOrderCanShip(*orderRecord); err != nil {
			return err
		}

		shipmentRecord, err := s.shipments.Create(ctx, tx, CreateShipmentParams{
			TenantID:        tenantID,
			StoreID:         storeID,
			OrderID:         orderID,
			CourierType:     normalized.CourierType,
			CourierName:     normalized.CourierName,
			TrackingNumber:  normalized.TrackingNumber,
			ShippingCost:    normalized.ShippingCost,
			AssignedToName:  normalized.AssignedToName,
			AssignedToPhone: normalized.AssignedToPhone,
			Note:            normalized.Note,
			CreatedBy:       input.ActorUserID,
		})
		if err != nil {
			return apperror.Internal(err)
		}

		if _, err := s.shipments.CreateStatusLog(ctx, tx, CreateStatusLogParams{
			TenantID:   tenantID,
			ShipmentID: shipmentRecord.ID,
			FromStatus: "",
			ToStatus:   StatusPending,
			Note:       defaultShipmentLogNote,
			CreatedBy:  input.ActorUserID,
		}); err != nil {
			return apperror.Internal(err)
		}
		if err := s.shipments.UpdateOrderShipmentStatus(ctx, tx, tenantID, storeID, orderID, StatusPending); err != nil {
			return orderReadError(err)
		}
		if err := s.insertShipmentCreatedEvents(ctx, tx, *shipmentRecord); err != nil {
			return err
		}

		created = shipmentRecord
		return nil
	})
	if err != nil {
		return CreateShipmentResponse{}, err
	}
	if created == nil {
		return CreateShipmentResponse{}, apperror.Internal(errors.New("created shipment is nil"))
	}

	return NewCreateShipmentResponse(*created), nil
}

func (s *Service) UpdateStatus(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, shipmentID uuid.UUID, input UpdateStatusInput) (UpdateStatusResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return UpdateStatusResponse{}, err
	}
	if shipmentID == uuid.Nil {
		return UpdateStatusResponse{}, invalidField("shipment_id", "Shipment is required")
	}
	if input.ActorUserID == uuid.Nil {
		return UpdateStatusResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	targetStatus := strings.TrimSpace(input.Status)
	if !knownStatus(targetStatus) {
		return UpdateStatusResponse{}, invalidShipmentStatus("Shipment status is not supported", "status", targetStatus)
	}
	note := strings.TrimSpace(input.Note)
	if len(note) > maxShipmentNoteLength {
		return UpdateStatusResponse{}, invalidField("note", "Note must be 500 characters or fewer")
	}

	var updated *Shipment
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.shipments.LockByID(ctx, tx, tenantID, storeID, shipmentID)
		if err != nil {
			return shipmentReadError(err)
		}

		orderRecord, err := s.shipments.LockOrderByID(ctx, tx, tenantID, storeID, current.OrderID)
		if err != nil {
			return orderReadError(err)
		}
		if isTerminalOrderStatus(orderRecord.Status) && targetStatus != current.Status {
			return invalidShipmentStatus("Order cannot be shipped in its current status", "order_status", orderRecord.Status)
		}

		if current.Status == targetStatus {
			updated = current
			return nil
		}
		if !canTransition(current.Status, targetStatus) {
			return invalidShipmentStatus("Invalid shipment status transition", "status", targetStatus)
		}

		shipmentRecord, err := s.shipments.UpdateStatus(ctx, tx, UpdateShipmentStatusParams{
			TenantID:   tenantID,
			StoreID:    storeID,
			ShipmentID: shipmentID,
			Status:     targetStatus,
			UpdatedBy:  input.ActorUserID,
		})
		if err != nil {
			return shipmentReadError(err)
		}

		logNote := note
		if targetStatus == StatusDelivered && logNote == "" {
			logNote = deliveredShipmentLogNote
		}
		if _, err := s.shipments.CreateStatusLog(ctx, tx, CreateStatusLogParams{
			TenantID:   tenantID,
			ShipmentID: shipmentID,
			FromStatus: current.Status,
			ToStatus:   targetStatus,
			Note:       logNote,
			CreatedBy:  input.ActorUserID,
		}); err != nil {
			return apperror.Internal(err)
		}
		if err := s.shipments.UpdateOrderShipmentStatus(ctx, tx, tenantID, storeID, shipmentRecord.OrderID, targetStatus); err != nil {
			return orderReadError(err)
		}
		if targetStatus == StatusDelivered {
			if err := s.markOrderDelivered(ctx, tx, *orderRecord, input.ActorUserID); err != nil {
				return err
			}
		}
		if err := s.insertShipmentStatusUpdatedEvents(ctx, tx, *shipmentRecord, current.Status, targetStatus); err != nil {
			return err
		}

		updated = shipmentRecord
		return nil
	})
	if err != nil {
		return UpdateStatusResponse{}, err
	}
	if updated == nil {
		return UpdateStatusResponse{}, apperror.Internal(errors.New("updated shipment is nil"))
	}

	return NewUpdateStatusResponse(*updated), nil
}

func (s *Service) PublicTracking(ctx context.Context, storeSlug string, orderNumber string, phone string) (PublicTrackingResponse, error) {
	phone = normalizePhone(phone)
	if phone == "" {
		return PublicTrackingResponse{}, invalidField("phone", "phone query param is required")
	}

	currentStore, err := s.publicStores.Resolve(ctx, strings.TrimSpace(storeSlug))
	if err != nil {
		return PublicTrackingResponse{}, err
	}

	orderRecord, err := s.shipments.FindPublicOrderByNumber(ctx, s.db, currentStore.TenantID, currentStore.StoreID, strings.TrimSpace(orderNumber))
	if err != nil {
		return PublicTrackingResponse{}, orderReadError(err)
	}
	if normalizePhone(orderRecord.CustomerPhone) != phone {
		return PublicTrackingResponse{}, apperror.NotFound("Order not found")
	}

	items, err := s.shipments.ListOrderItems(ctx, s.db, currentStore.TenantID, orderRecord.ID)
	if err != nil {
		return PublicTrackingResponse{}, apperror.Internal(err)
	}

	var shipmentRecord *Shipment
	var logs []StatusLog
	shipmentRecord, err = s.shipments.FindLatestByOrder(ctx, s.db, currentStore.TenantID, currentStore.StoreID, orderRecord.ID)
	if err != nil {
		if !errors.Is(err, ErrShipmentNotFound) {
			return PublicTrackingResponse{}, shipmentReadError(err)
		}
	} else {
		logs, err = s.shipments.ListStatusLogs(ctx, s.db, currentStore.TenantID, shipmentRecord.ID)
		if err != nil {
			return PublicTrackingResponse{}, apperror.Internal(err)
		}
	}

	return NewPublicTrackingResponse(*orderRecord, items, shipmentRecord, logs), nil
}

func (s *Service) markOrderDelivered(ctx context.Context, tx db.Tx, current order.Order, actorUserID uuid.UUID) error {
	if current.Status == order.StatusDelivered || current.Status == order.StatusCompleted {
		return nil
	}
	if current.Status == order.StatusCancelled || current.Status == order.StatusReturned || current.Status == order.StatusRefunded {
		return invalidShipmentStatus("Order cannot be delivered in its current status", "order_status", current.Status)
	}

	updated, err := s.shipments.UpdateOrderStatus(ctx, tx, current.TenantID, current.StoreID, current.ID, order.StatusDelivered)
	if err != nil {
		return orderReadError(err)
	}
	if err := s.shipments.CreateOrderStatusLog(ctx, tx, current.TenantID, current.ID, current.Status, updated.Status, "Shipment delivered", actorUserID); err != nil {
		return apperror.Internal(err)
	}
	if err := s.insertOrderDeliveredEvent(ctx, tx, *updated); err != nil {
		return err
	}
	return nil
}

func (s *Service) insertShipmentCreatedEvents(ctx context.Context, tx db.Tx, shipmentRecord Shipment) error {
	if err := s.insertShipmentEvent(ctx, tx, EventShipmentCreated, shipmentRecord, "shipment_created"); err != nil {
		return err
	}
	return s.insertNotificationEvent(ctx, tx, shipmentRecord, "shipment_created")
}

func (s *Service) insertShipmentStatusUpdatedEvents(ctx context.Context, tx db.Tx, shipmentRecord Shipment, fromStatus string, toStatus string) error {
	if err := s.insertShipmentEventWithTransition(ctx, tx, EventShipmentStatusUpdated, shipmentRecord, fromStatus, toStatus); err != nil {
		return err
	}
	return s.insertNotificationEvent(ctx, tx, shipmentRecord, "shipment_status_updated")
}

func (s *Service) insertShipmentEvent(ctx context.Context, tx db.Tx, eventType string, shipmentRecord Shipment, action string) error {
	return s.insertShipmentEventWithPayload(ctx, tx, eventType, shipmentRecord, map[string]any{
		"action": action,
	})
}

func (s *Service) insertShipmentEventWithTransition(ctx context.Context, tx db.Tx, eventType string, shipmentRecord Shipment, fromStatus string, toStatus string) error {
	return s.insertShipmentEventWithPayload(ctx, tx, eventType, shipmentRecord, map[string]any{
		"from_status": fromStatus,
		"to_status":   toStatus,
	})
}

func (s *Service) insertShipmentEventWithPayload(ctx context.Context, tx db.Tx, eventType string, shipmentRecord Shipment, extra map[string]any) error {
	payload := map[string]any{
		"tenant_id":       shipmentRecord.TenantID.String(),
		"store_id":        shipmentRecord.StoreID.String(),
		"order_id":        shipmentRecord.OrderID.String(),
		"shipment_id":     shipmentRecord.ID.String(),
		"status":          shipmentRecord.Status,
		"tracking_number": shipmentRecord.TrackingNumber,
	}
	for key, value := range extra {
		payload[key] = value
	}
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return apperror.Internal(err)
	}
	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      shipmentRecord.TenantID,
		EventType:     eventType,
		AggregateType: AggregateShipment,
		AggregateID:   shipmentRecord.ID,
		Payload:       rawPayload,
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) insertOrderDeliveredEvent(ctx context.Context, tx db.Tx, orderRecord order.Order) error {
	payload, err := json.Marshal(map[string]any{
		"tenant_id":    orderRecord.TenantID.String(),
		"store_id":     orderRecord.StoreID.String(),
		"order_id":     orderRecord.ID.String(),
		"order_number": orderRecord.OrderNumber,
		"status":       orderRecord.Status,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      orderRecord.TenantID,
		EventType:     EventOrderDelivered,
		AggregateType: AggregateOrder,
		AggregateID:   orderRecord.ID,
		Payload:       payload,
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) insertNotificationEvent(ctx context.Context, tx db.Tx, shipmentRecord Shipment, notificationType string) error {
	payload, err := json.Marshal(map[string]any{
		"tenant_id":   shipmentRecord.TenantID.String(),
		"store_id":    shipmentRecord.StoreID.String(),
		"order_id":    shipmentRecord.OrderID.String(),
		"shipment_id": shipmentRecord.ID.String(),
		"type":        notificationType,
	})
	if err != nil {
		return apperror.Internal(err)
	}
	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      shipmentRecord.TenantID,
		EventType:     EventNotificationRequested,
		AggregateType: AggregateShipment,
		AggregateID:   shipmentRecord.ID,
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
			return ListFilters{}, invalidShipmentStatus("Shipment status is not supported", "status", status)
		}
		filters.Status = &status
	}
	if filters.Limit <= 0 {
		filters.Limit = defaultListLimit
	}
	if filters.Limit > maxListLimit {
		filters.Limit = maxListLimit
	}
	return filters, nil
}

func normalizeCreateInput(input CreateInput) (normalizedCreateInput, error) {
	courierType := strings.TrimSpace(input.CourierType)
	if courierType == "" {
		courierType = CourierTypeManual
	}
	normalized := normalizedCreateInput{
		CourierType:     courierType,
		CourierName:     strings.TrimSpace(input.CourierName),
		TrackingNumber:  strings.TrimSpace(input.TrackingNumber),
		ShippingCost:    input.ShippingCost,
		AssignedToName:  strings.TrimSpace(input.AssignedToName),
		AssignedToPhone: strings.TrimSpace(input.AssignedToPhone),
		Note:            strings.TrimSpace(input.Note),
	}
	return normalized, validateCreateInput(normalized)
}

func validateCreateInput(input normalizedCreateInput) error {
	details := make([]map[string]string, 0)
	if input.CourierType != CourierTypeInternal && input.CourierType != CourierTypeManual {
		details = append(details, map[string]string{"field": "courier_type", "message": "courier_type must be internal or manual"})
	}
	if len(input.CourierName) > maxCourierNameLength {
		details = append(details, map[string]string{"field": "courier_name", "message": "courier_name must be 120 characters or fewer"})
	}
	if len(input.TrackingNumber) > maxTrackingNumberLength {
		details = append(details, map[string]string{"field": "tracking_number", "message": "tracking_number must be 120 characters or fewer"})
	}
	if input.ShippingCost < 0 {
		details = append(details, map[string]string{"field": "shipping_cost", "message": "shipping_cost must be greater than or equal to zero"})
	}
	if len(input.AssignedToName) > maxAssignedNameLength {
		details = append(details, map[string]string{"field": "assigned_to_name", "message": "assigned_to_name must be 120 characters or fewer"})
	}
	if len(input.AssignedToPhone) > maxAssignedPhoneLength {
		details = append(details, map[string]string{"field": "assigned_to_phone", "message": "assigned_to_phone must be 40 characters or fewer"})
	}
	if len(input.Note) > maxShipmentNoteLength {
		details = append(details, map[string]string{"field": "note", "message": "note must be 500 characters or fewer"})
	}
	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}
	return nil
}

func validateOrderCanShip(orderRecord order.Order) error {
	if orderRecord.PaymentStatus != order.PaymentStatusPaid {
		return invalidShipmentStatus("Order must be paid before shipment can be created", "payment_status", orderRecord.PaymentStatus)
	}
	switch orderRecord.Status {
	case order.StatusConfirmed, order.StatusProcessing, order.StatusReadyToShip:
		return nil
	default:
		return invalidShipmentStatus("Order cannot be shipped in its current status", "status", orderRecord.Status)
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

func knownStatus(status string) bool {
	switch status {
	case StatusPending, StatusReadyForPickup, StatusPickedUp, StatusOnDelivery, StatusDelivered, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

var allowedTransitions = map[string]map[string]bool{
	StatusPending: {
		StatusReadyForPickup: true,
		StatusCancelled:      true,
	},
	StatusReadyForPickup: {
		StatusPickedUp:  true,
		StatusCancelled: true,
	},
	StatusPickedUp: {
		StatusOnDelivery: true,
		StatusFailed:     true,
		StatusCancelled:  true,
	},
	StatusOnDelivery: {
		StatusDelivered: true,
		StatusFailed:    true,
		StatusCancelled: true,
	},
	StatusFailed: {
		StatusReadyForPickup: true,
		StatusCancelled:      true,
	},
}

func canTransition(from string, to string) bool {
	return allowedTransitions[from][to]
}

func isTerminalOrderStatus(status string) bool {
	switch status {
	case order.StatusCancelled, order.StatusReturned, order.StatusRefunded, order.StatusCompleted:
		return true
	default:
		return false
	}
}

func (s *Service) generateTrackingNumber() string {
	return fmt.Sprintf("TRK-%s-%s", s.now().UTC().Format("20060102"), strings.ToUpper(strings.ReplaceAll(uuid.NewString()[:8], "-", "")))
}

func normalizePhone(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range value {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		if r == '+' && b.Len() == 0 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{{"field": field, "message": message}})
}

func invalidShipmentStatus(message string, field string, status string) error {
	return apperror.InvalidOrderStatus(message, []map[string]string{{"field": field, "message": "unsupported or invalid status: " + status}})
}

func shipmentReadError(err error) error {
	if errors.Is(err, ErrShipmentNotFound) {
		return apperror.NotFound("Shipment not found")
	}
	return apperror.Internal(err)
}

func orderReadError(err error) error {
	if errors.Is(err, ErrOrderNotFound) {
		return apperror.NotFound("Order not found")
	}
	return apperror.Internal(err)
}
