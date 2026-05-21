package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestInsertUsesTransactionQueryerAndReturnsEvent(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	aggregateID := uuid.New()
	eventID := uuid.New()
	availableAt := time.Date(2026, 5, 19, 10, 0, 0, 0, time.UTC)
	fakeTx := &fakeOutboxTx{
		row: fakeOutboxRow{
			values: []any{
				eventID,
				tenantID,
				"OrderCreated",
				"order",
				aggregateID,
				[]byte(`{"order_id":"` + aggregateID.String() + `"}`),
				StatusPending,
				0,
				availableAt,
				(*time.Time)(nil),
				availableAt,
			},
		},
	}

	event, err := NewRepository().Insert(context.Background(), fakeTx, InsertEventParams{
		TenantID:      tenantID,
		EventType:     "OrderCreated",
		AggregateType: "order",
		AggregateID:   aggregateID,
		Payload:       json.RawMessage(`{"order_id":"` + aggregateID.String() + `"}`),
		AvailableAt:   &availableAt,
	})
	if err != nil {
		t.Fatalf("Insert error = %v", err)
	}
	if event.ID != eventID || event.TenantID != tenantID || event.AggregateID != aggregateID {
		t.Fatalf("Insert event = %+v", event)
	}
	if event.Status != StatusPending || event.Attempts != 0 {
		t.Fatalf("Insert status/attempts = %s/%d", event.Status, event.Attempts)
	}
	if !strings.Contains(fakeTx.query, "INSERT INTO outbox_events") {
		t.Fatalf("Insert query = %s", fakeTx.query)
	}
	if len(fakeTx.args) != 6 {
		t.Fatalf("Insert args len = %d, want 6", len(fakeTx.args))
	}
}

func TestFetchPendingUsesSkipLockedAndMarksProcessing(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	aggregateID := uuid.New()
	eventID := uuid.New()
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	fakeTx := &fakeOutboxTx{
		rows: &fakeOutboxRows{
			values: [][]any{
				eventValues(eventID, tenantID, "NotificationRequested", "order", aggregateID, StatusProcessing, 0, now, nil, now),
			},
		},
	}

	events, err := NewRepository().FetchPending(context.Background(), fakeTx, 10, 5)
	if err != nil {
		t.Fatalf("FetchPending error = %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("events len = %d, want 1", len(events))
	}
	if events[0].ID != eventID || events[0].Status != StatusProcessing {
		t.Fatalf("event = %+v", events[0])
	}
	if !strings.Contains(fakeTx.query, "FOR UPDATE SKIP LOCKED") {
		t.Fatalf("FetchPending query missing skip locked: %s", fakeTx.query)
	}
	if !strings.Contains(fakeTx.query, "SET status = 'processing'") {
		t.Fatalf("FetchPending query missing processing mark: %s", fakeTx.query)
	}
	if len(fakeTx.args) != 2 || fakeTx.args[0] != 10 || fakeTx.args[1] != 5 {
		t.Fatalf("FetchPending args = %#v", fakeTx.args)
	}
}

func TestMarkSucceededMarksProcessed(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	aggregateID := uuid.New()
	eventID := uuid.New()
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	processedAt := now.Add(time.Second)
	fakeTx := &fakeOutboxTx{
		row: fakeOutboxRow{
			values: eventValues(eventID, tenantID, "NotificationRequested", "order", aggregateID, StatusProcessed, 0, now, &processedAt, now),
		},
	}

	event, err := NewRepository().MarkSucceeded(context.Background(), fakeTx, eventID)
	if err != nil {
		t.Fatalf("MarkSucceeded error = %v", err)
	}
	if event.Status != StatusProcessed || event.ProcessedAt == nil {
		t.Fatalf("event = %+v, want processed with processed_at", event)
	}
	if !strings.Contains(fakeTx.query, "status = 'processed'") {
		t.Fatalf("MarkSucceeded query = %s", fakeTx.query)
	}
}

func TestMarkFailedIncrementsAttempts(t *testing.T) {
	t.Parallel()

	tenantID := uuid.New()
	aggregateID := uuid.New()
	eventID := uuid.New()
	now := time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC)
	retryAt := now.Add(30 * time.Second)
	fakeTx := &fakeOutboxTx{
		row: fakeOutboxRow{
			values: eventValues(eventID, tenantID, "MysteryEvent", "order", aggregateID, StatusPending, 2, retryAt, nil, now),
		},
	}

	event, err := NewRepository().MarkFailed(context.Background(), fakeTx, eventID, 5, retryAt)
	if err != nil {
		t.Fatalf("MarkFailed error = %v", err)
	}
	if event.Attempts != 2 || event.Status != StatusPending {
		t.Fatalf("event = %+v, want pending retry with attempts 2", event)
	}
	if !strings.Contains(fakeTx.query, "attempts = attempts + 1") {
		t.Fatalf("MarkFailed query = %s", fakeTx.query)
	}
	if len(fakeTx.args) != 3 || fakeTx.args[0] != eventID || fakeTx.args[1] != 5 || fakeTx.args[2] != retryAt {
		t.Fatalf("MarkFailed args = %#v", fakeTx.args)
	}
}

