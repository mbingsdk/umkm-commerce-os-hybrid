package admin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	ErrFeaturedItemNotFound    = errors.New("featured item not found")
	ErrFeaturedStoreNotFound   = errors.New("featured store not found")
	ErrFeaturedProductNotFound = errors.New("featured product not found")
)

func (r *Repository) ListFeaturedItems(ctx context.Context, q db.Queryer, filters FeaturedListFilters) ([]FeaturedItem, error) {
	args := make([]any, 0)
	conditions := []string{"d.deleted_at IS NULL"}

	if filters.ItemType != "" {
		args = append(args, filters.ItemType)
		conditions = append(conditions, fmt.Sprintf("d.item_type = $%d", len(args)))
	}
	if filters.Placement != "" {
		args = append(args, filters.Placement)
		conditions = append(conditions, fmt.Sprintf("d.placement = $%d", len(args)))
	}
	if filters.TenantID != nil {
		args = append(args, *filters.TenantID)
		conditions = append(conditions, fmt.Sprintf("d.tenant_id = $%d", len(args)))
	}
	if filters.IsActive != nil {
		args = append(args, *filters.IsActive)
		conditions = append(conditions, fmt.Sprintf("d.is_active = $%d", len(args)))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(d.created_at, d.id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			d.id,
			d.item_type,
			COALESCE(d.tenant_id, '00000000-0000-0000-0000-000000000000'::uuid),
			d.store_id,
			d.product_id,
			COALESCE(d.placement, 'home'),
			d.sort_order,
			d.starts_at,
			d.ends_at,
			d.is_active,
			d.created_by,
			d.created_at,
			d.updated_at,
			COALESCE(s.name, ''),
			COALESCE(s.slug, ''),
			COALESCE(p.name, ''),
			COALESCE(p.slug, '')
		FROM discovery_featured_items d
		LEFT JOIN stores s
		  ON s.id = COALESCE(d.store_id, CASE WHEN d.item_type = 'store' THEN d.item_id END)
		LEFT JOIN products p
		  ON p.id = COALESCE(d.product_id, CASE WHEN d.item_type = 'product' THEN d.item_id END)
		WHERE %s
		ORDER BY d.created_at DESC, d.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]FeaturedItem, 0)
	for rows.Next() {
		item, err := scanFeaturedItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) FindFeaturedItemByIDForUpdate(ctx context.Context, q db.Queryer, featuredID uuid.UUID) (*FeaturedItem, error) {
	const query = `
		SELECT
			d.id,
			d.item_type,
			COALESCE(d.tenant_id, '00000000-0000-0000-0000-000000000000'::uuid),
			d.store_id,
			d.product_id,
			COALESCE(d.placement, 'home'),
			d.sort_order,
			d.starts_at,
			d.ends_at,
			d.is_active,
			d.created_by,
			d.created_at,
			d.updated_at,
			COALESCE(s.name, ''),
			COALESCE(s.slug, ''),
			COALESCE(p.name, ''),
			COALESCE(p.slug, '')
		FROM discovery_featured_items d
		LEFT JOIN stores s
		  ON s.id = COALESCE(d.store_id, CASE WHEN d.item_type = 'store' THEN d.item_id END)
		LEFT JOIN products p
		  ON p.id = COALESCE(d.product_id, CASE WHEN d.item_type = 'product' THEN d.item_id END)
		WHERE d.id = $1
		  AND d.deleted_at IS NULL
		FOR UPDATE OF d
	`
	return scanFeaturedItemPtr(q.QueryRow(ctx, query, featuredID))
}

func (r *Repository) CreateFeaturedItem(ctx context.Context, q db.Queryer, params CreateFeaturedParams) (*FeaturedItem, error) {
	const query = `
		INSERT INTO discovery_featured_items (
			item_type,
			item_id,
			tenant_id,
			store_id,
			product_id,
			placement,
			sort_order,
			starts_at,
			ends_at,
			is_active,
			created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING
			id,
			item_type,
			tenant_id,
			store_id,
			product_id,
			placement,
			sort_order,
			starts_at,
			ends_at,
			is_active,
			created_by,
			created_at,
			updated_at,
			'',
			'',
			'',
			''
	`
	return scanFeaturedItemPtr(q.QueryRow(
		ctx,
		query,
		params.ItemType,
		params.ItemID,
		params.TenantID,
		uuidPtrArg(params.StoreID),
		uuidPtrArg(params.ProductID),
		params.Placement,
		params.SortOrder,
		timePtrArg(params.StartsAt),
		timePtrArg(params.EndsAt),
		params.IsActive,
		params.CreatedBy,
	))
}

func (r *Repository) UpdateFeaturedItem(ctx context.Context, q db.Queryer, params UpdateFeaturedParams) (*FeaturedItem, error) {
	const query = `
		UPDATE discovery_featured_items
		SET item_type = $2,
		    item_id = $3,
		    tenant_id = $4,
		    store_id = $5,
		    product_id = $6,
		    placement = $7,
		    sort_order = $8,
		    starts_at = $9,
		    ends_at = $10,
		    is_active = $11,
		    updated_at = now()
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			item_type,
			tenant_id,
			store_id,
			product_id,
			placement,
			sort_order,
			starts_at,
			ends_at,
			is_active,
			created_by,
			created_at,
			updated_at,
			'',
			'',
			'',
			''
	`
	return scanFeaturedItemPtr(q.QueryRow(
		ctx,
		query,
		params.ID,
		params.ItemType,
		params.ItemID,
		params.TenantID,
		uuidPtrArg(params.StoreID),
		uuidPtrArg(params.ProductID),
		params.Placement,
		params.SortOrder,
		timePtrArg(params.StartsAt),
		timePtrArg(params.EndsAt),
		params.IsActive,
	))
}

func (r *Repository) DeleteFeaturedItem(ctx context.Context, q db.Queryer, featuredID uuid.UUID) (*FeaturedItem, error) {
	const query = `
		UPDATE discovery_featured_items
		SET deleted_at = now(),
		    is_active = false,
		    updated_at = now()
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING
			id,
			item_type,
			COALESCE(tenant_id, '00000000-0000-0000-0000-000000000000'::uuid),
			store_id,
			product_id,
			COALESCE(placement, 'home'),
			sort_order,
			starts_at,
			ends_at,
			is_active,
			created_by,
			created_at,
			updated_at,
			'',
			'',
			'',
			''
	`
	return scanFeaturedItemPtr(q.QueryRow(ctx, query, featuredID))
}

func (r *Repository) FindFeaturedStoreTarget(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) (*FeaturedStoreTarget, error) {
	const query = `
		SELECT
			s.id,
			s.tenant_id,
			s.name,
			s.slug,
			t.status,
			s.status,
			s.is_discoverable
		FROM stores s
		JOIN tenants t
		  ON t.id = s.tenant_id
		 AND t.deleted_at IS NULL
		WHERE s.tenant_id = $1
		  AND s.id = $2
		  AND s.deleted_at IS NULL
		LIMIT 1
	`
	var target FeaturedStoreTarget
	if err := q.QueryRow(ctx, query, tenantID, storeID).Scan(
		&target.ID,
		&target.TenantID,
		&target.Name,
		&target.Slug,
		&target.TenantStatus,
		&target.Status,
		&target.IsDiscoverable,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFeaturedStoreNotFound
		}
		return nil, err
	}
	return &target, nil
}

