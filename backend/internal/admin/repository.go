package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var ErrAdminUserNotFound = errors.New("admin user not found")

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) FindUserByID(ctx context.Context, q db.Queryer, userID uuid.UUID) (*User, error) {
	const query = `
		SELECT id, name, email, platform_role, status
		FROM users
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	var user User
	if err := q.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PlatformRole,
		&user.Status,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) CreateAuditLog(ctx context.Context, q db.Queryer, entry AuditEntry) (*AuditLog, error) {
	beforeData, err := marshalOptionalJSON(entry.BeforeData)
	if err != nil {
		return nil, err
	}
	afterData, err := marshalOptionalJSON(entry.AfterData)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO admin_audit_logs (
			actor_user_id,
			action,
			target_type,
			target_id,
			before_data,
			after_data,
			ip_address,
			user_agent
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			$4,
			$5,
			$6,
			NULLIF($7, ''),
			NULLIF($8, '')
		)
		RETURNING id, actor_user_id, action, COALESCE(target_type, ''), target_id, created_at
	`

	var log AuditLog
	if err := q.QueryRow(
		ctx,
		query,
		entry.ActorUserID,
		entry.Action,
		entry.TargetType,
		entry.TargetID,
		beforeData,
		afterData,
		entry.IPAddress,
		entry.UserAgent,
	).Scan(
		&log.ID,
		&log.ActorUserID,
		&log.Action,
		&log.TargetType,
		&log.TargetID,
		&log.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &log, nil
}

func (r *Repository) ListTenants(ctx context.Context, q db.Queryer, filters TenantListFilters) ([]TenantListItem, error) {
	args := []any{}
	conditions := []string{"t.deleted_at IS NULL"}

	if filters.Status != "" {
		args = append(args, filters.Status)
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", len(args)))
	}
	if filters.PlanID != nil {
		args = append(args, *filters.PlanID)
		conditions = append(conditions, fmt.Sprintf("t.plan_id = $%d", len(args)))
	}
	if filters.Query != "" {
		args = append(args, filters.Query)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf(`(
			t.name ILIKE '%%' || $%d || '%%'
			OR t.slug ILIKE '%%' || $%d || '%%'
			OR COALESCE(primary_store.name, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(owner_user.email::text, '') ILIKE '%%' || $%d || '%%'
		)`, placeholder, placeholder, placeholder, placeholder))
	}
	if filters.CreatedFrom != nil {
		args = append(args, *filters.CreatedFrom)
		conditions = append(conditions, fmt.Sprintf("t.created_at >= $%d", len(args)))
	}
	if filters.CreatedTo != nil {
		args = append(args, *filters.CreatedTo)
		conditions = append(conditions, fmt.Sprintf("t.created_at < $%d", len(args)))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(t.created_at, t.id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			t.id,
			t.plan_id,
			t.name,
			t.slug,
			t.status,
			t.created_at,
			t.updated_at,
			COALESCE(p.id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(p.code, ''),
			COALESCE(p.name, ''),
			COALESCE(p.description, ''),
			COALESCE(p.price_monthly, 0)::bigint,
			p.product_limit,
			p.staff_limit,
			COALESCE(p.can_use_pos, false),
			COALESCE(p.can_use_discovery, false),
			COALESCE(p.can_use_courier, false),
			COALESCE(p.can_use_custom_domain, false),
			COALESCE(p.is_active, false),
			COALESCE(primary_store.id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(primary_store.name, ''),
			COALESCE(primary_store.slug, ''),
			COALESCE(primary_store.status, ''),
			COALESCE(primary_store.city, ''),
			COALESCE(primary_store.created_at, '0001-01-01T00:00:00Z'::timestamptz),
			COALESCE(owner_user.id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(owner_user.name, ''),
			COALESCE(owner_user.email::text, ''),
			COALESCE(owner_user.status, ''),
			COALESCE(counts.store_count, 0)::bigint,
			COALESCE(counts.product_count, 0)::bigint,
			COALESCE(counts.order_count, 0)::bigint,
			COALESCE(counts.user_count, 0)::bigint,
			COALESCE(counts.pos_transaction_count, 0)::bigint
		FROM tenants t
		LEFT JOIN plans p
		  ON p.id = t.plan_id
		LEFT JOIN LATERAL (
			SELECT id, name, slug, status, COALESCE(city, '') AS city, created_at
			FROM stores
			WHERE tenant_id = t.id
			  AND deleted_at IS NULL
			ORDER BY created_at ASC
			LIMIT 1
		) primary_store ON true
		LEFT JOIN LATERAL (
			SELECT u.id, u.name, u.email, u.status
			FROM user_tenants ut
			JOIN users u
			  ON u.id = ut.user_id
			 AND u.deleted_at IS NULL
			WHERE ut.tenant_id = t.id
			  AND ut.role = 'owner'
			  AND ut.status = 'active'
			ORDER BY ut.created_at ASC
			LIMIT 1
		) owner_user ON true
		LEFT JOIN LATERAL (
			SELECT
				(SELECT COUNT(*) FROM stores s WHERE s.tenant_id = t.id AND s.deleted_at IS NULL) AS store_count,
				(SELECT COUNT(*) FROM products p2 WHERE p2.tenant_id = t.id AND p2.deleted_at IS NULL) AS product_count,
				(SELECT COUNT(*) FROM orders o WHERE o.tenant_id = t.id) AS order_count,
				(SELECT COUNT(*) FROM user_tenants ut2 WHERE ut2.tenant_id = t.id AND ut2.status = 'active') AS user_count,
				(SELECT COUNT(*) FROM pos_transactions pt WHERE pt.tenant_id = t.id) AS pos_transaction_count
		) counts ON true
		WHERE %s
		ORDER BY t.created_at DESC, t.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TenantListItem, 0)
	for rows.Next() {
		item, err := scanTenantListItem(rows)
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

func (r *Repository) FindTenantByIDForUpdate(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (*Tenant, error) {
	const query = `
		SELECT id, plan_id, name, slug, status, created_at, updated_at
		FROM tenants
		WHERE id = $1
		  AND deleted_at IS NULL
		FOR UPDATE
	`
	return scanTenant(q.QueryRow(ctx, query, tenantID))
}

func (r *Repository) GetTenantDetail(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (*TenantDetail, error) {
	const query = `
		SELECT
			t.id,
			t.plan_id,
			t.name,
			t.slug,
			t.status,
			t.created_at,
			t.updated_at,
			COALESCE(p.id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(p.code, ''),
			COALESCE(p.name, ''),
			COALESCE(p.description, ''),
			COALESCE(p.price_monthly, 0)::bigint,
			p.product_limit,
			p.staff_limit,
			COALESCE(p.can_use_pos, false),
			COALESCE(p.can_use_discovery, false),
			COALESCE(p.can_use_courier, false),
			COALESCE(p.can_use_custom_domain, false),
			COALESCE(p.is_active, false),
			COALESCE(primary_store.id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(primary_store.name, ''),
			COALESCE(primary_store.slug, ''),
			COALESCE(primary_store.status, ''),
			COALESCE(primary_store.city, ''),
			COALESCE(primary_store.created_at, '0001-01-01T00:00:00Z'::timestamptz),
			COALESCE(owner_user.id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(owner_user.name, ''),
			COALESCE(owner_user.email::text, ''),
			COALESCE(owner_user.status, ''),
			COALESCE(counts.store_count, 0)::bigint,
			COALESCE(counts.product_count, 0)::bigint,
			COALESCE(counts.order_count, 0)::bigint,
			COALESCE(counts.user_count, 0)::bigint,
			COALESCE(counts.pos_transaction_count, 0)::bigint
		FROM tenants t
		LEFT JOIN plans p
		  ON p.id = t.plan_id
		LEFT JOIN LATERAL (
			SELECT id, name, slug, status, COALESCE(city, '') AS city, created_at
			FROM stores
			WHERE tenant_id = t.id
			  AND deleted_at IS NULL
			ORDER BY created_at ASC
			LIMIT 1
		) primary_store ON true
		LEFT JOIN LATERAL (
			SELECT u.id, u.name, u.email, u.status
			FROM user_tenants ut
			JOIN users u
			  ON u.id = ut.user_id
			 AND u.deleted_at IS NULL
			WHERE ut.tenant_id = t.id
			  AND ut.role = 'owner'
			  AND ut.status = 'active'
			ORDER BY ut.created_at ASC
			LIMIT 1
		) owner_user ON true
		LEFT JOIN LATERAL (
			SELECT
				(SELECT COUNT(*) FROM stores s WHERE s.tenant_id = t.id AND s.deleted_at IS NULL) AS store_count,
				(SELECT COUNT(*) FROM products p2 WHERE p2.tenant_id = t.id AND p2.deleted_at IS NULL) AS product_count,
				(SELECT COUNT(*) FROM orders o WHERE o.tenant_id = t.id) AS order_count,
				(SELECT COUNT(*) FROM user_tenants ut2 WHERE ut2.tenant_id = t.id AND ut2.status = 'active') AS user_count,
				(SELECT COUNT(*) FROM pos_transactions pt WHERE pt.tenant_id = t.id) AS pos_transaction_count
		) counts ON true
		WHERE t.id = $1
		  AND t.deleted_at IS NULL
		LIMIT 1
	`

	detail, err := scanTenantDetail(q.QueryRow(ctx, query, tenantID))
	if err != nil {
		return nil, err
	}

	audits, err := r.ListAuditSnippets(ctx, q, tenantID, 5)
	if err != nil {
		return nil, err
	}
	detail.LatestAudits = audits
	return detail, nil
}

func (r *Repository) UpdateTenantStatus(ctx context.Context, q db.Queryer, tenantID uuid.UUID, status string) (*Tenant, error) {
	const query = `
		UPDATE tenants
		SET status = $2,
		    updated_at = now()
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING id, plan_id, name, slug, status, created_at, updated_at
	`
	return scanTenant(q.QueryRow(ctx, query, tenantID, status))
}

func (r *Repository) FindActivePlanByID(ctx context.Context, q db.Queryer, planID uuid.UUID) (*Plan, error) {
	const query = `
		SELECT
			id,
			code,
			name,
			COALESCE(description, ''),
			price_monthly,
			product_limit,
			staff_limit,
			can_use_pos,
			can_use_discovery,
			can_use_courier,
			can_use_custom_domain,
			is_active
		FROM plans
		WHERE id = $1
		  AND is_active = true
		LIMIT 1
	`
	return scanPlan(q.QueryRow(ctx, query, planID))
}

func (r *Repository) ListPlans(ctx context.Context, q db.Queryer) ([]Plan, error) {
	const query = `
		SELECT
			id,
			code,
			name,
			COALESCE(description, ''),
			price_monthly,
			product_limit,
			staff_limit,
			can_use_pos,
			can_use_discovery,
			can_use_courier,
			can_use_custom_domain,
			is_active
		FROM plans
		ORDER BY price_monthly ASC, name ASC
	`

	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Plan, 0)
	for rows.Next() {
		item, err := scanPlan(rows)
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

func (r *Repository) FindPlanByIDForUpdate(ctx context.Context, q db.Queryer, planID uuid.UUID) (*Plan, error) {
	const query = `
		SELECT
			id,
			code,
			name,
			COALESCE(description, ''),
			price_monthly,
			product_limit,
			staff_limit,
			can_use_pos,
			can_use_discovery,
			can_use_courier,
			can_use_custom_domain,
			is_active
		FROM plans
		WHERE id = $1
		FOR UPDATE
	`
	return scanPlan(q.QueryRow(ctx, query, planID))
}

func (r *Repository) CreatePlan(ctx context.Context, q db.Queryer, params CreatePlanParams) (*Plan, error) {
	const query = `
		INSERT INTO plans (
			code,
			name,
			description,
			price_monthly,
			product_limit,
			staff_limit,
			can_use_pos,
			can_use_discovery,
			can_use_courier,
			can_use_custom_domain,
			is_active
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10,
			$11
		)
		RETURNING
			id,
			code,
			name,
			COALESCE(description, ''),
			price_monthly,
			product_limit,
			staff_limit,
			can_use_pos,
			can_use_discovery,
			can_use_courier,
			can_use_custom_domain,
			is_active
	`

	plan, err := scanPlan(q.QueryRow(
		ctx,
		query,
		params.Code,
		params.Name,
		params.Description,
		params.PriceMonthly,
		nullableIntArg(params.ProductLimit),
		nullableIntArg(params.StaffLimit),
		params.CanUsePOS,
		params.CanUseDiscovery,
		params.CanUseCourier,
		params.CanUseCustomDomain,
		params.IsActive,
	))
	if err != nil && isUniqueViolation(err) {
		return nil, ErrPlanCodeAlreadyInUse
	}
	return plan, err
}

func (r *Repository) UpdatePlan(ctx context.Context, q db.Queryer, params UpdatePlanParams) (*Plan, error) {
	const query = `
		UPDATE plans
		SET code = $2,
		    name = $3,
		    description = NULLIF($4, ''),
		    price_monthly = $5,
		    product_limit = $6,
		    staff_limit = $7,
		    can_use_pos = $8,
		    can_use_discovery = $9,
		    can_use_courier = $10,
		    can_use_custom_domain = $11,
		    is_active = $12,
		    updated_at = now()
		WHERE id = $1
		RETURNING
			id,
			code,
			name,
			COALESCE(description, ''),
			price_monthly,
			product_limit,
			staff_limit,
			can_use_pos,
			can_use_discovery,
			can_use_courier,
			can_use_custom_domain,
			is_active
	`

	plan, err := scanPlan(q.QueryRow(
		ctx,
		query,
		params.PlanID,
		params.Code,
		params.Name,
		params.Description,
		params.PriceMonthly,
		nullableIntArg(params.ProductLimit),
		nullableIntArg(params.StaffLimit),
		params.CanUsePOS,
		params.CanUseDiscovery,
		params.CanUseCourier,
		params.CanUseCustomDomain,
		params.IsActive,
	))
	if err != nil && isUniqueViolation(err) {
		return nil, ErrPlanCodeAlreadyInUse
	}
	return plan, err
}

func (r *Repository) UpdateTenantPlan(ctx context.Context, q db.Queryer, tenantID uuid.UUID, planID uuid.UUID) (*Tenant, error) {
	const query = `
		UPDATE tenants
		SET plan_id = $2,
		    updated_at = now()
		WHERE id = $1
		  AND deleted_at IS NULL
		RETURNING id, plan_id, name, slug, status, created_at, updated_at
	`
	return scanTenant(q.QueryRow(ctx, query, tenantID, planID))
}

func (r *Repository) ListAuditSnippets(ctx context.Context, q db.Queryer, tenantID uuid.UUID, limit int) ([]AuditSnippet, error) {
	const query = `
		SELECT
			a.id,
			COALESCE(a.actor_user_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(u.name, ''),
			a.action,
			COALESCE(a.target_type, ''),
			a.target_id,
			a.created_at
		FROM admin_audit_logs a
		LEFT JOIN users u
		  ON u.id = a.actor_user_id
		WHERE a.target_type = 'tenant'
		  AND a.target_id = $1
		ORDER BY a.created_at DESC
		LIMIT $2
	`

	rows, err := q.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]AuditSnippet, 0)
	for rows.Next() {
		var item AuditSnippet
		if err := rows.Scan(
			&item.ID,
			&item.ActorUserID,
			&item.ActorName,
			&item.Action,
			&item.TargetType,
			&item.TargetID,
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

func scanTenant(row pgx.Row) (*Tenant, error) {
	var tenant Tenant
	var planID uuid.NullUUID
	if err := row.Scan(
		&tenant.ID,
		&planID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	if planID.Valid {
		tenant.PlanID = &planID.UUID
	}
	return &tenant, nil
}

func scanPlan(row pgx.Row) (*Plan, error) {
	var plan Plan
	var productLimit sql.NullInt64
	var staffLimit sql.NullInt64
	if err := row.Scan(
		&plan.ID,
		&plan.Code,
		&plan.Name,
		&plan.Description,
		&plan.PriceMonthly,
		&productLimit,
		&staffLimit,
		&plan.CanUsePOS,
		&plan.CanUseDiscovery,
		&plan.CanUseCourier,
		&plan.CanUseCustomDomain,
		&plan.IsActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, err
	}
	plan.ProductLimit = nullableInt(productLimit)
	plan.StaffLimit = nullableInt(staffLimit)
	return &plan, nil
}

func scanTenantListItem(row pgx.Row) (TenantListItem, error) {
	var item TenantListItem
	var plan Plan
	var primaryStore StoreSummary
	var owner OwnerSummary

	if err := scanTenantOverview(row, &item.Tenant, &plan, &primaryStore, &owner, &item.Counts); err != nil {
		return TenantListItem{}, err
	}

	if plan.ID != uuid.Nil {
		item.Plan = &plan
	}
	if primaryStore.ID != uuid.Nil {
		item.PrimaryStore = &primaryStore
	}
	if owner.ID != uuid.Nil {
		item.Owner = &owner
	}
	return item, nil
}

func scanTenantDetail(row pgx.Row) (*TenantDetail, error) {
	var detail TenantDetail
	var plan Plan
	var primaryStore StoreSummary
	var owner OwnerSummary

	if err := scanTenantOverview(row, &detail.Tenant, &plan, &primaryStore, &owner, &detail.Counts); err != nil {
		return nil, err
	}

	if plan.ID != uuid.Nil {
		detail.Plan = &plan
	}
	if primaryStore.ID != uuid.Nil {
		detail.PrimaryStore = &primaryStore
	}
	if owner.ID != uuid.Nil {
		detail.Owner = &owner
	}
	return &detail, nil
}

func scanTenantOverview(
	row pgx.Row,
	tenant *Tenant,
	plan *Plan,
	primaryStore *StoreSummary,
	owner *OwnerSummary,
	counts *TenantCounts,
) error {
	var planID uuid.NullUUID
	var productLimit sql.NullInt64
	var staffLimit sql.NullInt64
	if err := row.Scan(
		&tenant.ID,
		&planID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Status,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
		&plan.ID,
		&plan.Code,
		&plan.Name,
		&plan.Description,
		&plan.PriceMonthly,
		&productLimit,
		&staffLimit,
		&plan.CanUsePOS,
		&plan.CanUseDiscovery,
		&plan.CanUseCourier,
		&plan.CanUseCustomDomain,
		&plan.IsActive,
		&primaryStore.ID,
		&primaryStore.Name,
		&primaryStore.Slug,
		&primaryStore.Status,
		&primaryStore.City,
		&primaryStore.CreatedAt,
		&owner.ID,
		&owner.Name,
		&owner.Email,
		&owner.Status,
		&counts.StoreCount,
		&counts.ProductCount,
		&counts.OrderCount,
		&counts.UserCount,
		&counts.POSTransactionCount,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrTenantNotFound
		}
		return err
	}
	if planID.Valid {
		tenant.PlanID = &planID.UUID
	}
	plan.ProductLimit = nullableInt(productLimit)
	plan.StaffLimit = nullableInt(staffLimit)
	return nil
}

func nullableInt(value sql.NullInt64) *int {
	if !value.Valid {
		return nil
	}
	converted := int(value.Int64)
	return &converted
}

func nullableIntArg(value *int) any {
	if value == nil {
		return nil
	}
	return *value
}

func marshalOptionalJSON(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	return json.Marshal(value)
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
