package upload

import (
	"io"
	"log/slog"
	"net/http"

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	fileBytes, folder, err := h.readUpload(w, r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Create(r.Context(), CreateInput{
		TenantID: currentTenant.TenantID,
		Folder:   folder,
		Data:     fileBytes,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteCreated(w, "File uploaded", result)
}

func (h *Handler) readUpload(w http.ResponseWriter, r *http.Request) ([]byte, string, error) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxUploadBytes+(1<<20))
	if err := r.ParseMultipartForm(h.maxUploadBytes); err != nil {
		return nil, "", apperror.Validation("Validation failed", []map[string]string{
			{"field": "file", "message": "Invalid multipart form or file is too large"},
		})
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		return nil, "", apperror.Validation("Validation failed", []map[string]string{
			{"field": "file", "message": "File is required"},
		})
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, h.maxUploadBytes+1))
	if err != nil {
		return nil, "", apperror.Validation("Validation failed", []map[string]string{
			{"field": "file", "message": "Unable to read file"},
		})
	}

	return data, r.FormValue("folder"), nil
}
