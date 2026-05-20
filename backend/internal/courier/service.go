package courier

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

const (
	maxZoneNameLength        = 120
	maxZoneDescriptionLength = 500
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type zoneStore interface {
	ListZones(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ListZoneFilters) ([]Zone, error)
	ListPublicActiveZones(context.Context, db.Queryer, uuid.UUID, uuid.UUID) ([]Zone, error)
	FindZoneByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Zone, error)
	CreateZone(context.Context, db.Queryer, CreateZoneParams) (*Zone, error)
	UpdateZone(context.Context, db.Queryer, UpdateZoneParams) (*Zone, error)
	SoftDeleteZone(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Zone, error)
}

type publicStoreResolver interface {
	Resolve(context.Context, string) (store.PublicContext, error)
}

type auditStore interface {
	Create(context.Context, db.Queryer, audit.Entry) error
}

type Service struct {
	db           database
	zones        zoneStore
	publicStores publicStoreResolver
	auditLogs    auditStore
}

type CreateZoneInput struct {
	ActorUserID uuid.UUID
	Name        string
	Description string
	Rate        int64
	IsActive    *bool
	SortOrder   *int
}

type UpdateZoneInput struct {
	ActorUserID uuid.UUID
	Name        *string
	Description *string
	Rate        *int64
	IsActive    *bool
	SortOrder   *int
}

type normalizedZoneInput struct {
	Name        string
	Description string
	Rate        int64
	IsActive    bool
	SortOrder   int
}

func NewService(database database, zones zoneStore, publicStores publicStoreResolver, auditLogs auditStore) *Service {
	return &Service{
		db:           database,
		zones:        zones,
		publicStores: publicStores,
		auditLogs:    auditLogs,
	}
}

func (s *Service) ListZones(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, filters ListZoneFilters) ([]ZoneResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, err
	}

	zones, err := s.zones.ListZones(ctx, s.db, tenantID, storeID, filters)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return NewZoneResponses(zones), nil
}

func (s *Service) ListPublicZones(ctx context.Context, storeSlug string) ([]ZoneResponse, error) {
	currentStore, err := s.publicStores.Resolve(ctx, storeSlug)
	if err != nil {
		return nil, err
	}

	zones, err := s.zones.ListPublicActiveZones(ctx, s.db, currentStore.TenantID, currentStore.StoreID)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return NewZoneResponses(zones), nil
}

func (s *Service) CreateZone(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, input CreateZoneInput) (ZoneResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ZoneResponse{}, err
	}
	if input.ActorUserID == uuid.Nil {
		return ZoneResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	normalized, err := normalizeCreateZone(input)
	if err != nil {
		return ZoneResponse{}, err
	}

	var created *Zone
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		zone, err := s.zones.CreateZone(ctx, tx, CreateZoneParams{
			TenantID:    tenantID,
			StoreID:     storeID,
			Name:        normalized.Name,
			Description: normalized.Description,
			Rate:        normalized.Rate,
			IsActive:    normalized.IsActive,
			SortOrder:   normalized.SortOrder,
		})
		if err != nil {
			return apperror.Internal(err)
		}
		if err := s.auditZoneChange(ctx, tx, AuditActionCourierZoneCreated, input.ActorUserID, nil, zone); err != nil {
			return err
		}
		created = zone
		return nil
	})
	if err != nil {
		return ZoneResponse{}, err
	}
	if created == nil {
		return ZoneResponse{}, apperror.Internal(errors.New("created courier zone is nil"))
	}

	return NewZoneResponse(*created), nil
}

func (s *Service) UpdateZone(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, zoneID uuid.UUID, input UpdateZoneInput) (ZoneResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ZoneResponse{}, err
	}
	if zoneID == uuid.Nil {
		return ZoneResponse{}, invalidField("zone_id", "Courier zone is required")
	}
	if input.ActorUserID == uuid.Nil {
		return ZoneResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	var updated *Zone
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.zones.FindZoneByID(ctx, tx, tenantID, storeID, zoneID)
		if err != nil {
			return zoneReadError(err)
		}

		normalized, err := normalizeUpdateZone(*current, input)
		if err != nil {
			return err
		}

		zone, err := s.zones.UpdateZone(ctx, tx, UpdateZoneParams{
			TenantID:    tenantID,
			StoreID:     storeID,
			ZoneID:      zoneID,
			Name:        normalized.Name,
			Description: normalized.Description,
			Rate:        normalized.Rate,
			IsActive:    normalized.IsActive,
			SortOrder:   normalized.SortOrder,
		})
		if err != nil {
			return zoneReadError(err)
		}
		if err := s.auditZoneChange(ctx, tx, AuditActionCourierZoneUpdated, input.ActorUserID, current, zone); err != nil {
			return err
		}
		updated = zone
		return nil
	})
	if err != nil {
		return ZoneResponse{}, err
	}
	if updated == nil {
		return ZoneResponse{}, apperror.Internal(errors.New("updated courier zone is nil"))
	}

	return NewZoneResponse(*updated), nil
}

