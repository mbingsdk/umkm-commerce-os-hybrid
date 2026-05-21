package plan

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type database interface {
	db.Queryer
}

type store interface {
	FindByTenantID(context.Context, db.Queryer, uuid.UUID) (*Plan, error)
	CountActiveProducts(context.Context, db.Queryer, uuid.UUID, uuid.UUID) (int, error)
	CountActiveStaff(context.Context, db.Queryer, uuid.UUID) (int, error)
}

type Service struct {
	db   database
	repo store
}

func NewService(database database, repo store) *Service {
	return &Service{db: database, repo: repo}
}

func (s *Service) RequireFeature(ctx context.Context, q db.Queryer, tenantID uuid.UUID, feature Feature) error {
	if tenantID == uuid.Nil {
		return invalidField("tenant_id", "Tenant is required")
	}

	currentPlan, err := s.repo.FindByTenantID(ctx, q, tenantID)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			return apperror.Forbidden("Tenant plan is required")
		}
		return apperror.Internal(err)
	}

	if !currentPlan.IsActive {
		return apperror.Forbidden("Tenant plan is not active")
	}

	allowed := false
	switch feature {
	case FeaturePOS:
		allowed = currentPlan.CanUsePOS
	case FeatureDiscovery:
		allowed = currentPlan.CanUseDiscovery
	case FeatureCourier:
		allowed = currentPlan.CanUseCourier
	default:
		return apperror.Forbidden("Feature is not available on current plan")
	}
	if !allowed {
		return apperror.Forbidden("Feature is not available on current plan")
	}

	return nil
}

func (s *Service) CheckProductLimit(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return invalidField("tenant_id", "Tenant is required")
	}
	if storeID == uuid.Nil {
		return invalidField("store_id", "Store is required")
	}

	currentPlan, err := s.repo.FindByTenantID(ctx, q, tenantID)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			return apperror.Forbidden("Tenant plan is required")
		}
		return apperror.Internal(err)
	}
	if !currentPlan.IsActive {
		return apperror.Forbidden("Tenant plan is not active")
	}
	if currentPlan.ProductLimit == nil {
		return nil
	}

	count, err := s.repo.CountActiveProducts(ctx, q, tenantID, storeID)
	if err != nil {
		return apperror.Internal(err)
	}
	if count >= *currentPlan.ProductLimit {
		return apperror.PlanLimitExceeded("Product limit exceeded", map[string]any{
			"feature":       "products",
			"current_count": count,
			"limit":         *currentPlan.ProductLimit,
			"plan_code":     currentPlan.Code,
		})
	}
	return nil
}

func (s *Service) CheckStaffLimit(ctx context.Context, q db.Queryer, tenantID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return invalidField("tenant_id", "Tenant is required")
	}

	currentPlan, err := s.repo.FindByTenantID(ctx, q, tenantID)
	if err != nil {
		if errors.Is(err, ErrPlanNotFound) {
			return apperror.Forbidden("Tenant plan is required")
		}
		return apperror.Internal(err)
	}
	if !currentPlan.IsActive {
		return apperror.Forbidden("Tenant plan is not active")
	}
	if currentPlan.StaffLimit == nil {
		return nil
	}

	count, err := s.repo.CountActiveStaff(ctx, q, tenantID)
	if err != nil {
		return apperror.Internal(err)
	}
	if count >= *currentPlan.StaffLimit {
		return apperror.PlanLimitExceeded("Staff limit exceeded", map[string]any{
			"feature":       "staff",
			"current_count": count,
			"limit":         *currentPlan.StaffLimit,
			"plan_code":     currentPlan.Code,
		})
	}
	return nil
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{{"field": field, "message": message}})
}