type fakeOutboxTx struct {
	query string
	args  []any
	row   fakeOutboxRow
	rows  pgx.Rows
}

func (f *fakeOutboxTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (f *fakeOutboxTx) Query(_ context.Context, query string, args ...any) (pgx.Rows, error) {
	if f.rows == nil {
		return nil, errors.New("unexpected Query")
	}
	f.query = query
	f.args = args
	return f.rows, nil
}

func (f *fakeOutboxTx) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	f.query = query
	f.args = args
	return f.row
}

type fakeOutboxRow struct {
	values []any
}

func (r fakeOutboxRow) Scan(dest ...any) error {
	for i := range dest {
		switch target := dest[i].(type) {
		case *uuid.UUID:
			*target = r.values[i].(uuid.UUID)
		case *uuid.NullUUID:
			if r.values[i] == nil {
				*target = uuid.NullUUID{}
			} else {
				*target = uuid.NullUUID{UUID: r.values[i].(uuid.UUID), Valid: true}
			}
		case *string:
			*target = r.values[i].(string)
		case *json.RawMessage:
			*target = append((*target)[0:0], r.values[i].([]byte)...)
		case *int:
			*target = r.values[i].(int)
		case *time.Time:
			*target = r.values[i].(time.Time)
		case **time.Time:
			*target = r.values[i].(*time.Time)
		default:
			return errors.New("unsupported scan target")
		}
	}
	return nil
}

type fakeOutboxRows struct {
	values [][]any
	index  int
}

func (r *fakeOutboxRows) Close() {}

func (r *fakeOutboxRows) Err() error { return nil }

func (r *fakeOutboxRows) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

func (r *fakeOutboxRows) FieldDescriptions() []pgconn.FieldDescription { return nil }

func (r *fakeOutboxRows) Next() bool {
	if r.index >= len(r.values) {
		return false
	}
	r.index++
	return true
}

func (r *fakeOutboxRows) Scan(dest ...any) error {
	return fakeOutboxRow{values: r.values[r.index-1]}.Scan(dest...)
}

func (r *fakeOutboxRows) Values() ([]any, error) { return r.values[r.index-1], nil }

func (r *fakeOutboxRows) RawValues() [][]byte { return nil }

func (r *fakeOutboxRows) Conn() *pgx.Conn { return nil }

func eventValues(
	eventID uuid.UUID,
	tenantID uuid.UUID,
	eventType string,
	aggregateType string,
	aggregateID uuid.UUID,
	status string,
	attempts int,
	availableAt time.Time,
	processedAt *time.Time,
	createdAt time.Time,
) []any {
	return []any{
		eventID,
		tenantID,
		eventType,
		aggregateType,
		aggregateID,
		[]byte(`{"safe":true}`),
		status,
		attempts,
		availableAt,
		processedAt,
		createdAt,
	}
}
