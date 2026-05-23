package store

import (
	"context"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/testutil"
)

func TestPublicStorefrontSafetyHidesUnpublishedStoreFixture(t *testing.T) {
	fixtures := testutil.NewSecurityFixtures()
	unpublished := Store{
		ID:       fixtures.Stores.A.ID,
		TenantID: fixtures.Tenants.A.ID,
		Name:     fixtures.Stores.A.Name,
		Slug:     fixtures.Stores.A.Slug,
		Status:   StatusUnpublished,
	}
	service := NewPublicService(publicNoopQueryer{}, &fakePublicStoreReader{
		record: &PublicStoreRecord{
			Store:        &unpublished,
			TenantStatus: tenantStatusActive,
		},
	})

	_, err := service.Resolve(context.Background(), fixtures.Stores.A.Slug)
	assertPublicStoreNotFound(t, err)
}

func assertPublicStoreNotFound(t *testing.T, err error) {
	t.Helper()

	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Resolve error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeNotFound {
		t.Fatalf("Resolve code = %s, want %s", appErr.Code, apperror.CodeNotFound)
	}
}
