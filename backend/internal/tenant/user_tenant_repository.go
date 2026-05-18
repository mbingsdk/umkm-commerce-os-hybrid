package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var ErrTenantAccessNotFound = errors.New("tenant access not found")

type UserTenantRepository struct{}

func NewUserTenantRepository() *UserTenantRepository {
	return &UserTenantRepository{}
}

func (r *UserTenantRepository) Create(
	ctx context.Context,
	q db.Queryer,
	params CreateMembershipParams,
) (*Membership, error) {
	const query = `
		INSERT INTO user_tenants (
			user_id,
			tenant_id,
			role,
			status,
			joined_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, tenant_id, role, status, joined_at
	`

	var membership Membership
	if err := q.QueryRow(
		ctx,
		query,
		params.UserID,
		params.TenantID,
		params.Role,
		params.Status,
		params.JoinedAt,
	).Scan(
		&membership.ID,
		&membership.UserID,
		&membership.TenantID,
		&membership.Role,
		&membership.Status,
		&membership.JoinedAt,
	); err != nil {
		return nil, err
	}

	return &membership, nil
}

func (r *UserTenantRepository) FindActiveAccess(
	ctx context.Context,
	q db.Queryer,
	userID uuid.UUID,
	tenantID uuid.UUID,
) (*AccessRecord, error) {
	const query = `
		SELECT
			t.id,
			t.status,
			s.id,
			ut.role
		FROM user_tenants ut
		JOIN tenants t
		  ON t.id = ut.tenant_id
		 AND t.deleted_at IS NULL
		JOIN stores s
		  ON s.tenant_id = t.id
		 AND s.deleted_at IS NULL
		WHERE ut.user_id = $1
		  AND ut.tenant_id = $2
		  AND ut.status = 'active'
		LIMIT 1
	`

	var access AccessRecord
	if err := q.QueryRow(ctx, query, userID, tenantID).Scan(
		&access.TenantID,
		&access.TenantStatus,
		&access.StoreID,
		&access.Role,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantAccessNotFound
		}
		return nil, err
	}

	return &access, nil
}

func (r *UserTenantRepository) ListByUserID(
	ctx context.Context,
	q db.Queryer,
	userID uuid.UUID,
) ([]TenantListItem, error) {
	const query = `
		SELECT
			t.id,
			t.name,
			t.slug,
			ut.role,
			t.status,
			s.id,
			s.name,
			s.slug,
			s.status
		FROM user_tenants ut
		JOIN tenants t
		  ON t.id = ut.tenant_id
		 AND t.deleted_at IS NULL
		JOIN stores s
		  ON s.tenant_id = t.id
		 AND s.deleted_at IS NULL
		WHERE ut.user_id = $1
		  AND ut.status = 'active'
		ORDER BY t.created_at ASC
	`

	rows, err := q.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TenantListItem, 0)
	for rows.Next() {
		var item TenantListItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Slug,
			&item.Role,
			&item.Status,
			&item.Store.ID,
			&item.Store.Name,
			&item.Store.Slug,
			&item.Store.Status,
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
