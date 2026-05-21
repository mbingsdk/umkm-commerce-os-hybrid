package shipment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) List(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, filters ListFilters) ([]Shipment, error) {
	var cursorCreatedAt any
	var cursorID any
	if filters.Cursor != nil {
		cursorCreatedAt = filters.Cursor.CreatedAt
		cursorID = filters.Cursor.ID
	}
	var dateFrom any
	var dateTo any
	if filters.DateFrom != nil {
		dateFrom = *filters.DateFrom
	}
	if filters.DateTo != nil {
		dateTo = *filters.DateTo
	}

	const query = `
		SELECT
			s.id,
			s.tenant_id,
			s.store_id,
			s.order_id,
			o.order_number,
			o.customer_name,
			o.customer_phone,
			s.courier_type,
			COALESCE(s.courier_name, ''),
			COALESCE(s.tracking_number, ''),
			s.status,
			s.shipping_cost,
			COALESCE(s.assigned_to_name, ''),
			COALESCE(s.assigned_to_phone, ''),
			COALESCE(s.note, ''),
			s.shipped_at,
			s.delivered_at,
			s.created_by,
			s.updated_by,
			s.created_at,
			s.updated_at
		FROM shipments s
		JOIN orders o
		  ON o.tenant_id = s.tenant_id
		 AND o.store_id = s.store_id
		 AND o.id = s.order_id
		WHERE s.tenant_id = $1
		  AND s.store_id = $2
		  AND ($3::text IS NULL OR s.status = $3)
		  AND (
			$4 = ''
			OR s.tracking_number ILIKE '%' || $4 || '%'
			OR o.order_number ILIKE '%' || $4 || '%'
			OR o.customer_name ILIKE '%' || $4 || '%'
			OR o.customer_phone ILIKE '%' || $4 || '%'
		  )
		  AND ($5::timestamptz IS NULL OR s.created_at >= $5)
		  AND ($6::timestamptz IS NULL OR s.created_at < $6)
		  AND (
			($7::timestamptz IS NULL AND $8::uuid IS NULL)
			OR (s.created_at, s.id) < ($7::timestamptz, $8::uuid)
		  )
		ORDER BY s.created_at DESC, s.id DESC
		LIMIT $9
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, filters.Status, filters.Query, dateFrom, dateTo, cursorCreatedAt, cursorID, filters.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Shipment, 0)
	for rows.Next() {
		item, err := scanShipment(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) FindByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, shipmentID uuid.UUID) (*Shipment, error) {
	const query = selectShipmentSQL + `
		WHERE s.tenant_id = $1
		  AND s.store_id = $2
		  AND s.id = $3
		LIMIT 1
	`

	return scanShipment(q.QueryRow(ctx, query, tenantID, storeID, shipmentID))
}

func (r *Repository) LockByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, shipmentID uuid.UUID) (*Shipment, error) {
	const query = selectShipmentSQL + `
		WHERE s.tenant_id = $1
		  AND s.store_id = $2
		  AND s.id = $3
		FOR UPDATE
	`

	return scanShipment(q.QueryRow(ctx, query, tenantID, storeID, shipmentID))
}

func (r *Repository) FindLatestByOrder(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*Shipment, error) {
	const query = selectShipmentSQL + `
		WHERE s.tenant_id = $1
		  AND s.store_id = $2
		  AND s.order_id = $3
		ORDER BY s.created_at DESC, s.id DESC
		LIMIT 1
	`

	return scanShipment(q.QueryRow(ctx, query, tenantID, storeID, orderID))
}

func (r *Repository) Create(ctx context.Context, q db.Queryer, params CreateShipmentParams) (*Shipment, error) {
	const query = `
		WITH inserted AS (
			INSERT INTO shipments (
				tenant_id,
				store_id,
				order_id,
				courier_type,
				courier_name,
				tracking_number,
				shipping_cost,
				assigned_to_name,
				assigned_to_phone,
				note,
				created_by,
				updated_by
			)
			VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), $11, $11)
			RETURNING *
		)
		SELECT
			i.id,
			i.tenant_id,
			i.store_id,
			i.order_id,
			o.order_number,
			o.customer_name,
			o.customer_phone,
			i.courier_type,
			COALESCE(i.courier_name, ''),
			COALESCE(i.tracking_number, ''),
			i.status,
			i.shipping_cost,
			COALESCE(i.assigned_to_name, ''),
			COALESCE(i.assigned_to_phone, ''),
			COALESCE(i.note, ''),
			i.shipped_at,
			i.delivered_at,
			i.created_by,
			i.updated_by,
			i.created_at,
			i.updated_at
		FROM inserted i
		JOIN orders o
		  ON o.tenant_id = i.tenant_id
		 AND o.store_id = i.store_id
		 AND o.id = i.order_id
	`

	return scanShipment(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.OrderID,
		params.CourierType,
		params.CourierName,
		params.TrackingNumber,
		params.ShippingCost,
		params.AssignedToName,
		params.AssignedToPhone,
		params.Note,
		params.CreatedBy,
	))
}

func (r *Repository) UpdateStatus(ctx context.Context, q db.Queryer, params UpdateShipmentStatusParams) (*Shipment, error) {
	const query = `
		WITH updated AS (
			UPDATE shipments
			SET status = $4,
			    shipped_at = CASE WHEN $4 IN ('picked_up', 'on_delivery', 'delivered') THEN COALESCE(shipped_at, now()) ELSE shipped_at END,
			    delivered_at = CASE WHEN $4 = 'delivered' THEN COALESCE(delivered_at, now()) ELSE delivered_at END,
			    updated_by = $5,
			    updated_at = now()
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND id = $3
			RETURNING *
		)
		SELECT
			u.id,
			u.tenant_id,
			u.store_id,
			u.order_id,
			o.order_number,
			o.customer_name,
			o.customer_phone,
			u.courier_type,
			COALESCE(u.courier_name, ''),
			COALESCE(u.tracking_number, ''),
			u.status,
			u.shipping_cost,
			COALESCE(u.assigned_to_name, ''),
			COALESCE(u.assigned_to_phone, ''),
			COALESCE(u.note, ''),
			u.shipped_at,
			u.delivered_at,
			u.created_by,
			u.updated_by,
			u.created_at,
			u.updated_at
		FROM updated u
		JOIN orders o
		  ON o.tenant_id = u.tenant_id
		 AND o.store_id = u.store_id
		 AND o.id = u.order_id
	`

	return scanShipment(q.QueryRow(ctx, query, params.TenantID, params.StoreID, params.ShipmentID, params.Status, params.UpdatedBy))
}

func (r *Repository) CreateStatusLog(ctx context.Context, q db.Queryer, params CreateStatusLogParams) (*StatusLog, error) {
	const query = `
		INSERT INTO shipment_status_logs (
			tenant_id,
			shipment_id,
			from_status,
			to_status,
			note,
			created_by
		)
		VALUES ($1, $2, NULLIF($3, ''), $4, NULLIF($5, ''), $6)
		RETURNING
			id,
			tenant_id,
			shipment_id,
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
		params.ShipmentID,
		params.FromStatus,
		params.ToStatus,
		params.Note,
		params.CreatedBy,
	).Scan(
		&item.ID,
		&item.TenantID,
		&item.ShipmentID,
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

func (r *Repository) ListStatusLogs(ctx context.Context, q db.Queryer, tenantID uuid.UUID, shipmentID uuid.UUID) ([]StatusLog, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			shipment_id,
			COALESCE(from_status, ''),
			to_status,
			COALESCE(note, ''),
			created_by,
			created_at
		FROM shipment_status_logs
		WHERE tenant_id = $1
		  AND shipment_id = $2
		ORDER BY created_at ASC, id ASC
	`

	rows, err := q.Query(ctx, query, tenantID, shipmentID)
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
			&item.ShipmentID,
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

func (r *Repository) LockOrderByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*order.Order, error) {
	const query = selectOrderSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		FOR UPDATE
	`

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderID))
}

