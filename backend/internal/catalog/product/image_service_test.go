package product

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/storage"
)

func TestAttachImageUsesTenantAndStoreScope(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()
	imageID := uuid.New()
	products := &imageScopedProductStore{
		findByID: func(_ context.Context, _ db.Queryer, gotTenantID uuid.UUID, gotStoreID uuid.UUID, gotProductID uuid.UUID) (*Product, error) {
			if gotTenantID != tenantID || gotStoreID != storeID || gotProductID != productID {
				t.Fatalf("FindByID scope = (%s, %s, %s), want (%s, %s, %s)", gotTenantID, gotStoreID, gotProductID, tenantID, storeID, productID)
			}
			return &Product{ID: productID}, nil
		},
	}
	images := &recordingImageStore{
		create: func(_ context.Context, _ db.Queryer, params CreateImageParams) (*Image, error) {
			if params.TenantID != tenantID || params.StoreID != storeID || params.ProductID != productID {
				t.Fatalf("Create scope = (%s, %s, %s), want (%s, %s, %s)", params.TenantID, params.StoreID, params.ProductID, tenantID, storeID, productID)
			}
			return &Image{ID: imageID, URL: params.URL, IsPrimary: params.IsPrimary}, nil
		},
	}
	assets := &recordingAssetStore{}
	service := NewService(fakeDatabase{}, products, &fakeCategoryReader{}, &fakeStockWriter{}, images, assets)

	result, err := service.AttachImage(context.Background(), tenantID, storeID, productID, AttachImageInput{
		IsPrimary: true,
		Data:      pngBytesForProductTest(),
	})
	if err != nil {
		t.Fatalf("AttachImage error = %v", err)
	}
	if result.ID != imageID {
		t.Fatalf("AttachImage image id = %s, want %s", result.ID, imageID)
	}
}

func TestDeleteImageUsesTenantAndStoreScope(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()
	imageID := uuid.New()
	images := &recordingImageStore{
		delete: func(_ context.Context, _ db.Queryer, gotTenantID uuid.UUID, gotStoreID uuid.UUID, gotProductID uuid.UUID, gotImageID uuid.UUID) error {
			if gotTenantID != tenantID || gotStoreID != storeID || gotProductID != productID || gotImageID != imageID {
				t.Fatalf("Delete scope = (%s, %s, %s, %s), want (%s, %s, %s, %s)", gotTenantID, gotStoreID, gotProductID, gotImageID, tenantID, storeID, productID, imageID)
			}
			return nil
		},
	}
	service := NewService(fakeDatabase{}, &imageScopedProductStore{}, &fakeCategoryReader{}, &fakeStockWriter{}, images, &recordingAssetStore{})

	if err := service.DeleteImage(context.Background(), tenantID, storeID, productID, imageID); err != nil {
		t.Fatalf("DeleteImage error = %v", err)
	}
}

type imageScopedProductStore struct {
	findByID func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Product, error)
}

func (s *imageScopedProductStore) List(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ListFilters) ([]Product, error) {
	return nil, nil
}

func (s *imageScopedProductStore) Create(context.Context, db.Queryer, CreateParams) (*Product, error) {
	return nil, errors.New("unexpected Create")
}

func (s *imageScopedProductStore) FindByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID) (*Product, error) {
	if s.findByID == nil {
		return &Product{ID: productID}, nil
	}
	return s.findByID(ctx, q, tenantID, storeID, productID)
}

func (s *imageScopedProductStore) ListImages(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) ([]Image, error) {
	return nil, nil
}

func (s *imageScopedProductStore) Update(context.Context, db.Queryer, UpdateParams) error {
	return errors.New("unexpected Update")
}

func (s *imageScopedProductStore) SoftDelete(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) error {
	return errors.New("unexpected SoftDelete")
}

type recordingImageStore struct {
	create func(context.Context, db.Queryer, CreateImageParams) (*Image, error)
	delete func(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) error
}

func (s *recordingImageStore) Create(ctx context.Context, q db.Queryer, params CreateImageParams) (*Image, error) {
	if s.create == nil {
		return nil, errors.New("unexpected Create")
	}
	return s.create(ctx, q, params)
}

func (s *recordingImageStore) Delete(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID, imageID uuid.UUID) error {
	if s.delete == nil {
		return errors.New("unexpected Delete")
	}
	return s.delete(ctx, q, tenantID, storeID, productID, imageID)
}

type recordingAssetStore struct{}

func (recordingAssetStore) Store(_ context.Context, input storage.StoreInput) (storage.Asset, error) {
	return storage.Asset{
		Key:      "tenants/" + input.TenantID.String() + "/products/test.png",
		URL:      "http://localhost:8080/uploads/tenants/" + input.TenantID.String() + "/products/test.png",
		MIMEType: storage.MIMEPNG,
		Size:     int64(len(input.Data)),
	}, nil
}

func (recordingAssetStore) Delete(context.Context, string) error {
	return nil
}

func pngBytesForProductTest() []byte {
	return []byte{
		0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n',
		0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01,
	}
}
