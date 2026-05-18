package store

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type storeStore interface {
	FindCurrentByTenantID(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (*Store, error)
	UpdateProfile(ctx context.Context, q db.Queryer, params UpdateProfileParams) (*Store, error)
	UpdateStatus(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, status string, publishedAt *time.Time) (*Store, error)
	ReplaceBusinessHours(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, items []BusinessHour) error
}

type auditRecorder interface {
	Create(ctx context.Context, q db.Queryer, entry audit.Entry) error
}

type Service struct {
	db        database
	stores    storeStore
	auditLogs auditRecorder
	now       func() time.Time
}

type Actor struct {
	UserID    uuid.UUID
	IPAddress string
	UserAgent string
}

type UpdateProfileInput struct {
	Name           string
	Description    string
	Phone          string
	Whatsapp       string
	Email          string
	Address        string
	City           string
	Province       string
	PostalCode     string
	IsDiscoverable bool
}

func NewService(database database, stores storeStore, auditLogs auditRecorder) *Service {
	return &Service{
		db:        database,
		stores:    stores,
		auditLogs: auditLogs,
		now:       time.Now,
	}
}

func (s *Service) GetCurrent(ctx context.Context, tenantID uuid.UUID) (Response, error) {
	currentStore, err := s.stores.FindCurrentByTenantID(ctx, s.db, tenantID)
	if err != nil {
		if errors.Is(err, ErrStoreNotFound) {
			return Response{}, apperror.NotFound("Store not found")
		}
		return Response{}, apperror.Internal(err)
	}

	return NewResponse(currentStore), nil
}

func (s *Service) UpdateProfile(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	actor Actor,
	input UpdateProfileInput,
) (Response, error) {
	normalized, err := validateUpdateProfile(input)
	if err != nil {
		return Response{}, err
	}

	var updated Response
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		before, err := s.stores.FindCurrentByTenantID(ctx, tx, tenantID)
		if err != nil {
			if errors.Is(err, ErrStoreNotFound) {
				return apperror.NotFound("Store not found")
			}
			return apperror.Internal(err)
		}

		after, err := s.stores.UpdateProfile(ctx, tx, UpdateProfileParams{
			TenantID:       tenantID,
			StoreID:        storeID,
			Name:           normalized.Name,
			Description:    normalized.Description,
			Phone:          normalized.Phone,
			Whatsapp:       normalized.Whatsapp,
			Email:          normalized.Email,
			Address:        normalized.Address,
			City:           normalized.City,
			Province:       normalized.Province,
			PostalCode:     normalized.PostalCode,
			IsDiscoverable: normalized.IsDiscoverable,
		})
		if err != nil {
			if errors.Is(err, ErrStoreNotFound) {
				return apperror.NotFound("Store not found")
			}
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    tenantID,
			StoreID:     &storeID,
			ActorUserID: &actor.UserID,
			Action:      "store.update_profile",
			EntityType:  "store",
			EntityID:    &storeID,
			BeforeData:  NewResponse(before),
			AfterData:   NewResponse(after),
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}

		updated = NewResponse(after)
		return nil
	})
	if err != nil {
		return Response{}, err
	}

	return updated, nil
}

func (s *Service) Publish(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	actor Actor,
) (Response, error) {
	return s.changeStatus(ctx, tenantID, storeID, actor, StatusPublished, "store.publish")
}

func (s *Service) Unpublish(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	actor Actor,
) (Response, error) {
	return s.changeStatus(ctx, tenantID, storeID, actor, StatusUnpublished, "store.unpublish")
}

func (s *Service) ReplaceBusinessHours(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	actor Actor,
	items []BusinessHour,
) (BusinessHoursResponse, error) {
	normalized, err := validateBusinessHours(items)
	if err != nil {
		return BusinessHoursResponse{}, err
	}

	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		if _, err := s.stores.FindCurrentByTenantID(ctx, tx, tenantID); err != nil {
			if errors.Is(err, ErrStoreNotFound) {
				return apperror.NotFound("Store not found")
			}
			return apperror.Internal(err)
		}

		if err := s.stores.ReplaceBusinessHours(ctx, tx, tenantID, storeID, normalized); err != nil {
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    tenantID,
			StoreID:     &storeID,
			ActorUserID: &actor.UserID,
			Action:      "store.update_business_hours",
			EntityType:  "store",
			EntityID:    &storeID,
			AfterData:   normalized,
			IPAddress:   actor.IPAddress,
			UserAgent:   actor.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}

		return nil
	})
	if err != nil {
		return BusinessHoursResponse{}, err
	}

	responseItems := make([]BusinessHourResponseItem, 0, len(normalized))
	for _, item := range normalized {
		responseItems = append(responseItems, BusinessHourResponseItem{
			DayOfWeek: item.DayOfWeek,
			OpenTime:  item.OpenTime,
			CloseTime: item.CloseTime,
			IsClosed:  item.IsClosed,
		})
	}

	return BusinessHoursResponse{Items: responseItems}, nil
}

