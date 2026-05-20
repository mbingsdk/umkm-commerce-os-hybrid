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
