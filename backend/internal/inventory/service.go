package inventory

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

const (
	defaultListLimit = 20
	maxListLimit     = 100
	maxReasonLength  = 200
	maxNoteLength    = 500
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type inventoryStore interface {
	ListStocks(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ListStockFilters) ([]StockListItem, error)
	FindProduct(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*ProductRef, error)
	LockProduct(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*ProductRef, error)
	ListMovements(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID, ListMovementFilters) ([]StockMovement, error)
	LockStockSnapshot(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*StockSnapshot, error)
	UpdateSnapshot(context.Context, db.Queryer, UpdateSnapshotParams) (*StockSnapshot, error)
	UpdateThreshold(context.Context, db.Queryer, UpdateThresholdParams) (*StockSnapshot, error)
	CreateMovement(context.Context, db.Queryer, CreateMovementParams) (*StockMovement, error)
}

type auditStore interface {
	Create(context.Context, db.Queryer, audit.Entry) error
}

type outboxStore interface {
	Insert(context.Context, db.Queryer, outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	db        database
	store     inventoryStore
	auditLogs auditStore
	outbox    outboxStore
}

type AdjustStockInput struct {
	ActorUserID    uuid.UUID
	AdjustmentType string
	Type           string
	Quantity       int
	Reason         string
	Note           string
}

type UpdateThresholdInput struct {
	ActorUserID       uuid.UUID
	LowStockThreshold int
}

func NewService(database database, store inventoryStore, auditLogs auditStore, outboxRepo outboxStore) *Service {
	return &Service{
		db:        database,
		store:     store,
		auditLogs: auditLogs,
		outbox:    outboxRepo,
	}
}

func (s *Service) ListStocks(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, filters ListStockFilters) ([]StockResponse, PaginationMeta, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, PaginationMeta{}, err
	}

	normalized := normalizeStockFilters(filters)
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.store.ListStocks(ctx, s.db, tenantID, storeID, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	response := make([]StockResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewStockResponse(item))
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeStockCursor(items[len(items)-1])
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

func (s *Service) ListMovements(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID, filters ListMovementFilters) ([]StockMovementResponse, PaginationMeta, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, PaginationMeta{}, err
	}
	if productID == uuid.Nil {
		return nil, PaginationMeta{}, invalidField("product_id", "Product is required")
	}

	if _, err := s.store.FindProduct(ctx, s.db, tenantID, storeID, productID); err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return nil, PaginationMeta{}, apperror.NotFound("Product not found")
		}
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	normalized := normalizeMovementFilters(filters)
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.store.ListMovements(ctx, s.db, tenantID, storeID, productID, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	response := make([]StockMovementResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewStockMovementResponse(item))
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeMovementCursor(items[len(items)-1])
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

func (s *Service) AdjustStock(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID, input AdjustStockInput) (AdjustStockResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return AdjustStockResponse{}, err
	}
	if productID == uuid.Nil {
		return AdjustStockResponse{}, invalidField("product_id", "Product is required")
	}
	if input.ActorUserID == uuid.Nil {
		return AdjustStockResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	movementType, err := normalizeAdjustmentType(input.AdjustmentType, input.Type)
	if err != nil {
		return AdjustStockResponse{}, err
	}
	quantity := input.Quantity
	if quantity <= 0 {
		return AdjustStockResponse{}, invalidField("quantity", "Quantity must be greater than zero")
	}
	reason := strings.TrimSpace(input.Reason)
	note := strings.TrimSpace(input.Note)
	if reason == "" {
		return AdjustStockResponse{}, invalidField("reason", "Reason is required")
	}
	if len(reason) > maxReasonLength {
		return AdjustStockResponse{}, invalidField("reason", "Reason must be 200 characters or fewer")
	}
	if len(note) > maxNoteLength {
		return AdjustStockResponse{}, invalidField("note", "Note must be 500 characters or fewer")
	}

	var updated *StockSnapshot
	var movement *StockMovement
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		product, err := s.store.LockProduct(ctx, tx, tenantID, storeID, productID)
		if err != nil {
			if errors.Is(err, ErrProductNotFound) {
				return apperror.NotFound("Product not found")
			}
			return apperror.Internal(err)
		}
		if product.TenantID != tenantID || product.StoreID != storeID {
			return apperror.NotFound("Product not found")
		}

		current, err := s.store.LockStockSnapshot(ctx, tx, tenantID, storeID, productID)
		if err != nil {
			if errors.Is(err, ErrStockSnapshotNotFound) {
				return apperror.NotFound("Stock snapshot not found")
			}
			return apperror.Internal(err)
		}

		newOnHand := current.QuantityOnHand
		newAvailable := current.QuantityAvailable
		movementQuantity := quantity
		if movementType == MovementTypeAdjustmentOut {
			if current.QuantityAvailable-quantity < 0 {
				return apperror.InsufficientStock("Insufficient stock for adjustment", []map[string]any{{
					"product_id":          productID.String(),
					"quantity_available":  current.QuantityAvailable,
					"requested_quantity":  quantity,
					"low_stock_threshold": current.LowStockThreshold,
				}})
			}
			newOnHand -= quantity
			newAvailable -= quantity
			movementQuantity = -quantity
		} else {
			newOnHand += quantity
			newAvailable += quantity
		}

		updatedSnapshot, err := s.store.UpdateSnapshot(ctx, tx, UpdateSnapshotParams{
			TenantID:          tenantID,
			StoreID:           storeID,
			ProductID:         productID,
			QuantityOnHand:    newOnHand,
			QuantityReserved:  current.QuantityReserved,
			QuantityAvailable: newAvailable,
		})
		if err != nil {
			if errors.Is(err, ErrStockSnapshotNotFound) {
				return apperror.NotFound("Stock snapshot not found")
			}
			return apperror.Internal(err)
		}

		createdMovement, err := s.store.CreateMovement(ctx, tx, CreateMovementParams{
			TenantID:      tenantID,
			StoreID:       storeID,
			ProductID:     productID,
			MovementType:  movementType,
			Quantity:      movementQuantity,
			BalanceAfter:  updatedSnapshot.QuantityAvailable,
			ReferenceType: ReferenceTypeManualAdjustment,
			Reason:        reason,
			Note:          note,
			CreatedBy:     &input.ActorUserID,
		})
		if err != nil {
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    tenantID,
			StoreID:     &storeID,
			ActorUserID: &input.ActorUserID,
			Action:      AuditActionStockAdjusted,
			EntityType:  AggregateProduct,
			EntityID:    &productID,
			BeforeData: map[string]any{
				"quantity_on_hand":    current.QuantityOnHand,
				"quantity_reserved":   current.QuantityReserved,
				"quantity_available":  current.QuantityAvailable,
				"low_stock_threshold": current.LowStockThreshold,
			},
			AfterData: map[string]any{
				"movement_type":          movementType,
				"quantity":               movementQuantity,
				"quantity_on_hand":       updatedSnapshot.QuantityOnHand,
				"quantity_reserved":      updatedSnapshot.QuantityReserved,
				"quantity_available":     updatedSnapshot.QuantityAvailable,
				"low_stock_threshold":    updatedSnapshot.LowStockThreshold,
				"stock_movement_id":      createdMovement.ID.String(),
				"manual_adjustment_note": note,
			},
			Reason: reason,
		}); err != nil {
			return apperror.Internal(err)
		}

		if err := s.insertStockAdjustedEvent(ctx, tx, *updatedSnapshot, *createdMovement, input.ActorUserID, reason); err != nil {
			return err
		}

		updated = updatedSnapshot
		movement = createdMovement
		return nil
	})
	if err != nil {
		return AdjustStockResponse{}, err
	}
	if updated == nil || movement == nil {
		return AdjustStockResponse{}, apperror.Internal(errors.New("stock adjustment response is nil"))
	}

	return NewAdjustStockResponse(*updated, *movement), nil
}

