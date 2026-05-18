package upload

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/storage"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type assetStore interface {
	Store(ctx context.Context, input storage.StoreInput) (storage.Asset, error)
}

type Service struct {
	assets assetStore
}

type CreateInput struct {
	TenantID uuid.UUID
	Folder   string
	Data     []byte
}

func NewService(assets assetStore) *Service {
	return &Service{assets: assets}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Response, error) {
	input.Folder = strings.TrimSpace(input.Folder)
	if !allowedFolder(input.Folder) {
		return Response{}, apperror.Validation("Validation failed", []map[string]string{
			{"field": "folder", "message": "Folder must be products, stores, or banners"},
		})
	}

	asset, err := s.assets.Store(ctx, storage.StoreInput{
		TenantID: input.TenantID,
		Folder:   input.Folder,
		Data:     input.Data,
	})
	if err != nil {
		return Response{}, err
	}

	return Response{
		URL:      asset.URL,
		MIMEType: asset.MIMEType,
		Size:     asset.Size,
	}, nil
}

func allowedFolder(folder string) bool {
	switch folder {
	case "products", "stores", "banners":
		return true
	default:
		return false
	}
}
