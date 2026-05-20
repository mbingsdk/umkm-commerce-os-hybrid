package finance

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Summary(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	dateRange DateRange,
) (FinanceTotals, error) {
	const query = `
		WITH online AS (
			SELECT
				COALESCE(SUM(grand_total), 0)::bigint AS total,
				COUNT(*)::bigint AS count
			FROM orders
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND source <> 'pos'
			  AND payment_status = 'paid'
			  AND status NOT IN ('cancelled', 'returned', 'refunded')
			  AND COALESCE(paid_at, updated_at, created_at) >= $3
			  AND COALESCE(paid_at, updated_at, created_at) < $4
		),
		pos AS (
			SELECT
				COALESCE(SUM(grand_total), 0)::bigint AS total,
				COUNT(*)::bigint AS count
			FROM pos_transactions
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND status = 'completed'
			  AND created_at >= $3
			  AND created_at < $4
		),
		expense AS (
			SELECT COALESCE(SUM(amount), 0)::bigint AS total
			FROM expenses
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND deleted_at IS NULL
			  AND expense_date >= $3::date
			  AND expense_date < $4::date
		)
		SELECT
			online.total,
			pos.total,
			expense.total,
			online.count,
			pos.count
		FROM online, pos, expense
	`

	return scanFinanceTotals(q.QueryRow(ctx, query, tenantID, storeID, dateRange.From, dateRange.To))
}

