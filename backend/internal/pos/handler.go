package pos

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

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

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	filters, err := productFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	result, err := h.service.ListProducts(r.Context(), currentTenant.TenantID, currentTenant.StoreID, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOK(w, "OK", result)
}

func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if idempotencyKey == "" {
		httpserver.WriteError(w, r, h.logger, apperror.Validation("Validation failed", []map[string]string{
			{"field": "Idempotency-Key", "message": "header is required"},
		}))
		return
	}

	rawBody, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxPOSTransactionBodyBytes))
	if err != nil {
		httpserver.WriteError(w, r, h.logger, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": "request body is too large or unreadable"},
		}))
		return
	}
	request, err := decodeCreateTransactionRequest(rawBody)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	result, err := h.service.CreateTransaction(r.Context(), CreateTransactionCommand{
		TenantID:       currentTenant.TenantID,
		StoreID:        currentTenant.StoreID,
		ActorUserID:    currentTenant.UserID,
		Role:           currentTenant.Role,
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
		Message: "POS transaction created",
		Data:    result.Response,
	})
}

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	filters, err := transactionFiltersFromRequest(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	items, meta, err := h.service.ListTransactions(r.Context(), currentTenant.TenantID, currentTenant.StoreID, currentTenant.UserID, currentTenant.Role, filters)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}

	httpserver.WriteOKWithMeta(w, "OK", items, meta)
}

func (h *Handler) TransactionDetail(w http.ResponseWriter, r *http.Request) {
	currentTenant, ok := tenantctx.FromContext(r.Context())
	if !ok {
		httpserver.WriteError(w, r, h.logger, apperror.Forbidden("Tenant context is required"))
		return
	}

	transactionID, err := parseTransactionID(r)
	if err != nil {
		httpserver.WriteError(w, r, h.logger, err)
		return
	}
	result, err := h.service.TransactionDetail(r.Context(), currentTenant.TenantID, currentTenant.StoreID, currentTenant.UserID, currentTenant.Role, transactionID)
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

func parseTransactionID(r *http.Request) (uuid.UUID, error) {
	transactionID, err := uuid.Parse(chi.URLParam(r, "transactionId"))
	if err != nil {
		return uuid.Nil, invalidField("transactionId", "transactionId must be a valid UUID")
	}
	return transactionID, nil
}

func productFiltersFromRequest(r *http.Request) (ProductSearchFilters, error) {
	query := r.URL.Query()
	filters := ProductSearchFilters{
		Query:   strings.TrimSpace(query.Get("q")),
		Barcode: strings.TrimSpace(query.Get("barcode")),
	}
	if filters.Query == "" {
		filters.Query = strings.TrimSpace(query.Get("search"))
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return ProductSearchFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	return filters, nil
}

func transactionFiltersFromRequest(r *http.Request) (TransactionListFilters, error) {
	query := r.URL.Query()
	var filters TransactionListFilters
	if raw := strings.TrimSpace(query.Get("date_from")); raw != "" {
		parsed, err := parseDateParam(raw, false)
		if err != nil {
			return TransactionListFilters{}, invalidField("date_from", "date_from must be YYYY-MM-DD or RFC3339")
		}
		filters.DateFrom = &parsed
	}
	if raw := strings.TrimSpace(query.Get("date_to")); raw != "" {
		parsed, err := parseDateParam(raw, true)
		if err != nil {
			return TransactionListFilters{}, invalidField("date_to", "date_to must be YYYY-MM-DD or RFC3339")
		}
		filters.DateTo = &parsed
	}
	if raw := strings.TrimSpace(query.Get("payment_method")); raw != "" {
		filters.PaymentMethod = &raw
	}
	if raw := strings.TrimSpace(query.Get("cashier_id")); raw != "" {
		cashierID, err := uuid.Parse(raw)
		if err != nil {
			return TransactionListFilters{}, invalidField("cashier_id", "cashier_id must be a valid UUID")
		}
		filters.CashierID = &cashierID
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		limit, err := strconv.Atoi(raw)
		if err != nil || limit <= 0 {
			return TransactionListFilters{}, invalidField("limit", "limit must be a positive integer")
		}
		filters.Limit = limit
	}
	if raw := strings.TrimSpace(query.Get("cursor")); raw != "" {
		cursor, err := DecodeCursor(raw)
		if err != nil {
			return TransactionListFilters{}, invalidField("cursor", "cursor is invalid")
		}
		filters.Cursor = cursor
	}
	return filters, nil
}

func parseDateParam(raw string, endExclusive bool) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, err
	}
	if endExclusive {
		return parsed.AddDate(0, 0, 1), nil
	}
	return parsed, nil
}

func decodeCreateTransactionRequest(rawBody []byte) (CreateTransactionRequest, error) {
	var request CreateTransactionRequest
	decoder := json.NewDecoder(bytes.NewReader(rawBody))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		return CreateTransactionRequest{}, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": err.Error()},
		})
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return CreateTransactionRequest{}, apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": "request body must contain a single JSON value"},
		})
	}
	return request, nil
}
