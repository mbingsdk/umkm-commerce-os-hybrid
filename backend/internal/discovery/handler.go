package discovery

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
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

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.Home(r.Context())
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) ListStores(w http.ResponseWriter, r *http.Request) {
	filters, err := storeFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, meta, err := h.service.ListStores(r.Context(), filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", result, meta)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	filters, err := productFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, meta, err := h.service.ListProducts(r.Context(), filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", result, meta)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	filters, err := searchFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Search(r.Context(), filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func storeFiltersFromRequest(r *http.Request) (ListStoresFilters, error) {
	query := r.URL.Query()
	filters := ListStoresFilters{
		Query:    strings.TrimSpace(query.Get("q")),
		City:     strings.TrimSpace(query.Get("city")),
		Category: strings.TrimSpace(query.Get("category")),
	}
	if filters.Query == "" {
		filters.Query = strings.TrimSpace(query.Get("search"))
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return ListStoresFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return ListStoresFilters{}, invalidField("cursor", "cursor is invalid")
		}
		filters.Cursor = cursor
	}
	return filters, nil
}

func productFiltersFromRequest(r *http.Request) (ListProductsFilters, error) {
	query := r.URL.Query()
	priceMin, err := parseOptionalMoney(query.Get("price_min"), "price_min")
	if err != nil {
		return ListProductsFilters{}, err
	}
	priceMax, err := parseOptionalMoney(query.Get("price_max"), "price_max")
	if err != nil {
		return ListProductsFilters{}, err
	}

	filters := ListProductsFilters{
		Query:    strings.TrimSpace(query.Get("q")),
		City:     strings.TrimSpace(query.Get("city")),
		Category: strings.TrimSpace(query.Get("category")),
		PriceMin: priceMin,
		PriceMax: priceMax,
	}
	if filters.Query == "" {
		filters.Query = strings.TrimSpace(query.Get("search"))
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return ListProductsFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return ListProductsFilters{}, invalidField("cursor", "cursor is invalid")
		}
		filters.Cursor = cursor
	}
	return filters, nil
}

func searchFiltersFromRequest(r *http.Request) (SearchFilters, error) {
	query := r.URL.Query()
	filters := SearchFilters{
		Query:    strings.TrimSpace(query.Get("q")),
		Type:     strings.TrimSpace(query.Get("type")),
		City:     strings.TrimSpace(query.Get("city")),
		Category: strings.TrimSpace(query.Get("category")),
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return SearchFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	return filters, nil
}
