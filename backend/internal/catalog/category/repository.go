package category

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	ErrCategoryNotFound         = errors.New("category not found")
	ErrCategorySlugAlreadyInUse = errors.New("category slug already in use")
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
) ([]Category, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			parent_id,
			name,
			slug,
			COALESCE(description, ''),
			sort_order,
			is_active
		FROM categories
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND deleted_at IS NULL
		  AND ($3::boolean IS NULL OR is_active = $3)
		ORDER BY sort_order ASC, name ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, filters.IsActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Category, 0)
	for rows.Next() {
		item, err := scanCategory(rows)
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

func (r *Repository) Create(ctx context.Context, q db.Queryer, params CreateParams) (*Category, error) {
	const query = `
		INSERT INTO categories (
			tenant_id,
			store_id,
			parent_id,
			name,
			slug,
			description,
			sort_order,
			is_active
		)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8)
		RETURNING
			id,
			tenant_id,
			store_id,
			parent_id,
			name,
			slug,
			COALESCE(description, ''),
			sort_order,
			is_active
	`

	category, err := scanCategory(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ParentID,
		params.Name,
		params.Slug,
		params.Description,
		params.SortOrder,
		params.IsActive,
	))
	if err != nil && isUniqueViolation(err) {
		return nil, ErrCategorySlugAlreadyInUse
	}
	return category, err
}

func (r *Repository) FindByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	categoryID uuid.UUID,
) (*Category, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			parent_id,
			name,
			slug,
			COALESCE(description, ''),
			sort_order,
			is_active
		FROM categories
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
		LIMIT 1
	`

	return scanCategory(q.QueryRow(ctx, query, tenantID, storeID, categoryID))
}

func (r *Repository) Update(ctx context.Context, q db.Queryer, params UpdateParams) (*Category, error) {
	const query = `
		UPDATE categories
		SET parent_id = $4,
		    name = $5,
		    slug = $6,
		    description = NULLIF($7, ''),
		    sort_order = $8,
		    is_active = $9,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
		RETURNING
			id,
			tenant_id,
			store_id,
			parent_id,
			name,
			slug,
			COALESCE(description, ''),
			sort_order,
			is_active
	`

	category, err := scanCategory(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.CategoryID,
		params.ParentID,
		params.Name,
		params.Slug,
		params.Description,
		params.SortOrder,
		params.IsActive,
	))
	if err != nil && isUniqueViolation(err) {
		return nil, ErrCategorySlugAlreadyInUse
	}
	return category, err
}

func (r *Repository) SoftDelete(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	categoryID uuid.UUID,
) error {
	const query = `
		UPDATE categories
		SET deleted_at = now(),
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
	`

	tag, err := q.Exec(ctx, query, tenantID, storeID, categoryID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrCategoryNotFound
	}

	return nil
}

func scanCategory(row pgx.Row) (*Category, error) {
	var category Category
	if err := row.Scan(
		&category.ID,
		&category.TenantID,
		&category.StoreID,
		&category.ParentID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.SortOrder,
		&category.IsActive,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}

	return &category, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