func (s *Service) changeStatus(
	ctx context.Context,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	actor Actor,
	nextStatus string,
	action string,
) (Response, error) {
	var updated Response
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		before, err := s.stores.FindCurrentByTenantID(ctx, tx, tenantID)
		if err != nil {
			if errors.Is(err, ErrStoreNotFound) {
				return apperror.NotFound("Store not found")
			}
			return apperror.Internal(err)
		}

		if nextStatus == StatusPublished {
			if err := validatePublishable(before); err != nil {
				return err
			}
		}

		var publishedAt *time.Time
		if nextStatus == StatusPublished {
			now := s.now().UTC()
			publishedAt = &now
		}

		after, err := s.stores.UpdateStatus(ctx, tx, tenantID, storeID, nextStatus, publishedAt)
		if err != nil {
			if errors.Is(err, ErrStoreNotFound) {
				return apperror.NotFound("Store not found")
			}
			return apperror.Internal(err)
		}

		if err := s.auditLogs.Create(ctx, tx, audit.Entry{
			TenantID:    tenantID,
			StoreID:     &storeID,
			ActorUserID: &actor.UserID,
			Action:      action,
			EntityType:  "store",
			EntityID:    &storeID,
			BeforeData: map[string]any{
				"status":       before.Status,
				"published_at": before.PublishedAt,
			},
			AfterData: map[string]any{
				"status":       after.Status,
				"published_at": after.PublishedAt,
			},
			IPAddress: actor.IPAddress,
			UserAgent: actor.UserAgent,
		}); err != nil {
			return apperror.Internal(err)
		}

		updated = NewResponse(after)
		return nil
	})
	if err != nil {
		return Response{}, err
	}

	return updated, nil
}

func validateUpdateProfile(input UpdateProfileInput) (UpdateProfileInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Phone = strings.TrimSpace(input.Phone)
	input.Whatsapp = strings.TrimSpace(input.Whatsapp)
	input.Email = strings.TrimSpace(input.Email)
	input.Address = strings.TrimSpace(input.Address)
	input.City = strings.TrimSpace(input.City)
	input.Province = strings.TrimSpace(input.Province)
	input.PostalCode = strings.TrimSpace(input.PostalCode)

	var details []map[string]string
	if input.Name == "" {
		details = append(details, map[string]string{"field": "name", "message": "Store name is required"})
	}
	if input.Email != "" && !isValidEmail(input.Email) {
		details = append(details, map[string]string{"field": "email", "message": "Store email is invalid"})
	}

	if len(details) > 0 {
		return UpdateProfileInput{}, apperror.Validation("Validation failed", details)
	}

	return input, nil
}

func validatePublishable(store *Store) error {
	var details []map[string]string
	if strings.TrimSpace(store.Name) == "" {
		details = append(details, map[string]string{"field": "name", "message": "Store name is required before publishing"})
	}
	if strings.TrimSpace(store.Slug) == "" {
		details = append(details, map[string]string{"field": "slug", "message": "Store slug is required before publishing"})
	}
	if strings.TrimSpace(store.Phone) == "" && strings.TrimSpace(store.Whatsapp) == "" {
		details = append(details, map[string]string{"field": "phone", "message": "Phone or whatsapp is required before publishing"})
	}
	if strings.TrimSpace(store.City) == "" {
		details = append(details, map[string]string{"field": "city", "message": "City is required before publishing"})
	}

	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}

	return nil
}

func validateBusinessHours(items []BusinessHour) ([]BusinessHour, error) {
	if len(items) == 0 {
		return nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "items", "message": "At least one business hour item is required"},
		})
	}

	normalized := make([]BusinessHour, 0, len(items))
	seenDays := make(map[int]struct{}, len(items))
	var details []map[string]string

	for _, item := range items {
		item.OpenTime = strings.TrimSpace(item.OpenTime)
		item.CloseTime = strings.TrimSpace(item.CloseTime)

		if item.DayOfWeek < 1 || item.DayOfWeek > 7 {
			details = append(details, map[string]string{
				"field":   "items",
				"message": "day_of_week must be between 1 and 7",
			})
			continue
		}
		if _, ok := seenDays[item.DayOfWeek]; ok {
			details = append(details, map[string]string{
				"field":   "items",
				"message": "day_of_week must be unique",
			})
			continue
		}
		seenDays[item.DayOfWeek] = struct{}{}

		if item.IsClosed {
			item.OpenTime = ""
			item.CloseTime = ""
			normalized = append(normalized, item)
			continue
		}

		openTime, openErr := time.Parse("15:04", item.OpenTime)
		closeTime, closeErr := time.Parse("15:04", item.CloseTime)
		if openErr != nil || closeErr != nil {
			details = append(details, map[string]string{
				"field":   "items",
				"message": "open_time and close_time must use HH:MM format",
			})
			continue
		}
		if !openTime.Before(closeTime) {
			details = append(details, map[string]string{
				"field":   "items",
				"message": "open_time must be before close_time",
			})
			continue
		}

		normalized = append(normalized, item)
	}

	if len(details) > 0 {
		return nil, apperror.Validation("Validation failed", details)
	}

	return normalized, nil
}

func isValidEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}