func (s *Service) DeleteZone(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, zoneID uuid.UUID, actorUserID uuid.UUID) (ZoneResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ZoneResponse{}, err
	}
	if zoneID == uuid.Nil {
		return ZoneResponse{}, invalidField("zone_id", "Courier zone is required")
	}
	if actorUserID == uuid.Nil {
		return ZoneResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	var deleted *Zone
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.zones.FindZoneByID(ctx, tx, tenantID, storeID, zoneID)
		if err != nil {
			return zoneReadError(err)
		}

		zone, err := s.zones.SoftDeleteZone(ctx, tx, tenantID, storeID, zoneID)
		if err != nil {
			return zoneReadError(err)
		}
		if err := s.auditZoneChange(ctx, tx, AuditActionCourierZoneDeleted, actorUserID, current, zone); err != nil {
			return err
		}
		deleted = zone
		return nil
	})
	if err != nil {
		return ZoneResponse{}, err
	}
	if deleted == nil {
		return ZoneResponse{}, apperror.Internal(errors.New("deleted courier zone is nil"))
	}

	return NewZoneResponse(*deleted), nil
}

func normalizeCreateZone(input CreateZoneInput) (normalizedZoneInput, error) {
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	sortOrder := 0
	if input.SortOrder != nil {
		sortOrder = *input.SortOrder
	}

	normalized := normalizedZoneInput{
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Rate:        input.Rate,
		IsActive:    isActive,
		SortOrder:   sortOrder,
	}
	return normalized, validateZoneInput(normalized)
}

func normalizeUpdateZone(current Zone, input UpdateZoneInput) (normalizedZoneInput, error) {
	normalized := normalizedZoneInput{
		Name:        current.Name,
		Description: current.Description,
		Rate:        current.Rate,
		IsActive:    current.IsActive,
		SortOrder:   current.SortOrder,
	}
	if input.Name != nil {
		normalized.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		normalized.Description = strings.TrimSpace(*input.Description)
	}
	if input.Rate != nil {
		normalized.Rate = *input.Rate
	}
	if input.IsActive != nil {
		normalized.IsActive = *input.IsActive
	}
	if input.SortOrder != nil {
		normalized.SortOrder = *input.SortOrder
	}
	return normalized, validateZoneInput(normalized)
}

func validateZoneInput(input normalizedZoneInput) error {
	details := make([]map[string]string, 0)
	if input.Name == "" {
		details = append(details, map[string]string{"field": "name", "message": "name is required"})
	}
	if len(input.Name) > maxZoneNameLength {
		details = append(details, map[string]string{"field": "name", "message": "name must be 120 characters or fewer"})
	}
	if len(input.Description) > maxZoneDescriptionLength {
		details = append(details, map[string]string{"field": "description", "message": "description must be 500 characters or fewer"})
	}
	if input.Rate < 0 {
		details = append(details, map[string]string{"field": "rate", "message": "rate must be greater than or equal to zero"})
	}
	if input.SortOrder < 0 {
		details = append(details, map[string]string{"field": "sort_order", "message": "sort_order must be greater than or equal to zero"})
	}
	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}
	return nil
}

func validateScope(tenantID uuid.UUID, storeID uuid.UUID) error {
	details := make([]map[string]string, 0)
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

func zoneReadError(err error) error {
	if errors.Is(err, ErrZoneNotFound) {
		return apperror.NotFound("Courier zone not found")
	}
	return apperror.Internal(err)
}

func (s *Service) auditZoneChange(ctx context.Context, tx db.Tx, action string, actorUserID uuid.UUID, before *Zone, after *Zone) error {
	var tenantID uuid.UUID
	var storeID uuid.UUID
	var entityID uuid.UUID
	if after != nil {
		tenantID = after.TenantID
		storeID = after.StoreID
		entityID = after.ID
	} else if before != nil {
		tenantID = before.TenantID
		storeID = before.StoreID
		entityID = before.ID
	}

	if err := s.auditLogs.Create(ctx, tx, audit.Entry{
		TenantID:    tenantID,
		StoreID:     &storeID,
		ActorUserID: &actorUserID,
		Action:      action,
		EntityType:  AggregateCourierZone,
		EntityID:    &entityID,
		BeforeData:  auditZoneSnapshot(before),
		AfterData:   auditZoneSnapshot(after),
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func auditZoneSnapshot(zone *Zone) map[string]any {
	if zone == nil {
		return nil
	}
	return map[string]any{
		"id":          zone.ID.String(),
		"name":        zone.Name,
		"rate":        zone.Rate,
		"is_active":   zone.IsActive,
		"sort_order":  zone.SortOrder,
		"description": zone.Description,
	}
}
