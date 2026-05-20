package courier

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

func (h *Handler) ListZones(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	filters, err := zoneFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	zones, err := h.service.ListZones(r.Context(), currentTenant.TenantID, currentTenant.StoreID, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", zones)
}

func (h *Handler) CreateZone(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	var req CreateZoneRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	zone, err := h.service.CreateZone(r.Context(), currentTenant.TenantID, currentTenant.StoreID, CreateZoneInput{
		ActorUserID: currentTenant.UserID,
		Name:        req.Name,
		Description: req.Description,
		Rate:        req.Rate,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Courier zone created", zone)
}

func (h *Handler) UpdateZone(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	zoneID, err := parseZoneID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateZoneRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	zone, err := h.service.UpdateZone(r.Context(), currentTenant.TenantID, currentTenant.StoreID, zoneID, UpdateZoneInput{
		ActorUserID: currentTenant.UserID,
		Name:        req.Name,
		Description: req.Description,
		Rate:        req.Rate,
		IsActive:    req.IsActive,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Courier zone updated", zone)
}

func (h *Handler) DeleteZone(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	zoneID, err := parseZoneID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	zone, err := h.service.DeleteZone(r.Context(), currentTenant.TenantID, currentTenant.StoreID, zoneID, currentTenant.UserID)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Courier zone deleted", zone)
}

func (h *Handler) ListPublicZones(w http.ResponseWriter, r *http.Request) {
	storeSlug := chi.URLParam(r, "storeSlug")

	zones, err := h.service.ListPublicZones(r.Context(), storeSlug)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", zones)
}

func zoneFiltersFromRequest(r *http.Request) (ListZoneFilters, error) {
	rawActive := strings.TrimSpace(r.URL.Query().Get("is_active"))
	if rawActive == "" {
		return ListZoneFilters{}, nil
	}

	parsed, err := strconv.ParseBool(rawActive)
	if err != nil {
		return ListZoneFilters{}, invalidField("is_active", "is_active must be true or false")
	}
	return ListZoneFilters{IsActive: &parsed}, nil
}

func parseZoneID(r *http.Request) (uuid.UUID, error) {
	zoneID, err := uuid.Parse(chi.URLParam(r, "zoneId"))
	if err != nil {
		return uuid.Nil, invalidField("zoneId", "zoneId must be a valid UUID")
	}
	return zoneID, nil
}
