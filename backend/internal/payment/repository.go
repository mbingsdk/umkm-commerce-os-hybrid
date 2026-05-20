package payment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/order"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	ErrOrderNotFound        = errors.New("payment order not found")
	ErrConfirmationNotFound = errors.New("payment confirmation not found")
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) FindOrderByPublicReference(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderNumber string,
	customerPhone string,
) (*order.Order, error) {
	const query = selectOrderSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND order_number = $3
		  AND customer_phone = $4
		LIMIT 1
	`

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderNumber, customerPhone))
}

func (r *Repository) LockOrderByID(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*order.Order, error) {
	const query = selectOrderSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		FOR UPDATE
	`

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderID))
}

func (r *Repository) CreateConfirmation(ctx context.Context, q db.Queryer, params CreateConfirmationParams) (*Confirmation, error) {
	const query = `
		INSERT INTO payment_confirmations (
			tenant_id,
			store_id,
			order_id,
			payer_name,
			bank_name,
			transfer_amount,
			transfer_date,
			proof_url,
			note
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''))
		RETURNING
			id,
			tenant_id,
			store_id,
			order_id,
			payer_name,
			bank_name,
			transfer_amount,
			transfer_date,
			COALESCE(proof_url, ''),
			COALESCE(note, ''),
			status,
			reviewed_by,
			reviewed_at,
			COALESCE(review_note, ''),
			created_at,
			updated_at
	`

	return scanConfirmation(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.OrderID,
		params.PayerName,
		params.BankName,
		params.TransferAmount,
		params.TransferDate,
		params.ProofURL,
		params.Note,
	))
}

