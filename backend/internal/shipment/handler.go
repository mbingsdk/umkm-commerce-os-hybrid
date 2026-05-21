package shipment

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	var req CreateShipmentRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Create(r.Context(), currentTenant.TenantID, currentTenant.StoreID, orderID, CreateInput{
		ActorUserID:     currentTenant.UserID,
		CourierType:     req.CourierType,
		CourierName:     req.CourierName,
		TrackingNumber:  req.TrackingNumber,
		ShippingCost:    req.ShippingCost,
		AssignedToName:  req.AssignedToName,
		AssignedToPhone: req.AssignedToPhone,
		Note:            req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Shipment created", result)
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

	shipmentID, err := parseShipmentID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Detail(r.Context(), currentTenant.TenantID, currentTenant.StoreID, shipmentID)
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

	shipmentID, err := parseShipmentID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateStatusRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.UpdateStatus(r.Context(), currentTenant.TenantID, currentTenant.StoreID, shipmentID, UpdateStatusInput{
		ActorUserID: currentTenant.UserID,
		Status:      req.Status,
		Note:        req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Shipment status updated", result)
}

func (h *Handler) PublicTracking(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.PublicTracking(
		r.Context(),
		chi.URLParam(r, "storeSlug"),
		chi.URLParam(r, "orderNumber"),
		r.URL.Query().Get("phone"),
	)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
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
	if raw := strings.TrimSpace(query.Get("date_from")); raw != "" {
		dateFrom, err := parseDateParam(raw, false)
		if err != nil {
			return ListFilters{}, invalidField("date_from", "date_from must be RFC3339 or YYYY-MM-DD")
		}
		filters.DateFrom = &dateFrom
	}
	if raw := strings.TrimSpace(query.Get("date_to")); raw != "" {
		dateTo, err := parseDateParam(raw, true)
		if err != nil {
			return ListFilters{}, invalidField("date_to", "date_to must be RFC3339 or YYYY-MM-DD")
		}
		filters.DateTo = &dateTo
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return ListFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return ListFilters{}, invalidField("cursor", "cursor is invalid")
		}
		filters.Cursor = cursor
	}
	return filters, nil
}

func parseOrderID(r *http.Request) (uuid.UUID, error) {
	orderID, err := uuid.Parse(chi.URLParam(r, "orderId"))
	if err != nil {
		return uuid.Nil, invalidField("orderId", "orderId must be a valid UUID")
	}
	return orderID, nil
}

func parseShipmentID(r *http.Request) (uuid.UUID, error) {
	shipmentID, err := uuid.Parse(chi.URLParam(r, "shipmentId"))
	if err != nil {
		return uuid.Nil, invalidField("shipmentId", "shipmentId must be a valid UUID")
	}
	return shipmentID, nil
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
