package httpserver

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func DecodeJSON(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dest); err != nil {
		return apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": err.Error()},
		})
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return apperror.Validation("Invalid JSON payload", []map[string]string{
			{"field": "body", "message": "request body must contain a single JSON value"},
		})
	}

	return nil
}
