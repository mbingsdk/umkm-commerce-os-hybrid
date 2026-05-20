package finance

import (
	"time"

	"github.com/google/uuid"
)

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type ExpenseResponse struct {
	ID            string         `json:"id"`
	CategoryID    *string        `json:"category_id,omitempty"`
	Category      string         `json:"category,omitempty"`
	CategoryName  string         `json:"category_name,omitempty"`
	Title         string         `json:"title"`
	Amount        int64          `json:"amount"`
	ExpenseDate   string         `json:"expense_date"`
	PaymentMethod string         `json:"payment_method,omitempty"`
	Note          string         `json:"note,omitempty"`
	CreatedBy     *ActorResponse `json:"created_by,omitempty"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type FinanceSummaryResponse struct {
	DateFrom             string `json:"date_from"`
	DateTo               string `json:"date_to"`
	GrossSales           int64  `json:"gross_sales"`
	OnlineSales          int64  `json:"online_sales"`
	POSSales             int64  `json:"pos_sales"`
	TotalExpenses        int64  `json:"total_expenses"`
	NetEstimate          int64  `json:"net_estimate"`
	OrderCount           int64  `json:"order_count"`
	POSTransactionCount  int64  `json:"pos_transaction_count"`
	TotalOrders          int64  `json:"total_orders"`
	TotalPOSTransactions int64  `json:"total_pos_transactions"`
	AverageOrderValue    int64  `json:"average_order_value"`
	Note                 string `json:"note"`
}

type DailyReportResponse struct {
	DateFrom string                   `json:"date_from"`
	DateTo   string                   `json:"date_to"`
	Days     []DailyReportDayResponse `json:"days"`
	Summary  FinanceMetricsResponse   `json:"summary"`
	Note     string                   `json:"note"`
}

type DailyReportDayResponse struct {
	Date string `json:"date"`
	FinanceMetricsResponse
}

type MonthlyReportResponse struct {
	Year    int                          `json:"year"`
	Month   *int                         `json:"month,omitempty"`
	Months  []MonthlyReportMonthResponse `json:"months"`
	Summary FinanceMetricsResponse       `json:"summary"`
	Note    string                       `json:"note"`
}

type MonthlyReportMonthResponse struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	FinanceMetricsResponse
}

type FinanceMetricsResponse struct {
	GrossSales           int64 `json:"gross_sales"`
	OnlineSales          int64 `json:"online_sales"`
	POSSales             int64 `json:"pos_sales"`
	TotalExpenses        int64 `json:"total_expenses"`
	NetEstimate          int64 `json:"net_estimate"`
	OrderCount           int64 `json:"order_count"`
	POSTransactionCount  int64 `json:"pos_transaction_count"`
	TotalOrders          int64 `json:"total_orders"`
	TotalPOSTransactions int64 `json:"total_pos_transactions"`
	AverageOrderValue    int64 `json:"average_order_value"`
}

type ActorResponse struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
}

func NewExpenseResponse(expense Expense) ExpenseResponse {
	var categoryID *string
	if expense.CategoryID != nil {
		value := expense.CategoryID.String()
		categoryID = &value
	}

	var actor *ActorResponse
	if expense.CreatedBy != nil && *expense.CreatedBy != uuid.Nil {
		actor = &ActorResponse{
			ID:   expense.CreatedBy.String(),
			Name: expense.CreatedByName,
		}
	}

	return ExpenseResponse{
		ID:            expense.ID.String(),
		CategoryID:    categoryID,
		Category:      expense.CategorySlug,
		CategoryName:  expense.CategoryName,
		Title:         expense.Title,
		Amount:        expense.Amount,
		ExpenseDate:   expense.ExpenseDate.Format(dateFormat),
		PaymentMethod: expense.PaymentMethod,
		Note:          expense.Note,
		CreatedBy:     actor,
		CreatedAt:     expense.CreatedAt.Format(timeFormatRFC3339),
		UpdatedAt:     expense.UpdatedAt.Format(timeFormatRFC3339),
	}
}

func NewFinanceSummaryResponse(dateRange DateRange, totals FinanceTotals) FinanceSummaryResponse {
	metrics := NewFinanceMetricsResponse(totals)
	return FinanceSummaryResponse{
		DateFrom:             dateRange.From.Format(dateFormat),
		DateTo:               inclusiveDateTo(dateRange).Format(dateFormat),
		GrossSales:           metrics.GrossSales,
		OnlineSales:          metrics.OnlineSales,
		POSSales:             metrics.POSSales,
		TotalExpenses:        metrics.TotalExpenses,
		NetEstimate:          metrics.NetEstimate,
		OrderCount:           metrics.OrderCount,
		POSTransactionCount:  metrics.POSTransactionCount,
		TotalOrders:          metrics.TotalOrders,
		TotalPOSTransactions: metrics.TotalPOSTransactions,
		AverageOrderValue:    metrics.AverageOrderValue,
		Note:                 netEstimateNote,
	}
}

func NewDailyReportResponse(dateRange DateRange, rows []DailyFinanceRow) DailyReportResponse {
	days := make([]DailyReportDayResponse, 0, len(rows))
	var total FinanceTotals
	for _, row := range rows {
		total = addFinanceTotals(total, row.FinanceTotals)
		days = append(days, DailyReportDayResponse{
			Date:                   row.Date.Format(dateFormat),
			FinanceMetricsResponse: NewFinanceMetricsResponse(row.FinanceTotals),
		})
	}

	return DailyReportResponse{
		DateFrom: dateRange.From.Format(dateFormat),
		DateTo:   inclusiveDateTo(dateRange).Format(dateFormat),
		Days:     days,
		Summary:  NewFinanceMetricsResponse(total),
		Note:     netEstimateNote,
	}
}

func NewMonthlyReportResponse(filter MonthlyReportFilter, rows []MonthlyFinanceRow) MonthlyReportResponse {
	months := make([]MonthlyReportMonthResponse, 0, len(rows))
	var total FinanceTotals
	for _, row := range rows {
		total = addFinanceTotals(total, row.FinanceTotals)
		months = append(months, MonthlyReportMonthResponse{
			Year:                   row.MonthStart.Year(),
			Month:                  int(row.MonthStart.Month()),
			FinanceMetricsResponse: NewFinanceMetricsResponse(row.FinanceTotals),
		})
	}

	return MonthlyReportResponse{
		Year:    filter.Year,
		Month:   filter.Month,
		Months:  months,
		Summary: NewFinanceMetricsResponse(total),
		Note:    netEstimateNote,
	}
}

func NewFinanceMetricsResponse(totals FinanceTotals) FinanceMetricsResponse {
	grossSales := totals.OnlineSales + totals.POSSales
	transactionCount := totals.OrderCount + totals.POSTransactionCount
	var average int64
	if transactionCount > 0 {
		average = grossSales / transactionCount
	}

	return FinanceMetricsResponse{
		GrossSales:           grossSales,
		OnlineSales:          totals.OnlineSales,
		POSSales:             totals.POSSales,
		TotalExpenses:        totals.TotalExpenses,
		NetEstimate:          grossSales - totals.TotalExpenses,
		OrderCount:           totals.OrderCount,
		POSTransactionCount:  totals.POSTransactionCount,
		TotalOrders:          totals.OrderCount,
		TotalPOSTransactions: totals.POSTransactionCount,
		AverageOrderValue:    average,
	}
}

func addFinanceTotals(left FinanceTotals, right FinanceTotals) FinanceTotals {
	return FinanceTotals{
		OnlineSales:         left.OnlineSales + right.OnlineSales,
		POSSales:            left.POSSales + right.POSSales,
		TotalExpenses:       left.TotalExpenses + right.TotalExpenses,
		OrderCount:          left.OrderCount + right.OrderCount,
		POSTransactionCount: left.POSTransactionCount + right.POSTransactionCount,
	}
}

func inclusiveDateTo(dateRange DateRange) time.Time {
	if dateRange.To.IsZero() {
		return dateRange.To
	}
	return dateRange.To.AddDate(0, 0, -1)
}

const (
	dateFormat        = "2006-01-02"
	timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
	netEstimateNote   = "Net estimate does not include detailed cost of goods sold yet."
)
