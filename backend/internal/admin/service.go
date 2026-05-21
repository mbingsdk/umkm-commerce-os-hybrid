package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
)

const (
	defaultTenantListLimit = 20
	maxTenantListLimit     = 100
	maxAdminReasonLength   = 500
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type store interface {
	FindUserByID(ctx context.Context, q db.Queryer, userID uuid.UUID) (*User, error)
	CreateAuditLog(ctx context.Context, q db.Queryer, entry AuditEntry) (*AuditLog, error)
	ListTenants(ctx context.Context, q db.Queryer, filters TenantListFilters) ([]TenantListItem, error)
	GetTenantDetail(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (*TenantDetail, error)
	FindTenantByIDForUpdate(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (*Tenant, error)
	UpdateTenantStatus(ctx context.Context, q db.Queryer, tenantID uuid.UUID, status string) (*Tenant, error)
	FindActivePlanByID(ctx context.Context, q db.Queryer, planID uuid.UUID) (*Plan, error)
	UpdateTenantPlan(ctx context.Context, q db.Queryer, tenantID uuid.UUID, planID uuid.UUID) (*Tenant, error)
}

type outboxStore interface {
	Insert(ctx context.Context, q db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	db     database
	repo   store
	outbox outboxStore
}

func NewService(database database, repo store, outboxStores ...outboxStore) *Service {
	var outboxRepo outboxStore
	if len(outboxStores) > 0 {
		outboxRepo = outboxStores[0]
	}
	return &Service{db: database, repo: repo, outbox: outboxRepo}
}

func (s *Service) ValidateSuperAdmin(ctx context.Context, userID uuid.UUID) (Context, error) {
	user, err := s.repo.FindUserByID(ctx, s.db, userID)
	if err != nil {
		if errors.Is(err, ErrAdminUserNotFound) {
			return Context{}, apperror.Unauthorized("Invalid access token")
		}
		return Context{}, apperror.Internal(err)
	}

	if user.Status != auth.UserStatusActive {
		return Context{}, apperror.Forbidden("User account is not active")
	}
	if user.PlatformRole != auth.PlatformRoleSuperAdmin {
		return Context{}, apperror.Forbidden("Super admin access required")
	}

	return Context{
		UserID:       user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PlatformRole: user.PlatformRole,
	}, nil
}

func (s *Service) RecordAudit(ctx context.Context, entry AuditEntry) (*AuditLog, error) {
	entry.Action = strings.TrimSpace(entry.Action)
	entry.TargetType = strings.TrimSpace(entry.TargetType)
	entry.IPAddress = strings.TrimSpace(entry.IPAddress)
	entry.UserAgent = strings.TrimSpace(entry.UserAgent)

	if entry.ActorUserID == uuid.Nil {
		return nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "actor_user_id", "message": "actor_user_id is required"},
		})
	}
	if entry.Action == "" {
		return nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "action", "message": "action is required"},
		})
	}

	log, err := s.repo.CreateAuditLog(ctx, s.db, entry)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return log, nil
}

type UpdateTenantStatusInput struct {
	ActorUserID uuid.UUID
	TenantID    uuid.UUID
	Status      string
	Reason      string
	IPAddress   string
	UserAgent   string
}

type UpdateTenantPlanInput struct {
	ActorUserID uuid.UUID
	TenantID    uuid.UUID
	PlanID      uuid.UUID
	Reason      string
	IPAddress   string
	UserAgent   string
}

func (s *Service) ListTenants(ctx context.Context, filters TenantListFilters) ([]AdminTenantListResponse, PaginationMeta, error) {
	normalized, err := normalizeTenantListFilters(filters)
	if err != nil {
		return nil, PaginationMeta{}, err
	}

	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.repo.ListTenants(ctx, s.db, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeTenantCursor(items[len(items)-1])
		if err != nil {
			return nil, PaginationMeta{}, apperror.Internal(err)
		}
		nextCursor = &encoded
	}

	return NewTenantListResponse(items), PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) GetTenantDetail(ctx context.Context, tenantID uuid.UUID) (AdminTenantDetailResponse, error) {
	if tenantID == uuid.Nil {
		return AdminTenantDetailResponse{}, invalidField("tenantId", "tenantId must be a valid UUID")
	}

	detail, err := s.repo.GetTenantDetail(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			return AdminTenantDetailResponse{}, apperror.NotFound("Tenant not found")
		}
		return AdminTenantDetailResponse{}, apperror.Internal(err)
	}

	return NewTenantDetailResponse(*detail), nil
}

