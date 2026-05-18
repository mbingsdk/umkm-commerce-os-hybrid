package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var ErrProductImageNotFound = errors.New("product image not found")

type ImageRepository struct{}

func NewImageRepository() *ImageRepository {
	return &ImageRepository{}
}

func (r *ImageRepository) Create(ctx context.Context, q db.Queryer, params CreateImageParams) (*Image, error) {
	if params.IsPrimary {
		const clearPrimaryQuery = `
			UPDATE product_images pi
			SET is_primary = false
			FROM products p
			WHERE pi.tenant_id = $1
			  AND pi.product_id = p.id
			  AND p.tenant_id = $1
			  AND p.store_id = $2
			  AND p.id = $3
			  AND p.deleted_at IS NULL
		`
		if _, err := q.Exec(ctx, clearPrimaryQuery, params.TenantID, params.StoreID, params.ProductID); err != nil {
			return nil, err
		}
	}

	const query = `
		INSERT INTO product_images (
			tenant_id,
			product_id,
			url,
			alt_text,
			sort_order,
			is_primary
		)
		SELECT
			$1,
			p.id,
			$4,
			NULLIF($5, ''),
			$6,
			$7
		FROM products p
		WHERE p.tenant_id = $1
		  AND p.store_id = $2
		  AND p.id = $3
		  AND p.deleted_at IS NULL
		RETURNING
			id,
			url,
			COALESCE(alt_text, ''),
			is_primary,
			sort_order
	`

	var image Image
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.URL,
		params.AltText,
		params.SortOrder,
		params.IsPrimary,
	).Scan(
		&image.ID,
		&image.URL,
		&image.AltText,
		&image.IsPrimary,
		&image.SortOrder,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return &image, nil
}

func (r *ImageRepository) Delete(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productID uuid.UUID,
	imageID uuid.UUID,
) error {
	const query = `
		DELETE FROM product_images pi
		USING products p
		WHERE pi.id = $4
		  AND pi.tenant_id = $1
		  AND pi.product_id = p.id
		  AND p.tenant_id = $1
		  AND p.store_id = $2
		  AND p.id = $3
		  AND p.deleted_at IS NULL
	`

	tag, err := q.Exec(ctx, query, tenantID, storeID, productID, imageID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrProductImageNotFound
	}
	return nil
}
