package category

import (
	"context"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

func (r *Repository) ListPublic(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
) ([]PublicCategory, error) {
	const query = `
		SELECT
			id,
			name,
			slug,
			COALESCE(image_url, '')
		FROM categories
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND is_active = true
		  AND deleted_at IS NULL
		ORDER BY sort_order ASC, name ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]PublicCategory, 0)
	for rows.Next() {
		var item PublicCategory
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Slug,
			&item.ImageURL,
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
