package order

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
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	filters, err := filtersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.List(r.Context(), currentTenant.TenantID, currentTenant.StoreID, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) Detail(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	orderID, err := parseOrderID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Detail(r.Context(), currentTenant.TenantID, currentTenant.StoreID, orderID)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	orderID, err := parseOrderID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateStatusRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.UpdateStatus(r.Context(), currentTenant.TenantID, currentTenant.StoreID, orderID, UpdateStatusInput{
		ActorUserID: currentTenant.UserID,
		Status:      req.Status,
		Note:        req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Order status updated", result)
}

func filtersFromRequest(r *http.Request) (ListFilters, error) {
	query := r.URL.Query()
	filters := ListFilters{
		Query: strings.TrimSpace(query.Get("q")),
	}
	if filters.Query == "" {
		filters.Query = strings.TrimSpace(query.Get("search"))
	}

	if raw := strings.TrimSpace(query.Get("status")); raw != "" {
		filters.Status = &raw
	}
	if raw := strings.TrimSpace(query.Get("payment_status")); raw != "" {
		filters.PaymentStatus = &raw
	}
	if raw := strings.TrimSpace(query.Get("source")); raw != "" {
		filters.Source = &raw
	}
	if raw := strings.TrimSpace(query.Get("date_from")); raw != "" {
		parsed, err := parseDateParam(raw, false)
		if err != nil {
			return ListFilters{}, apperror.Validation("Validation failed", []map[string]string{{"field": "date_from", "message": "date_from must be YYYY-MM-DD or RFC3339"}})
		}
		filters.DateFrom = &parsed
	}
	if raw := strings.TrimSpace(query.Get("date_to")); raw != "" {
		parsed, err := parseDateParam(raw, true)
		if err != nil {
			return ListFilters{}, apperror.Validation("Validation failed", []map[string]string{{"field": "date_to", "message": "date_to must be YYYY-MM-DD or RFC3339"}})
		}
		filters.DateTo = &parsed
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return ListFilters{}, apperror.Validation("Validation failed", []map[string]string{{"field": "limit", "message": "limit must be a positive integer"}})
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return ListFilters{}, apperror.Validation("Validation failed", []map[string]string{{"field": "cursor", "message": "cursor is invalid"}})
		}
		filters.Cursor = cursor
	}

	return filters, nil
}

func parseOrderID(r *http.Request) (uuid.UUID, error) {
	orderID, err := uuid.Parse(chi.URLParam(r, "orderId"))
	if err != nil {
		return uuid.Nil, apperror.Validation("Validation failed", []map[string]string{{"field": "orderId", "message": "orderId must be a valid UUID"}})
	}
	return orderID, nil
}

func parseDateParam(raw string, endExclusive bool) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}

	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	if endExclusive {
		return parsed.AddDate(0, 0, 1), nil
	}
	return parsed, nil
}
