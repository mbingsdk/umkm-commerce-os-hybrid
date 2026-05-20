package dashboard

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

const (
	defaultDashboardLimit = 5
	maxDashboardLimit     = 20
)

type database interface {
	db.Queryer
}

type metricsStore interface {
	Summary(context.Context, db.Queryer, uuid.UUID, uuid.UUID, DateRange) (SummaryMetrics, error)
	RecentOrders(context.Context, db.Queryer, uuid.UUID, uuid.UUID, int) ([]RecentOrder, error)
	LowStock(context.Context, db.Queryer, uuid.UUID, uuid.UUID, int) ([]LowStockItem, error)
}

type Service struct {
	db      database
	metrics metricsStore
	now     func() time.Time
}

func NewService(database database, metricsRepo metricsStore) *Service {
	return &Service{
		db:      database,
		metrics: metricsRepo,
		now:     time.Now,
	}
}

func (s *Service) Summary(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, role string) (DashboardSummaryResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return DashboardSummaryResponse{}, err
	}

	dateRange := s.todayRange()
	metrics, err := s.metrics.Summary(ctx, s.db, tenantID, storeID, dateRange)
	if err != nil {
		return DashboardSummaryResponse{}, apperror.Internal(err)
	}

	return NewDashboardSummaryResponse(dateRange, metrics, role), nil
}

func (s *Service) RecentOrders(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, limit int) ([]RecentOrderResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, err
	}

	normalizedLimit, err := normalizeLimit(limit)
	if err != nil {
		return nil, err
	}

	items, err := s.metrics.RecentOrders(ctx, s.db, tenantID, storeID, normalizedLimit)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	response := make([]RecentOrderResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewRecentOrderResponse(item))
	}
	return response, nil
}

func (s *Service) LowStock(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, limit int) ([]LowStockResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, err
	}

	normalizedLimit, err := normalizeLimit(limit)
	if err != nil {
		return nil, err
	}

	items, err := s.metrics.LowStock(ctx, s.db, tenantID, storeID, normalizedLimit)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	response := make([]LowStockResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewLowStockResponse(item))
	}
	return response, nil
}

func (s *Service) todayRange() DateRange {
	now := s.now().UTC()
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return DateRange{From: from, To: from.AddDate(0, 0, 1)}
}

func normalizeLimit(limit int) (int, error) {
	if limit == 0 {
		return defaultDashboardLimit, nil
	}
	if limit < 0 {
		return 0, invalidField("limit", "Limit must be greater than zero")
	}
	if limit > maxDashboardLimit {
		return maxDashboardLimit, nil
	}
	return limit, nil
}

func validateScope(tenantID uuid.UUID, storeID uuid.UUID) error {
	if tenantID == uuid.Nil {
		return invalidField("tenant_id", "Tenant is required")
	}
	if storeID == uuid.Nil {
		return invalidField("store_id", "Store is required")
	}
	return nil
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{
		{"field": field, "message": message},
	})
}
