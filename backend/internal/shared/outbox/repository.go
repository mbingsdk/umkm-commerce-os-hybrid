package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Insert(ctx context.Context, q db.Queryer, params InsertEventParams) (*Event, error) {
	availableAt := time.Now().UTC()
	if params.AvailableAt != nil {
		availableAt = *params.AvailableAt
	}

	payload := params.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	const query = `
		INSERT INTO outbox_events (
			tenant_id,
			event_type,
			aggregate_type,
			aggregate_id,
			payload,
			available_at
		)
		VALUES ($1, $2, $3, $4, $5::jsonb, $6)
		RETURNING
			id,
			tenant_id,
			event_type,
			aggregate_type,
			aggregate_id,
			payload,
			status,
			attempts,
			available_at,
			processed_at,
			created_at
	`

	var event Event
	if err := q.QueryRow(
		ctx,
		query,
		params.TenantID,
		params.EventType,
		params.AggregateType,
		params.AggregateID,
		string(payload),
		availableAt,
	).Scan(
		&event.ID,
		&event.TenantID,
		&event.EventType,
		&event.AggregateType,
		&event.AggregateID,
		&event.Payload,
		&event.Status,
		&event.Attempts,
		&event.AvailableAt,
		&event.ProcessedAt,
		&event.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &event, nil
}
