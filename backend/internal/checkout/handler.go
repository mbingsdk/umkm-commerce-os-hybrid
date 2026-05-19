package checkout

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

const maxCheckoutBodyBytes = 1 << 20

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

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		httpserver.WriteError(w, r, h.logger, apperror.Validation("Validation failed", []map[string]string{
			{"field": "Idempotency-Key", "message": "header is required"},
		}))
		return
	}

	rawBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxCheckoutBodyBytes))
	if err != nil {
		httpserver.WriteError(w, r, h.logger, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": "request body is too large or unreadable"},
		}))
		return
	}

	request, err := decodeCheckoutRequest(rawBody)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.Checkout(r.Context(), Command{
		StoreSlug:      chi.URLParam(r, "storeSlug"),
		IdempotencyKey: idempotencyKey,
		Method:         r.Method,
		Path:           r.URL.Path,
		RawBody:        rawBody,
		Request:        request,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteJSON(w, result.StatusCode, httpserver.SuccessResponse{
		Success: true,
		Message: "Checkout berhasil dibuat",
		Data:    result.Response,
	})
}

func decodeCheckoutRequest(rawBody []byte) (CheckoutRequest, error) {
	var request CheckoutRequest
	decoder := json.NewDecoder(bytes.NewReader(rawBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return CheckoutRequest{}, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": err.Error()},
		})
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return CheckoutRequest{}, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": "request body must contain a single JSON value"},
		})
	}
	return request, nil
}
