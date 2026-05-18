package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

func (r *Repository) ListPublic(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters PublicListFilters,
) ([]PublicListItem, error) {
	const query = `
		SELECT
			p.id,
			p.tenant_id,
			p.store_id,
			p.name,
			p.slug,
			p.price,
			p.compare_at_price,
			p.status,
			COALESCE(img.url, ''),
			COALESCE(s.quantity_on_hand, 0),
			COALESCE(s.quantity_reserved, 0),
			COALESCE(s.quantity_available, 0),
			COALESCE(s.low_stock_threshold, 5),
			p.created_at
		FROM products p
		LEFT JOIN product_stock_snapshots s
		  ON s.tenant_id = p.tenant_id
		 AND s.store_id = p.store_id
		 AND s.product_id = p.id
		LEFT JOIN LATERAL (
			SELECT pi.url
			FROM product_images pi
			WHERE pi.tenant_id = p.tenant_id
			  AND pi.product_id = p.id
			ORDER BY pi.is_primary DESC, pi.sort_order ASC, pi.created_at ASC
			LIMIT 1
		) img ON true
		WHERE p.tenant_id = $1
		  AND p.store_id = $2
		  AND p.status = 'active'
		  AND p.deleted_at IS NULL
		  AND (
			$3 = ''
			OR p.name ILIKE '%' || $3 || '%'
			OR COALESCE(p.description, '') ILIKE '%' || $3 || '%'
		  )
		  AND (
			$4 = ''
			OR EXISTS (
				SELECT 1
				FROM categories c
				WHERE c.tenant_id = p.tenant_id
				  AND c.store_id = p.store_id
				  AND c.id = p.category_id
				  AND c.slug = $4
				  AND c.is_active = true
				  AND c.deleted_at IS NULL
			)
		  )
		  AND (
			$5::boolean IS NULL
			OR ($5 = true AND COALESCE(s.quantity_available, 0) > 0)
			OR ($5 = false AND COALESCE(s.quantity_available, 0) <= 0)
		  )
		  AND (
			$6::timestamptz IS NULL
			OR (p.created_at, p.id) < ($6, $7::uuid)
		  )
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $8
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
		filters.Query,
		filters.CategorySlug,
		filters.InStock,
		cursorCreatedAt,
		cursorID,
		filters.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PublicListItem, 0)
	for rows.Next() {
		var item PublicListItem
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.StoreID,
			&item.Name,
			&item.Slug,
			&item.Price,
			&item.CompareAtPrice,
			&item.Status,
			&item.PrimaryImageURL,
			&item.Stock.QuantityOnHand,
			&item.Stock.QuantityReserved,
			&item.Stock.QuantityAvailable,
			&item.Stock.LowStockThreshold,
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

func (r *Repository) FindPublicBySlug(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	slug string,
) (*PublicProduct, error) {
	const query = `
		SELECT
			p.id,
			p.tenant_id,
			p.store_id,
			p.name,
			p.slug,
			COALESCE(p.description, ''),
			p.price,
			p.compare_at_price,
			p.weight_gram,
			p.status,
			COALESCE(s.quantity_on_hand, 0),
			COALESCE(s.quantity_reserved, 0),
			COALESCE(s.quantity_available, 0),
			COALESCE(s.low_stock_threshold, 5),
			c.id,
			c.name,
			c.slug
		FROM products p
		LEFT JOIN product_stock_snapshots s
		  ON s.tenant_id = p.tenant_id
		 AND s.store_id = p.store_id
		 AND s.product_id = p.id
		LEFT JOIN categories c
		  ON c.tenant_id = p.tenant_id
		 AND c.store_id = p.store_id
		 AND c.id = p.category_id
		 AND c.is_active = true
		 AND c.deleted_at IS NULL
		WHERE p.tenant_id = $1
		  AND p.store_id = $2
		  AND p.slug = $3
		  AND p.status = 'active'
		  AND p.deleted_at IS NULL
		LIMIT 1
	`

	var item PublicProduct
	var categoryID *uuid.UUID
	var categoryName *string
	var categorySlug *string
	if err := q.QueryRow(ctx, query, tenantID, storeID, slug).Scan(
		&item.ID,
		&item.TenantID,
		&item.StoreID,
		&item.Name,
		&item.Slug,
		&item.Description,
		&item.Price,
		&item.CompareAtPrice,
		&item.WeightGram,
		&item.Status,
		&item.Stock.QuantityOnHand,
		&item.Stock.QuantityReserved,
		&item.Stock.QuantityAvailable,
		&item.Stock.LowStockThreshold,
		&categoryID,
		&categoryName,
		&categorySlug,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if categoryID != nil && categoryName != nil && categorySlug != nil {
		item.Category = &PublicCategorySummary{
			ID:   *categoryID,
			Name: *categoryName,
			Slug: *categorySlug,
		}
	}

	return &item, nil
}