func (s *Service) UpdateTenantStatus(ctx context.Context, input UpdateTenantStatusInput) (AdminTenantMutationResponse, error) {
	normalized, err := normalizeTenantStatusInput(input)
	if err != nil {
		return AdminTenantMutationResponse{}, err
	}

	var updated *Tenant
	var plan *Plan
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.repo.FindTenantByIDForUpdate(ctx, tx, normalized.TenantID)
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return apperror.NotFound("Tenant not found")
			}
			return apperror.Internal(err)
		}

		tenant, err := s.repo.UpdateTenantStatus(ctx, tx, normalized.TenantID, normalized.Status)
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return apperror.NotFound("Tenant not found")
			}
			return apperror.Internal(err)
		}

		if tenant.PlanID != nil {
			foundPlan, err := s.repo.FindActivePlanByID(ctx, tx, *tenant.PlanID)
			if err == nil {
				plan = foundPlan
			} else if !errors.Is(err, ErrPlanNotFound) {
				return apperror.Internal(err)
			}
		}

		if _, err := s.repo.CreateAuditLog(ctx, tx, AuditEntry{
			ActorUserID: normalized.ActorUserID,
			Action:      AuditActionTenantStatusUpdated,
			TargetType:  AggregateTenant,
			TargetID:    &normalized.TenantID,
			BeforeData:  tenantStatusAuditSnapshot(current),
			AfterData:   tenantStatusAuditSnapshot(tenant),
			IPAddress:   normalized.IPAddress,
			UserAgent:   normalized.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}
		if err := s.insertTenantEvent(ctx, tx, normalized.TenantID, EventTenantStatusChanged, map[string]any{
			"tenant_id":       normalized.TenantID.String(),
			"actor_user_id":   normalized.ActorUserID.String(),
			"previous_status": current.Status,
			"new_status":      tenant.Status,
			"reason":          normalized.Reason,
		}); err != nil {
			return err
		}

		updated = tenant
		return nil
	})
	if err != nil {
		return AdminTenantMutationResponse{}, err
	}
	if updated == nil {
		return AdminTenantMutationResponse{}, apperror.Internal(errors.New("updated tenant is nil"))
	}
	return NewTenantMutationResponse(*updated, plan), nil
}

func (s *Service) UpdateTenantPlan(ctx context.Context, input UpdateTenantPlanInput) (AdminTenantMutationResponse, error) {
	normalized, err := normalizeTenantPlanInput(input)
	if err != nil {
		return AdminTenantMutationResponse{}, err
	}

	var updated *Tenant
	var plan *Plan
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.repo.FindTenantByIDForUpdate(ctx, tx, normalized.TenantID)
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return apperror.NotFound("Tenant not found")
			}
			return apperror.Internal(err)
		}

		foundPlan, err := s.repo.FindActivePlanByID(ctx, tx, normalized.PlanID)
		if err != nil {
			if errors.Is(err, ErrPlanNotFound) {
				return apperror.NotFound("Plan not found")
			}
			return apperror.Internal(err)
		}

		tenant, err := s.repo.UpdateTenantPlan(ctx, tx, normalized.TenantID, normalized.PlanID)
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return apperror.NotFound("Tenant not found")
			}
			return apperror.Internal(err)
		}

		if _, err := s.repo.CreateAuditLog(ctx, tx, AuditEntry{
			ActorUserID: normalized.ActorUserID,
			Action:      AuditActionTenantPlanUpdated,
			TargetType:  AggregateTenant,
			TargetID:    &normalized.TenantID,
			BeforeData:  tenantPlanAuditSnapshot(current),
			AfterData:   map[string]any{"plan_id": foundPlan.ID.String(), "plan_code": foundPlan.Code},
			IPAddress:   normalized.IPAddress,
			UserAgent:   normalized.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}
		if err := s.insertTenantEvent(ctx, tx, normalized.TenantID, EventTenantPlanChanged, map[string]any{
			"tenant_id":     normalized.TenantID.String(),
			"actor_user_id": normalized.ActorUserID.String(),
			"previous_plan": uuidPtrString(current.PlanID),
			"new_plan_id":   foundPlan.ID.String(),
			"new_plan_code": foundPlan.Code,
			"reason":        normalized.Reason,
		}); err != nil {
			return err
		}

		updated = tenant
		plan = foundPlan
		return nil
	})
	if err != nil {
		return AdminTenantMutationResponse{}, err
	}
	if updated == nil {
		return AdminTenantMutationResponse{}, apperror.Internal(errors.New("updated tenant is nil"))
	}
	return NewTenantMutationResponse(*updated, plan), nil
}

