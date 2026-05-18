package store

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	ErrStoreNotFound         = errors.New("store not found")
	ErrStoreSlugAlreadyInUse = errors.New("store slug already in use")
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, q db.Queryer, params CreateParams) (*Store, error) {
	const query = `
		INSERT INTO stores (
			tenant_id,
			name,
			slug,
			description,
			phone,
			whatsapp,
			email,
			address,
			city,
			province,
			postal_code
		)
		VALUES (
			$1,
			$2,
			$3,
			NULLIF($4, ''),
			NULLIF($5, ''),
			NULLIF($6, ''),
			NULLIF($7, ''),
			NULLIF($8, ''),
			NULLIF($9, ''),
			NULLIF($10, ''),
			NULLIF($11, '')
		)
		RETURNING
			id,
			tenant_id,
			name,
			slug,
			COALESCE(description, ''),
			COALESCE(logo_url, ''),
			COALESCE(banner_url, ''),
			COALESCE(phone, ''),
			COALESCE(whatsapp, ''),
			COALESCE(email, ''),
			COALESCE(address, ''),
			COALESCE(city, ''),
			COALESCE(province, ''),
			COALESCE(postal_code, ''),
			status,
			is_discoverable,
			published_at
	`

	return scanStore(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.Name,
		params.Slug,
		params.Description,
		params.Phone,
		params.Whatsapp,
		params.Email,
		params.Address,
		params.City,
		params.Province,
		params.PostalCode,
	))
}

func (r *Repository) FindCurrentByTenantID(ctx context.Context, q db.Queryer, tenantID uuid.UUID) (*Store, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			name,
			slug,
			COALESCE(description, ''),
			COALESCE(logo_url, ''),
			COALESCE(banner_url, ''),
			COALESCE(phone, ''),
			COALESCE(whatsapp, ''),
			COALESCE(email, ''),
			COALESCE(address, ''),
			COALESCE(city, ''),
			COALESCE(province, ''),
			COALESCE(postal_code, ''),
			status,
			is_discoverable,
			published_at
		FROM stores
		WHERE tenant_id = $1
		  AND deleted_at IS NULL
		LIMIT 1
	`

	return scanStore(q.QueryRow(ctx, query, tenantID))
}

func (r *Repository) UpdateProfile(ctx context.Context, q db.Queryer, params UpdateProfileParams) (*Store, error) {
	const query = `
		UPDATE stores
		SET name = $3,
		    description = NULLIF($4, ''),
		    phone = NULLIF($5, ''),
		    whatsapp = NULLIF($6, ''),
		    email = NULLIF($7, ''),
		    address = NULLIF($8, ''),
		    city = NULLIF($9, ''),
		    province = NULLIF($10, ''),
		    postal_code = NULLIF($11, ''),
		    is_discoverable = $12,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND id = $2
		  AND deleted_at IS NULL
		RETURNING
			id,
			tenant_id,
			name,
			slug,
			COALESCE(description, ''),
			COALESCE(logo_url, ''),
			COALESCE(banner_url, ''),
			COALESCE(phone, ''),
			COALESCE(whatsapp, ''),
			COALESCE(email, ''),
			COALESCE(address, ''),
			COALESCE(city, ''),
			COALESCE(province, ''),
			COALESCE(postal_code, ''),
			status,
			is_discoverable,
			published_at
	`

	return scanStore(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.Name,
		params.Description,
		params.Phone,
		params.Whatsapp,
		params.Email,
		params.Address,
		params.City,
		params.Province,
		params.PostalCode,
		params.IsDiscoverable,
	))
}

func (r *Repository) UpdateStatus(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	status string,
	publishedAt *time.Time,
) (*Store, error) {
	const query = `
		UPDATE stores
		SET status = $3,
		    published_at = $4,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND id = $2
		  AND deleted_at IS NULL
		RETURNING
			id,
			tenant_id,
			name,
			slug,
			COALESCE(description, ''),
			COALESCE(logo_url, ''),
			COALESCE(banner_url, ''),
			COALESCE(phone, ''),
			COALESCE(whatsapp, ''),
			COALESCE(email, ''),
			COALESCE(address, ''),
			COALESCE(city, ''),
			COALESCE(province, ''),
			COALESCE(postal_code, ''),
			status,
			is_discoverable,
			published_at
	`

	return scanStore(q.QueryRow(ctx, query, tenantID, storeID, status, publishedAt))
}

func (r *Repository) ReplaceBusinessHours(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	items []BusinessHour,
) error {
	const deleteQuery = `
		DELETE FROM store_business_hours
		WHERE tenant_id = $1
		  AND store_id = $2
	`
	if _, err := q.Exec(ctx, deleteQuery, tenantID, storeID); err != nil {
		return err
	}

	const insertQuery = `
		INSERT INTO store_business_hours (
			tenant_id,
			store_id,
			day_of_week,
			open_time,
			close_time,
			is_closed
		)
		VALUES ($1, $2, $3, NULLIF($4, '')::time, NULLIF($5, '')::time, $6)
	`

	for _, item := range items {
		if _, err := q.Exec(
			ctx,
			insertQuery,
			tenantID,
			storeID,
			item.DayOfWeek,
			item.OpenTime,
			item.CloseTime,
			item.IsClosed,
		); err != nil {
			return err
		}
	}

	return nil
}

func scanStore(row pgx.Row) (*Store, error) {
	var store Store
	if err := row.Scan(
		&store.ID,
		&store.TenantID,
		&store.Name,
		&store.Slug,
		&store.Description,
		&store.LogoURL,
		&store.BannerURL,
		&store.Phone,
		&store.Whatsapp,
		&store.Email,
		&store.Address,
		&store.City,
		&store.Province,
		&store.PostalCode,
		&store.Status,
		&store.IsDiscoverable,
		&store.PublishedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStoreNotFound
		}
		if isUniqueViolation(err) {
			return nil, ErrStoreSlugAlreadyInUse
		}
		return nil, err
	}

	return &store, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
