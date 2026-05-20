package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) ListStocks(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters ListStockFilters,
) ([]StockListItem, error) {
	args := []any{tenantID, storeID}
	conditions := []string{
		"s.tenant_id = $1",
		"s.store_id = $2",
		"p.tenant_id = $1",
		"p.store_id = $2",
		"p.deleted_at IS NULL",
	}

	if filters.Query != "" {
		args = append(args, filters.Query)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf(`(
			p.name ILIKE '%%' || $%d || '%%'
			OR COALESCE(p.sku, '') ILIKE '%%' || $%d || '%%'
		)`, placeholder, placeholder))
	}
	if filters.LowStock != nil && *filters.LowStock {
		conditions = append(conditions, "s.quantity_available <= s.low_stock_threshold")
	}
	if filters.OutOfStock != nil && *filters.OutOfStock {
		conditions = append(conditions, "s.quantity_available = 0")
	}
	if filters.CategoryID != nil {
		args = append(args, *filters.CategoryID)
		conditions = append(conditions, fmt.Sprintf("p.category_id = $%d", len(args)))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(s.updated_at, p.id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			p.id,
			p.tenant_id,
			p.store_id,
			p.name,
			COALESCE(p.sku, ''),
			p.category_id,
			COALESCE(c.name, ''),
			COALESCE(img.url, ''),
			s.quantity_on_hand,
			s.quantity_reserved,
			s.quantity_available,
			s.low_stock_threshold,
			s.updated_at
		FROM product_stock_snapshots s
		JOIN products p
		  ON p.id = s.product_id
		 AND p.tenant_id = s.tenant_id
		 AND p.store_id = s.store_id
		LEFT JOIN categories c
		  ON c.id = p.category_id
		 AND c.tenant_id = p.tenant_id
		 AND c.store_id = p.store_id
		LEFT JOIN LATERAL (
			SELECT pi.url
			FROM product_images pi
			WHERE pi.tenant_id = p.tenant_id
			  AND pi.product_id = p.id
			ORDER BY pi.is_primary DESC, pi.sort_order ASC, pi.created_at ASC
			LIMIT 1
		) img ON true
		WHERE %s
		ORDER BY s.updated_at DESC, p.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StockListItem, 0)
	for rows.Next() {
		var item StockListItem
		if err := rows.Scan(
			&item.ProductID,
			&item.TenantID,
			&item.StoreID,
			&item.ProductName,
			&item.SKU,
			&item.CategoryID,
			&item.CategoryName,
			&item.PrimaryImageURL,
			&item.QuantityOnHand,
			&item.QuantityReserved,
			&item.QuantityAvailable,
			&item.LowStockThreshold,
			&item.UpdatedAt,
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

func (r *Repository) FindProduct(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
) (*ProductRef, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			category_id,
			name,
			COALESCE(sku, ''),
			allow_backorder
		FROM products
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
		LIMIT 1
	`

	return scanProductRef(q.QueryRow(ctx, query, tenantID, storeID, productID))
}

func (r *Repository) LockProduct(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
) (*ProductRef, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			category_id,
			name,
			COALESCE(sku, ''),
			allow_backorder
		FROM products
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
		FOR UPDATE
	`

	return scanProductRef(q.QueryRow(ctx, query, tenantID, storeID, productID))
}

func (r *Repository) ListMovements(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
	filters ListMovementFilters,
) ([]StockMovement, error) {
	args := []any{tenantID, storeID, productID}
	conditions := []string{
		"sm.tenant_id = $1",
		"sm.store_id = $2",
		"sm.product_id = $3",
		"p.tenant_id = $1",
		"p.store_id = $2",
		"p.deleted_at IS NULL",
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(sm.created_at, sm.id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			sm.id,
			sm.tenant_id,
			sm.store_id,
			sm.product_id,
			sm.movement_type,
			sm.quantity,
			COALESCE(sm.balance_after, 0),
			COALESCE(sm.reference_type, ''),
			sm.reference_id,
			COALESCE(sm.note, ''),
			sm.created_by,
			COALESCE(u.name, ''),
			sm.created_at
		FROM stock_movements sm
		JOIN products p
		  ON p.id = sm.product_id
		 AND p.tenant_id = sm.tenant_id
		 AND p.store_id = sm.store_id
		LEFT JOIN users u
		  ON u.id = sm.created_by
		WHERE %s
		ORDER BY sm.created_at DESC, sm.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StockMovement, 0)
	for rows.Next() {
		movement, err := scanMovement(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *movement)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *Repository) LockStockSnapshot(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
) (*StockSnapshot, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold,
			updated_at
		FROM product_stock_snapshots
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND product_id = $3
		FOR UPDATE
	`

	return scanSnapshot(q.QueryRow(ctx, query, tenantID, storeID, productID))
}

func (r *Repository) UpdateSnapshot(
	ctx context.Context,
	q db.Queryer,
	params UpdateSnapshotParams,
) (*StockSnapshot, error) {
	const query = `
		UPDATE product_stock_snapshots
		SET quantity_on_hand = $4,
		    quantity_reserved = $5,
		    quantity_available = $6,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND product_id = $3
		RETURNING
			id,
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold,
			updated_at
	`

	return scanSnapshot(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.QuantityOnHand,
		params.QuantityReserved,
		params.QuantityAvailable,
	))
}

func (r *Repository) UpdateThreshold(
	ctx context.Context,
	q db.Queryer,
	params UpdateThresholdParams,
) (*StockSnapshot, error) {
	const query = `
		UPDATE product_stock_snapshots
		SET low_stock_threshold = $4,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND product_id = $3
		RETURNING
			id,
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold,
			updated_at
	`

	return scanSnapshot(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.LowStockThreshold,
	))
}

func (r *Repository) CreateSnapshot(
	ctx context.Context,
	q db.Queryer,
	params CreateSnapshotParams,
) (*StockSnapshot, error) {
	const query = `
		INSERT INTO product_stock_snapshots (
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold,
			updated_at
	`

	return scanSnapshot(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.QuantityOnHand,
		params.QuantityReserved,
		params.QuantityAvailable,
		params.LowStockThreshold,
	))
}

func (r *Repository) CreateMovement(
	ctx context.Context,
	q db.Queryer,
	params CreateMovementParams,
) (*StockMovement, error) {
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
		RETURNING
			id,
			tenant_id,
			store_id,
			product_id,
			movement_type,
			quantity,
			COALESCE(balance_after, 0),
			COALESCE(reference_type, ''),
			reference_id,
			COALESCE(note, ''),
			created_by,
			'',
			created_at
	`

	movement, err := scanMovement(q.QueryRow(
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
		formatAdjustmentNote(params.Reason, params.Note),
		params.CreatedBy,
	))
	if err != nil {
		return nil, err
	}
	if movement.Reason == "" {
		movement.Reason = strings.TrimSpace(params.Reason)
	}
	return movement, nil
}

type UpdateSnapshotParams struct {
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductID         uuid.UUID
	QuantityOnHand    int
	QuantityReserved  int
	QuantityAvailable int
}

type UpdateThresholdParams struct {
	TenantID          uuid.UUID
	StoreID           uuid.UUID
	ProductID         uuid.UUID
	LowStockThreshold int
}

func scanProductRef(row pgx.Row) (*ProductRef, error) {
	var product ProductRef
	if err := row.Scan(
		&product.ID,
		&product.TenantID,
		&product.StoreID,
		&product.CategoryID,
		&product.Name,
		&product.SKU,
		&product.AllowBackorder,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

func scanSnapshot(row pgx.Row) (*StockSnapshot, error) {
	var snapshot StockSnapshot
	if err := row.Scan(
		&snapshot.ID,
		&snapshot.TenantID,
		&snapshot.StoreID,
		&snapshot.ProductID,
		&snapshot.QuantityOnHand,
		&snapshot.QuantityReserved,
		&snapshot.QuantityAvailable,
		&snapshot.LowStockThreshold,
		&snapshot.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStockSnapshotNotFound
		}
		return nil, err
	}
	return &snapshot, nil
}

func scanMovement(row pgx.Row) (*StockMovement, error) {
	var movement StockMovement
	if err := row.Scan(
		&movement.ID,
		&movement.TenantID,
		&movement.StoreID,
		&movement.ProductID,
		&movement.MovementType,
		&movement.Quantity,
		&movement.BalanceAfter,
		&movement.ReferenceType,
		&movement.ReferenceID,
		&movement.Note,
		&movement.CreatedBy,
		&movement.CreatedByName,
		&movement.CreatedAt,
	); err != nil {
		return nil, err
	}
	movement.Reason, movement.Note = splitAdjustmentNote(movement.Note)
	return &movement, nil
}