func (s *Service) UpdateThreshold(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID, input UpdateThresholdInput) (UpdateThresholdResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return UpdateThresholdResponse{}, err
	}
	if productID == uuid.Nil {
		return UpdateThresholdResponse{}, invalidField("product_id", "Product is required")
	}
	if input.ActorUserID == uuid.Nil {
		return UpdateThresholdResponse{}, invalidField("actor_user_id", "Actor is required")
	}
	if input.LowStockThreshold < 0 {
		return UpdateThresholdResponse{}, invalidField("low_stock_threshold", "Low stock threshold must be zero or greater")
	}

	var updated *StockSnapshot
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		if _, err := s.store.FindProduct(ctx, tx, tenantID, storeID, productID); err != nil {
			if errors.Is(err, ErrProductNotFound) {
				return apperror.NotFound("Product not found")
			}
			return apperror.Internal(err)
		}

		current, err := s.store.LockStockSnapshot(ctx, tx, tenantID, storeID, productID)
		if err != nil {
			if errors.Is(err, ErrStockSnapshotNotFound) {
				return apperror.NotFound("Stock snapshot not found")
			}
			return apperror.Internal(err)
		}

		updatedSnapshot, err := s.store.UpdateThreshold(ctx, tx, UpdateThresholdParams{
			TenantID:          tenantID,
			StoreID:           storeID,
			ProductID:         productID,
			LowStockThreshold: input.LowStockThreshold,
		})
		if err != nil {
			if errors.Is(err, ErrStockSnapshotNotFound) {
				return apperror.NotFound("Stock snapshot not found")
			}
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    tenantID,
			StoreID:     &storeID,
			ActorUserID: &input.ActorUserID,
			Action:      AuditActionThresholdUpdated,
			EntityType:  AggregateProduct,
			EntityID:    &productID,
			BeforeData: map[string]any{
				"low_stock_threshold": current.LowStockThreshold,
			},
			AfterData: map[string]any{
				"low_stock_threshold": updatedSnapshot.LowStockThreshold,
			},
		}); err != nil {
			return apperror.Internal(err)
		}

		updated = updatedSnapshot
		return nil
	})
	if err != nil {
		return UpdateThresholdResponse{}, err
	}
	if updated == nil {
		return UpdateThresholdResponse{}, apperror.Internal(errors.New("threshold response is nil"))
	}

	return NewUpdateThresholdResponse(*updated), nil
}

