package courier

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) ListZones(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters ListZoneFilters,
) ([]Zone, error) {
	var isActive any
	if filters.IsActive != nil {
		isActive = *filters.IsActive
	}

	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			name,
			COALESCE(description, ''),
			rate,
			is_active,
			sort_order,
			created_at,
			updated_at,
			deleted_at
		FROM courier_zones
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND deleted_at IS NULL
		  AND ($3::boolean IS NULL OR is_active = $3)
		ORDER BY sort_order ASC, name ASC, id ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, isActive)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	zones := make([]Zone, 0)
	for rows.Next() {
		zone, err := scanZone(rows)
		if err != nil {
			return nil, err
		}
		zones = append(zones, *zone)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return zones, nil
}

func (r *Repository) ListPublicActiveZones(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID) ([]Zone, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			name,
			COALESCE(description, ''),
			rate,
			is_active,
			sort_order,
			created_at,
			updated_at,
			deleted_at
		FROM courier_zones
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND is_active = true
		  AND deleted_at IS NULL
		ORDER BY sort_order ASC, name ASC, id ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	zones := make([]Zone, 0)
	for rows.Next() {
		zone, err := scanZone(rows)
		if err != nil {
			return nil, err
		}
		zones = append(zones, *zone)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return zones, nil
}

func (r *Repository) FindZoneByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, zoneID uuid.UUID) (*Zone, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			name,
			COALESCE(description, ''),
			rate,
			is_active,
			sort_order,
			created_at,
			updated_at,
			deleted_at
		FROM courier_zones
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
		LIMIT 1
	`

	return scanZone(q.QueryRow(ctx, query, tenantID, storeID, zoneID))
}

func (r *Repository) CreateZone(ctx context.Context, q db.Queryer, params CreateZoneParams) (*Zone, error) {
	const query = `
		INSERT INTO courier_zones (
			tenant_id,
			store_id,
			name,
			description,
			rate,
			is_active,
			sort_order
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)
		RETURNING
			id,
			tenant_id,
			store_id,
			name,
			COALESCE(description, ''),
			rate,
			is_active,
			sort_order,
			created_at,
			updated_at,
			deleted_at
	`

	return scanZone(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.Name,
		params.Description,
		params.Rate,
		params.IsActive,
		params.SortOrder,
	))
}

func (r *Repository) UpdateZone(ctx context.Context, q db.Queryer, params UpdateZoneParams) (*Zone, error) {
	const query = `
		UPDATE courier_zones
		SET name = $4,
		    description = NULLIF($5, ''),
		    rate = $6,
		    is_active = $7,
		    sort_order = $8,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
		RETURNING
			id,
			tenant_id,
			store_id,
			name,
			COALESCE(description, ''),
			rate,
			is_active,
			sort_order,
			created_at,
			updated_at,
			deleted_at
	`

	return scanZone(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ZoneID,
		params.Name,
		params.Description,
		params.Rate,
		params.IsActive,
		params.SortOrder,
	))
}

func (r *Repository) SoftDeleteZone(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	zoneID uuid.UUID,
) (*Zone, error) {
	const query = `
		UPDATE courier_zones
		SET deleted_at = now(),
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND deleted_at IS NULL
		RETURNING
			id,
			tenant_id,
			store_id,
			name,
			COALESCE(description, ''),
			rate,
			is_active,
			sort_order,
			created_at,
			updated_at,
			deleted_at
	`

	return scanZone(q.QueryRow(ctx, query, tenantID, storeID, zoneID))
}

func scanZone(row pgx.Row) (*Zone, error) {
	var zone Zone
	if err := row.Scan(
		&zone.ID,
		&zone.TenantID,
		&zone.StoreID,
		&zone.Name,
		&zone.Description,
		&zone.Rate,
		&zone.IsActive,
		&zone.SortOrder,
		&zone.CreatedAt,
		&zone.UpdatedAt,
		&zone.DeletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrZoneNotFound
		}
		return nil, err
	}
	return &zone, nil
}
