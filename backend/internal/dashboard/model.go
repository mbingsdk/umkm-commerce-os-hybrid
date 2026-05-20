package dashboard

import (
	"time"

	"github.com/google/uuid"
)

type DateRange struct {
	From time.Time
	To   time.Time
}

type SummaryMetrics struct {
	OnlineSalesToday   int64
	POSSalesToday      int64
	ExpenseToday       int64
	TodayOrderCount    int64
	PendingOrdersCount int64
	LowStockCount      int64
}

type RecentOrder struct {
	OrderID       uuid.UUID
	OrderNumber   string
	CustomerName  string
	TotalAmount   int64
	Status        string
	PaymentStatus string
	CreatedAt     time.Time
}

type LowStockItem struct {
	ProductID         uuid.UUID
	ProductName       string
	SKU               string
	AvailableQuantity int
	LowStockThreshold int
}
