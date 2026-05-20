package pos

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) FindCurrentOpenByCashier(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	cashierID uuid.UUID,
) (*CashierSession, error) {
	const query = selectSessionSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND cashier_id = $3
		  AND status = 'open'
		ORDER BY opened_at DESC
		LIMIT 1
	`

	return scanSession(q.QueryRow(ctx, query, tenantID, storeID, cashierID))
}

func (r *Repository) CreateSession(
	ctx context.Context,
	q db.Queryer,
	params CreateSessionParams,
) (*CashierSession, error) {
	const query = `
		INSERT INTO cashier_sessions (
			tenant_id,
			store_id,
			cashier_id,
			session_number,
			opening_cash,
			status
		)
		VALUES ($1, $2, $3, $4, $5, 'open')
		RETURNING
			id,
			tenant_id,
			store_id,
			cashier_id,
			session_number,
			opening_cash,
			closing_cash,
			expected_cash,
			difference,
			status,
			opened_at,
			closed_at,
			created_at,
			updated_at
	`

	session, err := scanSession(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.CashierID,
		params.SessionNumber,
		params.OpeningCash,
	))
	if err != nil && isUniqueViolation(err) {
		return nil, ErrOpenSessionExists
	}
	return session, err
}

func (r *Repository) LockSessionByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	sessionID uuid.UUID,
) (*CashierSession, error) {
	const query = selectSessionSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		FOR UPDATE
	`

	return scanSession(q.QueryRow(ctx, query, tenantID, storeID, sessionID))
}

func (r *Repository) SumCompletedCashTransactions(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	sessionID uuid.UUID,
) (int64, error) {
	const query = `
		SELECT COALESCE(SUM(payment_amount - change_amount), 0)::bigint
		FROM pos_transactions
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND cashier_session_id = $3
		  AND payment_method = 'cash'
		  AND status = 'completed'
	`

	var total int64
	if err := q.QueryRow(ctx, query, tenantID, storeID, sessionID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repository) CloseSession(
	ctx context.Context,
	q db.Queryer,
	params CloseSessionParams,
) (*CashierSession, error) {
	const query = `
		UPDATE cashier_sessions
		SET closing_cash = $4,
		    expected_cash = $5,
		    difference = $6,
		    status = 'closed',
		    closed_at = now(),
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		  AND status = 'open'
		RETURNING
			id,
			tenant_id,
			store_id,
			cashier_id,
			session_number,
			opening_cash,
			closing_cash,
			expected_cash,
			difference,
			status,
			opened_at,
			closed_at,
			created_at,
			updated_at
	`

	session, err := scanSession(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.SessionID,
		params.ClosingCash,
		params.ExpectedCash,
		params.Difference,
	))
	if err != nil && errors.Is(err, ErrSessionNotFound) {
		return nil, ErrSessionAlreadyDone
	}
	return session, err
}

func (r *Repository) ListProducts(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters ProductSearchFilters,
) ([]POSProduct, error) {
	args := []any{tenantID, storeID}
	conditions := []string{
		"p.tenant_id = $1",
		"p.store_id = $2",
		"p.status = 'active'",
		"p.deleted_at IS NULL",
	}
	if filters.Query != "" {
		args = append(args, filters.Query)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf(`(
			p.name ILIKE '%%' || $%d || '%%'
			OR COALESCE(p.sku, '') ILIKE '%%' || $%d || '%%'
		)`, placeholder, placeholder))
	}
	if filters.Barcode != "" {
		args = append(args, filters.Barcode)
		conditions = append(conditions, fmt.Sprintf("COALESCE(p.barcode, '') = $%d", len(args)))
	}
	args = append(args, filters.Limit)

	query := fmt.Sprintf(selectPOSProductSQL+`
		WHERE %s
		ORDER BY p.name ASC, p.id ASC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (r *Repository) ListProductsByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productIDs []uuid.UUID,
) ([]POSProduct, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	const query = selectPOSProductSQL + `
		WHERE p.tenant_id = $1
		  AND p.store_id = $2
		  AND p.id = ANY($3::uuid[])
		  AND p.status = 'active'
		  AND p.deleted_at IS NULL
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanProducts(rows)
}

