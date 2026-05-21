package plan

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) FindByTenantID(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (*Plan, error) {
	const query = `
		SELECT
			p.id,
			p.code,
			p.name,
			COALESCE(p.description, ''),
			p.price_monthly,
			p.product_limit,
			p.staff_limit,
			p.can_use_pos,
			p.can_use_discovery,
			p.can_use_courier,
			p.can_use_custom_domain,
			p.is_active
		FROM tenants t
		JOIN plans p
		  ON p.id = t.plan_id
		WHERE t.id = $1
		  AND t.deleted_at IS NULL
		LIMIT 1
	`
	return scanPlan(q.QueryRow(ctx, query, tenantID))
}

func (r *Repository) CountActiveProducts(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM products
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND deleted_at IS NULL
		  AND status <> 'archived'
	`
	var count int
	if err := q.QueryRow(ctx, query, tenantID, storeID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *Repository) CountActiveStaff(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (int, error) {
	const query = `
		SELECT COUNT(*)
		FROM user_tenants
		WHERE tenant_id = $1
		  AND status = 'active'
		  AND role <> 'owner'
	`
	var count int
	if err := q.QueryRow(ctx, query, tenantID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func scanPlan(row pgx.Row) (*Plan, error) {
	var item Plan
	var productLimit sql.NullInt64
	var staffLimit sql.NullInt64
	if err := row.Scan(
		&item.ID,
		&item.Code,
		&item.Name,
		&item.Description,
		&item.PriceMonthly,
		&productLimit,
		&staffLimit,
		&item.CanUsePOS,
		&item.CanUseDiscovery,
		&item.CanUseCourier,
		&item.CanUseCustomDomain,
		&item.IsActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	if productLimit.Valid {
		value := int(productLimit.Int64)
		item.ProductLimit = &value
	}
	if staffLimit.Valid {
		value := int(staffLimit.Int64)
		item.StaffLimit = &value
	}
	return &item, nil
}
