package order

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var ErrOrderNotFound = errors.New("order not found")
var ErrStockSnapshotNotFound = errors.New("stock snapshot not found")

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) List(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters ListFilters,
) ([]Order, []int, error) {
	const query = `
		SELECT
			o.id,
			o.tenant_id,
			o.store_id,
			o.customer_id,
			o.order_number,
			o.source,
			o.status,
			o.payment_status,
			COALESCE(o.shipment_status, ''),
			o.subtotal,
			o.discount_total,
			o.shipping_cost,
			o.tax_total,
			o.grand_total,
			o.customer_name,
			o.customer_phone,
			COALESCE(o.customer_email, ''),
			COALESCE(o.shipping_address, ''),
			COALESCE(o.shipping_city, ''),
			COALESCE(o.shipping_province, ''),
			COALESCE(o.shipping_postal_code, ''),
			COALESCE(o.customer_note, ''),
			COALESCE(o.internal_note, ''),
			o.confirmed_at,
			o.paid_at,
			o.completed_at,
			o.cancelled_at,
			o.created_at,
			o.updated_at,
			COUNT(oi.id)::int AS item_count
		FROM orders o
		LEFT JOIN order_items oi
		  ON oi.tenant_id = o.tenant_id
		 AND oi.order_id = o.id
		WHERE o.tenant_id = $1
		  AND o.store_id = $2
		  AND ($3::text IS NULL OR o.status = $3)
		  AND ($4::text IS NULL OR o.payment_status = $4)
		  AND ($5::text IS NULL OR o.source = $5)
		  AND (
			$6 = ''
			OR o.order_number ILIKE '%' || $6 || '%'
			OR o.customer_phone ILIKE '%' || $6 || '%'
			OR o.customer_name ILIKE '%' || $6 || '%'
		  )
		  AND ($7::timestamptz IS NULL OR o.created_at >= $7)
		  AND ($8::timestamptz IS NULL OR o.created_at < $8)
		  AND (
			($9::timestamptz IS NULL AND $10::uuid IS NULL)
			OR (o.created_at, o.id) < ($9::timestamptz, $10::uuid)
		  )
		GROUP BY o.id
		ORDER BY o.created_at DESC, o.id DESC
		LIMIT $11
	`

	var cursorCreatedAt any
	var cursorID any
	if filters.Cursor != nil {
		cursorCreatedAt = filters.Cursor.CreatedAt
		cursorID = filters.Cursor.ID
	}

	rows, err := q.Query(
		ctx,
		query,
		tenantID,
		storeID,
		filters.Status,
		filters.PaymentStatus,
		filters.Source,
		filters.Query,
		filters.DateFrom,
		filters.DateTo,
		cursorCreatedAt,
		cursorID,
		filters.Limit,
	)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	orders := make([]Order, 0)
	itemCounts := make([]int, 0)
	for rows.Next() {
		item, itemCount, err := scanOrderWithItemCount(rows)
		if err != nil {
			return nil, nil, err
		}
		orders = append(orders, *item)
		itemCounts = append(itemCounts, itemCount)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return orders, itemCounts, nil
}

func (r *Repository) FindByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
) (*Order, error) {
	const query = selectOrderSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		LIMIT 1
	`

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderID))
}

func (r *Repository) LockByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
) (*Order, error) {
	const query = selectOrderSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		FOR UPDATE
	`

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderID))
}