func (r *Repository) LockStockSnapshots(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	productIDs []uuid.UUID,
) ([]StockSnapshot, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			product_id,
			quantity_on_hand,
			quantity_reserved,
			quantity_available
		FROM product_stock_snapshots
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND product_id = ANY($3::uuid[])
		ORDER BY product_id ASC
		FOR UPDATE
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, productIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]StockSnapshot, 0, len(productIDs))
	for rows.Next() {
		var item StockSnapshot
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.StoreID,
			&item.ProductID,
			&item.QuantityOnHand,
			&item.QuantityReserved,
			&item.QuantityAvailable,
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

func (r *Repository) CreateTransaction(ctx context.Context, q db.Queryer, params CreateTransactionParams) (*POSTransaction, error) {
	const query = `
		INSERT INTO pos_transactions (
			tenant_id,
			store_id,
			cashier_session_id,
			cashier_id,
			transaction_number,
			subtotal,
			discount_total,
			tax_total,
			grand_total,
			payment_method,
			payment_amount,
			change_amount,
			status
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 'completed')
		RETURNING
			id,
			tenant_id,
			store_id,
			cashier_session_id,
			cashier_id,
			transaction_number,
			subtotal,
			discount_total,
			tax_total,
			grand_total,
			payment_method,
			payment_amount,
			change_amount,
			status,
			created_at,
			updated_at
	`

	return scanTransaction(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.CashierSessionID,
		params.CashierID,
		params.TransactionNumber,
		params.Subtotal,
		params.DiscountTotal,
		params.TaxTotal,
		params.GrandTotal,
		params.PaymentMethod,
		params.PaymentAmount,
		params.ChangeAmount,
	))
}

func (r *Repository) CreateTransactionItem(ctx context.Context, q db.Queryer, params CreateTransactionItemParams) error {
	const query = `
		INSERT INTO pos_transaction_items (
			tenant_id,
			pos_transaction_id,
			product_id,
			product_name,
			sku,
			quantity,
			unit_price,
			discount_total,
			subtotal
		)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8, $9)
	`

	_, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.POSTransactionID,
		params.ProductID,
		params.ProductName,
		params.SKU,
		params.Quantity,
		params.UnitPrice,
		params.DiscountTotal,
		params.Subtotal,
	)
	return err
}

func (r *Repository) UpdateStockSnapshot(ctx context.Context, q db.Queryer, params UpdateStockSnapshotParams) error {
	const query = `
		UPDATE product_stock_snapshots
		SET quantity_on_hand = $4,
		    quantity_reserved = $5,
		    quantity_available = $6,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND product_id = $3
	`

	tag, err := q.Exec(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ProductID,
		params.QuantityOnHand,
		params.QuantityReserved,
		params.QuantityAvailable,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrStockSnapshotNotFound
	}
	return nil
}

func (r *Repository) CreateStockMovement(ctx context.Context, q db.Queryer, params CreateStockMovementParams) error {
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
	`

	_, err := q.Exec(
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
	)
	return err
}

func (r *Repository) ListTransactions(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters TransactionListFilters,
) ([]POSTransaction, error) {
	args := []any{tenantID, storeID}
	conditions := []string{
		"tenant_id = $1",
		"store_id = $2",
	}
	if filters.DateFrom != nil {
		args = append(args, *filters.DateFrom)
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", len(args)))
	}
	if filters.DateTo != nil {
		args = append(args, *filters.DateTo)
		conditions = append(conditions, fmt.Sprintf("created_at < $%d", len(args)))
	}
	if filters.PaymentMethod != nil {
		args = append(args, *filters.PaymentMethod)
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", len(args)))
	}
	if filters.CashierID != nil {
		args = append(args, *filters.CashierID)
		conditions = append(conditions, fmt.Sprintf("cashier_id = $%d", len(args)))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(conditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", len(args)-1, len(args)))
	}
	args = append(args, filters.Limit)

	query := fmt.Sprintf(selectTransactionSQL+`
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]POSTransaction, 0)
	for rows.Next() {
		item, err := scanTransaction(rows)
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

func (r *Repository) FindTransactionByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	transactionID uuid.UUID,
) (*POSTransaction, error) {
	const query = selectTransactionSQL + `
		WHERE tenant_id = $1
		  AND store_id = $2
		  AND id = $3
		LIMIT 1
	`

	return scanTransaction(q.QueryRow(ctx, query, tenantID, storeID, transactionID))
}

