package tenant

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestValidateCreateStore(t *testing.T) {
	t.Parallel()

	valid := CreateStoreInput{
		UserID:     uuid.New(),
		TenantName: "Toko Bunga Ayu",
		TenantSlug: "toko-bunga-ayu",
		Store: StoreCreateInput{
			Name:  "Toko Bunga Ayu",
			Slug:  "toko-bunga-ayu",
			Email: "hello@example.com",
		},
	}

	if _, err := validateCreateStore(valid); err != nil {
		t.Fatalf("validateCreateStore(valid) error = %v", err)
	}

	invalid := valid
	invalid.TenantSlug = "Toko Bunga"
	invalid.Store.Email = "not-an-email"

	_, err := validateCreateStore(invalid)
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("validateCreateStore(invalid) error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeValidation {
		t.Fatalf("validateCreateStore(invalid) code = %s, want %s", appErr.Code, apperror.CodeValidation)
	}
}

func TestValidateAccessRejectsSuspendedTenant(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	userID := uuid.New()
	service := NewService(nil, nil, &fakeUserTenantStore{
		access: AccessRecord{
			TenantID:     tenantID,
			TenantStatus: "suspended",
			StoreID:      uuid.New(),
			Role:         MembershipRoleOwner,
		},
	}, nil, nil)

	_, err := service.ValidateAccess(context.Background(), userID, tenantID)
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("ValidateAccess error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeTenantAccessDenied {
		t.Fatalf("ValidateAccess code = %s, want %s", appErr.Code, apperror.CodeTenantAccessDenied)
	}
}

type fakeUserTenantStore struct {
	access AccessRecord
}

func (f *fakeUserTenantStore) Create(context.Context, db.Queryer, CreateMembershipParams) (*Membership, error) {
	return nil, nil
}

func (f *fakeUserTenantStore) FindActiveAccess(context.Context, db.Queryer, uuid.UUID, uuid.UUID) (*AccessRecord, error) {
	return &f.access, nil
}

func (f *fakeUserTenantStore) ListByUserID(context.Context, db.Queryer, uuid.UUID) ([]TenantListItem, error) {
	return nil, nil
}
