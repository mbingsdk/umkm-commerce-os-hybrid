package idempotency

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	ScopeCheckout            = "checkout"
	ScopePOS                 = "pos"
	ScopePaymentConfirmation = "payment_confirmation"
)

var ErrKeyNotFound = errKeyNotFound{}

type errKeyNotFound struct{}

func (errKeyNotFound) Error() string {
	return "idempotency key not found"
}

type Record struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Scope        string
	Key          string
	RequestHash  string
	ResponseBody json.RawMessage
	StatusCode   *int
	LockedUntil  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r Record) IsCompleted() bool {
	return len(r.ResponseBody) > 0 && r.StatusCode != nil
}

type State struct {
	Record       *Record
	Created      bool
	CanReplay    bool
	IsProcessing bool
	ResponseBody json.RawMessage
	StatusCode   int
	LockedUntil  *time.Time
}