func (r *Repository) ListTransactionItems(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	transactionID uuid.UUID,
) ([]POSTransactionItem, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			pos_transaction_id,
			product_id,
			product_name,
			COALESCE(sku, ''),
			quantity,
			unit_price,
			discount_total,
			subtotal,
			created_at
		FROM pos_transaction_items
		WHERE tenant_id = $1
		  AND pos_transaction_id = $2
		ORDER BY created_at ASC, id ASC
	`

	rows, err := q.Query(ctx, query, tenantID, transactionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]POSTransactionItem, 0)
	for rows.Next() {
		var item POSTransactionItem
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.POSTransactionID,
			&item.ProductID,
			&item.ProductName,
			&item.SKU,
			&item.Quantity,
			&item.UnitPrice,
			&item.DiscountTotal,
			&item.Subtotal,
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

const selectSessionSQL = `
	SELECT
		id,
		tenant_id,
		store_id,
		cashier_id,
		session_number,
		opening_cash,
		closing_cash,
		expected_cash,
		difference,
		status,
		opened_at,
		closed_at,
		created_at,
		updated_at
	FROM cashier_sessions
`

const selectPOSProductSQL = `
	SELECT
		p.id,
		p.tenant_id,
		p.store_id,
		p.category_id,
		COALESCE(c.name, ''),
		p.name,
		COALESCE(p.sku, ''),
		COALESCE(p.barcode, ''),
		p.price,
		p.status,
		COALESCE(img.url, ''),
		COALESCE(s.quantity_available, 0)
	FROM products p
	LEFT JOIN categories c
	  ON c.id = p.category_id
	 AND c.tenant_id = p.tenant_id
	 AND c.store_id = p.store_id
	LEFT JOIN product_stock_snapshots s
	  ON s.tenant_id = p.tenant_id
	 AND s.store_id = p.store_id
	 AND s.product_id = p.id
	LEFT JOIN LATERAL (
		SELECT pi.url
		FROM product_images pi
		WHERE pi.tenant_id = p.tenant_id
		  AND pi.product_id = p.id
		ORDER BY pi.is_primary DESC, pi.sort_order ASC, pi.created_at ASC
		LIMIT 1
	) img ON true
`

const selectTransactionSQL = `
	SELECT
		id,
		tenant_id,
		store_id,
		cashier_session_id,
		cashier_id,
		transaction_number,
		subtotal,
		discount_total,
		tax_total,
		grand_total,
		payment_method,
		payment_amount,
		change_amount,
		status,
		created_at,
		updated_at
	FROM pos_transactions
`

type productRows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}

func scanProducts(rows productRows) ([]POSProduct, error) {
	items := make([]POSProduct, 0)
	for rows.Next() {
		var item POSProduct
		if err := rows.Scan(
			&item.ID,
			&item.TenantID,
			&item.StoreID,
			&item.CategoryID,
			&item.CategoryName,
			&item.Name,
			&item.SKU,
			&item.Barcode,
			&item.Price,
			&item.Status,
			&item.PrimaryImageURL,
			&item.StockAvailable,
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

func scanSession(row pgx.Row) (*CashierSession, error) {
	var session CashierSession
	if err := row.Scan(
		&session.ID,
		&session.TenantID,
		&session.StoreID,
		&session.CashierID,
		&session.SessionNumber,
		&session.OpeningCash,
		&session.ClosingCash,
		&session.ExpectedCash,
		&session.Difference,
		&session.Status,
		&session.OpenedAt,
		&session.ClosedAt,
		&session.CreatedAt,
		&session.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

func scanTransaction(row pgx.Row) (*POSTransaction, error) {
	var transaction POSTransaction
	if err := row.Scan(
		&transaction.ID,
		&transaction.TenantID,
		&transaction.StoreID,
		&transaction.CashierSessionID,
		&transaction.CashierID,
		&transaction.TransactionNumber,
		&transaction.Subtotal,
		&transaction.DiscountTotal,
		&transaction.TaxTotal,
		&transaction.GrandTotal,
		&transaction.PaymentMethod,
		&transaction.PaymentAmount,
		&transaction.ChangeAmount,
		&transaction.Status,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, err
	}
	return &transaction, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
