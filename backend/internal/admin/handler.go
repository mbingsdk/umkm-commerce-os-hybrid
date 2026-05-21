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

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListPlans(r.Context())
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", items)
}

func (h *Handler) ListFeaturedItems(w http.ResponseWriter, r *http.Request) {
	filters, err := featuredFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.ListFeaturedItems(r.Context(), filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) CreateFeaturedItem(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	var req CreateFeaturedRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	tenantID, err := uuid.Parse(strings.TrimSpace(req.TenantID))
	if err != nil {
		httpserver.WriteError(w, r, h.logger, invalidField("tenant_id", "tenant_id must be a valid UUID"))
		return
	}
	storeID, err := parseOptionalUUIDString(req.StoreID, "store_id")
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	productID, err := parseOptionalUUIDString(req.ProductID, "product_id")
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	ipAddress, userAgent := RequestMetadata(r)
	result, err := h.service.CreateFeaturedItem(r.Context(), CreateFeaturedInput{
		ActorUserID: adminCtx.UserID,
		ItemType:    req.ItemType,
		TenantID:    tenantID,
		StoreID:     storeID,
		ProductID:   productID,
		Placement:   req.Placement,
		SortOrder:   req.SortOrder,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		IsActive:    req.IsActive,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Featured item created", result)
}

func (h *Handler) UpdateFeaturedItem(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	featuredID, err := parseFeaturedID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateFeaturedRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	tenantID, err := parseOptionalUUIDPointer(req.TenantID, "tenant_id")
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	storeID, err := parseOptionalUUIDPointer(req.StoreID, "store_id")
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	productID, err := parseOptionalUUIDPointer(req.ProductID, "product_id")
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	ipAddress, userAgent := RequestMetadata(r)
	result, err := h.service.UpdateFeaturedItem(r.Context(), UpdateFeaturedInput{
		ActorUserID:  adminCtx.UserID,
		FeaturedID:   featuredID,
		ItemType:     req.ItemType,
		TenantID:     tenantID,
		StoreID:      storeID,
		StoreIDSet:   req.StoreID != nil,
		ProductID:    productID,
		ProductIDSet: req.ProductID != nil,
		Placement:    req.Placement,
		SortOrder:    req.SortOrder,
		StartsAt:     req.StartsAt,
		EndsAt:       req.EndsAt,
		IsActive:     req.IsActive,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Featured item updated", result)
}

func (h *Handler) DeleteFeaturedItem(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	featuredID, err := parseFeaturedID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	ipAddress, userAgent := RequestMetadata(r)
	result, err := h.service.DeleteFeaturedItem(r.Context(), DeleteFeaturedInput{
		ActorUserID: adminCtx.UserID,
		FeaturedID:  featuredID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Featured item deleted", result)
}

func (h *Handler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	filters, err := auditLogFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.ListAuditLogs(r.Context(), filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	var req CreatePlanRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	ipAddress, userAgent := RequestMetadata(r)
	result, err := h.service.CreatePlan(r.Context(), CreatePlanInput{
		ActorUserID:     adminCtx.UserID,
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		PriceMonthly:    req.PriceMonthly,
		ProductLimit:    req.ProductLimit,
		StaffLimit:      req.StaffLimit,
		CanUsePOS:       req.CanUsePOS,
		CanUseDiscovery: req.CanUseDiscovery,
		CanUseCourier:   req.CanUseCourier,
		IsActive:        req.IsActive,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Plan created", result)
}

func (h *Handler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	adminCtx, ok := FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Super admin context is required"))
		return
	}

	planID, err := parsePlanID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdatePlanRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	ipAddress, userAgent := RequestMetadata(r)
	result, err := h.service.UpdatePlan(r.Context(), UpdatePlanInput{
		ActorUserID:     adminCtx.UserID,
		PlanID:          planID,
		Code:            req.Code,
		Name:            req.Name,
		Description:     req.Description,
		PriceMonthly:    req.PriceMonthly,
		ProductLimit:    req.ProductLimit.Value,
		ProductLimitSet: req.ProductLimit.Set,
		StaffLimit:      req.StaffLimit.Value,
		StaffLimitSet:   req.StaffLimit.Set,
		CanUsePOS:       req.CanUsePOS,
		CanUseDiscovery: req.CanUseDiscovery,
		CanUseCourier:   req.CanUseCourier,
		IsActive:        req.IsActive,
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Plan updated", result)
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

func featuredFiltersFromRequest(r *http.Request) (FeaturedListFilters, error) {
	query := r.URL.Query()
	filters := FeaturedListFilters{
		ItemType:  strings.TrimSpace(query.Get("item_type")),
		Placement: strings.TrimSpace(query.Get("placement")),
	}
	if raw := strings.TrimSpace(query.Get("tenant_id")); raw != "" {
		tenantID, err := uuid.Parse(raw)
		if err != nil {
			return FeaturedListFilters{}, invalidField("tenant_id", "tenant_id must be a valid UUID")
		}
		filters.TenantID = &tenantID
	}
	if raw := strings.TrimSpace(query.Get("is_active")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			return FeaturedListFilters{}, invalidField("is_active", "is_active must be true or false")
		}
		filters.IsActive = &value
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return FeaturedListFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeAdminListCursor(raw)
		if err != nil {
			return FeaturedListFilters{}, invalidField("cursor", "cursor is invalid")
		}
		filters.Cursor = cursor
	}
	return filters, nil
}

func auditLogFiltersFromRequest(r *http.Request) (AuditLogListFilters, error) {
	query := r.URL.Query()
	filters := AuditLogListFilters{
		Action:     strings.TrimSpace(query.Get("action")),
		TargetType: strings.TrimSpace(query.Get("target_type")),
	}
	if raw := strings.TrimSpace(query.Get("actor_user_id")); raw != "" {
		actorID, err := uuid.Parse(raw)
		if err != nil {
			return AuditLogListFilters{}, invalidField("actor_user_id", "actor_user_id must be a valid UUID")
		}
		filters.ActorUserID = &actorID
	}
	if raw := strings.TrimSpace(query.Get("target_id")); raw != "" {
		targetID, err := uuid.Parse(raw)
		if err != nil {
			return AuditLogListFilters{}, invalidField("target_id", "target_id must be a valid UUID")
		}
		filters.TargetID = &targetID
	}
	if raw := strings.TrimSpace(query.Get("date_from")); raw != "" {
		parsed, err := parseDate(raw, false)
		if err != nil {
			return AuditLogListFilters{}, invalidField("date_from", "date_from must be YYYY-MM-DD or RFC3339")
		}
		filters.DateFrom = &parsed
	}
	if raw := strings.TrimSpace(query.Get("date_to")); raw != "" {
		parsed, err := parseDate(raw, true)
		if err != nil {
			return AuditLogListFilters{}, invalidField("date_to", "date_to must be YYYY-MM-DD or RFC3339")
		}
		filters.DateTo = &parsed
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return AuditLogListFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeAdminListCursor(raw)
		if err != nil {
			return AuditLogListFilters{}, invalidField("cursor", "cursor is invalid")
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

func parsePlanID(r *http.Request) (uuid.UUID, error) {
	planID, err := uuid.Parse(chi.URLParam(r, "planId"))
	if err != nil {
		return uuid.Nil, invalidField("planId", "planId must be a valid UUID")
	}
	return planID, nil
}

func parseFeaturedID(r *http.Request) (uuid.UUID, error) {
	featuredID, err := uuid.Parse(chi.URLParam(r, "featuredId"))
	if err != nil {
		return uuid.Nil, invalidField("featuredId", "featuredId must be a valid UUID")
	}
	return featuredID, nil
}

func parseOptionalUUIDString(raw string, field string) (*uuid.UUID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return nil, invalidField(field, field+" must be a valid UUID")
	}
	return &parsed, nil
}

func parseOptionalUUIDPointer(raw *string, field string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}
	return parseOptionalUUIDString(*raw, field)
}
