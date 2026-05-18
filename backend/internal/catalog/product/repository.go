package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	ErrProductNotFound         = errors.New("product not found")
	ErrProductSlugAlreadyInUse = errors.New("product slug already in use")
)

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
) ([]Product, error) {
	const query = `
		SELECT
			p.id,
			p.tenant_id,
			p.store_id,
			p.category_id,
			p.name,
			p.slug,
			COALESCE(p.description, ''),
			COALESCE(p.sku, ''),
			COALESCE(p.barcode, ''),
			p.price,
			p.compare_at_price,
			p.cost_price,
			p.weight_gram,
			p.length_cm,
			p.width_cm,
			p.height_cm,
			p.status,
			p.is_discoverable,
			p.track_inventory,
			p.allow_backorder,
			COALESCE(img.url, ''),
			COALESCE(s.quantity_on_hand, 0),
			COALESCE(s.quantity_reserved, 0),
			COALESCE(s.quantity_available, 0),
			COALESCE(s.low_stock_threshold, 5)
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
		  AND p.deleted_at IS NULL
		  AND ($3::text IS NULL OR p.status = $3)
		  AND ($4::uuid IS NULL OR p.category_id = $4)
		  AND (
			$5 = ''
			OR p.name ILIKE '%' || $5 || '%'
			OR COALESCE(p.description, '') ILIKE '%' || $5 || '%'
		  )
		ORDER BY p.created_at DESC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, filters.Status, filters.CategoryID, filters.Query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Product, 0)
	for rows.Next() {
		item, err := scanProduct(rows)
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

func (r *Repository) Create(ctx context.Context, q db.Queryer, params CreateParams) (*Product, error) {
	const query = `
		INSERT INTO products (
			tenant_id,
			store_id,
			category_id,
			name,
			slug,
			description,
			sku,
			barcode,
			price,
			compare_at_price,
			cost_price,
			weight_gram,
			length_cm,
			width_cm,
			height_cm,
			status,
			is_discoverable,
			track_inventory,
			allow_backorder
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			NULLIF($6, ''),
			NULLIF($7, ''),
			NULLIF($8, ''),
			$9,
			$10,
			$11,
			$12,
			$13,
			$14,
			$15,
			$16,
			$17,
			$18,
			$19
		)
		RETURNING
			id,
			tenant_id,
			store_id,
			category_id,
			name,
			slug,
			COALESCE(description, ''),
			COALESCE(sku, ''),
			COALESCE(barcode, ''),
			price,
			compare_at_price,
			cost_price,
			weight_gram,
			length_cm,
			width_cm,
			height_cm,
			status,
			is_discoverable,
			track_inventory,
			allow_backorder,
			'',
			0,
			0,
			0,
			5
	`

	product, err := scanProduct(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.CategoryID,
		params.Name,
		params.Slug,
		params.Description,
		params.SKU,
		params.Barcode,
		params.Price,
		params.CompareAtPrice,
		params.CostPrice,
		params.WeightGram,
		params.LengthCM,
		params.WidthCM,
		params.HeightCM,
		params.Status,
		params.IsDiscoverable,
		params.TrackInventory,
		params.AllowBackorder,
	))
	if err != nil && isUniqueViolation(err) {
		return nil, ErrProductSlugAlreadyInUse
	}
	return product, err
}

func (r *Repository) FindByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
) (*Product, error) {
	const query = `
		SELECT
			p.id,
			p.tenant_id,
			p.store_id,
			p.category_id,
			p.name,
			p.slug,
			COALESCE(p.description, ''),
			COALESCE(p.sku, ''),
			COALESCE(p.barcode, ''),
			p.price,
			p.compare_at_price,
			p.cost_price,
			p.weight_gram,
			p.length_cm,
			p.width_cm,
			p.height_cm,
			p.status,
			p.is_discoverable,
			p.track_inventory,
			p.allow_backorder,
			COALESCE(img.url, ''),
			COALESCE(s.quantity_on_hand, 0),
			COALESCE(s.quantity_reserved, 0),
			COALESCE(s.quantity_available, 0),
			COALESCE(s.low_stock_threshold, 5)
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
		  AND p.id = $3
		  AND p.deleted_at IS NULL
		LIMIT 1
	`

	return scanProduct(q.QueryRow(ctx, query, tenantID, storeID, productID))
}

func (r *Repository) ListImages(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
) ([]Image, error) {
	const query = `
		SELECT
			pi.id,
			pi.url,
			COALESCE(pi.alt_text, ''),
			pi.is_primary,
			pi.sort_order
		FROM product_images pi
		JOIN products p
		  ON p.id = pi.product_id
		 AND p.tenant_id = pi.tenant_id
		WHERE pi.tenant_id = $1
		  AND p.tenant_id = $1
		  AND p.store_id = $2
		  AND pi.product_id = $3
		  AND p.deleted_at IS NULL
		ORDER BY pi.is_primary DESC, pi.sort_order ASC, pi.created_at ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Image, 0)
	for rows.Next() {
		var item Image
		if err := rows.Scan(
			&item.ID,
			&item.URL,
			&item.AltText,
			&item.IsPrimary,
			&item.SortOrder,
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

func (r *Repository) Update(ctx context.Context, q db.Queryer, params UpdateParams) error {
	const query = `
		UPDATE products
		SET category_id = $4,
		    name = $5,
		    slug = $6,
		    description = NULLIF($7, ''),
		    sku = NULLIF($8, ''),
		    barcode = NULLIF($9, ''),
		    price = $10,
		    compare_at_price = $11,
		    cost_price = $12,
		    weight_gram = $13,
		    length_cm = $14,
		    width_cm = $15,
		    height_cm = $16,
		    status = $17,
		    is_discoverable = $18,
		    track_inventory = $19,
		    allow_backorder = $20,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
	`

	tag, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.CategoryID,
		params.Name,
		params.Slug,
		params.Description,
		params.SKU,
		params.Barcode,
		params.Price,
		params.CompareAtPrice,
		params.CostPrice,
		params.WeightGram,
		params.LengthCM,
		params.WidthCM,
		params.HeightCM,
		params.Status,
		params.IsDiscoverable,
		params.TrackInventory,
		params.AllowBackorder,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return ErrProductSlugAlreadyInUse
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}

func (r *Repository) SoftDelete(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
) error {
	const query = `
		UPDATE products
		SET deleted_at = now(),
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
	`

	tag, err := q.Exec(ctx, query, tenantID, storeID, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}

func scanProduct(row pgx.Row) (*Product, error) {
	var product Product
	if err := row.Scan(
		&product.ID,
		&product.TenantID,
		&product.StoreID,
		&product.CategoryID,
		&product.Name,
		&product.Slug,
		&product.Description,
		&product.SKU,
		&product.Barcode,
		&product.Price,
		&product.CompareAtPrice,
		&product.CostPrice,
		&product.WeightGram,
		&product.LengthCM,
		&product.WidthCM,
		&product.HeightCM,
		&product.Status,
		&product.IsDiscoverable,
		&product.TrackInventory,
		&product.AllowBackorder,
		&product.PrimaryImageURL,
		&product.Stock.QuantityOnHand,
		&product.Stock.QuantityReserved,
		&product.Stock.QuantityAvailable,
		&product.Stock.LowStockThreshold,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &product, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