func (s *Service) insertTenantEvent(ctx context.Context, tx db.Tx, tenantID uuid.UUID, eventType string, payload map[string]any) error {
	if s.outbox == nil {
		return nil
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return apperror.Internal(err)
	}

	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      tenantID,
		EventType:     eventType,
		AggregateType: AggregateTenant,
		AggregateID:   tenantID,
		Payload:       rawPayload,
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func normalizeTenantListFilters(filters TenantListFilters) (TenantListFilters, error) {
	filters.Status = strings.TrimSpace(filters.Status)
	filters.Query = strings.TrimSpace(filters.Query)
	if filters.Status != "" && !allowedTenantStatus(filters.Status) {
		return TenantListFilters{}, invalidField("status", "status must be active, trialing, suspended, or cancelled")
	}
	if filters.Limit <= 0 {
		filters.Limit = defaultTenantListLimit
	}
	if filters.Limit > maxTenantListLimit {
		filters.Limit = maxTenantListLimit
	}
	return filters, nil
}

func normalizeTenantStatusInput(input UpdateTenantStatusInput) (UpdateTenantStatusInput, error) {
	input.Status = strings.TrimSpace(input.Status)
	input.Reason = strings.TrimSpace(input.Reason)
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)

	var details []map[string]string
	if input.ActorUserID == uuid.Nil {
		details = append(details, map[string]string{"field": "actor_user_id", "message": "Actor is required"})
	}
	if input.TenantID == uuid.Nil {
		details = append(details, map[string]string{"field": "tenant_id", "message": "Tenant is required"})
	}
	if !allowedTenantStatus(input.Status) {
		details = append(details, map[string]string{"field": "status", "message": "status must be active, trialing, suspended, or cancelled"})
	}
	if len(input.Reason) > maxAdminReasonLength {
		details = append(details, map[string]string{"field": "reason", "message": "reason must be 500 characters or fewer"})
	}
	if len(details) > 0 {
		return UpdateTenantStatusInput{}, apperror.Validation("Validation failed", details)
	}
	return input, nil
}

func normalizeTenantPlanInput(input UpdateTenantPlanInput) (UpdateTenantPlanInput, error) {
	input.Reason = strings.TrimSpace(input.Reason)
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)

	var details []map[string]string
	if input.ActorUserID == uuid.Nil {
		details = append(details, map[string]string{"field": "actor_user_id", "message": "Actor is required"})
	}
	if input.TenantID == uuid.Nil {
		details = append(details, map[string]string{"field": "tenant_id", "message": "Tenant is required"})
	}
	if input.PlanID == uuid.Nil {
		details = append(details, map[string]string{"field": "plan_id", "message": "plan_id must be a valid UUID"})
	}
	if len(input.Reason) > maxAdminReasonLength {
		details = append(details, map[string]string{"field": "reason", "message": "reason must be 500 characters or fewer"})
	}
	if len(details) > 0 {
		return UpdateTenantPlanInput{}, apperror.Validation("Validation failed", details)
	}
	return input, nil
}

func allowedTenantStatus(status string) bool {
	switch status {
	case TenantStatusActive, TenantStatusTrialing, TenantStatusSuspended, TenantStatusCancelled:
		return true
	default:
		return false
	}
}

func tenantStatusAuditSnapshot(tenant *Tenant) map[string]any {
	if tenant == nil {
		return nil
	}
	return map[string]any{
		"id":     tenant.ID.String(),
		"name":   tenant.Name,
		"slug":   tenant.Slug,
		"status": tenant.Status,
	}
}

func tenantPlanAuditSnapshot(tenant *Tenant) map[string]any {
	if tenant == nil {
		return nil
	}
	return map[string]any{
		"id":      tenant.ID.String(),
		"plan_id": uuidPtrString(tenant.PlanID),
	}
}

func uuidPtrString(value *uuid.UUID) any {
	if value == nil || *value == uuid.Nil {
		return nil
	}
	return value.String()
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{{"field": field, "message": message}})
}

func parseDate(raw string, endExclusive bool) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	if endExclusive {
		return parsed.AddDate(0, 0, 1), nil
	}
	return parsed, nil
}
