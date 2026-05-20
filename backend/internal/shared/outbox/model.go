package outbox

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusProcessed  = "processed"
	StatusFailed     = "failed"
)

const (
	EventNotificationRequested       = "NotificationRequested"
	EventFinanceSummaryUpdateRequest = "FinanceSummaryUpdateRequested"
)

type Event struct {
	ID            uuid.UUID
	TenantID      uuid.UUID
	EventType     string
	AggregateType string
	AggregateID   uuid.UUID
	Payload       json.RawMessage
	Status        string
	Attempts      int
	AvailableAt   time.Time
	ProcessedAt   *time.Time
	CreatedAt     time.Time
}

type InsertEventParams struct {
	TenantID      uuid.UUID
	EventType     string
	AggregateType string
	AggregateID   uuid.UUID
	Payload       json.RawMessage
	AvailableAt   *time.Time
}
