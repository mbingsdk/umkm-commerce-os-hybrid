package product

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
)

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
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

	input, err := h.readImageUpload(w, r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.AttachImage(r.Context(), currentTenant.TenantID, currentTenant.StoreID, productID, input)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "Image uploaded", result)
}

func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
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
	imageID, err := parseImageID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	if err := h.service.DeleteImage(r.Context(), currentTenant.TenantID, currentTenant.StoreID, productID, imageID); err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "Image deleted", nil)
}

func (h *Handler) readImageUpload(w http.ResponseWriter, r *http.Request) (AttachImageInput, error) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(h.maxUploadBytes); err != nil {
		return AttachImageInput{}, apperror.Validation("Validation failed", []map[string]string{
			{"field": "file", "message": "Invalid multipart form or file is too large"},
		})
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return AttachImageInput{}, apperror.Validation("Validation failed", []map[string]string{
			{"field": "file", "message": "File is required"},
		})
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, h.maxUploadBytes+1))
	if err != nil {
		return AttachImageInput{}, apperror.Validation("Validation failed", []map[string]string{
			{"field": "file", "message": "Unable to read file"},
		})
	}

	isPrimary, err := parseOptionalBool(r.FormValue("is_primary"))
	if err != nil {
		return AttachImageInput{}, err
	}
	sortOrder, err := parseOptionalInt(r.FormValue("sort_order"))
	if err != nil {
		return AttachImageInput{}, err
	}

	return AttachImageInput{
		AltText:   strings.TrimSpace(r.FormValue("alt_text")),
		IsPrimary: isPrimary,
		SortOrder: sortOrder,
		Data:      data,
	}, nil
}

func parseImageID(r *http.Request) (uuid.UUID, error) {
	imageID, err := uuid.Parse(chi.URLParam(r, "imageId"))
	if err != nil {
		return uuid.Nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "imageId", "message": "imageId must be a valid UUID"},
		})
	}
	return imageID, nil
}

func parseOptionalBool(value string) (bool, error) {
	if strings.TrimSpace(value) == "" {
		return false, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, apperror.Validation("Validation failed", []map[string]string{
			{"field": "is_primary", "message": "is_primary must be true or false"},
		})
	}
	return parsed, nil
}

func parseOptionalInt(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, apperror.Validation("Validation failed", []map[string]string{
			{"field": "sort_order", "message": "sort_order must be an integer"},
		})
	}
	return parsed, nil
}
