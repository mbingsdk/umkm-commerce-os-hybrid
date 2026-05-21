package discovery

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) ListFeaturedStores(ctx context.Context, q db.Queryer, limit int) ([]Store, error) {
	const query = `
		SELECT
			s.id,
			COALESCE(NULLIF(d.title, ''), s.name),
			s.slug,
			COALESCE(NULLIF(d.description, ''), s.description, ''),
			COALESCE(NULLIF(d.image_url, ''), s.logo_url, ''),
			COALESCE(s.banner_url, ''),
			COALESCE(s.city, ''),
			COALESCE(s.province, ''),
			s.created_at
		FROM discovery_featured_items d
		JOIN stores s ON s.id = d.item_id
		JOIN tenants t ON t.id = s.tenant_id
		WHERE d.item_type = 'store'
		  AND d.is_active = true
		  AND (d.starts_at IS NULL OR d.starts_at <= now())
		  AND (d.ends_at IS NULL OR d.ends_at > now())
		  AND t.status IN ('active', 'trialing')
		  AND t.deleted_at IS NULL
		  AND s.status = 'published'
		  AND s.is_discoverable = true
		  AND s.deleted_at IS NULL
		ORDER BY d.sort_order ASC, d.created_at DESC, d.id DESC
		LIMIT $1
	`

	rows, err := q.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStores(rows)
}

