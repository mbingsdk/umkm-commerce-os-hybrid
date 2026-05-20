package finance

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/audit"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/outbox"
)

const (
	defaultExpenseListLimit = 20
	maxExpenseListLimit     = 100
	maxExpenseTitleLength   = 200
	maxExpenseNoteLength    = 500
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type expenseStore interface {
	ListExpenses(context.Context, db.Queryer, uuid.UUID, uuid.UUID, ListExpenseFilters) ([]Expense, error)
	FindExpenseByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*Expense, error)
	FindCategoryByID(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID) (*ExpenseCategory, error)
	FindCategoryBySlug(context.Context, db.Queryer, uuid.UUID, uuid.UUID, string) (*ExpenseCategory, error)
	CreateExpense(context.Context, db.Queryer, CreateExpenseParams) (*Expense, error)
	UpdateExpense(context.Context, db.Queryer, UpdateExpenseParams) (*Expense, error)
	SoftDeleteExpense(context.Context, db.Queryer, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) (*Expense, error)
}

type auditStore interface {
	Create(context.Context, db.Queryer, audit.Entry) error
}

type outboxStore interface {
	Insert(context.Context, db.Queryer, outbox.InsertEventParams) (*outbox.Event, error)
}

type Service struct {
	db        database
	expenses  expenseStore
	auditLogs auditStore
	outbox    outboxStore
}

type CreateExpenseInput struct {
	ActorUserID   uuid.UUID
	CategoryID    *uuid.UUID
	Category      string
	Title         string
	Description   string
	Amount        int64
	ExpenseDate   string
	PaymentMethod string
	Note          string
}

type UpdateExpenseInput struct {
	ActorUserID   uuid.UUID
	CategoryID    *uuid.UUID
	Category      *string
	Title         *string
	Description   *string
	Amount        *int64
	ExpenseDate   *string
	PaymentMethod *string
	Note          *string
}

type normalizedExpenseInput struct {
	CategoryID    *uuid.UUID
	CategorySlug  string
	Title         string
	Amount        int64
	ExpenseDate   time.Time
	PaymentMethod string
	Note          string
}

func NewService(database database, expenseRepo expenseStore, auditLogs auditStore, outboxRepo outboxStore) *Service {
	return &Service{
		db:        database,
		expenses:  expenseRepo,
		auditLogs: auditLogs,
		outbox:    outboxRepo,
	}
}

func (s *Service) ListExpenses(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, filters ListExpenseFilters) ([]ExpenseResponse, PaginationMeta, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return nil, PaginationMeta{}, err
	}

	normalized := normalizeListFilters(filters)
	queryFilters := normalized
	queryFilters.Limit = normalized.Limit + 1

	items, err := s.expenses.ListExpenses(ctx, s.db, tenantID, storeID, queryFilters)
	if err != nil {
		return nil, PaginationMeta{}, apperror.Internal(err)
	}

	hasMore := len(items) > normalized.Limit
	if hasMore {
		items = items[:normalized.Limit]
	}

	response := make([]ExpenseResponse, 0, len(items))
	for _, item := range items {
		response = append(response, NewExpenseResponse(item))
	}

	var nextCursor *string
	if hasMore && len(items) > 0 {
		encoded, err := EncodeExpenseCursor(items[len(items)-1])
		if err != nil {
			return nil, PaginationMeta{}, apperror.Internal(err)
		}
		nextCursor = &encoded
	}

	return response, PaginationMeta{Pagination: Pagination{
		Limit:      normalized.Limit,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}}, nil
}

func (s *Service) CreateExpense(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, input CreateExpenseInput) (ExpenseResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ExpenseResponse{}, err
	}
	if input.ActorUserID == uuid.Nil {
		return ExpenseResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	normalized, err := normalizeCreateExpense(input)
	if err != nil {
		return ExpenseResponse{}, err
	}

	var created *Expense
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		categoryID, err := s.resolveCategoryID(ctx, tx, tenantID, storeID, normalized.CategoryID, normalized.CategorySlug)
		if err != nil {
			return err
		}

		expense, err := s.expenses.CreateExpense(ctx, tx, CreateExpenseParams{
			TenantID:      tenantID,
			StoreID:       storeID,
			CategoryID:    categoryID,
			Title:         normalized.Title,
			Amount:        normalized.Amount,
			ExpenseDate:   normalized.ExpenseDate,
			PaymentMethod: normalized.PaymentMethod,
			Note:          normalized.Note,
			CreatedBy:     input.ActorUserID,
		})
		if err != nil {
			return apperror.Internal(err)
		}

		if err := s.auditExpenseChange(ctx, tx, AuditActionExpenseCreated, input.ActorUserID, nil, expense); err != nil {
			return err
		}
		if err := s.insertExpenseChangedEvent(ctx, tx, *expense, AuditActionExpenseCreated); err != nil {
			return err
		}

		created = expense
		return nil
	})
	if err != nil {
		return ExpenseResponse{}, err
	}
	if created == nil {
		return ExpenseResponse{}, apperror.Internal(errors.New("created expense is nil"))
	}

	return NewExpenseResponse(*created), nil
}

