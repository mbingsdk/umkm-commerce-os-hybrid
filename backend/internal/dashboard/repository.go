package dashboard

import (
	"context"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Summary(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	dateRange DateRange,
) (SummaryMetrics, error) {
	const query = `
		WITH online_sales AS (
			SELECT COALESCE(SUM(grand_total), 0)::bigint AS total
			FROM orders
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND source <> 'pos'
			  AND payment_status = 'paid'
			  AND status NOT IN ('cancelled', 'returned', 'refunded')
			  AND COALESCE(paid_at, updated_at, created_at) >= $3
			  AND COALESCE(paid_at, updated_at, created_at) < $4
		),
		pos_sales AS (
			SELECT COALESCE(SUM(grand_total), 0)::bigint AS total
			FROM pos_transactions
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND status = 'completed'
			  AND created_at >= $3
			  AND created_at < $4
		),
		today_orders AS (
			SELECT COUNT(*)::bigint AS total
			FROM orders
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND created_at >= $3
			  AND created_at < $4
		),
		pending_orders AS (
			SELECT COUNT(*)::bigint AS total
			FROM orders
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND status IN ('pending', 'confirmed', 'processing', 'ready_to_ship')
		),
		low_stock AS (
			SELECT COUNT(*)::bigint AS total
			FROM product_stock_snapshots s
			JOIN products p
			  ON p.id = s.product_id
			 AND p.tenant_id = s.tenant_id
			 AND p.store_id = s.store_id
			WHERE s.tenant_id = $1
			  AND s.store_id = $2
			  AND p.deleted_at IS NULL
			  AND s.quantity_available <= s.low_stock_threshold
		),
		expenses_today AS (
			SELECT COALESCE(SUM(amount), 0)::bigint AS total
			FROM expenses
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND deleted_at IS NULL
			  AND expense_date >= $3::date
			  AND expense_date < $4::date
		)
		SELECT
			online_sales.total,
			pos_sales.total,
			expenses_today.total,
			today_orders.total,
			pending_orders.total,
			low_stock.total
		FROM online_sales, pos_sales, expenses_today, today_orders, pending_orders, low_stock
	`

	var metrics SummaryMetrics
	err := q.QueryRow(ctx, query, tenantID, storeID, dateRange.From, dateRange.To).Scan(
		&metrics.OnlineSalesToday,
		&metrics.POSSalesToday,
		&metrics.ExpenseToday,
		&metrics.TodayOrderCount,
		&metrics.PendingOrdersCount,
		&metrics.LowStockCount,
	)
	if err != nil {
		return SummaryMetrics{}, err
	}

	return metrics, nil
}

func (r *Repository) RecentOrders(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	limit int,
) ([]RecentOrder, error) {
	const query = `
		SELECT
			id,
			order_number,
			customer_name,
			grand_total,
			status,
			payment_status,
			created_at
		FROM orders
		WHERE tenant_id = $1
		  AND store_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]RecentOrder, 0)
	for rows.Next() {
		var item RecentOrder
		if err := rows.Scan(
			&item.OrderID,
			&item.OrderNumber,
			&item.CustomerName,
			&item.TotalAmount,
			&item.Status,
			&item.PaymentStatus,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) LowStock(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	limit int,
) ([]LowStockItem, error) {
	const query = `
		SELECT
			p.id,
			p.name,
			COALESCE(p.sku, ''),
			s.quantity_available,
			s.low_stock_threshold
		FROM product_stock_snapshots s
		JOIN products p
		  ON p.id = s.product_id
		 AND p.tenant_id = s.tenant_id
		 AND p.store_id = s.store_id
		WHERE s.tenant_id = $1
		  AND s.store_id = $2
		  AND p.deleted_at IS NULL
		  AND s.quantity_available <= s.low_stock_threshold
		ORDER BY s.quantity_available ASC, p.name ASC, p.id ASC
		LIMIT $3
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]LowStockItem, 0)
	for rows.Next() {
		var item LowStockItem
		if err := rows.Scan(
			&item.ProductID,
			&item.ProductName,
			&item.SKU,
			&item.AvailableQuantity,
			&item.LowStockThreshold,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
