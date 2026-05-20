package dashboard

import "github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"

const (
	dateFormat        = "2006-01-02"
	timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"
	netEstimateNote   = "net_estimate_today belum termasuk perhitungan HPP/COGS detail."
)

type DashboardSummaryResponse struct {
	Date               string   `json:"date"`
	TodaySales         int64    `json:"today_sales"`
	TodayOrderCount    *int64   `json:"today_order_count,omitempty"`
	PendingOrdersCount *int64   `json:"pending_orders_count,omitempty"`
	LowStockCount      *int64   `json:"low_stock_count,omitempty"`
	POSSalesToday      *int64   `json:"pos_sales_today,omitempty"`
	OnlineSalesToday   *int64   `json:"online_sales_today,omitempty"`
	ExpenseToday       *int64   `json:"expense_today,omitempty"`
	NetEstimateToday   *int64   `json:"net_estimate_today,omitempty"`
	VisibleCards       []string `json:"visible_cards"`
	HiddenCards        []string `json:"hidden_cards,omitempty"`
	Note               string   `json:"note,omitempty"`
}

type RecentOrderResponse struct {
	OrderID       string `json:"order_id"`
	OrderNumber   string `json:"order_number"`
	CustomerName  string `json:"customer_name"`
	TotalAmount   int64  `json:"total_amount"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	CreatedAt     string `json:"created_at"`
}

type LowStockResponse struct {
	ProductID         string `json:"product_id"`
	ProductName       string `json:"product_name"`
	SKU               string `json:"sku,omitempty"`
	AvailableQuantity int    `json:"available_quantity"`
	LowStockThreshold int    `json:"low_stock_threshold"`
}

func NewDashboardSummaryResponse(dateRange DateRange, metrics SummaryMetrics, role string) DashboardSummaryResponse {
	response := DashboardSummaryResponse{
		Date:         dateRange.From.Format(dateFormat),
		VisibleCards: []string{"today_sales"},
		HiddenCards:  make([]string, 0, 8),
	}

	canSeeOrders := permission.Allowed(role, permission.OrderRead) || permission.Allowed(role, permission.DashboardReadRecentOrders)
	canSeeLowStock := permission.Allowed(role, permission.InventoryRead) || permission.Allowed(role, permission.DashboardReadLowStock)
	canSeePOS := permission.Allowed(role, permission.POSReadTransaction)
	canSeeFinance := permission.Allowed(role, permission.FinanceReadSummary)

	if canSeeOrders {
		response.TodaySales += metrics.OnlineSalesToday
		response.TodayOrderCount = int64Ptr(metrics.TodayOrderCount)
		response.PendingOrdersCount = int64Ptr(metrics.PendingOrdersCount)
		response.OnlineSalesToday = int64Ptr(metrics.OnlineSalesToday)
		response.VisibleCards = append(response.VisibleCards, "today_order_count", "pending_orders_count", "online_sales_today")
	} else {
		response.HiddenCards = append(response.HiddenCards, "today_order_count", "pending_orders_count", "online_sales_today")
	}

	if canSeeLowStock {
		response.LowStockCount = int64Ptr(metrics.LowStockCount)
		response.VisibleCards = append(response.VisibleCards, "low_stock_count")
	} else {
		response.HiddenCards = append(response.HiddenCards, "low_stock_count")
	}

	if canSeePOS {
		response.TodaySales += metrics.POSSalesToday
		response.POSSalesToday = int64Ptr(metrics.POSSalesToday)
		response.VisibleCards = append(response.VisibleCards, "pos_sales_today")
	} else {
		response.HiddenCards = append(response.HiddenCards, "pos_sales_today")
	}

	if canSeeFinance {
		netEstimate := response.TodaySales - metrics.ExpenseToday
		response.ExpenseToday = int64Ptr(metrics.ExpenseToday)
		response.NetEstimateToday = int64Ptr(netEstimate)
		response.VisibleCards = append(response.VisibleCards, "expense_today", "net_estimate_today")
		response.Note = netEstimateNote
	} else {
		response.HiddenCards = append(response.HiddenCards, "expense_today", "net_estimate_today")
	}

	return response
}

func NewRecentOrderResponse(item RecentOrder) RecentOrderResponse {
	return RecentOrderResponse{
		OrderID:       item.OrderID.String(),
		OrderNumber:   item.OrderNumber,
		CustomerName:  item.CustomerName,
		TotalAmount:   item.TotalAmount,
		Status:        item.Status,
		PaymentStatus: item.PaymentStatus,
		CreatedAt:     item.CreatedAt.Format(timeFormatRFC3339),
	}
}

func NewLowStockResponse(item LowStockItem) LowStockResponse {
	return LowStockResponse{
		ProductID:         item.ProductID.String(),
		ProductName:       item.ProductName,
		SKU:               item.SKU,
		AvailableQuantity: item.AvailableQuantity,
		LowStockThreshold: item.LowStockThreshold,
	}
}

func int64Ptr(value int64) *int64 {
	return &value
}
