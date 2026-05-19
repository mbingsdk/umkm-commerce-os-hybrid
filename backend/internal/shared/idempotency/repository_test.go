package idempotency

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

func TestResolveExistingSameHashCanReplayCompletedResponse(t *testing.T) {
	t.Parallel()

	statusCode := 201
	state, err := ResolveExisting(Record{
		ID:           uuid.New(),
		TenantID:     uuid.New(),
		Scope:        ScopeCheckout,
		Key:          "checkout_123",
		RequestHash:  "same-hash",
		ResponseBody: []byte(`{"order_id":"order-1"}`),
		StatusCode:   &statusCode,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, "same-hash")
	if err != nil {
		t.Fatalf("ResolveExisting error = %v", err)
	}
	if !state.CanReplay {
		t.Fatal("ResolveExisting CanReplay = false, want true")
	}
	if state.StatusCode != statusCode {
		t.Fatalf("ResolveExisting StatusCode = %d, want %d", state.StatusCode, statusCode)
	}
	if string(state.ResponseBody) != `{"order_id":"order-1"}` {
		t.Fatalf("ResolveExisting ResponseBody = %s", state.ResponseBody)
	}
}

func TestResolveExistingDifferentHashReturnsIdempotencyConflict(t *testing.T) {
	t.Parallel()

	_, err := ResolveExisting(Record{
		ID:          uuid.New(),
		TenantID:    uuid.New(),
		Scope:       ScopeCheckout,
		Key:         "checkout_123",
		RequestHash: "hash-a",
	}, "hash-b")

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("ResolveExisting error type = %T, want *apperror.AppError", err)
	}
	if appErr.Code != apperror.CodeIdempotencyConflict {
		t.Fatalf("ResolveExisting code = %s, want %s", appErr.Code, apperror.CodeIdempotencyConflict)
	}
}

func TestRequestHashCanonicalizesJSONObjects(t *testing.T) {
	t.Parallel()

	left, err := RequestHash("POST", "/api/v1/public/stores/toko/checkout", []byte(`{"items":[{"quantity":2,"product_id":"p1"}],"customer":{"phone":"081","name":"Andi"}}`))
	if err != nil {
		t.Fatalf("RequestHash left error = %v", err)
	}

	right, err := RequestHash("post", "/api/v1/public/stores/toko/checkout", []byte(`{
		"customer": {"name": "Andi", "phone": "081"},
		"items": [{"product_id": "p1", "quantity": 2}]
	}`))
	if err != nil {
		t.Fatalf("RequestHash right error = %v", err)
	}

	if left != right {
		t.Fatalf("RequestHash mismatch: %s != %s", left, right)
	}
}
