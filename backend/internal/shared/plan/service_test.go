package plan

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestCheckProductLimitPreventsCreateAtLimit(t *testing.T) {
	t.Parallel()

	limit := 2
	service := NewService(noopDB{}, &fakePlanRepository{
		plan:         &Plan{ID: uuid.New(), Code: "starter", ProductLimit: &limit, IsActive: true},
		productCount: 2,
	})

	err := service.CheckProductLimit(context.Background(), noopDB{}, uuid.New(), uuid.New())
	assertCode(t, err, apperror.CodePlanLimitExceeded)
}

func TestCheckProductLimitAllowsUnlimited(t *testing.T) {
	t.Parallel()

	service := NewService(noopDB{}, &fakePlanRepository{
		plan:         &Plan{ID: uuid.New(), Code: "enterprise", ProductLimit: nil, IsActive: true},
		productCount: 999,
	})

	if err := service.CheckProductLimit(context.Background(), noopDB{}, uuid.New(), uuid.New()); err != nil {
		t.Fatalf("CheckProductLimit error = %v", err)
	}
}

func TestRequireFeatureBlocksDisabledPOS(t *testing.T) {
	t.Parallel()

	service := NewService(noopDB{}, &fakePlanRepository{
		plan: &Plan{ID: uuid.New(), Code: "starter", IsActive: true, CanUsePOS: false},
	})

	assertCode(t, service.RequireFeature(context.Background(), noopDB{}, uuid.New(), FeaturePOS), apperror.CodeForbidden)
}

func TestRequireFeatureBlocksDisabledCourier(t *testing.T) {
	t.Parallel()

	service := NewService(noopDB{}, &fakePlanRepository{
		plan: &Plan{ID: uuid.New(), Code: "starter", IsActive: true, CanUseCourier: false},
	})

	assertCode(t, service.RequireFeature(context.Background(), noopDB{}, uuid.New(), FeatureCourier), apperror.CodeForbidden)
}

func TestRequireFeatureBlocksDisabledDiscovery(t *testing.T) {
	t.Parallel()

	service := NewService(noopDB{}, &fakePlanRepository{
		plan: &Plan{ID: uuid.New(), Code: "starter", IsActive: true, CanUseDiscovery: false},
	})

	assertCode(t, service.RequireFeature(context.Background(), noopDB{}, uuid.New(), FeatureDiscovery), apperror.CodeForbidden)
}

type fakePlanRepository struct {
	plan         *Plan
	productCount int
	staffCount   int
}

func (f *fakePlanRepository) FindByTenantID(context.Context, db.Queryer, uuid.UUID) (*Plan, error) {
	if f.plan == nil {
		return nil, ErrPlanNotFound
	}
	return f.plan, nil
}

func (f *fakePlanRepository) CountActiveProducts(context.Context, db.Queryer, uuid.UUID, uuid.UUID) (int, error) {
	return f.productCount, nil
}

func (f *fakePlanRepository) CountActiveStaff(context.Context, db.Queryer, uuid.UUID) (int, error) {
	return f.staffCount, nil
}

type noopDB struct{}

func (noopDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (noopDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (noopDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func assertCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error, got %T: %v", err, err)
	}
	if appErr.Code != code {
		t.Fatalf("code = %s, want %s", appErr.Code, code)
	}
}
