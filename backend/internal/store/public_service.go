package store

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

const (
	tenantStatusActive   = "active"
	tenantStatusTrialing = "trialing"
)

type publicStoreReader interface {
	FindPublicBySlug(ctx context.Context, q db.Queryer, slug string) (*PublicStoreRecord, error)
	ListBusinessHours(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) ([]BusinessHour, error)
}

type PublicService struct {
	db     db.Queryer
	stores publicStoreReader
}

type PublicContext struct {
	TenantID uuid.UUID
	StoreID  uuid.UUID
	Store    Store
}

func NewPublicService(database db.Queryer, stores publicStoreReader) *PublicService {
	return &PublicService{
		db:     database,
		stores: stores,
	}
}

func (s *PublicService) Resolve(ctx context.Context, slug string) (PublicContext, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return PublicContext{}, apperror.NotFound("Store not found")
	}

	record, err := s.stores.FindPublicBySlug(ctx, s.db, slug)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			return PublicContext{}, apperror.NotFound("Store not found")
		}
		return PublicContext{}, apperror.Internal(err)
	}
	if record.Store == nil ||
		record.Store.Status != StatusPublished ||
		(record.TenantStatus != tenantStatusActive && record.TenantStatus != tenantStatusTrialing) {
		return PublicContext{}, apperror.NotFound("Store not found")
	}

	return PublicContext{
		TenantID: record.Store.TenantID,
		StoreID:  record.Store.ID,
		Store:    *record.Store,
	}, nil
}

func (s *PublicService) Get(ctx context.Context, slug string) (PublicResponse, error) {
	currentStore, err := s.Resolve(ctx, slug)
	if err != nil {
		return PublicResponse{}, err
	}

	hours, err := s.stores.ListBusinessHours(ctx, s.db, currentStore.TenantID, currentStore.StoreID)
	if err != nil {
		return PublicResponse{}, apperror.Internal(err)
	}

	return NewPublicResponse(&currentStore.Store, hours), nil
}
