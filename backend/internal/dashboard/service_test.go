package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

var (
	dashTenantA  = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	dashStoreA   = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	dashTenantB  = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	dashStoreB   = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	dashOrderA   = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	dashOrderB   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	dashProductA = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	dashProductB = uuid.MustParse("88888888-8888-8888-8888-888888888888")
)

func TestDashboardSummaryIsTenantScoped(t *testing.T) {
	service, _ := newDashboardTestService()

	result, err := service.Summary(context.Background(), dashTenantA, dashStoreA, string(permission.RoleOwner))
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.TodaySales != 150000 {
		t.Fatalf("today_sales = %d, want tenant A total 150000", result.TodaySales)
	}
	if result.TodayOrderCount == nil || *result.TodayOrderCount != 2 {
		t.Fatalf("today_order_count = %#v, want 2", result.TodayOrderCount)
	}
	if result.ExpenseToday == nil || *result.ExpenseToday != 20000 {
		t.Fatalf("expense_today = %#v, want 20000", result.ExpenseToday)
	}
}

func TestRecentOrdersTenantScoped(t *testing.T) {
	service, _ := newDashboardTestService()

	items, err := service.RecentOrders(context.Background(), dashTenantA, dashStoreA, 5)
	if err != nil {
		t.Fatalf("RecentOrders error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("recent orders len = %d, want 1", len(items))
	}
	if items[0].OrderID != dashOrderA.String() {
		t.Fatalf("order id = %s, want tenant A order %s", items[0].OrderID, dashOrderA)
	}
}

func TestLowStockTenantScoped(t *testing.T) {
	service, _ := newDashboardTestService()

	items, err := service.LowStock(context.Background(), dashTenantA, dashStoreA, 10)
	if err != nil {
		t.Fatalf("LowStock error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("low stock len = %d, want 1", len(items))
	}
	if items[0].ProductID != dashProductA.String() {
		t.Fatalf("product id = %s, want tenant A product %s", items[0].ProductID, dashProductA)
	}
}

func TestFinanceCardsHiddenForStaff(t *testing.T) {
	service, _ := newDashboardTestService()

	result, err := service.Summary(context.Background(), dashTenantA, dashStoreA, string(permission.RoleStaff))
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.ExpenseToday != nil || result.NetEstimateToday != nil {
		t.Fatalf("staff response leaked finance cards: %#v", result)
	}
	if result.OnlineSalesToday == nil || *result.OnlineSalesToday != 100000 {
		t.Fatalf("staff should still see operational online sales, got %#v", result.OnlineSalesToday)
	}
}

func TestCashierCannotSeeFinanceMetricsIfNotAllowed(t *testing.T) {
	if permission.Allowed(string(permission.RoleCashier), permission.DashboardReadSummary) {
		t.Fatalf("cashier should not be allowed to open dashboard summary by default")
	}
	if permission.Allowed(string(permission.RoleCashier), permission.FinanceReadSummary) {
		t.Fatalf("cashier should not be allowed to read finance summary")
	}

	service, _ := newDashboardTestService()
	result, err := service.Summary(context.Background(), dashTenantA, dashStoreA, string(permission.RoleCashier))
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.ExpenseToday != nil || result.NetEstimateToday != nil {
		t.Fatalf("cashier response leaked finance cards: %#v", result)
	}
	if result.POSSalesToday == nil || *result.POSSalesToday != 50000 {
		t.Fatalf("cashier POS metric = %#v, want 50000 if summary is ever enabled", result.POSSalesToday)
	}
}

func newDashboardTestService() (*Service, *fakeDashboardRepository) {
	repo := &fakeDashboardRepository{
		summaries: map[dashboardScope]SummaryMetrics{
			{tenantID: dashTenantA, storeID: dashStoreA}: {
				OnlineSalesToday:   100000,
				POSSalesToday:      50000,
				ExpenseToday:       20000,
				TodayOrderCount:    2,
				PendingOrdersCount: 3,
				LowStockCount:      1,
			},
			{tenantID: dashTenantB, storeID: dashStoreB}: {
				OnlineSalesToday:   999000,
				POSSalesToday:      999000,
				ExpenseToday:       999000,
				TodayOrderCount:    99,
				PendingOrdersCount: 99,
				LowStockCount:      99,
			},
		},
		recentOrders: map[dashboardScope][]RecentOrder{
			{tenantID: dashTenantA, storeID: dashStoreA}: {
				{
					OrderID:       dashOrderA,
					OrderNumber:   "ORD-A",
					CustomerName:  "Ayu",
					TotalAmount:   100000,
					Status:        "pending",
					PaymentStatus: "unpaid",
					CreatedAt:     time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
				},
			},
			{tenantID: dashTenantB, storeID: dashStoreB}: {
				{
					OrderID:       dashOrderB,
					OrderNumber:   "ORD-B",
					CustomerName:  "Budi",
					TotalAmount:   999000,
					Status:        "paid",
					PaymentStatus: "paid",
					CreatedAt:     time.Date(2026, 5, 21, 11, 0, 0, 0, time.UTC),
				},
			},
		},
		lowStock: map[dashboardScope][]LowStockItem{
			{tenantID: dashTenantA, storeID: dashStoreA}: {
				{
					ProductID:         dashProductA,
					ProductName:       "Wrapping Paper",
					SKU:               "WRP-A",
					AvailableQuantity: 2,
					LowStockThreshold: 5,
				},
			},
			{tenantID: dashTenantB, storeID: dashStoreB}: {
				{
					ProductID:         dashProductB,
					ProductName:       "Ribbon",
					SKU:               "RBN-B",
					AvailableQuantity: 1,
					LowStockThreshold: 5,
				},
			},
		},
	}

	service := NewService(nil, repo)
	service.now = func() time.Time {
		return time.Date(2026, 5, 21, 12, 0, 0, 0, time.UTC)
	}
	return service, repo
}

type dashboardScope struct {
	tenantID uuid.UUID
	storeID  uuid.UUID
}

type fakeDashboardRepository struct {
	summaries    map[dashboardScope]SummaryMetrics
	recentOrders map[dashboardScope][]RecentOrder
	lowStock     map[dashboardScope][]LowStockItem
}

func (r *fakeDashboardRepository) Summary(
	_ context.Context,
	_ db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	_ DateRange,
) (SummaryMetrics, error) {
	return r.summaries[dashboardScope{tenantID: tenantID, storeID: storeID}], nil
}

func (r *fakeDashboardRepository) RecentOrders(
	_ context.Context,
	_ db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	limit int,
) ([]RecentOrder, error) {
	items := append([]RecentOrder(nil), r.recentOrders[dashboardScope{tenantID: tenantID, storeID: storeID}]...)
	if limit > 0 && len(items) > limit {
		return items[:limit], nil
	}
	return items, nil
}

func (r *fakeDashboardRepository) LowStock(
	_ context.Context,
	_ db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	limit int,
) ([]LowStockItem, error) {
	items := append([]LowStockItem(nil), r.lowStock[dashboardScope{tenantID: tenantID, storeID: storeID}]...)
	if limit > 0 && len(items) > limit {
		return items[:limit], nil
	}
	return items, nil
}