func (r *Repository) DailyReport(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	dateRange DateRange,
) ([]DailyFinanceRow, error) {
	const query = `
		WITH days AS (
			SELECT generate_series($3::date, ($4::date - INTERVAL '1 day'), INTERVAL '1 day')::date AS day
		),
		online AS (
			SELECT
				COALESCE(paid_at, updated_at, created_at)::date AS day,
				COALESCE(SUM(grand_total), 0)::bigint AS total,
				COUNT(*)::bigint AS count
			FROM orders
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND source <> 'pos'
			  AND payment_status = 'paid'
			  AND status NOT IN ('cancelled', 'returned', 'refunded')
			  AND COALESCE(paid_at, updated_at, created_at) >= $3
			  AND COALESCE(paid_at, updated_at, created_at) < $4
			GROUP BY COALESCE(paid_at, updated_at, created_at)::date
		),
		pos AS (
			SELECT
				created_at::date AS day,
				COALESCE(SUM(grand_total), 0)::bigint AS total,
				COUNT(*)::bigint AS count
			FROM pos_transactions
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND status = 'completed'
			  AND created_at >= $3
			  AND created_at < $4
			GROUP BY created_at::date
		),
		expense AS (
			SELECT
				expense_date AS day,
				COALESCE(SUM(amount), 0)::bigint AS total
			FROM expenses
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND deleted_at IS NULL
			  AND expense_date >= $3::date
			  AND expense_date < $4::date
			GROUP BY expense_date
		)
		SELECT
			days.day,
			COALESCE(online.total, 0)::bigint,
			COALESCE(pos.total, 0)::bigint,
			COALESCE(expense.total, 0)::bigint,
			COALESCE(online.count, 0)::bigint,
			COALESCE(pos.count, 0)::bigint
		FROM days
		LEFT JOIN online ON online.day = days.day
		LEFT JOIN pos ON pos.day = days.day
		LEFT JOIN expense ON expense.day = days.day
		ORDER BY days.day ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, dateRange.From, dateRange.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]DailyFinanceRow, 0)
	for rows.Next() {
		var item DailyFinanceRow
		if err := rows.Scan(
			&item.Date,
			&item.OnlineSales,
			&item.POSSales,
			&item.TotalExpenses,
			&item.OrderCount,
			&item.POSTransactionCount,
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

func (r *Repository) MonthlyReport(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	dateRange DateRange,
) ([]MonthlyFinanceRow, error) {
	const query = `
		WITH months AS (
			SELECT generate_series(
				date_trunc('month', $3::date),
				date_trunc('month', ($4::date - INTERVAL '1 day')),
				INTERVAL '1 month'
			)::date AS month_start
		),
		online AS (
			SELECT
				date_trunc('month', COALESCE(paid_at, updated_at, created_at))::date AS month_start,
				COALESCE(SUM(grand_total), 0)::bigint AS total,
				COUNT(*)::bigint AS count
			FROM orders
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND source <> 'pos'
			  AND payment_status = 'paid'
			  AND status NOT IN ('cancelled', 'returned', 'refunded')
			  AND COALESCE(paid_at, updated_at, created_at) >= $3
			  AND COALESCE(paid_at, updated_at, created_at) < $4
			GROUP BY date_trunc('month', COALESCE(paid_at, updated_at, created_at))::date
		),
		pos AS (
			SELECT
				date_trunc('month', created_at)::date AS month_start,
				COALESCE(SUM(grand_total), 0)::bigint AS total,
				COUNT(*)::bigint AS count
			FROM pos_transactions
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND status = 'completed'
			  AND created_at >= $3
			  AND created_at < $4
			GROUP BY date_trunc('month', created_at)::date
		),
		expense AS (
			SELECT
				date_trunc('month', expense_date)::date AS month_start,
				COALESCE(SUM(amount), 0)::bigint AS total
			FROM expenses
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND deleted_at IS NULL
			  AND expense_date >= $3::date
			  AND expense_date < $4::date
			GROUP BY date_trunc('month', expense_date)::date
		)
		SELECT
			months.month_start,
			COALESCE(online.total, 0)::bigint,
			COALESCE(pos.total, 0)::bigint,
			COALESCE(expense.total, 0)::bigint,
			COALESCE(online.count, 0)::bigint,
			COALESCE(pos.count, 0)::bigint
		FROM months
		LEFT JOIN online ON online.month_start = months.month_start
		LEFT JOIN pos ON pos.month_start = months.month_start
		LEFT JOIN expense ON expense.month_start = months.month_start
		ORDER BY months.month_start ASC
	`

	rows, err := q.Query(ctx, query, tenantID, storeID, dateRange.From, dateRange.To)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]MonthlyFinanceRow, 0)
	for rows.Next() {
		var item MonthlyFinanceRow
		if err := rows.Scan(
			&item.MonthStart,
			&item.OnlineSales,
			&item.POSSales,
			&item.TotalExpenses,
			&item.OrderCount,
			&item.POSTransactionCount,
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

func (r *Repository) ListExpenses(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	filters ListExpenseFilters,
) ([]Expense, error) {
	args := []any{tenantID, storeID}
	conditions := []string{
		"e.tenant_id = $1",
		"e.store_id = $2",
		"e.deleted_at IS NULL",
	}

	if filters.DateFrom != nil {
		args = append(args, *filters.DateFrom)
		conditions = append(conditions, fmt.Sprintf("e.expense_date >= $%d::date", len(args)))
	}
	if filters.DateTo != nil {
		args = append(args, *filters.DateTo)
		conditions = append(conditions, fmt.Sprintf("e.expense_date < $%d::date", len(args)))
	}
	if filters.CategoryID != nil {
		args = append(args, *filters.CategoryID)
		conditions = append(conditions, fmt.Sprintf("e.category_id = $%d", len(args)))
	}
	if filters.CategorySlug != "" {
		args = append(args, filters.CategorySlug)
		conditions = append(conditions, fmt.Sprintf("c.slug = $%d", len(args)))
	}
	if filters.Query != "" {
		args = append(args, filters.Query)
		placeholder := len(args)
		conditions = append(conditions, fmt.Sprintf(`(
			e.title ILIKE '%%' || $%d || '%%'
			OR COALESCE(e.note, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(c.name, '') ILIKE '%%' || $%d || '%%'
			OR COALESCE(c.slug, '') ILIKE '%%' || $%d || '%%'
		)`, placeholder, placeholder, placeholder, placeholder))
	}
	if filters.Cursor != nil {
		args = append(args, filters.Cursor.ExpenseDate, filters.Cursor.CreatedAt, filters.Cursor.ID)
		conditions = append(
			conditions,
			fmt.Sprintf("(e.expense_date, e.created_at, e.id) < ($%d::date, $%d, $%d)", len(args)-2, len(args)-1, len(args)),
		)
	}

	args = append(args, filters.Limit)
	query := fmt.Sprintf(`
		SELECT
			e.id,
			e.tenant_id,
			e.store_id,
			e.category_id,
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			e.title,
			e.amount,
			e.expense_date,
			COALESCE(e.payment_method, ''),
			COALESCE(e.note, ''),
			e.created_by,
			COALESCE(u.name, ''),
			e.updated_by,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM expenses e
		LEFT JOIN expense_categories c
		  ON c.id = e.category_id
		 AND c.deleted_at IS NULL
		 AND (
			(c.tenant_id IS NULL AND c.store_id IS NULL)
			OR (c.tenant_id = e.tenant_id AND (c.store_id IS NULL OR c.store_id = e.store_id))
		 )
		LEFT JOIN users u
		  ON u.id = e.created_by
		WHERE %s
		ORDER BY e.expense_date DESC, e.created_at DESC, e.id DESC
		LIMIT $%d
	`, strings.Join(conditions, "\n\t\t  AND "), len(args))

	rows, err := q.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Expense, 0)
	for rows.Next() {
		item, err := scanExpense(rows)
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

func (r *Repository) FindExpenseByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	expenseID uuid.UUID,
) (*Expense, error) {
	const query = `
		SELECT
			e.id,
			e.tenant_id,
			e.store_id,
			e.category_id,
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			e.title,
			e.amount,
			e.expense_date,
			COALESCE(e.payment_method, ''),
			COALESCE(e.note, ''),
			e.created_by,
			COALESCE(u.name, ''),
			e.updated_by,
			e.created_at,
			e.updated_at,
			e.deleted_at
		FROM expenses e
		LEFT JOIN expense_categories c
		  ON c.id = e.category_id
		 AND c.deleted_at IS NULL
		 AND (
			(c.tenant_id IS NULL AND c.store_id IS NULL)
			OR (c.tenant_id = e.tenant_id AND (c.store_id IS NULL OR c.store_id = e.store_id))
		 )
		LEFT JOIN users u
		  ON u.id = e.created_by
		WHERE e.tenant_id = $1
		  AND e.store_id = $2
		  AND e.id = $3
		  AND e.deleted_at IS NULL
		LIMIT 1
	`

	return scanExpense(q.QueryRow(ctx, query, tenantID, storeID, expenseID))
}

func (r *Repository) FindCategoryByID(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	categoryID uuid.UUID,
) (*ExpenseCategory, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			name,
			slug,
			is_system,
			created_at,
			updated_at
		FROM expense_categories
		WHERE id = $3
		  AND deleted_at IS NULL
		  AND (
			(tenant_id IS NULL AND store_id IS NULL)
			OR (tenant_id = $1 AND (store_id IS NULL OR store_id = $2))
		  )
		LIMIT 1
	`

	return scanCategory(q.QueryRow(ctx, query, tenantID, storeID, categoryID))
}

func (r *Repository) FindCategoryBySlug(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	slug string,
) (*ExpenseCategory, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			store_id,
			name,
			slug,
			is_system,
			created_at,
			updated_at
		FROM expense_categories
		WHERE slug = $3
		  AND deleted_at IS NULL
		  AND (
			(tenant_id IS NULL AND store_id IS NULL)
			OR (tenant_id = $1 AND (store_id IS NULL OR store_id = $2))
		  )
		ORDER BY is_system ASC, created_at DESC
		LIMIT 1
	`

	return scanCategory(q.QueryRow(ctx, query, tenantID, storeID, slug))
}

func (r *Repository) CreateExpense(ctx context.Context, q db.Queryer, params CreateExpenseParams) (*Expense, error) {
	const query = `
		WITH inserted AS (
			INSERT INTO expenses (
				tenant_id,
				store_id,
				category_id,
				title,
				amount,
				expense_date,
				payment_method,
				note,
				created_by,
				updated_by
			)
			VALUES ($1, $2, $3, $4, $5, $6::date, NULLIF($7, ''), NULLIF($8, ''), $9, $9)
			RETURNING *
		)
		SELECT
			i.id,
			i.tenant_id,
			i.store_id,
			i.category_id,
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			i.title,
			i.amount,
			i.expense_date,
			COALESCE(i.payment_method, ''),
			COALESCE(i.note, ''),
			i.created_by,
			COALESCE(u.name, ''),
			i.updated_by,
			i.created_at,
			i.updated_at,
			i.deleted_at
		FROM inserted i
		LEFT JOIN expense_categories c ON c.id = i.category_id
		LEFT JOIN users u ON u.id = i.created_by
	`

	return scanExpense(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.CategoryID,
		params.Title,
		params.Amount,
		params.ExpenseDate,
		params.PaymentMethod,
		params.Note,
		params.CreatedBy,
	))
}

func (r *Repository) UpdateExpense(ctx context.Context, q db.Queryer, params UpdateExpenseParams) (*Expense, error) {
	const query = `
		WITH updated AS (
			UPDATE expenses
			SET category_id = $4,
			    title = $5,
			    amount = $6,
			    expense_date = $7::date,
			    payment_method = NULLIF($8, ''),
			    note = NULLIF($9, ''),
			    updated_by = $10,
			    updated_at = now()
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND id = $3
			  AND deleted_at IS NULL
			RETURNING *
		)
		SELECT
			u.id,
			u.tenant_id,
			u.store_id,
			u.category_id,
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			u.title,
			u.amount,
			u.expense_date,
			COALESCE(u.payment_method, ''),
			COALESCE(u.note, ''),
			u.created_by,
			COALESCE(creator.name, ''),
			u.updated_by,
			u.created_at,
			u.updated_at,
			u.deleted_at
		FROM updated u
		LEFT JOIN expense_categories c ON c.id = u.category_id
		LEFT JOIN users creator ON creator.id = u.created_by
	`

	return scanExpense(q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.StoreID,
		params.ExpenseID,
		params.CategoryID,
		params.Title,
		params.Amount,
		params.ExpenseDate,
		params.PaymentMethod,
		params.Note,
		params.UpdatedBy,
	))
}

func (r *Repository) SoftDeleteExpense(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	expenseID uuid.UUID,
	actorUserID uuid.UUID,
) (*Expense, error) {
	const query = `
		WITH deleted AS (
			UPDATE expenses
			SET deleted_at = now(),
			    updated_by = $4,
			    updated_at = now()
			WHERE tenant_id = $1
			  AND store_id = $2
			  AND id = $3
			  AND deleted_at IS NULL
			RETURNING *
		)
		SELECT
			d.id,
			d.tenant_id,
			d.store_id,
			d.category_id,
			COALESCE(c.name, ''),
			COALESCE(c.slug, ''),
			d.title,
			d.amount,
			d.expense_date,
			COALESCE(d.payment_method, ''),
			COALESCE(d.note, ''),
			d.created_by,
			COALESCE(creator.name, ''),
			d.updated_by,
			d.created_at,
			d.updated_at,
			d.deleted_at
		FROM deleted d
		LEFT JOIN expense_categories c ON c.id = d.category_id
		LEFT JOIN users creator ON creator.id = d.created_by
	`

	return scanExpense(q.QueryRow(ctx, query, tenantID, storeID, expenseID, actorUserID))
}

func scanFinanceTotals(row pgx.Row) (FinanceTotals, error) {
	var totals FinanceTotals
	if err := row.Scan(
		&totals.OnlineSales,
		&totals.POSSales,
		&totals.TotalExpenses,
		&totals.OrderCount,
		&totals.POSTransactionCount,
	); err != nil {
		return FinanceTotals{}, err
	}
	return totals, nil
}

func scanExpense(row pgx.Row) (*Expense, error) {
	var expense Expense
	if err := row.Scan(
		&expense.ID,
		&expense.TenantID,
		&expense.StoreID,
		&expense.CategoryID,
		&expense.CategoryName,
		&expense.CategorySlug,
		&expense.Title,
		&expense.Amount,
		&expense.ExpenseDate,
		&expense.PaymentMethod,
		&expense.Note,
		&expense.CreatedBy,
		&expense.CreatedByName,
		&expense.UpdatedBy,
		&expense.CreatedAt,
		&expense.UpdatedAt,
		&expense.DeletedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}
	return &expense, nil
}

func scanCategory(row pgx.Row) (*ExpenseCategory, error) {
	var category ExpenseCategory
	if err := row.Scan(
		&category.ID,
		&category.TenantID,
		&category.StoreID,
		&category.Name,
		&category.Slug,
		&category.IsSystem,
		&category.CreatedAt,
		&category.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCategoryNotFound
		}
		return nil, err
	}
	return &category, nil
}