func (r *Repository) FindFeaturedProductTarget(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID *uuid.UUID, productID uuid.UUID) (*FeaturedProductTarget, error) {
	const query = `
		SELECT
			p.id,
			p.tenant_id,
			p.store_id,
			p.name,
			p.slug,
			t.status,
			p.status,
			p.is_discoverable,
			s.status,
			s.is_discoverable
		FROM products p
		JOIN tenants t
		  ON t.id = p.tenant_id
		 AND t.deleted_at IS NULL
		JOIN stores s
		  ON s.id = p.store_id
		 AND s.tenant_id = p.tenant_id
		 AND s.deleted_at IS NULL
		WHERE p.tenant_id = $1
		  AND ($2::uuid IS NULL OR p.store_id = $2)
		  AND p.id = $3
		  AND p.deleted_at IS NULL
		LIMIT 1
	`
	var target FeaturedProductTarget
	if err := q.QueryRow(ctx, query, tenantID, uuidPtrArg(storeID), productID).Scan(
		&target.ID,
		&target.TenantID,
		&target.StoreID,
		&target.Name,
		&target.Slug,
		&target.TenantStatus,
		&target.Status,
		&target.IsDiscoverable,
		&target.StoreStatus,
		&target.StoreDiscoverable,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrFeaturedProductNotFound
		}
		return nil, err
	}
	return &target, nil
}