func (r *Repository) ListFeaturedProducts(ctx context.Context, q db.Queryer, limit int) ([]Product, error) {
	const query = `
		SELECT
			p.id,
			COALESCE(NULLIF(d.title, ''), p.name),
			p.slug,
			COALESCE(NULLIF(d.description, ''), p.description, ''),
			p.price,
			COALESCE(NULLIF(d.image_url, ''), pi.url, ''),
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			s.id,
			s.name,
			s.slug,
			COALESCE(s.city, ''),
			COALESCE(s.province, ''),
			p.created_at
		FROM discovery_featured_items d
		JOIN products p ON p.id = d.item_id
		JOIN stores s ON s.id = p.store_id AND s.tenant_id = p.tenant_id
		JOIN tenants t ON t.id = p.tenant_id
		LEFT JOIN categories c
		  ON c.id = p.category_id
		 AND c.tenant_id = p.tenant_id
		 AND c.store_id = p.store_id
		 AND c.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT url
			FROM product_images
			WHERE tenant_id = p.tenant_id
			  AND product_id = p.id
			ORDER BY is_primary DESC, sort_order ASC, created_at ASC
			LIMIT 1
		) pi ON true
		WHERE d.item_type = 'product'
		  AND d.is_active = true
		  AND (d.starts_at IS NULL OR d.starts_at <= now())
		  AND (d.ends_at IS NULL OR d.ends_at > now())
		  AND t.status IN ('active', 'trialing')
		  AND t.deleted_at IS NULL
		  AND s.status = 'published'
		  AND s.is_discoverable = true
		  AND s.deleted_at IS NULL
		  AND p.status = 'active'
		  AND p.is_discoverable = true
		  AND p.deleted_at IS NULL
		ORDER BY d.sort_order ASC, d.created_at DESC, d.id DESC
		LIMIT $1
	`

	rows, err := q.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *Repository) ListStores(ctx context.Context, q db.Queryer, filters ListStoresFilters) ([]Store, error) {
	args := make([]any, 0)
	conditions := publicStoreConditions()

	if filters.Query != "" {
		args = append(args, filters.Query)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf(`(
			s.name ILIKE '%%' || $%d || '%%'
			OR COALESCE(s.description, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(s.city, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(s.province, '') ILIKE '%%' || $%d || '%%'
		)`, placeholder, placeholder, placeholder, placeholder))
	}
	if filters.City != "" {
		args = append(args, filters.City)
		conditions = append(conditions, fmt.Sprintf("LOWER(COALESCE(s.city, '')) = LOWER($%d)", len(args)))
	}
	if filters.Category != "" {
		args = append(args, filters.Category)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM products p
			JOIN categories c
			  ON c.id = p.category_id
			 AND c.tenant_id = p.tenant_id
			 AND c.store_id = p.store_id
			 AND c.deleted_at IS NULL
			WHERE p.tenant_id = s.tenant_id
			  AND p.store_id = s.id
			  AND p.status = 'active'
			  AND p.is_discoverable = true
			  AND p.deleted_at IS NULL
			  AND (c.slug = $%d OR c.name ILIKE '%%' || $%d || '%%')
		)`, placeholder, placeholder))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(s.created_at, s.id) < ($%d::timestamptz, $%d::uuid)", len(args)-1, len(args)))
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			s.id,
			s.name,
			s.slug,
			COALESCE(s.description, ''),
			COALESCE(s.logo_url, ''),
			COALESCE(s.banner_url, ''),
			COALESCE(s.city, ''),
			COALESCE(s.province, ''),
			s.created_at
		FROM stores s
		JOIN tenants t ON t.id = s.tenant_id
		WHERE %s
		ORDER BY s.created_at DESC, s.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStores(rows)
}

func (r *Repository) ListProducts(ctx context.Context, q db.Queryer, filters ListProductsFilters) ([]Product, error) {
	args := make([]any, 0)
	conditions := publicProductConditions()

	if filters.Query != "" {
		args = append(args, filters.Query)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf(`(
			p.name ILIKE '%%' || $%d || '%%'
			OR COALESCE(p.description, '') ILIKE '%%' || $%d || '%%'
			OR s.name ILIKE '%%' || $%d || '%%'
			OR COALESCE(c.name, '') ILIKE '%%' || $%d || '%%'
		)`, placeholder, placeholder, placeholder, placeholder))
	}
	if filters.City != "" {
		args = append(args, filters.City)
		conditions = append(conditions, fmt.Sprintf("LOWER(COALESCE(s.city, '')) = LOWER($%d)", len(args)))
	}
	if filters.Category != "" {
		args = append(args, filters.Category)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf("(c.slug = $%d OR c.name ILIKE '%%' || $%d || '%%')", placeholder, placeholder))
	}
	if filters.PriceMin != nil {
		args = append(args, *filters.PriceMin)
		conditions = append(conditions, fmt.Sprintf("p.price >= $%d", len(args)))
	}
	if filters.PriceMax != nil {
		args = append(args, *filters.PriceMax)
		conditions = append(conditions, fmt.Sprintf("p.price <= $%d", len(args)))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(p.created_at, p.id) < ($%d::timestamptz, $%d::uuid)", len(args)-1, len(args)))
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			p.id,
			p.name,
			p.slug,
			COALESCE(p.description, ''),
			p.price,
			COALESCE(pi.url, ''),
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			s.id,
			s.name,
			s.slug,
			COALESCE(s.city, ''),
			COALESCE(s.province, ''),
			p.created_at
		FROM products p
		JOIN stores s ON s.id = p.store_id AND s.tenant_id = p.tenant_id
		JOIN tenants t ON t.id = p.tenant_id
		LEFT JOIN categories c
		  ON c.id = p.category_id
		 AND c.tenant_id = p.tenant_id
		 AND c.store_id = p.store_id
		 AND c.deleted_at IS NULL
		LEFT JOIN LATERAL (
			SELECT url
			FROM product_images
			WHERE tenant_id = p.tenant_id
			  AND product_id = p.id
			ORDER BY is_primary DESC, sort_order ASC, created_at ASC
			LIMIT 1
		) pi ON true
		WHERE %s
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProducts(rows)
}

func (r *Repository) PopularCategories(ctx context.Context, q db.Queryer, limit int) ([]CategoryAggregate, error) {
	const query = `
		SELECT c.name, c.slug, COUNT(p.id)::int
		FROM categories c
		JOIN products p
		  ON p.category_id = c.id
		 AND p.tenant_id = c.tenant_id
		 AND p.store_id = c.store_id
		JOIN stores s ON s.id = p.store_id AND s.tenant_id = p.tenant_id
		JOIN tenants t ON t.id = p.tenant_id
		WHERE c.deleted_at IS NULL
		  AND c.is_active = true
		  AND t.status IN ('active', 'trialing')
		  AND t.deleted_at IS NULL
		  AND s.status = 'published'
		  AND s.is_discoverable = true
		  AND s.deleted_at IS NULL
		  AND p.status = 'active'
		  AND p.is_discoverable = true
		  AND p.deleted_at IS NULL
		GROUP BY c.id, c.name, c.slug
		ORDER BY COUNT(p.id) DESC, c.name ASC
		LIMIT $1
	`

	rows, err := q.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CategoryAggregate, 0)
	for rows.Next() {
		var item CategoryAggregate
		if err := rows.Scan(&item.Name, &item.Slug, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) PopularCities(ctx context.Context, q db.Queryer, limit int) ([]CityAggregate, error) {
	const query = `
		SELECT s.city, COUNT(s.id)::int
		FROM stores s
		JOIN tenants t ON t.id = s.tenant_id
		WHERE t.status IN ('active', 'trialing')
		  AND t.deleted_at IS NULL
		  AND s.status = 'published'
		  AND s.is_discoverable = true
		  AND s.deleted_at IS NULL
		  AND COALESCE(s.city, '') <> ''
		GROUP BY s.city
		ORDER BY COUNT(s.id) DESC, s.city ASC
		LIMIT $1
	`

	rows, err := q.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CityAggregate, 0)
	for rows.Next() {
		var item CityAggregate
		if err := rows.Scan(&item.City, &item.Count); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func publicStoreConditions() []string {
	return []string{
		"t.status IN ('active', 'trialing')",
		"t.deleted_at IS NULL",
		"s.status = 'published'",
		"s.is_discoverable = true",
		"s.deleted_at IS NULL",
	}
}

func publicProductConditions() []string {
	conditions := publicStoreConditions()
	conditions = append(conditions,
		"p.status = 'active'",
		"p.is_discoverable = true",
		"p.deleted_at IS NULL",
	)
	return conditions
}

func scanStores(rows pgx.Rows) ([]Store, error) {
	items := make([]Store, 0)
	for rows.Next() {
		var item Store
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Slug,
			&item.Description,
			&item.LogoURL,
			&item.BannerURL,
			&item.City,
			&item.Province,
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

func scanProducts(rows pgx.Rows) ([]Product, error) {
	items := make([]Product, 0)
	for rows.Next() {
		var item Product
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Slug,
			&item.Description,
			&item.Price,
			&item.PrimaryImageURL,
			&item.CategoryName,
			&item.CategorySlug,
			&item.StoreID,
			&item.StoreName,
			&item.StoreSlug,
			&item.StoreCity,
			&item.StoreProvince,
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
