package store

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestPublicResolveRejectsUnpublishedStore(t *testing.T) {
	t.Parallel()

	service := NewPublicService(publicNoopQueryer{}, &fakePublicStoreReader{
		record: &PublicStoreRecord{
			Store: &Store{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Slug:     "toko-bunga-ayu",
				Status:   StatusUnpublished,
			},
			TenantStatus: tenantStatusActive,
		},
	})

	_, err := service.Resolve(context.Background(), "toko-bunga-ayu")
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Resolve error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeNotFound {
		t.Fatalf("Resolve code = %s, want %s", appErr.Code, apperror.CodeNotFound)
	}
}

func TestPublicResolveRejectsSuspendedTenant(t *testing.T) {
	t.Parallel()

	service := NewPublicService(publicNoopQueryer{}, &fakePublicStoreReader{
		record: &PublicStoreRecord{
			Store: &Store{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Slug:     "toko-bunga-ayu",
				Status:   StatusPublished,
			},
			TenantStatus: "suspended",
		},
	})

	_, err := service.Resolve(context.Background(), "toko-bunga-ayu")
	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("Resolve error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeNotFound {
		t.Fatalf("Resolve code = %s, want %s", appErr.Code, apperror.CodeNotFound)
	}
}

func TestPublicResolveReturnsPublishedStore(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	storeID := uuid.New()
	service := NewPublicService(publicNoopQueryer{}, &fakePublicStoreReader{
		record: &PublicStoreRecord{
			Store: &Store{
				ID:       storeID,
				TenantID: tenantID,
				Name:     "Toko Bunga Ayu",
				Slug:     "toko-bunga-ayu",
				Status:   StatusPublished,
			},
			TenantStatus: tenantStatusTrialing,
		},
	})

	result, err := service.Resolve(context.Background(), "toko-bunga-ayu")
	if err != nil {
		t.Fatalf("Resolve error = %v", err)
	}
	if result.TenantID != tenantID || result.StoreID != storeID {
		t.Fatalf("Resolve scope = (%s, %s), want (%s, %s)", result.TenantID, result.StoreID, tenantID, storeID)
	}
	if result.Store.Slug != "toko-bunga-ayu" {
		t.Fatalf("Resolve store slug = %s, want toko-bunga-ayu", result.Store.Slug)
	}
}

type fakePublicStoreReader struct {
	record *PublicStoreRecord
}

func (f *fakePublicStoreReader) FindPublicBySlug(context.Context, db.Queryer, string) (*PublicStoreRecord, error) {
	if f.record == nil {
		return nil, ErrStoreNotFound
	}
	return f.record, nil
}

func (f *fakePublicStoreReader) ListBusinessHours(context.Context, db.Queryer, uuid.UUID, uuid.UUID) ([]BusinessHour, error) {
	return nil, nil
}

type publicNoopQueryer struct{}

func (publicNoopQueryer) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	panic("unexpected Exec")
}

func (publicNoopQueryer) Query(context.Context, string, ...any) (pgx.Rows, error) {
	panic("unexpected Query")
}

func (publicNoopQueryer) QueryRow(context.Context, string, ...any) pgx.Row {
	panic("unexpected QueryRow")
}