func (s *Service) UpdateExpense(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, expenseID uuid.UUID, input UpdateExpenseInput) (ExpenseResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ExpenseResponse{}, err
	}
	if expenseID == uuid.Nil {
		return ExpenseResponse{}, invalidField("expense_id", "Expense is required")
	}
	if input.ActorUserID == uuid.Nil {
		return ExpenseResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	var updated *Expense
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.expenses.FindExpenseByID(ctx, tx, tenantID, storeID, expenseID)
		if err != nil {
			if errors.Is(err, ErrExpenseNotFound) {
				return apperror.NotFound("Expense not found")
			}
			return apperror.Internal(err)
		}

		normalized, err := normalizeUpdateExpense(*current, input)
		if err != nil {
			return err
		}
		categoryID, err := s.resolveCategoryID(ctx, tx, tenantID, storeID, normalized.CategoryID, normalized.CategorySlug)
		if err != nil {
			return err
		}

		expense, err := s.expenses.UpdateExpense(ctx, tx, UpdateExpenseParams{
			TenantID:      tenantID,
			StoreID:       storeID,
			ExpenseID:     expenseID,
			CategoryID:    categoryID,
			Title:         normalized.Title,
			Amount:        normalized.Amount,
			ExpenseDate:   normalized.ExpenseDate,
			PaymentMethod: normalized.PaymentMethod,
			Note:          normalized.Note,
			UpdatedBy:     input.ActorUserID,
		})
		if err != nil {
			if errors.Is(err, ErrExpenseNotFound) {
				return apperror.NotFound("Expense not found")
			}
			return apperror.Internal(err)
		}

		if err := s.auditExpenseChange(ctx, tx, AuditActionExpenseUpdated, input.ActorUserID, current, expense); err != nil {
			return err
		}
		if err := s.insertExpenseChangedEvent(ctx, tx, *expense, AuditActionExpenseUpdated); err != nil {
			return err
		}

		updated = expense
		return nil
	})
	if err != nil {
		return ExpenseResponse{}, err
	}
	if updated == nil {
		return ExpenseResponse{}, apperror.Internal(errors.New("updated expense is nil"))
	}

	return NewExpenseResponse(*updated), nil
}

func (s *Service) DeleteExpense(ctx context.Context, tenantID uuid.UUID, storeID uuid.UUID, expenseID uuid.UUID, actorUserID uuid.UUID) (ExpenseResponse, error) {
	if err := validateScope(tenantID, storeID); err != nil {
		return ExpenseResponse{}, err
	}
	if expenseID == uuid.Nil {
		return ExpenseResponse{}, invalidField("expense_id", "Expense is required")
	}
	if actorUserID == uuid.Nil {
		return ExpenseResponse{}, invalidField("actor_user_id", "Actor is required")
	}

	var deleted *Expense
	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		current, err := s.expenses.FindExpenseByID(ctx, tx, tenantID, storeID, expenseID)
		if err != nil {
			if errors.Is(err, ErrExpenseNotFound) {
				return apperror.NotFound("Expense not found")
			}
			return apperror.Internal(err)
		}

		expense, err := s.expenses.SoftDeleteExpense(ctx, tx, tenantID, storeID, expenseID, actorUserID)
		if err != nil {
			if errors.Is(err, ErrExpenseNotFound) {
				return apperror.NotFound("Expense not found")
			}
			return apperror.Internal(err)
		}

		if err := s.auditExpenseChange(ctx, tx, AuditActionExpenseDeleted, actorUserID, current, expense); err != nil {
			return err
		}
		if err := s.insertExpenseChangedEvent(ctx, tx, *expense, AuditActionExpenseDeleted); err != nil {
			return err
		}

		deleted = expense
		return nil
	})
	if err != nil {
		return ExpenseResponse{}, err
	}
	if deleted == nil {
		return ExpenseResponse{}, apperror.Internal(errors.New("deleted expense is nil"))
	}

	return NewExpenseResponse(*deleted), nil
}

