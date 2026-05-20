package inventory

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

func (h *Handler) ListStocks(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	filters, err := stockFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.ListStocks(r.Context(), currentTenant.TenantID, currentTenant.StoreID, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) ListProductMovements(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	productID, err := parseProductID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	filters, err := movementFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.ListMovements(r.Context(), currentTenant.TenantID, currentTenant.StoreID, productID, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	productID, err := parseProductID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req AdjustStockRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.AdjustStock(r.Context(), currentTenant.TenantID, currentTenant.StoreID, productID, AdjustStockInput{
		ActorUserID:    currentTenant.UserID,
		AdjustmentType: req.AdjustmentType,
		Type:           req.Type,
		Quantity:       req.Quantity,
		Reason:         req.Reason,
		Note:           req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Stock adjusted", result)
}

func (h *Handler) UpdateThreshold(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	productID, err := parseProductID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateThresholdRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.UpdateThreshold(r.Context(), currentTenant.TenantID, currentTenant.StoreID, productID, UpdateThresholdInput{
		ActorUserID:       currentTenant.UserID,
		LowStockThreshold: req.LowStockThreshold,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Low stock threshold updated", result)
}

func stockFiltersFromRequest(r *http.Request) (ListStockFilters, error) {
	query := r.URL.Query()
	filters := ListStockFilters{
		Query: strings.TrimSpace(query.Get("q")),
	}
	if filters.Query == "" {
		filters.Query = strings.TrimSpace(query.Get("search"))
	}

	if raw := strings.TrimSpace(query.Get("low_stock")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return ListStockFilters{}, invalidField("low_stock", "low_stock must be true or false")
		}
		filters.LowStock = &value
	}
	if raw := strings.TrimSpace(query.Get("out_of_stock")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return ListStockFilters{}, invalidField("out_of_stock", "out_of_stock must be true or false")
		}
		filters.OutOfStock = &value
	}
	if raw := strings.TrimSpace(query.Get("category_id")); raw != "" {
		categoryID, err := uuid.Parse(raw)
		if err != nil {
			return ListStockFilters{}, invalidField("category_id", "category_id must be a valid UUID")
		}
		filters.CategoryID = &categoryID
	}

	limit, cursor, err := paginationFromRequest(r)
	if err != nil {
		return ListStockFilters{}, err
	}
	filters.Limit = limit
	filters.Cursor = cursor

	return filters, nil
}

func movementFiltersFromRequest(r *http.Request) (ListMovementFilters, error) {
	limit, cursor, err := paginationFromRequest(r)
	if err != nil {
		return ListMovementFilters{}, err
	}
	return ListMovementFilters{Limit: limit, Cursor: cursor}, nil
}

func paginationFromRequest(r *http.Request) (int, *Cursor, error) {
	query := r.URL.Query()
	var limit int
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			return 0, nil, invalidField("limit", "limit must be a positive integer")
		}
		limit = parsed
	}

	var cursor *Cursor
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		parsed, err := DecodeCursor(raw)
		if err != nil {
			return 0, nil, invalidField("cursor", "cursor is invalid")
		}
		cursor = parsed
	}

	return limit, cursor, nil
}

func parseProductID(r *http.Request) (uuid.UUID, error) {
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		return uuid.Nil, invalidField("productId", "productId must be a valid UUID")
	}
	return productID, nil
}
