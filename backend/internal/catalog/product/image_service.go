package product

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/storage"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type imageStore interface {
	Create(ctx context.Context, q db.Queryer, params CreateImageParams) (*Image, error)
	Delete(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, productID uuid.UUID, imageID uuid.UUID) error
}

type assetStore interface {
	Store(ctx context.Context, input storage.StoreInput) (storage.Asset, error)
	Delete(ctx context.Context, key string) error
}

type AttachImageInput struct {
	AltText   string
	IsPrimary bool
	SortOrder int
	Data      []byte
}

func (s *Service) AttachImage(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
	input AttachImageInput,
) (ImageResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ImageResponse{}, err
	}
	if productID == uuid.Nil {
		return ImageResponse{}, invalidField("product_id", "Product is required")
	}

	input.AltText = strings.TrimSpace(input.AltText)
	if input.SortOrder < 0 {
		return ImageResponse{}, invalidField("sort_order", "Sort order must be zero or greater")
	}

	if _, err := s.products.FindByID(ctx, s.db, tenantID, storeID, productID); err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return ImageResponse{}, apperror.NotFound("Product not found")
		}
		return ImageResponse{}, apperror.Internal(err)
	}

	asset, err := s.assets.Store(ctx, storage.StoreInput{
		TenantID: tenantID,
		Folder:   "products",
		Data:     input.Data,
	})
	if err != nil {
		return ImageResponse{}, err
	}

	var created *Image
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		var createErr error
		created, createErr = s.images.Create(ctx, tx, CreateImageParams{
			TenantID:  tenantID,
			StoreID:   storeID,
			ProductID: productID,
			URL:       asset.URL,
			AltText:   input.AltText,
			IsPrimary: input.IsPrimary,
			SortOrder: input.SortOrder,
		})
		if createErr != nil {
			if errors.Is(createErr, ErrProductNotFound) {
				return apperror.NotFound("Product not found")
			}
			return apperror.Internal(createErr)
		}
		return nil
	})
	if err != nil {
		_ = s.assets.Delete(ctx, asset.Key)
		return ImageResponse{}, err
	}

	return ImageResponse{
		ID:        created.ID,
		URL:       created.URL,
		AltText:   created.AltText,
		IsPrimary: created.IsPrimary,
		SortOrder: created.SortOrder,
	}, nil
}

func (s *Service) DeleteImage(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
	imageID uuid.UUID,
) error {
	if err := validateScope(tenantID, storeID); err != nil {
		return err
	}
	if productID == uuid.Nil {
		return invalidField("product_id", "Product is required")
	}
	if imageID == uuid.Nil {
		return invalidField("image_id", "Image is required")
	}

	if err := s.images.Delete(ctx, s.db, tenantID, storeID, productID, imageID); err != nil {
		if errors.Is(err, ErrProductImageNotFound) {
			return apperror.NotFound("Product image not found")
		}
		return apperror.Internal(err)
	}

	return nil
}
