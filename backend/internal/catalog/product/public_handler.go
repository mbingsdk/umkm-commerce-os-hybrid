package product

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
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

func (h *PublicHandler) List(w http.ResponseWriter, r *http.Request) {
	filters, err := publicFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, meta, err := h.service.List(r.Context(), chi.URLParam(r, "storeSlug"), filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", result, meta)
}

func (h *PublicHandler) Get(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Get(
		r.Context(),
		chi.URLParam(r, "storeSlug"),
		chi.URLParam(r, "productSlug"),
	)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func publicFiltersFromRequest(r *http.Request) (PublicListFilters, error) {
	filters := PublicListFilters{
		Query:        strings.TrimSpace(r.URL.Query().Get("q")),
		CategorySlug: strings.TrimSpace(r.URL.Query().Get("category")),
	}

	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			return PublicListFilters{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "limit", "message": "limit must be a positive integer"},
			})
		}
		filters.Limit = limit
	}

	if rawCursor := strings.TrimSpace(r.URL.Query().Get("cursor")); rawCursor != "" {
		cursor, err := DecodePublicCursor(rawCursor)
		if err != nil {
			return PublicListFilters{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "cursor", "message": "cursor is invalid"},
			})
		}
		filters.Cursor = cursor
	}

	if rawInStock := strings.TrimSpace(r.URL.Query().Get("in_stock")); rawInStock != "" {
		inStock, err := strconv.ParseBool(rawInStock)
		if err != nil {
			return PublicListFilters{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "in_stock", "message": "in_stock must be true or false"},
			})
		}
		filters.InStock = &inStock
	}

	return filters, nil
}
