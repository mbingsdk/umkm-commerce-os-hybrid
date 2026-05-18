package storage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestStoreRejectsUnsupportedMIME(t *testing.T) {
	t.Parallel()

	store := NewLocal(t.TempDir(), "http://localhost:8080/uploads", 1024)
	_, err := store.Store(context.Background(), StoreInput{
		TenantID: uuid.New(),
		Folder:   "products",
		Data:     []byte("plain text"),
	})
	assertValidationError(t, err)
}

func TestStoreRejectsFilesOverConfiguredLimit(t *testing.T) {
	t.Parallel()

	store := NewLocal(t.TempDir(), "http://localhost:8080/uploads", 4)
	_, err := store.Store(context.Background(), StoreInput{
		TenantID: uuid.New(),
		Folder:   "products",
		Data:     pngBytes(),
	})
	assertValidationError(t, err)
}

func TestStoreRejectsTraversalFolder(t *testing.T) {
	t.Parallel()

	store := NewLocal(t.TempDir(), "http://localhost:8080/uploads", 1024)
	_, err := store.Store(context.Background(), StoreInput{
		TenantID: uuid.New(),
		Folder:   "../products",
		Data:     pngBytes(),
	})
	assertValidationError(t, err)
}

func assertValidationError(t *testing.T, err error) {
	t.Helper()

	appErr, ok := err.(*apperror.AppError)
	if !ok {
		t.Fatalf("error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeValidation {
		t.Fatalf("error code = %s, want %s", appErr.Code, apperror.CodeValidation)
	}
}

func pngBytes() []byte {
	return []byte{
		0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n',
		0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01,
	}
}
