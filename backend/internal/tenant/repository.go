package tenant

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	ErrDefaultPlanNotFound    = errors.New("default plan not found")
	ErrTenantSlugAlreadyInUse = errors.New("tenant slug already in use")
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) FindDefaultPlanID(ctx context.Context, q db.Queryer) (uuid.UUID, error) {
	const query = `
		SELECT id
		FROM plans
		WHERE code = 'starter'
		  AND is_active = true
		LIMIT 1
	`

	var planID uuid.UUID
	if err := q.QueryRow(ctx, query).Scan(&planID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrDefaultPlanNotFound
		}
		return uuid.Nil, err
	}

	return planID, nil
}

func (r *Repository) Create(ctx context.Context, q db.Queryer, params CreateTenantParams) (*Tenant, error) {
	const query = `
		INSERT INTO tenants (plan_id, name, slug)
		VALUES ($1, $2, $3)
		RETURNING id, name, slug, status
	`

	var tenant Tenant
	err := q.QueryRow(ctx, query, params.PlanID, params.Name, params.Slug).Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Slug,
		&tenant.Status,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrTenantSlugAlreadyInUse
		}
		return nil, err
	}

	return &tenant, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
