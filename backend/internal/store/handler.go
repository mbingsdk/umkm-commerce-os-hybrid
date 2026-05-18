package store

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

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

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	result, err := h.service.GetCurrent(r.Context(), currentTenant.TenantID)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	var req UpdateProfileRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.UpdateProfile(
		r.Context(),
		currentTenant.TenantID,
		currentTenant.StoreID,
		Actor{
			UserID:    currentTenant.UserID,
			IPAddress: clientIP(r),
			UserAgent: r.UserAgent(),
		},
		UpdateProfileInput{
			Name:           req.Name,
			Description:    req.Description,
			Phone:          req.Phone,
			Whatsapp:       req.Whatsapp,
			Email:          req.Email,
			Address:        req.Address,
			City:           req.City,
			Province:       req.Province,
			PostalCode:     req.PostalCode,
			IsDiscoverable: req.IsDiscoverable,
		},
	)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Store updated", result)
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, "Store published", h.service.Publish)
}

func (h *Handler) Unpublish(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, "Store unpublished", h.service.Unpublish)
}

func (h *Handler) ReplaceBusinessHours(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	var req BusinessHoursRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items := make([]BusinessHour, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, BusinessHour{
			DayOfWeek: item.DayOfWeek,
			OpenTime:  item.OpenTime,
			CloseTime: item.CloseTime,
			IsClosed:  item.IsClosed,
		})
	}

	result, err := h.service.ReplaceBusinessHours(
		r.Context(),
		currentTenant.TenantID,
		currentTenant.StoreID,
		Actor{
			UserID:    currentTenant.UserID,
			IPAddress: clientIP(r),
			UserAgent: r.UserAgent(),
		},
		items,
	)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Business hours updated", result)
}

func (h *Handler) changeStatus(
	w http.ResponseWriter,
	r *http.Request,
	message string,
	update func(context.Context, uuid.UUID, uuid.UUID, Actor) (Response, error),
) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	result, err := update(
		r.Context(),
		currentTenant.TenantID,
		currentTenant.StoreID,
		Actor{
			UserID:    currentTenant.UserID,
			IPAddress: clientIP(r),
			UserAgent: r.UserAgent(),
		},
	)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, message, result)
}

func clientIP(r *http.Request) string {
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
