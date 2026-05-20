package pos

import (
	"log/slog"
	"net/http"

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

func (h *Handler) OpenSession(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	var req OpenSessionRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	openingCash, err := resolveMoney(req.OpeningCashAmount, req.OpeningCash, "opening_cash_amount")
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.OpenSession(r.Context(), currentTenant.TenantID, currentTenant.StoreID, OpenSessionInput{
		ActorUserID:       currentTenant.UserID,
		OpeningCashAmount: openingCash,
		Note:              req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Cashier session opened", result)
}

func (h *Handler) CurrentSession(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	result, err := h.service.CurrentSession(r.Context(), currentTenant.TenantID, currentTenant.StoreID, CurrentSessionInput{
		ActorUserID: currentTenant.UserID,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) CloseSession(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	sessionID, err := parseSessionID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req CloseSessionRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	closingCash, err := resolveMoney(req.ClosingCashAmount, req.ClosingCash, "closing_cash_amount")
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.CloseSession(r.Context(), currentTenant.TenantID, currentTenant.StoreID, sessionID, CloseSessionInput{
		ActorUserID:       currentTenant.UserID,
		Role:              currentTenant.Role,
		ClosingCashAmount: closingCash,
		Note:              req.Note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Cashier session closed", result)
}

func resolveMoney(primary *int64, legacy *int64, field string) (int64, error) {
	if primary != nil {
		return *primary, nil
	}
	if legacy != nil {
		return *legacy, nil
	}
	return 0, invalidField(field, field+" is required")
}

func parseSessionID(r *http.Request) (uuid.UUID, error) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "sessionId"))
	if err != nil {
		return uuid.Nil, invalidField("sessionId", "sessionId must be a valid UUID")
	}
	return sessionID, nil
}