func (r *Repository) ListItems(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID) ([]Item, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			order_id,
			product_id,
			product_name,
			COALESCE(sku, ''),
			quantity,
			unit_price,
			discount_total,
			subtotal,
			created_at
		FROM order_items
		WHERE tenant_id = $1
		  AND order_id = $2
		ORDER BY created_at ASC, id ASC
	`

	rows, err := q.Query(ctx, query, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.SKU,
			&item.Quantity,
			&item.UnitPrice,
			&item.DiscountTotal,
			&item.Subtotal,
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

func (r *Repository) ListStatusLogs(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID) ([]StatusLog, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			order_id,
			COALESCE(from_status, ''),
			to_status,
			COALESCE(note, ''),
			created_by,
			created_at
		FROM order_status_logs
		WHERE tenant_id = $1
		  AND order_id = $2
		ORDER BY created_at ASC, id ASC
	`

	rows, err := q.Query(ctx, query, tenantID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StatusLog, 0)
	for rows.Next() {
		var item StatusLog
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.OrderID,
			&item.FromStatus,
			&item.ToStatus,
			&item.Note,
			&item.CreatedBy,
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

func (r *Repository) ListReservationSummary(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
) ([]ReservationSummary, error) {
	const query = `
		SELECT status, COALESCE(SUM(quantity), 0)::int, COUNT(*)::int
		FROM stock_reservations
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND order_id = $3
		GROUP BY status
		ORDER BY status ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ReservationSummary, 0)
	for rows.Next() {
		var item ReservationSummary
		if err := rows.Scan(&item.Status, &item.Quantity, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) LockActiveReservationsByOrder(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
) ([]StockReservation, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			product_id,
			order_id,
			quantity,
			status,
			created_at
		FROM stock_reservations
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND order_id = $3
		  AND status IN ('active', 'confirmed')
		ORDER BY product_id ASC, id ASC
		FOR UPDATE
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StockReservation, 0)
	for rows.Next() {
		var item StockReservation
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.StoreID,
			&item.ProductID,
			&item.OrderID,
			&item.Quantity,
			&item.Status,
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

func (r *Repository) LockStockSnapshots(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productIDs []uuid.UUID,
) ([]StockSnapshot, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available
		FROM product_stock_snapshots
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND product_id = ANY($3::uuid[])
		ORDER BY product_id ASC
		FOR UPDATE
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StockSnapshot, 0, len(productIDs))
	for rows.Next() {
		var item StockSnapshot
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.StoreID,
			&item.ProductID,
			&item.QuantityOnHand,
			&item.QuantityReserved,
			&item.QuantityAvailable,
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

func (r *Repository) UpdateStatus(ctx context.Context, q db.Queryer, params UpdateStatusParams) (*Order, error) {
	const query = `
		UPDATE orders
		SET status = $4,
		    confirmed_at = CASE WHEN $4 = 'confirmed' THEN COALESCE(confirmed_at, now()) ELSE confirmed_at END,
		    completed_at = CASE WHEN $4 = 'completed' THEN COALESCE(completed_at, now()) ELSE completed_at END,
		    cancelled_at = CASE WHEN $4 = 'cancelled' THEN COALESCE(cancelled_at, now()) ELSE cancelled_at END,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		RETURNING
			id,
			tenant_id,
			store_id,
			customer_id,
			order_number,
			source,
			status,
			payment_status,
			COALESCE(shipment_status, ''),
			subtotal,
			discount_total,
			shipping_cost,
			tax_total,
			grand_total,
			customer_name,
			customer_phone,
			COALESCE(customer_email, ''),
			COALESCE(shipping_address, ''),
			COALESCE(shipping_city, ''),
			COALESCE(shipping_province, ''),
			COALESCE(shipping_postal_code, ''),
			COALESCE(customer_note, ''),
			COALESCE(internal_note, ''),
			confirmed_at,
			paid_at,
			completed_at,
			cancelled_at,
			created_at,
			updated_at
	`

	return scanOrder(q.QueryRow(ctx, query, params.TenantID, params.StoreID, params.OrderID, params.Status))
}

func (r *Repository) UpdateStockSnapshot(ctx context.Context, q db.Queryer, params UpdateStockSnapshotParams) error {
	const query = `
		UPDATE product_stock_snapshots
		SET quantity_reserved = $4,
		    quantity_available = $5,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND product_id = $3
	`

	tag, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.QuantityReserved,
		params.QuantityAvailable,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrStockSnapshotNotFound
	}
	return nil
}

func (r *Repository) ReleaseReservations(ctx context.Context, q db.Queryer, params ReleaseReservationsParams) error {
	if len(params.ReservationIDs) == 0 {
		return nil
	}

	const query = `
		UPDATE stock_reservations
		SET status = $4,
		    released_at = COALESCE(released_at, now())
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = ANY($3::uuid[])
		  AND status IN ('active', 'confirmed')
	`

	_, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ReservationIDs,
		params.Status,
	)
	return err
}

func (r *Repository) CreateStockMovement(ctx context.Context, q db.Queryer, params CreateStockMovementParams) error {
	const query = `
		INSERT INTO stock_movements (
			tenant_id,
			store_id,
			product_id,
			movement_type,
			quantity,
			balance_after,
			reference_type,
			reference_id,
			note,
			created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, NULLIF($9, ''), $10)
	`

	_, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.MovementType,
		params.Quantity,
		params.BalanceAfter,
		params.ReferenceType,
		params.ReferenceID,
		params.Note,
		params.CreatedBy,
	)
	return err
}

func (r *Repository) CreateStatusLog(ctx context.Context, q db.Queryer, params CreateStatusLogParams) (*StatusLog, error) {
	const query = `
		INSERT INTO order_status_logs (
			tenant_id,
			order_id,
			from_status,
			to_status,
			note,
			created_by
		)
		VALUES ($1, $2, NULLIF($3, ''), $4, NULLIF($5, ''), $6)
		RETURNING
			id,
			tenant_id,
			order_id,
			COALESCE(from_status, ''),
			to_status,
			COALESCE(note, ''),
			created_by,
			created_at
	`

	var item StatusLog
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.OrderID,
		params.FromStatus,
		params.ToStatus,
		params.Note,
		params.CreatedBy,
	).Scan(
		&item.ID,
		&item.TenantID,
		&item.OrderID,
		&item.FromStatus,
		&item.ToStatus,
		&item.Note,
		&item.CreatedBy,
		&item.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &item, nil
}

const selectOrderSQL = `
	SELECT
		id,
		tenant_id,
		store_id,
		customer_id,
		order_number,
		source,
		status,
		payment_status,
		COALESCE(shipment_status, ''),
		subtotal,
		discount_total,
		shipping_cost,
		tax_total,
		grand_total,
		customer_name,
		customer_phone,
		COALESCE(customer_email, ''),
		COALESCE(shipping_address, ''),
		COALESCE(shipping_city, ''),
		COALESCE(shipping_province, ''),
		COALESCE(shipping_postal_code, ''),
		COALESCE(customer_note, ''),
		COALESCE(internal_note, ''),
		confirmed_at,
		paid_at,
		completed_at,
		cancelled_at,
		created_at,
		updated_at
	FROM orders
`

func scanOrder(row pgx.Row) (*Order, error) {
	var item Order
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.StoreID,
		&item.CustomerID,
		&item.OrderNumber,
		&item.Source,
		&item.Status,
		&item.PaymentStatus,
		&item.ShipmentStatus,
		&item.Subtotal,
		&item.DiscountTotal,
		&item.ShippingCost,
		&item.TaxTotal,
		&item.GrandTotal,
		&item.CustomerName,
		&item.CustomerPhone,
		&item.CustomerEmail,
		&item.ShippingAddress,
		&item.ShippingCity,
		&item.ShippingProvince,
		&item.ShippingPostalCode,
		&item.CustomerNote,
		&item.InternalNote,
		&item.ConfirmedAt,
		&item.PaidAt,
		&item.CompletedAt,
		&item.CancelledAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		if isCheckViolation(err) {
			return nil, err
		}
		return nil, err
	}
	return &item, nil
}

func scanOrderWithItemCount(row pgx.Row) (*Order, int, error) {
	var item Order
	var itemCount int
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.StoreID,
		&item.CustomerID,
		&item.OrderNumber,
		&item.Source,
		&item.Status,
		&item.PaymentStatus,
		&item.ShipmentStatus,
		&item.Subtotal,
		&item.DiscountTotal,
		&item.ShippingCost,
		&item.TaxTotal,
		&item.GrandTotal,
		&item.CustomerName,
		&item.CustomerPhone,
		&item.CustomerEmail,
		&item.ShippingAddress,
		&item.ShippingCity,
		&item.ShippingProvince,
		&item.ShippingPostalCode,
		&item.CustomerNote,
		&item.InternalNote,
		&item.ConfirmedAt,
		&item.PaidAt,
		&item.CompletedAt,
		&item.CancelledAt,
		&item.CreatedAt,
		&item.UpdatedAt,
		&itemCount,
	); err != nil {
		return nil, 0, err
	}
	return &item, itemCount, nil
}

func isCheckViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23514"
}
