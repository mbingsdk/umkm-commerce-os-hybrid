package tenant

import (
	"context"
	"errors"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type tenantStore interface {
	FindDefaultPlanID(ctx context.Context, q db.Queryer) (uuid.UUID, error)
	Create(ctx context.Context, q db.Queryer, params CreateTenantParams) (*Tenant, error)
}

type userTenantStore interface {
	Create(ctx context.Context, q db.Queryer, params CreateMembershipParams) (*Membership, error)
	FindActiveAccess(ctx context.Context, q db.Queryer, userID uuid.UUID, tenantID uuid.UUID) (*AccessRecord, error)
	ListByUserID(ctx context.Context, q db.Queryer, userID uuid.UUID) ([]TenantListItem, error)
}

type storeCreator interface {
	Create(ctx context.Context, q db.Queryer, params store.CreateParams) (*store.Store, error)
}

type auditRecorder interface {
	Create(ctx context.Context, q db.Queryer, entry audit.Entry) error
}

type Service struct {
	db          database
	tenants     tenantStore
	memberships userTenantStore
	stores      storeCreator
	auditLogs   auditRecorder
	now         func() time.Time
}

type CreateStoreInput struct {
	UserID     uuid.UUID
	TenantName string
	TenantSlug string
	Store      StoreCreateInput
	IPAddress  string
	UserAgent  string
}

type StoreCreateInput struct {
	Name        string
	Slug        string
	Description string
	Phone       string
	Whatsapp    string
	Email       string
	Address     string
	City        string
	Province    string
	PostalCode  string
}

func NewService(
	database database,
	tenants tenantStore,
	memberships userTenantStore,
	stores storeCreator,
	auditLogs auditRecorder,
) *Service {
	return &Service{
		db:          database,
		tenants:     tenants,
		memberships: memberships,
		stores:      stores,
		auditLogs:   auditLogs,
		now:         time.Now,
	}
}

func (s *Service) CreateStore(ctx context.Context, input CreateStoreInput) (*OnboardingResponse, error) {
	normalized, err := validateCreateStore(input)
	if err != nil {
		return nil, err
	}

	var response *OnboardingResponse
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		planID, err := s.tenants.FindDefaultPlanID(ctx, tx)
		if err != nil {
			return apperror.Internal(err)
		}

		createdTenant, err := s.tenants.Create(ctx, tx, CreateTenantParams{
			PlanID: planID,
			Name:   normalized.TenantName,
			Slug:   normalized.TenantSlug,
		})
		if err != nil {
			if errors.Is(err, ErrTenantSlugAlreadyInUse) {
				return apperror.Validation("Validation failed", []map[string]string{
					{"field": "tenant_slug", "message": "Tenant slug is already in use"},
				})
			}
			return apperror.Internal(err)
		}

		now := s.now().UTC()
		membership, err := s.memberships.Create(ctx, tx, CreateMembershipParams{
			UserID:   normalized.UserID,
			TenantID: createdTenant.ID,
			Role:     MembershipRoleOwner,
			Status:   MembershipStatusActive,
			JoinedAt: now,
		})
		if err != nil {
			return apperror.Internal(err)
		}

		createdStore, err := s.stores.Create(ctx, tx, store.CreateParams{
			TenantID:    createdTenant.ID,
			Name:        normalized.Store.Name,
			Slug:        normalized.Store.Slug,
			Description: normalized.Store.Description,
			Phone:       normalized.Store.Phone,
			Whatsapp:    normalized.Store.Whatsapp,
			Email:       normalized.Store.Email,
			Address:     normalized.Store.Address,
			City:        normalized.Store.City,
			Province:    normalized.Store.Province,
			PostalCode:  normalized.Store.PostalCode,
		})
		if err != nil {
			if errors.Is(err, store.ErrStoreSlugAlreadyInUse) {
				return apperror.Validation("Validation failed", []map[string]string{
					{"field": "store.slug", "message": "Store slug is already in use"},
				})
			}
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    createdTenant.ID,
			StoreID:     &createdStore.ID,
			ActorUserID: &normalized.UserID,
			Action:      "onboarding.create_store",
			EntityType:  "store",
			EntityID:    &createdStore.ID,
			AfterData: map[string]any{
				"tenant_id": createdTenant.ID,
				"store_id":  createdStore.ID,
				"role":      membership.Role,
			},
			IPAddress: normalized.IPAddress,
			UserAgent: normalized.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}

		response = &OnboardingResponse{
			Tenant: TenantResponse{
				ID:     createdTenant.ID,
				Name:   createdTenant.Name,
				Slug:   createdTenant.Slug,
				Status: createdTenant.Status,
			},
			Store: StoreSummaryResponse{
				ID:     createdStore.ID,
				Name:   createdStore.Name,
				Slug:   createdStore.Slug,
				Status: createdStore.Status,
			},
			Membership: MembershipResponse{
				Role:        membership.Role,
				Permissions: []string{},
			},
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) ListMyTenants(ctx context.Context, userID uuid.UUID) ([]ListItemResponse, error) {
	items, err := s.memberships.ListByUserID(ctx, s.db, userID)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	response := make([]ListItemResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewListItemResponse(item))
	}

	return response, nil
}

func (s *Service) ValidateAccess(
	ctx context.Context,
	userID uuid.UUID,
	tenantID uuid.UUID,
) (tenantctx.TenantContext, error) {
	access, err := s.memberships.FindActiveAccess(ctx, s.db, userID, tenantID)
	if err != nil {
		if errors.Is(err, ErrTenantAccessNotFound) {
			return tenantctx.TenantContext{}, apperror.TenantAccessDenied("Tenant access denied")
		}
		return tenantctx.TenantContext{}, apperror.Internal(err)
	}

	if access.TenantStatus != StatusActive && access.TenantStatus != StatusTrialing {
		return tenantctx.TenantContext{}, apperror.TenantAccessDenied("Tenant access denied")
	}

	return tenantctx.TenantContext{
		TenantID: access.TenantID,
		StoreID:  access.StoreID,
		UserID:   userID,
		Role:     access.Role,
	}, nil
}

func validateCreateStore(input CreateStoreInput) (CreateStoreInput, error) {
	input.TenantName = strings.TrimSpace(input.TenantName)
	input.TenantSlug = strings.TrimSpace(input.TenantSlug)
	input.Store.Name = strings.TrimSpace(input.Store.Name)
	input.Store.Slug = strings.TrimSpace(input.Store.Slug)
	input.Store.Description = strings.TrimSpace(input.Store.Description)
	input.Store.Phone = strings.TrimSpace(input.Store.Phone)
	input.Store.Whatsapp = strings.TrimSpace(input.Store.Whatsapp)
	input.Store.Email = strings.TrimSpace(input.Store.Email)
	input.Store.Address = strings.TrimSpace(input.Store.Address)
	input.Store.City = strings.TrimSpace(input.Store.City)
	input.Store.Province = strings.TrimSpace(input.Store.Province)
	input.Store.PostalCode = strings.TrimSpace(input.Store.PostalCode)
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)

	var details []map[string]string
	if input.UserID == uuid.Nil {
		details = append(details, map[string]string{"field": "user_id", "message": "User is required"})
	}
	if input.TenantName == "" {
		details = append(details, map[string]string{"field": "tenant_name", "message": "Tenant name is required"})
	}
	if !isValidSlug(input.TenantSlug) {
		details = append(details, map[string]string{"field": "tenant_slug", "message": "Tenant slug is invalid"})
	}
	if input.Store.Name == "" {
		details = append(details, map[string]string{"field": "store.name", "message": "Store name is required"})
	}
	if !isValidSlug(input.Store.Slug) {
		details = append(details, map[string]string{"field": "store.slug", "message": "Store slug is invalid"})
	}
	if input.Store.Email != "" && !isValidEmail(input.Store.Email) {
		details = append(details, map[string]string{"field": "store.email", "message": "Store email is invalid"})
	}

	if len(details) > 0 {
		return CreateStoreInput{}, apperror.Validation("Validation failed", details)
	}

	return input, nil
}

func isValidSlug(value string) bool {
	return slugPattern.MatchString(value)
}

func isValidEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
