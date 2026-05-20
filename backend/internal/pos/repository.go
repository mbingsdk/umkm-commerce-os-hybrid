package pos

import (
	"context"
	"errors"

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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
