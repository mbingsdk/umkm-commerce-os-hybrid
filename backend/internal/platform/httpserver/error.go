package httpserver

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type ErrorPayload struct {
	Code    apperror.Code `json:"code"`
	Details any           `json:"details,omitempty"`
}

type ErrorResponse struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Error   ErrorPayload `json:"error"`
}

func WriteError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		logger.Error("unhandled request error",
			"request_id", RequestIDFromContext(r.Context()),
			"error", err,
		)
		appErr = apperror.Internal(err)
	}

	status := statusFromCode(appErr.Code)
	if status >= http.StatusInternalServerError {
		logger.Error("request failed",
			"request_id", RequestIDFromContext(r.Context()),
			"code", appErr.Code,
			"error", appErr.Err,
		)
	}

	WriteJSON(w, status, ErrorResponse{
		Success: false,
		Message: appErr.Message,
		Error: ErrorPayload{
			Code:    appErr.Code,
			Details: appErr.Details,
		},
	})
}

func statusFromCode(code apperror.Code) int {
	switch code {
	case apperror.CodeValidation:
		return http.StatusBadRequest
	case apperror.CodeUnauthorized:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeRateLimited:
		return http.StatusTooManyRequests
	case apperror.CodeServiceUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
