package tenant

import (
	"testing"

	"github.com/google/uuid"
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