func (r *Repository) FindPublicOrderByNumber(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderNumber string) (*order.Order, error) {
	const query = selectOrderSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND order_number = $3
		LIMIT 1
	`

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderNumber))
}

func (r *Repository) ListOrderItems(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID) ([]order.Item, error) {
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

	items := make([]order.Item, 0)
	for rows.Next() {
		var item order.Item
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

func (r *Repository) UpdateOrderShipmentStatus(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, status string) error {
	const query = `
		UPDATE orders
		SET shipment_status = $4,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
	`

	tag, err := q.Exec(ctx, query, tenantID, storeID, orderID, status)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *Repository) UpdateOrderStatus(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID, status string) (*order.Order, error) {
	const query = `
		UPDATE orders
		SET status = $4,
		    completed_at = CASE WHEN $4 = 'completed' THEN COALESCE(completed_at, now()) ELSE completed_at END,
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

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderID, status))
}

func (r *Repository) CreateOrderStatusLog(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID, fromStatus string, toStatus string, note string, createdBy uuid.UUID) error {
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
	`

	_, err := q.Exec(ctx, query, tenantID, orderID, fromStatus, toStatus, note, createdBy)
	return err
}

const selectShipmentSQL = `
	SELECT
		s.id,
		s.tenant_id,
		s.store_id,
		s.order_id,
		o.order_number,
		o.customer_name,
		o.customer_phone,
		s.courier_type,
		COALESCE(s.courier_name, ''),
		COALESCE(s.tracking_number, ''),
		s.status,
		s.shipping_cost,
		COALESCE(s.assigned_to_name, ''),
		COALESCE(s.assigned_to_phone, ''),
		COALESCE(s.note, ''),
		s.shipped_at,
		s.delivered_at,
		s.created_by,
		s.updated_by,
		s.created_at,
		s.updated_at
	FROM shipments s
	JOIN orders o
	  ON o.tenant_id = s.tenant_id
	 AND o.store_id = s.store_id
	 AND o.id = s.order_id
`

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

func scanShipment(row pgx.Row) (*Shipment, error) {
	var item Shipment
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.StoreID,
		&item.OrderID,
		&item.OrderNumber,
		&item.CustomerName,
		&item.CustomerPhone,
		&item.CourierType,
		&item.CourierName,
		&item.TrackingNumber,
		&item.Status,
		&item.ShippingCost,
		&item.AssignedToName,
		&item.AssignedToPhone,
		&item.Note,
		&item.ShippedAt,
		&item.DeliveredAt,
		&item.CreatedBy,
		&item.UpdatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShipmentNotFound
		}
		return nil, err
	}
	return &item, nil
}

func scanOrder(row pgx.Row) (*order.Order, error) {
	var item order.Order
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
		return nil, err
	}
	return &item, nil
}