func (r *Repository) ListConfirmations(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) ([]Confirmation, error) {
	const query = `
		SELECT
			pc.id,
			pc.tenant_id,
			pc.store_id,
			pc.order_id,
			pc.payer_name,
			pc.bank_name,
			pc.transfer_amount,
			pc.transfer_date,
			COALESCE(pc.proof_url, ''),
			COALESCE(pc.note, ''),
			pc.status,
			pc.reviewed_by,
			pc.reviewed_at,
			COALESCE(pc.review_note, ''),
			pc.created_at,
			pc.updated_at
		FROM payment_confirmations pc
		JOIN orders o
		  ON o.id = pc.order_id
		 AND o.tenant_id = pc.tenant_id
		 AND o.store_id = pc.store_id
		WHERE pc.tenant_id = $1
		  AND pc.store_id = $2
		  AND pc.order_id = $3
		ORDER BY pc.created_at DESC, pc.id DESC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Confirmation, 0)
	for rows.Next() {
		item, err := scanConfirmation(rows)
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

func (r *Repository) FindPendingConfirmation(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	orderID uuid.UUID,
	confirmationID *uuid.UUID,
) (*Confirmation, error) {
	const query = `
		SELECT
			pc.id,
			pc.tenant_id,
			pc.store_id,
			pc.order_id,
			pc.payer_name,
			pc.bank_name,
			pc.transfer_amount,
			pc.transfer_date,
			COALESCE(pc.proof_url, ''),
			COALESCE(pc.note, ''),
			pc.status,
			pc.reviewed_by,
			pc.reviewed_at,
			COALESCE(pc.review_note, ''),
			pc.created_at,
			pc.updated_at
		FROM payment_confirmations pc
		WHERE pc.tenant_id = $1
		  AND pc.store_id = $2
		  AND pc.order_id = $3
		  AND pc.status = 'pending'
		  AND ($4::uuid IS NULL OR pc.id = $4)
		ORDER BY pc.created_at DESC, pc.id DESC
		LIMIT 1
		FOR UPDATE
	`

	return scanConfirmation(q.QueryRow(ctx, query, tenantID, storeID, orderID, confirmationID))
}

func (r *Repository) MarkConfirmationReviewed(ctx context.Context, q db.Queryer, params ReviewConfirmationParams) (*Confirmation, error) {
	const query = `
		UPDATE payment_confirmations
		SET status = $5,
		    reviewed_by = $6,
		    reviewed_at = now(),
		    review_note = NULLIF($7, ''),
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND order_id = $3
		  AND id = $4
		  AND status = 'pending'
		RETURNING
			id,
			tenant_id,
			store_id,
			order_id,
			payer_name,
			bank_name,
			transfer_amount,
			transfer_date,
			COALESCE(proof_url, ''),
			COALESCE(note, ''),
			status,
			reviewed_by,
			reviewed_at,
			COALESCE(review_note, ''),
			created_at,
			updated_at
	`

	return scanConfirmation(q.QueryRow(ctx, query, params.TenantID, params.StoreID, params.OrderID, params.ConfirmationID, params.Status, params.ReviewedBy, params.ReviewNote))
}

func (r *Repository) UpdateOrderPaymentWaiting(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) error {
	const query = `
		UPDATE orders
		SET payment_status = CASE WHEN payment_status = 'unpaid' THEN 'waiting_confirmation' ELSE payment_status END,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
	`

	_, err := q.Exec(ctx, query, tenantID, storeID, orderID)
	return err
}

func (r *Repository) UpdateOrderPaid(ctx context.Context, q db.Queryer, tenantID uuid.UUID, storeID uuid.UUID, orderID uuid.UUID) (*order.Order, error) {
	const query = `
		UPDATE orders
		SET payment_status = 'paid',
		    paid_at = COALESCE(paid_at, now()),
		    status = CASE WHEN status = 'pending' THEN 'confirmed' ELSE status END,
		    confirmed_at = CASE WHEN status = 'pending' THEN COALESCE(confirmed_at, now()) ELSE confirmed_at END,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		RETURNING
			id,
			tenant_id,
			store_id,
			customer_id,
			order_number,
			source,
			status,
			payment_status,
			COALESCE(shipment_status, ''),
			subtotal,
			discount_total,
			shipping_cost,
			tax_total,
			grand_total,
			customer_name,
			customer_phone,
			COALESCE(customer_email, ''),
			COALESCE(shipping_address, ''),
			COALESCE(shipping_city, ''),
			COALESCE(shipping_province, ''),
			COALESCE(shipping_postal_code, ''),
			COALESCE(customer_note, ''),
			COALESCE(internal_note, ''),
			confirmed_at,
			paid_at,
			completed_at,
			cancelled_at,
			created_at,
			updated_at
	`

	return scanOrder(q.QueryRow(ctx, query, tenantID, storeID, orderID))
}

func (r *Repository) CreatePayment(ctx context.Context, q db.Queryer, params CreatePaymentParams) (*Payment, error) {
	const query = `
		INSERT INTO payments (
			tenant_id,
			store_id,
			order_id,
			payment_confirmation_id,
			method,
			status,
			amount,
			payer_name,
			bank_name,
			proof_url,
			note,
			paid_at,
			confirmed_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NULLIF($9, ''), NULLIF($10, ''), NULLIF($11, ''), now(), $12)
		RETURNING
			id,
			tenant_id,
			store_id,
			order_id,
			payment_confirmation_id,
			method,
			status,
			amount,
			COALESCE(payer_name, ''),
			COALESCE(bank_name, ''),
			COALESCE(proof_url, ''),
			COALESCE(note, ''),
			paid_at,
			confirmed_by,
			created_at,
			updated_at
	`

	var payment Payment
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.OrderID,
		params.PaymentConfirmationID,
		params.Method,
		params.Status,
		params.Amount,
		params.PayerName,
		params.BankName,
		params.ProofURL,
		params.Note,
		params.ConfirmedBy,
	).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.StoreID,
		&payment.OrderID,
		&payment.PaymentConfirmationID,
		&payment.Method,
		&payment.Status,
		&payment.Amount,
		&payment.PayerName,
		&payment.BankName,
		&payment.ProofURL,
		&payment.Note,
		&payment.PaidAt,
		&payment.ConfirmedBy,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *Repository) CreateOrderStatusLog(ctx context.Context, q db.Queryer, tenantID uuid.UUID, orderID uuid.UUID, fromStatus string, toStatus string, note string, actorID uuid.UUID) error {
	const query = `
		INSERT INTO order_status_logs (tenant_id, order_id, from_status, to_status, note, created_by)
		VALUES ($1, $2, NULLIF($3, ''), $4, NULLIF($5, ''), $6)
	`

	_, err := q.Exec(ctx, query, tenantID, orderID, fromStatus, toStatus, note, actorID)
	return err
}

const selectOrderSQL = `
	SELECT
		id,
		tenant_id,
		store_id,
		customer_id,
		order_number,
		source,
		status,
		payment_status,
		COALESCE(shipment_status, ''),
		subtotal,
		discount_total,
		shipping_cost,
		tax_total,
		grand_total,
		customer_name,
		customer_phone,
		COALESCE(customer_email, ''),
		COALESCE(shipping_address, ''),
		COALESCE(shipping_city, ''),
		COALESCE(shipping_province, ''),
		COALESCE(shipping_postal_code, ''),
		COALESCE(customer_note, ''),
		COALESCE(internal_note, ''),
		confirmed_at,
		paid_at,
		completed_at,
		cancelled_at,
		created_at,
		updated_at
	FROM orders
`

func scanOrder(row pgx.Row) (*order.Order, error) {
	var item order.Order
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.StoreID,
		&item.CustomerID,
		&item.OrderNumber,
		&item.Source,
		&item.Status,
		&item.PaymentStatus,
		&item.ShipmentStatus,
		&item.Subtotal,
		&item.DiscountTotal,
		&item.ShippingCost,
		&item.TaxTotal,
		&item.GrandTotal,
		&item.CustomerName,
		&item.CustomerPhone,
		&item.CustomerEmail,
		&item.ShippingAddress,
		&item.ShippingCity,
		&item.ShippingProvince,
		&item.ShippingPostalCode,
		&item.CustomerNote,
		&item.InternalNote,
		&item.ConfirmedAt,
		&item.PaidAt,
		&item.CompletedAt,
		&item.CancelledAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return &item, nil
}

func scanConfirmation(row pgx.Row) (*Confirmation, error) {
	var item Confirmation
	if err := row.Scan(
		&item.ID,
		&item.TenantID,
		&item.StoreID,
		&item.OrderID,
		&item.PayerName,
		&item.BankName,
		&item.TransferAmount,
		&item.TransferDate,
		&item.ProofURL,
		&item.Note,
		&item.Status,
		&item.ReviewedBy,
		&item.ReviewedAt,
		&item.ReviewNote,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrConfirmationNotFound
		}
		return nil, err
	}
	return &item, nil
}