func (s *Service) resolveCategoryID(
	ctx context.Context,
	tx db.Tx,
	tenantID uuid.UUID,
	storeID uuid.UUID,
	categoryID *uuid.UUID,
	categorySlug string,
) (*uuid.UUID, error) {
	if categoryID != nil && *categoryID != uuid.Nil {
		category, err := s.expenses.FindCategoryByID(ctx, tx, tenantID, storeID, *categoryID)
		if err != nil {
			if errors.Is(err, ErrCategoryNotFound) {
				return nil, apperror.NotFound("Expense category not found")
			}
			return nil, apperror.Internal(err)
		}
		return &category.ID, nil
	}

	categorySlug = strings.TrimSpace(categorySlug)
	if categorySlug == "" {
		return nil, nil
	}
	category, err := s.expenses.FindCategoryBySlug(ctx, tx, tenantID, storeID, categorySlug)
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			return nil, apperror.NotFound("Expense category not found")
		}
		return nil, apperror.Internal(err)
	}
	return &category.ID, nil
}

func (s *Service) auditExpenseChange(ctx context.Context, tx db.Tx, action string, actorUserID uuid.UUID, before *Expense, after *Expense) error {
	var tenantID uuid.UUID
	var storeID uuid.UUID
	var entityID uuid.UUID
	if after != nil {
		tenantID = after.TenantID
		storeID = after.StoreID
		entityID = after.ID
	} else if before != nil {
		tenantID = before.TenantID
		storeID = before.StoreID
		entityID = before.ID
	}

	if err := s.auditLogs.Create(ctx, tx, audit.Entry{
		TenantID:    tenantID,
		StoreID:     &storeID,
		ActorUserID: &actorUserID,
		Action:      action,
		EntityType:  AggregateExpense,
		EntityID:    &entityID,
		BeforeData:  auditExpenseSnapshot(before),
		AfterData:   auditExpenseSnapshot(after),
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func (s *Service) insertExpenseChangedEvent(ctx context.Context, tx db.Tx, expense Expense, action string) error {
	payload, err := json.Marshal(map[string]any{
		"tenant_id":      expense.TenantID.String(),
		"store_id":       expense.StoreID.String(),
		"expense_id":     expense.ID.String(),
		"category_id":    uuidPtrString(expense.CategoryID),
		"action":         action,
		"amount":         expense.Amount,
		"expense_date":   expense.ExpenseDate.Format(dateFormat),
		"payment_method": expense.PaymentMethod,
	})
	if err != nil {
		return apperror.Internal(err)
	}

	if _, err := s.outbox.Insert(ctx, tx, outbox.InsertEventParams{
		TenantID:      expense.TenantID,
		EventType:     EventExpenseChanged,
		AggregateType: AggregateExpense,
		AggregateID:   expense.ID,
		Payload:       payload,
	}); err != nil {
		return apperror.Internal(err)
	}
	return nil
}

func auditExpenseSnapshot(expense *Expense) map[string]any {
	if expense == nil {
		return nil
	}
	return map[string]any{
		"id":             expense.ID.String(),
		"category_id":    uuidPtrString(expense.CategoryID),
		"category":       expense.CategorySlug,
		"title":          expense.Title,
		"amount":         expense.Amount,
		"expense_date":   expense.ExpenseDate.Format(dateFormat),
		"payment_method": expense.PaymentMethod,
	}
}

func uuidPtrString(value *uuid.UUID) any {
	if value == nil || *value == uuid.Nil {
		return nil
	}
	return value.String()
}

func normalizeListFilters(filters ListExpenseFilters) ListExpenseFilters {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.CategorySlug = strings.TrimSpace(filters.CategorySlug)
	if filters.Limit <= 0 {
		filters.Limit = defaultExpenseListLimit
	}
	if filters.Limit > maxExpenseListLimit {
		filters.Limit = maxExpenseListLimit
	}
	return filters
}

func normalizeCreateExpense(input CreateExpenseInput) (normalizedExpenseInput, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		title = strings.TrimSpace(input.Description)
	}
	expenseDate, dateErr := parseExpenseDate(strings.TrimSpace(input.ExpenseDate))
	normalized := normalizedExpenseInput{
		CategoryID:    input.CategoryID,
		CategorySlug:  strings.TrimSpace(input.Category),
		Title:         title,
		Amount:        input.Amount,
		ExpenseDate:   expenseDate,
		PaymentMethod: strings.TrimSpace(input.PaymentMethod),
		Note:          strings.TrimSpace(input.Note),
	}

	return normalized, validateExpenseInput(normalized, dateErr)
}

func normalizeUpdateExpense(current Expense, input UpdateExpenseInput) (normalizedExpenseInput, error) {
	normalized := normalizedExpenseInput{
		CategoryID:    current.CategoryID,
		CategorySlug:  "",
		Title:         current.Title,
		Amount:        current.Amount,
		ExpenseDate:   current.ExpenseDate,
		PaymentMethod: current.PaymentMethod,
		Note:          current.Note,
	}

	var dateErr error
	if input.CategoryID != nil {
		normalized.CategoryID = input.CategoryID
		normalized.CategorySlug = ""
	}
	if input.Category != nil {
		normalized.CategoryID = nil
		normalized.CategorySlug = strings.TrimSpace(*input.Category)
	}
	if input.Title != nil {
		normalized.Title = strings.TrimSpace(*input.Title)
	}
	if input.Description != nil && normalized.Title == "" {
		normalized.Title = strings.TrimSpace(*input.Description)
	}
	if input.Amount != nil {
		normalized.Amount = *input.Amount
	}
	if input.ExpenseDate != nil {
		normalized.ExpenseDate, dateErr = parseExpenseDate(strings.TrimSpace(*input.ExpenseDate))
	}
	if input.PaymentMethod != nil {
		normalized.PaymentMethod = strings.TrimSpace(*input.PaymentMethod)
	}
	if input.Note != nil {
		normalized.Note = strings.TrimSpace(*input.Note)
	}

	return normalized, validateExpenseInput(normalized, dateErr)
}

func validateExpenseInput(input normalizedExpenseInput, dateErr error) error {
	details := make([]map[string]string, 0)
	if input.Title == "" {
		details = append(details, map[string]string{"field": "title", "message": "title is required"})
	}
	if len(input.Title) > maxExpenseTitleLength {
		details = append(details, map[string]string{"field": "title", "message": "title must be 200 characters or fewer"})
	}
	if input.Amount <= 0 {
		details = append(details, map[string]string{"field": "amount", "message": "amount must be greater than zero"})
	}
	if dateErr != nil || input.ExpenseDate.IsZero() {
		details = append(details, map[string]string{"field": "expense_date", "message": "expense_date must be YYYY-MM-DD or RFC3339"})
	}
	if input.PaymentMethod != "" && !allowedPaymentMethod(input.PaymentMethod) {
		details = append(details, map[string]string{"field": "payment_method", "message": "payment_method must be cash, bank_transfer, qris_manual, or other"})
	}
	if len(input.Note) > maxExpenseNoteLength {
		details = append(details, map[string]string{"field": "note", "message": "note must be 500 characters or fewer"})
	}
	if input.CategoryID != nil && *input.CategoryID == uuid.Nil {
		details = append(details, map[string]string{"field": "category_id", "message": "category_id must be a valid UUID"})
	}

	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}
	return nil
}

func parseExpenseDate(raw string) (time.Time, error) {
	if raw == "" {
		return time.Time{}, errors.New("expense date is required")
	}
	if parsed, err := time.Parse(dateFormat, raw); err == nil {
		return parsed, nil
	}
	return time.Parse(time.RFC3339, raw)
}

func allowedPaymentMethod(method string) bool {
	switch method {
	case PaymentMethodCash, PaymentMethodBankTransfer, PaymentMethodQRISManual, PaymentMethodOther:
		return true
	default:
		return false
	}
}

func validateScope(tenantID uuid.UUID, storeID uuid.UUID) error {
	var details []map[string]string
	if tenantID == uuid.Nil {
		details = append(details, map[string]string{"field": "tenant_id", "message": "Tenant is required"})
	}
	if storeID == uuid.Nil {
		details = append(details, map[string]string{"field": "store_id", "message": "Store is required"})
	}
	if len(details) > 0 {
		return apperror.Validation("Validation failed", details)
	}
	return nil
}

func invalidField(field string, message string) error {
	return apperror.Validation("Validation failed", []map[string]string{{"field": field, "message": message}})
}