func (s *Service) insertStockAdjustedEvent(ctx context.Context, tx db.Tx, snapshot StockSnapshot, movement StockMovement, actorUserID uuid.UUID, reason string) error {
	payload, err := json.Marshal(map[string]any{
		"tenant_id":          snapshot.TenantID.String(),
		"store_id":           snapshot.StoreID.String(),
		"product_id":         snapshot.ProductID.String(),
		"stock_movement_id":  movement.ID.String(),
		"movement_type":      movement.MovementType,
		"quantity":           movement.Quantity,
		"quantity_on_hand":   snapshot.QuantityOnHand,
		"quantity_reserved":  snapshot.QuantityReserved,
		"quantity_available": snapshot.QuantityAvailable,
		"actor_user_id":      actorUserID.String(),
		"reason":             reason,
	})
	if err != nil {
		return apperror.Internal(err)
	}

	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      snapshot.TenantID,
		EventType:     EventStockAdjusted,
		AggregateType: AggregateProduct,
		AggregateID:   snapshot.ProductID,
		Payload:       payload,
	}); err != nil {
		return apperror.Internal(err)
	}

	return nil
}

func normalizeStockFilters(filters ListStockFilters) ListStockFilters {
	filters.Query = querytext.NormalizeSearch(filters.Query)
	if filters.Limit <= 0 {
		filters.Limit = defaultListLimit
	}
	if filters.Limit > maxListLimit {
		filters.Limit = maxListLimit
	}
	return filters
}

func normalizeMovementFilters(filters ListMovementFilters) ListMovementFilters {
	if filters.Limit <= 0 {
		filters.Limit = defaultListLimit
	}
	if filters.Limit > maxListLimit {
		filters.Limit = maxListLimit
	}
	return filters
}

func normalizeAdjustmentType(adjustmentType string, legacyType string) (string, error) {
	raw := strings.ToLower(strings.TrimSpace(adjustmentType))
	if raw == "" {
		raw = strings.ToLower(strings.TrimSpace(legacyType))
	}

	switch raw {
	case AdjustmentTypeIn, MovementTypeAdjustmentIn:
		return MovementTypeAdjustmentIn, nil
	case AdjustmentTypeOut, MovementTypeAdjustmentOut:
		return MovementTypeAdjustmentOut, nil
	default:
		return "", invalidField("adjustment_type", "adjustment_type must be in or out")
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

func formatAdjustmentNote(reason string, note string) string {
	reason = strings.TrimSpace(reason)
	note = strings.TrimSpace(note)
	if reason == "" {
		return note
	}
	if note == "" {
		return "Reason: " + reason
	}
	return "Reason: " + reason + ". " + note
}

func splitAdjustmentNote(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "Reason: ") {
		return "", raw
	}

	trimmed := strings.TrimPrefix(raw, "Reason: ")
	parts := strings.SplitN(trimmed, ". ", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
