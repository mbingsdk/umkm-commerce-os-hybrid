package admin

import (
	"log/slog"
	"net/http"

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
