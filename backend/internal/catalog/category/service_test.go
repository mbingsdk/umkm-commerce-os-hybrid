package category

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestCreateRejectsParentOutsideTenantScope(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	storeID := uuid.New()
	parentID := uuid.New()
	repo := &fakeCategoryStore{
		findByID: func(_ context.Context, _ db.Queryer, gotTenantID uuid.UUID, gotStoreID uuid.UUID, gotCategoryID uuid.UUID) (*Category, error) {
			if gotTenantID != tenantID || gotStoreID != storeID || gotCategoryID != parentID {
				t.Fatalf("FindByID scope = (%s, %s, %s), want (%s, %s, %s)", gotTenantID, gotStoreID, gotCategoryID, tenantID, storeID, parentID)
			}
			return nil, ErrCategoryNotFound
		},
	}
	service := NewService(noopQueryer{}, repo)

	_, err := service.Create(context.Background(), tenantID, storeID, CreateInput{
		ParentID: &parentID,
		Name:     "Bouquet",
		Slug:     "bouquet",
		IsActive: true,
	})
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Create error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeValidation {
		t.Fatalf("Create code = %s, want %s", appErr.Code, apperror.CodeValidation)
	}
}

type fakeCategoryStore struct {
	findByID func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Category, error)
}

func (f *fakeCategoryStore) List(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ListFilters) ([]Category, error) {
	return nil, nil
}

func (f *fakeCategoryStore) Create(context.Context, db.Queryer, CreateParams) (*Category, error) {
	return nil, nil
}

func (f *fakeCategoryStore) FindByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	categoryID uuid.UUID,
) (*Category, error) {
	if f.findByID == nil {
		return nil, ErrCategoryNotFound
	}
	return f.findByID(ctx, q, tenantID, storeID, categoryID)
}

func (f *fakeCategoryStore) Update(context.Context, db.Queryer, UpdateParams) (*Category, error) {
	return nil, nil
}

func (f *fakeCategoryStore) SoftDelete(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return nil
}

type noopQueryer struct{}

func (noopQueryer) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func (noopQueryer) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (noopQueryer) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}
