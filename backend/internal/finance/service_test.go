package finance

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
)

var (
	finTenantA    = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	finStoreA     = uuid.MustParse("22222222-2222-2222-2222-222222222222")
	finTenantB    = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	finStoreB     = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	finExpenseA   = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	finExpenseB   = uuid.MustParse("66666666-6666-6666-6666-666666666666")
	finActorID    = uuid.MustParse("77777777-7777-7777-7777-777777777777")
	finCategoryID = uuid.MustParse("88888888-8888-8888-8888-888888888888")
)

func TestCreateExpenseValidation(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newFinanceTestService()

	_, err := service.CreateExpense(context.Background(), finTenantA, finStoreA, CreateExpenseInput{
		ActorUserID: finActorID,
		Amount:      0,
	})
	assertFinanceAppErrorCode(t, err, apperror.CodeValidation)
	if len(repo.expenses) != 2 {
		t.Fatalf("expense count = %d, want unchanged 2", len(repo.expenses))
	}
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestExpenseAmountMustBePositive(t *testing.T) {
	service, _, _, _ := newFinanceTestService()

	_, err := service.CreateExpense(context.Background(), finTenantA, finStoreA, CreateExpenseInput{
		ActorUserID: finActorID,
		Title:       "Beli plastik wrapping",
		Amount:      -1,
		ExpenseDate: "2026-05-20",
	})
	assertFinanceAppErrorCode(t, err, apperror.CodeValidation)
}

func TestCreateExpenseCreatesAuditAndOutbox(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newFinanceTestService()

	result, err := service.CreateExpense(context.Background(), finTenantA, finStoreA, CreateExpenseInput{
		ActorUserID:   finActorID,
		Category:      "operasional",
		Title:         "Beli plastik wrapping",
		Amount:        150000,
		ExpenseDate:   "2026-05-20",
		PaymentMethod: PaymentMethodCash,
		Note:          "Catatan finansial tidak ikut audit/outbox",
	})
	if err != nil {
		t.Fatalf("CreateExpense error = %v", err)
	}
	if result.ID == "" || result.Amount != 150000 || result.Category != "operasional" {
		t.Fatalf("CreateExpense result = %#v", result)
	}
	if len(repo.expenses) != 3 {
		t.Fatalf("expense count = %d, want 3", len(repo.expenses))
	}
	if len(auditRepo.entries) != 1 || auditRepo.entries[0].Action != AuditActionExpenseCreated {
		t.Fatalf("audit entries = %#v", auditRepo.entries)
	}
	if strings.Contains(mustMarshalFinance(t, auditRepo.entries[0].AfterData), "Catatan finansial") {
		t.Fatalf("audit after_data leaked financial note: %#v", auditRepo.entries[0].AfterData)
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventExpenseChanged {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
	if strings.Contains(string(outboxRepo.events[0].Payload), "Catatan finansial") {
		t.Fatalf("outbox payload leaked financial note: %s", outboxRepo.events[0].Payload)
	}
}

func TestExpenseListIsTenantScoped(t *testing.T) {
	service, _, _, _ := newFinanceTestService()

	items, _, err := service.ListExpenses(context.Background(), finTenantA, finStoreA, ListExpenseFilters{})
	if err != nil {
		t.Fatalf("ListExpenses error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expenses len = %d, want 1", len(items))
	}
	if items[0].ID != finExpenseA.String() {
		t.Fatalf("expense id = %s, want %s", items[0].ID, finExpenseA)
	}
}

func TestTenantCannotUpdateOtherTenantExpense(t *testing.T) {
	service, _, auditRepo, outboxRepo := newFinanceTestService()
	amount := int64(250000)

	_, err := service.UpdateExpense(context.Background(), finTenantA, finStoreA, finExpenseB, UpdateExpenseInput{
		ActorUserID: finActorID,
		Amount:      &amount,
	})
	assertFinanceAppErrorCode(t, err, apperror.CodeNotFound)
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestDeleteIsTenantScoped(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newFinanceTestService()

	_, err := service.DeleteExpense(context.Background(), finTenantA, finStoreA, finExpenseB, finActorID)
	assertFinanceAppErrorCode(t, err, apperror.CodeNotFound)
	if repo.expenses[finExpenseB].DeletedAt != nil {
		t.Fatalf("tenant B expense was deleted")
	}
	if len(auditRepo.entries) != 0 {
		t.Fatalf("audit entries = %d, want 0", len(auditRepo.entries))
	}
	if len(outboxRepo.events) != 0 {
		t.Fatalf("outbox events = %d, want 0", len(outboxRepo.events))
	}
}

func TestDeleteExpenseSoftDeletesWithAuditAndOutbox(t *testing.T) {
	service, repo, auditRepo, outboxRepo := newFinanceTestService()

	result, err := service.DeleteExpense(context.Background(), finTenantA, finStoreA, finExpenseA, finActorID)
	if err != nil {
		t.Fatalf("DeleteExpense error = %v", err)
	}
	if result.ID != finExpenseA.String() {
		t.Fatalf("deleted id = %s, want %s", result.ID, finExpenseA)
	}
	if repo.expenses[finExpenseA].DeletedAt == nil {
		t.Fatalf("expense was not soft deleted")
	}
	if len(auditRepo.entries) != 1 || auditRepo.entries[0].Action != AuditActionExpenseDeleted {
		t.Fatalf("audit entries = %#v", auditRepo.entries)
	}
	if len(outboxRepo.events) != 1 || outboxRepo.events[0].EventType != EventExpenseChanged {
		t.Fatalf("outbox events = %#v", outboxRepo.events)
	}
}

func TestFinanceSummaryUsesPaidOnlineOrdersOnly(t *testing.T) {
	service, _, _, _ := newFinanceTestService()
	mayRange := DateRange{
		From: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	result, err := service.Summary(context.Background(), finTenantA, finStoreA, mayRange)
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.OnlineSales != 100000 {
		t.Fatalf("online_sales = %d, want 100000", result.OnlineSales)
	}
	if result.OrderCount != 1 {
		t.Fatalf("order_count = %d, want 1", result.OrderCount)
	}
}

func TestFinanceSummaryUsesCompletedPOSTransactionsOnly(t *testing.T) {
	service, _, _, _ := newFinanceTestService()
	mayRange := DateRange{
		From: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	result, err := service.Summary(context.Background(), finTenantA, finStoreA, mayRange)
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.POSSales != 50000 {
		t.Fatalf("pos_sales = %d, want 50000", result.POSSales)
	}
	if result.POSTransactionCount != 1 {
		t.Fatalf("pos_transaction_count = %d, want 1", result.POSTransactionCount)
	}
}

func TestFinanceSummaryExpensesReduceNetEstimate(t *testing.T) {
	service, _, _, _ := newFinanceTestService()
	mayRange := DateRange{
		From: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	result, err := service.Summary(context.Background(), finTenantA, finStoreA, mayRange)
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.GrossSales != 150000 || result.TotalExpenses != 150000 || result.NetEstimate != 0 {
		t.Fatalf("summary = %#v, want gross 150000 expenses 150000 net 0", result)
	}
	if result.Note == "" {
		t.Fatalf("summary note is empty")
	}
}

func TestFinanceSummaryIsTenantScoped(t *testing.T) {
	service, _, _, _ := newFinanceTestService()
	mayRange := DateRange{
		From: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	}

	result, err := service.Summary(context.Background(), finTenantA, finStoreA, mayRange)
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.GrossSales == 1050000 || result.TotalExpenses == 350000 {
		t.Fatalf("tenant A summary appears to include tenant B data: %#v", result)
	}
	if result.GrossSales != 150000 {
		t.Fatalf("gross_sales = %d, want tenant A only 150000", result.GrossSales)
	}
}

func TestFinanceSummaryDateRangeFiltersWork(t *testing.T) {
	service, _, _, _ := newFinanceTestService()
	nextDayRange := DateRange{
		From: time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 5, 22, 0, 0, 0, 0, time.UTC),
	}

	result, err := service.Summary(context.Background(), finTenantA, finStoreA, nextDayRange)
	if err != nil {
		t.Fatalf("Summary error = %v", err)
	}
	if result.GrossSales != 0 || result.TotalExpenses != 0 {
		t.Fatalf("summary for empty range = %#v, want zero totals", result)
	}
}

func TestFinanceDailyAndMonthlyReports(t *testing.T) {
	service, _, _, _ := newFinanceTestService()
	dateRange := DateRange{
		From: time.Date(2026, 5, 20, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, 5, 21, 0, 0, 0, 0, time.UTC),
	}

	daily, err := service.DailyReport(context.Background(), finTenantA, finStoreA, dateRange)
	if err != nil {
		t.Fatalf("DailyReport error = %v", err)
	}
	if len(daily.Days) != 1 || daily.Summary.GrossSales != 150000 || daily.Summary.TotalExpenses != 150000 {
		t.Fatalf("daily report = %#v", daily)
	}

	month := 5
	monthly, err := service.MonthlyReport(context.Background(), finTenantA, finStoreA, MonthlyReportFilter{Year: 2026, Month: &month})
	if err != nil {
		t.Fatalf("MonthlyReport error = %v", err)
	}
	if len(monthly.Months) != 1 || monthly.Summary.GrossSales != 150000 || monthly.Summary.TotalExpenses != 150000 {
		t.Fatalf("monthly report = %#v", monthly)
	}
}

func newFinanceTestService() (*Service, *fakeFinanceRepository, *fakeFinanceAuditRepository, *fakeFinanceOutboxRepository) {
	now := time.Date(2026, 5, 20, 8, 0, 0, 0, time.UTC)
	repo := &fakeFinanceRepository{
		categories: map[uuid.UUID]ExpenseCategory{
			finCategoryID: {
				ID:        finCategoryID,
				Name:      "Operasional",
				Slug:      "operasional",
				IsSystem:  true,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		expenses: map[uuid.UUID]Expense{
			finExpenseA: {
				ID:            finExpenseA,
				TenantID:      finTenantA,
				StoreID:       finStoreA,
				CategoryID:    &finCategoryID,
				CategoryName:  "Operasional",
				CategorySlug:  "operasional",
				Title:         "Plastik wrapping",
				Amount:        150000,
				ExpenseDate:   now,
				PaymentMethod: PaymentMethodCash,
				CreatedBy:     &finActorID,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			finExpenseB: {
				ID:          finExpenseB,
				TenantID:    finTenantB,
				StoreID:     finStoreB,
				Title:       "Biaya tenant B",
				Amount:      200000,
				ExpenseDate: now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
		onlineOrders: []fakeFinanceOrder{
			{
				TenantID:      finTenantA,
				StoreID:       finStoreA,
				Source:        "storefront",
				Status:        "completed",
				PaymentStatus: "paid",
				GrandTotal:    100000,
				PaidAt:        now,
			},
			{
				TenantID:      finTenantA,
				StoreID:       finStoreA,
				Source:        "storefront",
				Status:        "pending",
				PaymentStatus: "unpaid",
				GrandTotal:    300000,
				PaidAt:        now,
			},
			{
				TenantID:      finTenantA,
				StoreID:       finStoreA,
				Source:        "storefront",
				Status:        "cancelled",
				PaymentStatus: "paid",
				GrandTotal:    400000,
				PaidAt:        now,
			},
			{
				TenantID:      finTenantA,
				StoreID:       finStoreA,
				Source:        "storefront",
				Status:        "completed",
				PaymentStatus: "paid",
				GrandTotal:    250000,
				PaidAt:        now.AddDate(0, 0, -30),
			},
			{
				TenantID:      finTenantB,
				StoreID:       finStoreB,
				Source:        "storefront",
				Status:        "completed",
				PaymentStatus: "paid",
				GrandTotal:    900000,
				PaidAt:        now,
			},
		},
		posTransactions: []fakeFinancePOSTransaction{
			{
				TenantID:   finTenantA,
				StoreID:    finStoreA,
				Status:     "completed",
				GrandTotal: 50000,
				CreatedAt:  now,
			},
			{
				TenantID:   finTenantA,
				StoreID:    finStoreA,
				Status:     "cancelled",
				GrandTotal: 70000,
				CreatedAt:  now,
			},
			{
				TenantID:   finTenantB,
				StoreID:    finStoreB,
				Status:     "completed",
				GrandTotal: 150000,
				CreatedAt:  now,
			},
		},
		now: now,
	}
	auditRepo := &fakeFinanceAuditRepository{}
	outboxRepo := &fakeFinanceOutboxRepository{}
	service := NewService(fakeFinanceDB{}, repo, auditRepo, outboxRepo)
	return service, repo, auditRepo, outboxRepo
}

func assertFinanceAppErrorCode(t *testing.T, err error, code apperror.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error code %s, got nil", code)
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("expected app error %s, got %T: %v", code, err, err)
	}
	if appErr.Code != code {
		t.Fatalf("error code = %s, want %s", appErr.Code, code)
	}
}

func mustMarshalFinance(t *testing.T, value any) string {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal value: %v", err)
	}
	return string(raw)
}

type fakeFinanceDB struct{}

func (fakeFinanceDB) WithTx(_ context.Context, fn func(tx db.Tx) error) error {
	return fn(fakeFinanceDB{})
}

func (fakeFinanceDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (fakeFinanceDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
}

func (fakeFinanceDB) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakeFinanceRepository struct {
	categories      map[uuid.UUID]ExpenseCategory
	expenses        map[uuid.UUID]Expense
	onlineOrders    []fakeFinanceOrder
	posTransactions []fakeFinancePOSTransaction
	now             time.Time
}

type fakeFinanceOrder struct {
	TenantID      uuid.UUID
	StoreID       uuid.UUID
	Source        string
	Status        string
	PaymentStatus string
	GrandTotal    int64
	PaidAt        time.Time
}

type fakeFinancePOSTransaction struct {
	TenantID   uuid.UUID
	StoreID    uuid.UUID
	Status     string
	GrandTotal int64
	CreatedAt  time.Time
}

func (f *fakeFinanceRepository) Summary(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, dateRange DateRange) (FinanceTotals, error) {
	return f.totalsForRange(tenantID, storeID, dateRange), nil
}

func (f *fakeFinanceRepository) DailyReport(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, dateRange DateRange) ([]DailyFinanceRow, error) {
	rows := make([]DailyFinanceRow, 0)
	for day := dateRange.From; day.Before(dateRange.To); day = day.AddDate(0, 0, 1) {
		rows = append(rows, DailyFinanceRow{
			Date:          day,
			FinanceTotals: f.totalsForRange(tenantID, storeID, DateRange{From: day, To: day.AddDate(0, 0, 1)}),
		})
	}
	return rows, nil
}

func (f *fakeFinanceRepository) MonthlyReport(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, dateRange DateRange) ([]MonthlyFinanceRow, error) {
	rows := make([]MonthlyFinanceRow, 0)
	for month := dateRange.From; month.Before(dateRange.To); month = month.AddDate(0, 1, 0) {
		rows = append(rows, MonthlyFinanceRow{
			MonthStart:    month,
			FinanceTotals: f.totalsForRange(tenantID, storeID, DateRange{From: month, To: month.AddDate(0, 1, 0)}),
		})
	}
	return rows, nil
}

func (f *fakeFinanceRepository) totalsForRange(tenantID uuid.UUID, storeID uuid.UUID, dateRange DateRange) FinanceTotals {
	var totals FinanceTotals
	for _, order := range f.onlineOrders {
		if order.TenantID != tenantID || order.StoreID != storeID {
			continue
		}
		if order.Source == "pos" || order.PaymentStatus != "paid" || !financeOrderStatusCounts(order.Status) {
			continue
		}
		if !timeInRange(order.PaidAt, dateRange) {
			continue
		}
		totals.OnlineSales += order.GrandTotal
		totals.OrderCount++
	}
	for _, transaction := range f.posTransactions {
		if transaction.TenantID != tenantID || transaction.StoreID != storeID || transaction.Status != "completed" {
			continue
		}
		if !timeInRange(transaction.CreatedAt, dateRange) {
			continue
		}
		totals.POSSales += transaction.GrandTotal
		totals.POSTransactionCount++
	}
	for _, expense := range f.expenses {
		if expense.TenantID != tenantID || expense.StoreID != storeID || expense.DeletedAt != nil {
			continue
		}
		if !dateInRange(expense.ExpenseDate, dateRange) {
			continue
		}
		totals.TotalExpenses += expense.Amount
	}
	return totals
}

func (f *fakeFinanceRepository) ListExpenses(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListExpenseFilters) ([]Expense, error) {
	items := make([]Expense, 0)
	for _, expense := range f.expenses {
		if expense.TenantID != tenantID || expense.StoreID != storeID || expense.DeletedAt != nil {
			continue
		}
		items = append(items, expense)
	}
	return items, nil
}

func (f *fakeFinanceRepository) FindExpenseByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, expenseID uuid.UUID) (*Expense, error) {
	expense, ok := f.expenses[expenseID]
	if !ok || expense.TenantID != tenantID || expense.StoreID != storeID || expense.DeletedAt != nil {
		return nil, ErrExpenseNotFound
	}
	return cloneExpense(expense), nil
}

func (f *fakeFinanceRepository) FindCategoryByID(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, categoryID uuid.UUID) (*ExpenseCategory, error) {
	category, ok := f.categories[categoryID]
	if !ok || !categoryAllowed(category, tenantID, storeID) {
		return nil, ErrCategoryNotFound
	}
	return &category, nil
}

func (f *fakeFinanceRepository) FindCategoryBySlug(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, slug string) (*ExpenseCategory, error) {
	for _, category := range f.categories {
		if category.Slug == slug && categoryAllowed(category, tenantID, storeID) {
			return &category, nil
		}
	}
	return nil, ErrCategoryNotFound
}

func (f *fakeFinanceRepository) CreateExpense(_ context.Context, _ db.Queryer, params CreateExpenseParams) (*Expense, error) {
	expense := Expense{
		ID:            uuid.New(),
		TenantID:      params.TenantID,
		StoreID:       params.StoreID,
		CategoryID:    params.CategoryID,
		Title:         params.Title,
		Amount:        params.Amount,
		ExpenseDate:   params.ExpenseDate,
		PaymentMethod: params.PaymentMethod,
		Note:          params.Note,
		CreatedBy:     &params.CreatedBy,
		UpdatedBy:     &params.CreatedBy,
		CreatedAt:     f.now,
		UpdatedAt:     f.now,
	}
	f.applyCategory(&expense)
	f.expenses[expense.ID] = expense
	return cloneExpense(expense), nil
}

func (f *fakeFinanceRepository) UpdateExpense(_ context.Context, _ db.Queryer, params UpdateExpenseParams) (*Expense, error) {
	expense, ok := f.expenses[params.ExpenseID]
	if !ok || expense.TenantID != params.TenantID || expense.StoreID != params.StoreID || expense.DeletedAt != nil {
		return nil, ErrExpenseNotFound
	}
	expense.CategoryID = params.CategoryID
	expense.CategoryName = ""
	expense.CategorySlug = ""
	expense.Title = params.Title
	expense.Amount = params.Amount
	expense.ExpenseDate = params.ExpenseDate
	expense.PaymentMethod = params.PaymentMethod
	expense.Note = params.Note
	expense.UpdatedBy = &params.UpdatedBy
	expense.UpdatedAt = f.now.Add(time.Minute)
	f.applyCategory(&expense)
	f.expenses[params.ExpenseID] = expense
	return cloneExpense(expense), nil
}

func (f *fakeFinanceRepository) SoftDeleteExpense(_ context.Context, _ db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, expenseID uuid.UUID, actorUserID uuid.UUID) (*Expense, error) {
	expense, ok := f.expenses[expenseID]
	if !ok || expense.TenantID != tenantID || expense.StoreID != storeID || expense.DeletedAt != nil {
		return nil, ErrExpenseNotFound
	}
	deletedAt := f.now.Add(time.Minute)
	expense.DeletedAt = &deletedAt
	expense.UpdatedBy = &actorUserID
	expense.UpdatedAt = deletedAt
	f.expenses[expenseID] = expense
	return cloneExpense(expense), nil
}

func (f *fakeFinanceRepository) applyCategory(expense *Expense) {
	if expense.CategoryID == nil {
		return
	}
	category, ok := f.categories[*expense.CategoryID]
	if !ok {
		return
	}
	expense.CategoryName = category.Name
	expense.CategorySlug = category.Slug
}

func categoryAllowed(category ExpenseCategory, tenantID uuid.UUID, storeID uuid.UUID) bool {
	if category.TenantID == nil && category.StoreID == nil {
		return true
	}
	if category.TenantID == nil || *category.TenantID != tenantID {
		return false
	}
	return category.StoreID == nil || *category.StoreID == storeID
}

func cloneExpense(expense Expense) *Expense {
	clone := expense
	return &clone
}

func financeOrderStatusCounts(status string) bool {
	switch status {
	case "cancelled", "returned", "refunded":
		return false
	default:
		return true
	}
}

func timeInRange(value time.Time, dateRange DateRange) bool {
	return !value.Before(dateRange.From) && value.Before(dateRange.To)
}

func dateInRange(value time.Time, dateRange DateRange) bool {
	day := normalizeDate(value)
	return !day.Before(dateRange.From) && day.Before(dateRange.To)
}

type fakeFinanceAuditRepository struct {
	entries []audit.Entry
}

func (f *fakeFinanceAuditRepository) Create(_ context.Context, _ db.Queryer, entry audit.Entry) error {
	f.entries = append(f.entries, entry)
	return nil
}

type fakeFinanceOutboxRepository struct {
	events []outbox.InsertEventParams
}

func (f *fakeFinanceOutboxRepository) Insert(_ context.Context, _ db.Queryer, params outbox.InsertEventParams) (*outbox.Event, error) {
	if !json.Valid(params.Payload) {
		return nil, errors.New("invalid json payload")
	}
	f.events = append(f.events, params)
	return &outbox.Event{}, nil
}