func (r *Repository) ListAdminAuditLogs(ctx context.Context, q db.Queryer, filters AuditLogListFilters) ([]AdminAuditLogItem, error) {
	args := make([]any, 0)
	conditions := []string{"1 = 1"}

	if filters.ActorUserID != nil {
		args = append(args, *filters.ActorUserID)
		conditions = append(conditions, fmt.Sprintf("a.actor_user_id = $%d", len(args)))
	}
	if filters.Action != "" {
		args = append(args, filters.Action)
		conditions = append(conditions, fmt.Sprintf("a.action = $%d", len(args)))
	}
	if filters.TargetType != "" {
		args = append(args, filters.TargetType)
		conditions = append(conditions, fmt.Sprintf("a.target_type = $%d", len(args)))
	}
	if filters.TargetID != nil {
		args = append(args, *filters.TargetID)
		conditions = append(conditions, fmt.Sprintf("a.target_id = $%d", len(args)))
	}
	if filters.DateFrom != nil {
		args = append(args, *filters.DateFrom)
		conditions = append(conditions, fmt.Sprintf("a.created_at >= $%d", len(args)))
	}
	if filters.DateTo != nil {
		args = append(args, *filters.DateTo)
		conditions = append(conditions, fmt.Sprintf("a.created_at < $%d", len(args)))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(a.created_at, a.id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			a.id,
			COALESCE(a.actor_user_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(u.name, ''),
			a.action,
			COALESCE(a.target_type, ''),
			a.target_id,
			COALESCE(a.before_data, 'null'::jsonb),
			COALESCE(a.after_data, 'null'::jsonb),
			COALESCE(a.ip_address, ''),
			COALESCE(a.user_agent, ''),
			a.created_at
		FROM admin_audit_logs a
		LEFT JOIN users u ON u.id = a.actor_user_id
		WHERE %s
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AdminAuditLogItem, 0)
	for rows.Next() {
		var item AdminAuditLogItem
		if err := rows.Scan(
			&item.ID,
			&item.ActorUserID,
			&item.ActorName,
			&item.Action,
			&item.TargetType,
			&item.TargetID,
			&item.BeforeData,
			&item.AfterData,
			&item.IPAddress,
			&item.UserAgent,
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

func scanFeaturedItemPtr(row pgx.Row) (*FeaturedItem, error) {
	item, err := scanFeaturedItem(row)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func scanFeaturedItem(row pgx.Row) (FeaturedItem, error) {
	var item FeaturedItem
	var storeID uuid.NullUUID
	var productID uuid.NullUUID
	var startsAt sql.NullTime
	var endsAt sql.NullTime
	var createdBy uuid.NullUUID
	if err := row.Scan(
		&item.ID,
		&item.ItemType,
		&item.TenantID,
		&storeID,
		&productID,
		&item.Placement,
		&item.SortOrder,
		&startsAt,
		&endsAt,
		&item.IsActive,
		&createdBy,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.StoreName,
		&item.StoreSlug,
		&item.ProductName,
		&item.ProductSlug,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FeaturedItem{}, ErrFeaturedItemNotFound
		}
		return FeaturedItem{}, err
	}
	if storeID.Valid {
		item.StoreID = &storeID.UUID
	}
	if productID.Valid {
		item.ProductID = &productID.UUID
	}
	if startsAt.Valid {
		item.StartsAt = &startsAt.Time
	}
	if endsAt.Valid {
		item.EndsAt = &endsAt.Time
	}
	if createdBy.Valid {
		item.CreatedBy = &createdBy.UUID
	}
	return item, nil
}

func uuidPtrArg(value *uuid.UUID) any {
	if value == nil || *value == uuid.Nil {
		return nil
	}
	return *value
}

func timePtrArg(value *time.Time) any {
	if value == nil || value.IsZero() {
		return nil
	}
	return *value
}
