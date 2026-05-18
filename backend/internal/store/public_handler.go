package store

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
)

type PublicHandler struct {
	service *PublicService
	logger  *slog.Logger
}

func NewPublicHandler(service *PublicService, logger *slog.Logger) *PublicHandler {
	return &PublicHandler{
		service: service,
		logger:  logger,
	}
}

func (h *PublicHandler) Get(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Get(r.Context(), chi.URLParam(r, "storeSlug"))
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}
