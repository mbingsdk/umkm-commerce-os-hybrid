package finance

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	filters, err := expenseFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.ListExpenses(r.Context(), currentTenant.TenantID, currentTenant.StoreID, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	var req CreateExpenseRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.CreateExpense(r.Context(), currentTenant.TenantID, currentTenant.StoreID, CreateExpenseInput{
		ActorUserID:   currentTenant.UserID,
		CategoryID:    req.CategoryID,
		Category:      req.Category,
		Title:         req.Title,
		Description:   req.Description,
		Amount:        req.Amount,
		ExpenseDate:   req.ExpenseDate,
		PaymentMethod: req.PaymentMethod,
		Note:          req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Expense created", result)
}

func (h *Handler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	expenseID, err := parseExpenseID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateExpenseRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.UpdateExpense(r.Context(), currentTenant.TenantID, currentTenant.StoreID, expenseID, UpdateExpenseInput{
		ActorUserID:   currentTenant.UserID,
		CategoryID:    req.CategoryID,
		Category:      req.Category,
		Title:         req.Title,
		Description:   req.Description,
		Amount:        req.Amount,
		ExpenseDate:   req.ExpenseDate,
		PaymentMethod: req.PaymentMethod,
		Note:          req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Expense updated", result)
}

func (h *Handler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	expenseID, err := parseExpenseID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.DeleteExpense(r.Context(), currentTenant.TenantID, currentTenant.StoreID, expenseID, currentTenant.UserID)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Expense deleted", result)
}

func expenseFiltersFromRequest(r *http.Request) (ListExpenseFilters, error) {
	query := r.URL.Query()
	filters := ListExpenseFilters{
		Query:        strings.TrimSpace(query.Get("q")),
		CategorySlug: strings.TrimSpace(query.Get("category")),
	}
	if filters.Query == "" {
		filters.Query = strings.TrimSpace(query.Get("search"))
	}

	if raw := strings.TrimSpace(query.Get("date_from")); raw != "" {
		parsed, err := parseDateParam(raw, false)
		if err != nil {
			return ListExpenseFilters{}, invalidField("date_from", "date_from must be YYYY-MM-DD or RFC3339")
		}
		filters.DateFrom = &parsed
	}
	if raw := strings.TrimSpace(query.Get("date_to")); raw != "" {
		parsed, err := parseDateParam(raw, true)
		if err != nil {
			return ListExpenseFilters{}, invalidField("date_to", "date_to must be YYYY-MM-DD or RFC3339")
		}
		filters.DateTo = &parsed
	}
	if raw := strings.TrimSpace(query.Get("category_id")); raw != "" {
		categoryID, err := uuid.Parse(raw)
		if err != nil {
			return ListExpenseFilters{}, invalidField("category_id", "category_id must be a valid UUID")
		}
		filters.CategoryID = &categoryID
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return ListExpenseFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return ListExpenseFilters{}, invalidField("cursor", "cursor is invalid")
		}
		filters.Cursor = cursor
	}

	return filters, nil
}

func parseExpenseID(r *http.Request) (uuid.UUID, error) {
	expenseID, err := uuid.Parse(chi.URLParam(r, "expenseId"))
	if err != nil {
		return uuid.Nil, invalidField("expenseId", "expenseId must be a valid UUID")
	}
	return expenseID, nil
}

func parseDateParam(raw string, endExclusive bool) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse(dateFormat, raw)
	if err != nil {
		return time.Time{}, err
	}
	if endExclusive {
		return parsed.AddDate(0, 0, 1), nil
	}
	return parsed, nil
}
