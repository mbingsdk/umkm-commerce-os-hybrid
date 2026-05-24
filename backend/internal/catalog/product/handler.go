package product

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
	service        *Service
	logger         *slog.Logger
	maxUploadBytes int64
}

func NewHandler(service *Service, logger *slog.Logger, maxUploadBytes int64) *Handler {
	return &Handler{
		service:        service,
		logger:         logger,
		maxUploadBytes: maxUploadBytes,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	filters, err := filtersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, meta, err := h.service.List(r.Context(), currentTenant.TenantID, currentTenant.StoreID, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", result, meta)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	var req CreateRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Create(r.Context(), currentTenant.TenantID, currentTenant.StoreID, currentTenant.UserID, CreateInput{
		CategoryID:     req.CategoryID,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		SKU:            req.SKU,
		Barcode:        req.Barcode,
		Price:          req.Price,
		CompareAtPrice: req.CompareAtPrice,
		CostPrice:      req.CostPrice,
		WeightGram:     req.WeightGram,
		LengthCM:       req.LengthCM,
		WidthCM:        req.WidthCM,
		HeightCM:       req.HeightCM,
		Status:         req.Status,
		IsDiscoverable: req.IsDiscoverable,
		TrackInventory: req.TrackInventory,
		AllowBackorder: req.AllowBackorder,
		InitialStock:   req.InitialStock,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Product created", result)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	productID, err := parseProductID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Get(
		r.Context(),
		currentTenant.TenantID,
		currentTenant.StoreID,
		productID,
		CanReadCostPrice(currentTenant.Role),
	)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	productID, err := parseProductID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	var req UpdateRequest
	if err := httpserver.DecodeJSON(r, &req); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Update(r.Context(), currentTenant.TenantID, currentTenant.StoreID, productID, UpdateInput{
		CategoryID:     req.CategoryID,
		Name:           req.Name,
		Slug:           req.Slug,
		Description:    req.Description,
		SKU:            req.SKU,
		Barcode:        req.Barcode,
		Price:          req.Price,
		CompareAtPrice: req.CompareAtPrice,
		CostPrice:      req.CostPrice,
		WeightGram:     req.WeightGram,
		LengthCM:       req.LengthCM,
		WidthCM:        req.WidthCM,
		HeightCM:       req.HeightCM,
		Status:         req.Status,
		IsDiscoverable: req.IsDiscoverable,
		TrackInventory: req.TrackInventory,
		AllowBackorder: req.AllowBackorder,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Product updated", result)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	productID, err := parseProductID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	if err := h.service.Delete(r.Context(), currentTenant.TenantID, currentTenant.StoreID, productID); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Product deleted", nil)
}

func filtersFromRequest(r *http.Request) (ListFilters, error) {
	filters := ListFilters{
		Query: strings.TrimSpace(r.URL.Query().Get("q")),
	}
	if status := strings.TrimSpace(r.URL.Query().Get("status")); status != "" {
		filters.Status = &status
	}
	if rawCategoryID := strings.TrimSpace(r.URL.Query().Get("category_id")); rawCategoryID != "" {
		categoryID, err := uuid.Parse(rawCategoryID)
		if err != nil {
			return ListFilters{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "category_id", "message": "category_id must be a valid UUID"},
			})
		}
		filters.CategoryID = &categoryID
	}
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			return ListFilters{}, apperror.Validation("Validation failed", []map[string]string{
				{"field": "limit", "message": "limit must be a positive integer"},
			})
		}
		filters.Limit = limit
	}
	return filters, nil
}

func parseProductID(r *http.Request) (uuid.UUID, error) {
	productID, err := uuid.Parse(chi.URLParam(r, "productId"))
	if err != nil {
		return uuid.Nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "productId", "message": "productId must be a valid UUID"},
		})
	}
	return productID, nil
}
