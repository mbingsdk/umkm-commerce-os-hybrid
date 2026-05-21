package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
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

	var tenantID any = params.TenantID
	if params.TenantID == uuid.Nil {
		tenantID = nil
	}

	return scanEvent(q.QueryRow(
		ctx,
		query,
		tenantID,
		params.EventType,
		params.AggregateType,
		params.AggregateID,
		string(payload),
		availableAt,
	))
}

func (r *Repository) FetchPending(ctx context.Context, q db.Queryer, limit int, maxAttempts int) ([]Event, error) {
	const query = `
		WITH picked AS (
			SELECT id
			FROM outbox_events
			WHERE status = 'pending'
			  AND attempts < $2
			  AND available_at <= now()
			ORDER BY available_at ASC, created_at ASC
			LIMIT $1
			FOR UPDATE SKIP LOCKED
		)
		UPDATE outbox_events e
		SET status = 'processing'
		FROM picked
		WHERE e.id = picked.id
		RETURNING
			e.id,
			e.tenant_id,
			e.event_type,
			e.aggregate_type,
			e.aggregate_id,
			e.payload,
			e.status,
			e.attempts,
			e.available_at,
			e.processed_at,
			e.created_at
	`

	rows, err := q.Query(ctx, query, limit, maxAttempts)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (r *Repository) MarkSucceeded(ctx context.Context, q db.Queryer, eventID uuid.UUID) (*Event, error) {
	const query = `
		UPDATE outbox_events
		SET status = 'processed',
		    processed_at = now()
		WHERE id = $1
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

	return scanEvent(q.QueryRow(ctx, query, eventID))
}

func (r *Repository) MarkFailed(
	ctx context.Context,
	q db.Queryer,
	eventID uuid.UUID,
	maxAttempts int,
	retryAt time.Time,
) (*Event, error) {
	const query = `
		UPDATE outbox_events
		SET attempts = attempts + 1,
		    status = CASE
				WHEN attempts + 1 >= $2 THEN 'failed'
				ELSE 'pending'
		    END,
		    available_at = CASE
				WHEN attempts + 1 >= $2 THEN available_at
				ELSE $3
		    END,
		    processed_at = NULL
		WHERE id = $1
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

	return scanEvent(q.QueryRow(ctx, query, eventID, maxAttempts, retryAt))
}

func scanEvent(row interface {
	Scan(dest ...any) error
}) (*Event, error) {
	var event Event
	var tenantID uuid.NullUUID
	if err := row.Scan(
		&event.ID,
		&tenantID,
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
	if tenantID.Valid {
		event.TenantID = tenantID.UUID
	}
	return &event, nil
}

func scanEvents(rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
}) ([]Event, error) {
	events := make([]Event, 0)
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, *event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}
