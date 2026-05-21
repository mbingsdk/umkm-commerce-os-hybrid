package admin

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	httpserver.WriteOK(w, "OK", NewMeResponse(adminCtx))
}

func (h *Handler) ListTenants(w http.ResponseWriter, r *http.Request) {
	filters, err := tenantFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.ListTenants(r.Context(), filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) GetTenant(w http.ResponseWriter, r *http.Request) {
	tenantID, err := parseTenantID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.GetTenantDetail(r.Context(), tenantID)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) UpdateTenantStatus(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	tenantID, err := parseTenantID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateTenantStatusRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	ipAddress, userAgent := RequestMetadata(r)
	result, err := h.service.UpdateTenantStatus(r.Context(), UpdateTenantStatusInput{
		ActorUserID: adminCtx.UserID,
		TenantID:    tenantID,
		Status:      req.Status,
		Reason:      req.Reason,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Tenant status updated", result)
}

func (h *Handler) UpdateTenantPlan(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	tenantID, err := parseTenantID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateTenantPlanRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	planID, err := uuid.Parse(strings.TrimSpace(req.PlanID))
	if err != nil {
		httpserver.WriteError(w, r, h.logger, invalidField("plan_id", "plan_id must be a valid UUID"))
		return
	}

	ipAddress, userAgent := RequestMetadata(r)
	result, err := h.service.UpdateTenantPlan(r.Context(), UpdateTenantPlanInput{
		ActorUserID: adminCtx.UserID,
		TenantID:    tenantID,
		PlanID:      planID,
		Reason:      req.Reason,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Tenant plan updated", result)
}

func tenantFiltersFromRequest(r *http.Request) (TenantListFilters, error) {
	query := r.URL.Query()
	filters := TenantListFilters{
		Status: strings.TrimSpace(query.Get("status")),
		Query:  strings.TrimSpace(query.Get("q")),
	}
	if filters.Query == "" {
		filters.Query = strings.TrimSpace(query.Get("search"))
	}
	if raw := strings.TrimSpace(query.Get("plan_id")); raw != "" {
		planID, err := uuid.Parse(raw)
		if err != nil {
			return TenantListFilters{}, invalidField("plan_id", "plan_id must be a valid UUID")
		}
		filters.PlanID = &planID
	}
	if raw := strings.TrimSpace(query.Get("created_from")); raw != "" {
		parsed, err := parseDate(raw, false)
		if err != nil {
			return TenantListFilters{}, invalidField("created_from", "created_from must be YYYY-MM-DD or RFC3339")
		}
		filters.CreatedFrom = &parsed
	} else if raw := strings.TrimSpace(query.Get("date_from")); raw != "" {
		parsed, err := parseDate(raw, false)
		if err != nil {
			return TenantListFilters{}, invalidField("date_from", "date_from must be YYYY-MM-DD or RFC3339")
		}
		filters.CreatedFrom = &parsed
	}
	if raw := strings.TrimSpace(query.Get("created_to")); raw != "" {
		parsed, err := parseDate(raw, true)
		if err != nil {
			return TenantListFilters{}, invalidField("created_to", "created_to must be YYYY-MM-DD or RFC3339")
		}
		filters.CreatedTo = &parsed
	} else if raw := strings.TrimSpace(query.Get("date_to")); raw != "" {
		parsed, err := parseDate(raw, true)
		if err != nil {
			return TenantListFilters{}, invalidField("date_to", "date_to must be YYYY-MM-DD or RFC3339")
		}
		filters.CreatedTo = &parsed
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return TenantListFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeTenantCursor(raw)
		if err != nil {
			return TenantListFilters{}, invalidField("cursor", "cursor is invalid")
		}
		filters.Cursor = cursor
	}
	return filters, nil
}

func parseTenantID(r *http.Request) (uuid.UUID, error) {
	tenantID, err := uuid.Parse(chi.URLParam(r, "tenantId"))
	if err != nil {
		return uuid.Nil, invalidField("tenantId", "tenantId must be a valid UUID")
	}
	return tenantID, nil
}
