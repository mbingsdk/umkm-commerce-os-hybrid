package payment

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/httpserver"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/tenantctx"
)

const maxPaymentConfirmationBodyBytes = 1 << 20

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) PublicConfirm(w http.ResponseWriter, r *http.Request) {
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		httpserver.WriteError(w, r, h.logger, apperror.Validation("Validation failed", []map[string]string{{"field": "Idempotency-Key", "message": "header is required"}}))
		return
	}

	rawBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxPaymentConfirmationBodyBytes))
	if err != nil {
		httpserver.WriteError(w, r, h.logger, apperror.Validation("Invalid JSON payload", []map[string]string{{"field": "body", "message": "request body is too large or unreadable"}}))
		return
	}
	request, err := decodePublicConfirmationRequest(rawBody)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, statusCode, err := h.service.PublicConfirm(r.Context(), PublicConfirmationCommand{
		StoreSlug:      chi.URLParam(r, "storeSlug"),
		OrderNumber:    chi.URLParam(r, "orderNumber"),
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

	httpserver.WriteJSON(w, statusCode, httpserver.SuccessResponse{Success: true, Message: "Payment confirmation submitted", Data: result})
}

func (h *Handler) ListConfirmations(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}
	orderID, err := parseOrderID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.ListConfirmations(r.Context(), currentTenant.TenantID, currentTenant.StoreID, orderID)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) ConfirmPayment(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}
	orderID, err := parseOrderID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	input, err := reviewInputFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	result, err := h.service.ConfirmPayment(r.Context(), currentTenant.TenantID, currentTenant.StoreID, orderID, ConfirmInput{
		ActorUserID:    currentTenant.UserID,
		ConfirmationID: input.confirmationID,
		Note:           input.note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	httpserver.WriteOK(w, "Payment confirmed", result)
}

func (h *Handler) RejectPayment(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}
	orderID, err := parseOrderID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	input, err := reviewInputFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	result, err := h.service.RejectPayment(r.Context(), currentTenant.TenantID, currentTenant.StoreID, orderID, RejectInput{
		ActorUserID:    currentTenant.UserID,
		ConfirmationID: input.confirmationID,
		Note:           input.note,
	})
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	httpserver.WriteOK(w, "Payment rejected", result)
}

func decodePublicConfirmationRequest(rawBody []byte) (PublicConfirmationRequest, error) {
	var request PublicConfirmationRequest
	decoder := json.NewDecoder(bytes.NewReader(rawBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return PublicConfirmationRequest{}, apperror.Validation("Invalid JSON payload", []map[string]string{{"field": "body", "message": err.Error()}})
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return PublicConfirmationRequest{}, apperror.Validation("Invalid JSON payload", []map[string]string{{"field": "body", "message": "request body must contain a single JSON value"}})
	}
	return request, nil
}

type reviewInput struct {
	confirmationID *uuid.UUID
	note           string
}

func reviewInputFromRequest(r *http.Request) (reviewInput, error) {
	if r.Body == nil || r.ContentLength == 0 {
		return reviewInput{}, nil
	}
	var request ReviewPaymentRequest
	if err := httpserver.DecodeJSON(r, &request); err != nil {
		return reviewInput{}, err
	}
	result := reviewInput{note: strings.TrimSpace(request.Note)}
	if strings.TrimSpace(request.PaymentConfirmationID) != "" {
		confirmationID, err := uuid.Parse(strings.TrimSpace(request.PaymentConfirmationID))
		if err != nil {
			return reviewInput{}, apperror.Validation("Validation failed", []map[string]string{{"field": "payment_confirmation_id", "message": "payment_confirmation_id must be a valid UUID"}})
		}
		result.confirmationID = &confirmationID
	}
	return result, nil
}

func parseOrderID(r *http.Request) (uuid.UUID, error) {
	orderID, err := uuid.Parse(chi.URLParam(r, "orderId"))
	if err != nil {
		return uuid.Nil, apperror.Validation("Validation failed", []map[string]string{{"field": "orderId", "message": "orderId must be a valid UUID"}})
	}
	return orderID, nil
}
