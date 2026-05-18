package inventory

import (
	"context"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) CreateSnapshot(
	ctx context.Context,
	q db.Queryer,
	params CreateSnapshotParams,
) (*StockSnapshot, error) {
	const query = `
		INSERT INTO product_stock_snapshots (
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING
			id,
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available,
			low_stock_threshold
	`

	var snapshot StockSnapshot
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.QuantityOnHand,
		params.QuantityReserved,
		params.QuantityAvailable,
		params.LowStockThreshold,
	).Scan(
		&snapshot.ID,
		&snapshot.TenantID,
		&snapshot.StoreID,
		&snapshot.ProductID,
		&snapshot.QuantityOnHand,
		&snapshot.QuantityReserved,
		&snapshot.QuantityAvailable,
		&snapshot.LowStockThreshold,
	); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (r *Repository) CreateMovement(
	ctx context.Context,
	q db.Queryer,
	params CreateMovementParams,
) (*StockMovement, error) {
	const query = `
		INSERT INTO stock_movements (
			tenant_id,
			store_id,
			product_id,
			movement_type,
			quantity,
			balance_after,
			reference_type,
			reference_id,
			note,
			created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, NULLIF($9, ''), $10)
		RETURNING
			id,
			tenant_id,
			store_id,
			product_id,
			movement_type,
			quantity,
			COALESCE(balance_after, 0),
			COALESCE(reference_type, ''),
			reference_id,
			COALESCE(note, ''),
			created_by
	`

	var movement StockMovement
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.MovementType,
		params.Quantity,
		params.BalanceAfter,
		params.ReferenceType,
		params.ReferenceID,
		params.Note,
		params.CreatedBy,
	).Scan(
		&movement.ID,
		&movement.TenantID,
		&movement.StoreID,
		&movement.ProductID,
		&movement.MovementType,
		&movement.Quantity,
		&movement.BalanceAfter,
		&movement.ReferenceType,
		&movement.ReferenceID,
		&movement.Note,
		&movement.CreatedBy,
	); err != nil {
		return nil, err
	}

	return &movement, nil
}
