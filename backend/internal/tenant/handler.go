package tenant

import (
	"log/slog"
	"net"
	"net/http"
	"strings"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
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

func (h *Handler) CreateStore(w http.ResponseWriter, r *http.Request) {
	authCtx, ok := auth.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Unauthorized("Authentication required"))
		return
	}

	var req CreateStoreRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.CreateStore(r.Context(), CreateStoreInput{
		UserID:     authCtx.UserID,
		TenantName: req.TenantName,
		TenantSlug: req.TenantSlug,
		Store: StoreCreateInput{
			Name:        req.Store.Name,
			Slug:        req.Store.Slug,
			Description: req.Store.Description,
			Phone:       req.Store.Phone,
			Whatsapp:    req.Store.Whatsapp,
			Email:       req.Store.Email,
			Address:     req.Store.Address,
			City:        req.Store.City,
			Province:    req.Store.Province,
			PostalCode:  req.Store.PostalCode,
		},
		IPAddress: clientIP(r),
		UserAgent: r.UserAgent(),
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Store created", result)
}

func (h *Handler) ListMyTenants(w http.ResponseWriter, r *http.Request) {
	authCtx, ok := auth.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Unauthorized("Authentication required"))
		return
	}

	result, err := h.service.ListMyTenants(r.Context(), authCtx.UserID)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
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
