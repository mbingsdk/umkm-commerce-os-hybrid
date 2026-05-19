package checkout

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var ErrStockSnapshotNotFound = errors.New("stock snapshot not found")

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) ListProductsForCheckout(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productIDs []uuid.UUID,
) ([]ProductForCheckout, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			name,
			COALESCE(sku, ''),
			price,
			status,
			track_inventory,
			allow_backorder
		FROM products
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = ANY($3::uuid[])
		  AND deleted_at IS NULL
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]ProductForCheckout, 0, len(productIDs))
	for rows.Next() {
		var item ProductForCheckout
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.StoreID,
			&item.Name,
			&item.SKU,
			&item.Price,
			&item.Status,
			&item.TrackInventory,
			&item.AllowBackorder,
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

func (r *Repository) FindOrCreateCustomer(
	ctx context.Context,
	q db.Queryer,
	params FindOrCreateCustomerParams,
) (*CustomerRecord, error) {
	const query = `
		INSERT INTO customers (
			tenant_id,
			store_id,
			name,
			phone,
			email
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))
		ON CONFLICT (tenant_id, phone) DO UPDATE
		SET store_id = EXCLUDED.store_id,
		    name = EXCLUDED.name,
		    email = EXCLUDED.email,
		    deleted_at = NULL,
		    updated_at = now()
		RETURNING
			id,
			tenant_id,
			store_id,
			name,
			phone,
			COALESCE(email, '')
	`

	var customer CustomerRecord
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.Name,
		params.Phone,
		params.Email,
	).Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.StoreID,
		&customer.Name,
		&customer.Phone,
		&customer.Email,
	); err != nil {
		return nil, err
	}

	return &customer, nil
}

func (r *Repository) CreateCustomerAddress(
	ctx context.Context,
	q db.Queryer,
	params CreateAddressParams,
) (*AddressRecord, error) {
	const query = `
		INSERT INTO customer_addresses (
			tenant_id,
			customer_id,
			label,
			recipient_name,
			recipient_phone,
			address,
			city,
			province,
			postal_code,
			is_default
		)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''), true)
		RETURNING id, tenant_id, customer_id
	`

	var address AddressRecord
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.CustomerID,
		params.Label,
		params.RecipientName,
		params.RecipientPhone,
		params.Address,
		params.City,
		params.Province,
		params.PostalCode,
	).Scan(
		&address.ID,
		&address.TenantID,
		&address.CustomerID,
	); err != nil {
		return nil, err
	}

	return &address, nil
}

func (r *Repository) CreateOrder(ctx context.Context, q db.Queryer, params CreateOrderParams) (*OrderRecord, error) {
	const query = `
		INSERT INTO orders (
			tenant_id,
			store_id,
			customer_id,
			order_number,
			source,
			status,
			payment_status,
			subtotal,
			discount_total,
			shipping_cost,
			tax_total,
			grand_total,
			customer_name,
			customer_phone,
			customer_email,
			shipping_address,
			shipping_city,
			shipping_province,
			shipping_postal_code,
			customer_note
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			NULLIF($15, ''),
			NULLIF($16, ''),
			NULLIF($17, ''),
			NULLIF($18, ''),
			NULLIF($19, ''),
			NULLIF($20, '')
		)
		RETURNING
			id,
			tenant_id,
			store_id,
			order_number,
			status,
			payment_status,
			subtotal,
			discount_total,
			shipping_cost,
			tax_total,
			grand_total,
			created_at
	`

	var order OrderRecord
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.CustomerID,
		params.OrderNumber,
		params.Source,
		params.Status,
		params.PaymentStatus,
		params.Subtotal,
		params.DiscountTotal,
		params.ShippingCost,
		params.TaxTotal,
		params.GrandTotal,
		params.CustomerName,
		params.CustomerPhone,
		params.CustomerEmail,
		params.ShippingAddress,
		params.ShippingCity,
		params.ShippingProvince,
		params.ShippingPostalCode,
		params.CustomerNote,
	).Scan(
		&order.ID,
		&order.TenantID,
		&order.StoreID,
		&order.OrderNumber,
		&order.Status,
		&order.PaymentStatus,
		&order.Subtotal,
		&order.DiscountTotal,
		&order.ShippingCost,
		&order.TaxTotal,
		&order.GrandTotal,
		&order.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &order, nil
}

func (r *Repository) CreateOrderItem(ctx context.Context, q db.Queryer, params CreateOrderItemParams) error {
	const query = `
		INSERT INTO order_items (
			tenant_id,
			order_id,
			product_id,
			product_name,
			sku,
			quantity,
			unit_price,
			discount_total,
			subtotal
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9)
	`

	_, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.OrderID,
		params.ProductID,
		params.ProductName,
		params.SKU,
		params.Quantity,
		params.UnitPrice,
		params.DiscountTotal,
		params.Subtotal,
	)
	return err
}

func (r *Repository) CreateStockReservation(
	ctx context.Context,
	q db.Queryer,
	params CreateReservationParams,
) error {
	const query = `
		INSERT INTO stock_reservations (
			tenant_id,
			store_id,
			product_id,
			order_id,
			quantity,
			status,
			expires_at
		)
		VALUES ($1, $2, $3, $4, $5, 'active', $6)
	`

	_, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.OrderID,
		params.Quantity,
		params.ExpiresAt,
	)
	return err
}

func (r *Repository) UpdateStockSnapshot(
	ctx context.Context,
	q db.Queryer,
	params UpdateSnapshotParams,
) error {
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

func (r *Repository) CreateStockMovement(
	ctx context.Context,
	q db.Queryer,
	params CreateStockMovementParams,
) error {
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
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, NULLIF($9, ''), NULL)
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
	)
	return err
}

func (r *Repository) CreateOrderStatusLog(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	orderID uuid.UUID,
	toStatus string,
	note string,
) error {
	const query = `
		INSERT INTO order_status_logs (
			tenant_id,
			order_id,
			from_status,
			to_status,
			note,
			created_by
		)
		VALUES ($1, $2, NULL, $3, NULLIF($4, ''), NULL)
	`

	_, err := q.Exec(ctx, query, tenantID, orderID, toStatus, note)
	return err
}

func (r *Repository) UpdateCustomerStats(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	customerID uuid.UUID,
	orderTotal int64,
) error {
	const query = `
		UPDATE customers
		SET total_orders = total_orders + 1,
		    total_spent = total_spent + $4,
		    last_order_at = now(),
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
	`

	_, err := q.Exec(ctx, query, tenantID, storeID, customerID, orderTotal)
	return err
}
