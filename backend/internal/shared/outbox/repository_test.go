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

type fakeOutboxTx struct {
	query string
	args  []any
	row   fakeOutboxRow
}

func (f *fakeOutboxTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, errors.New("unexpected Exec")
}

func (f *fakeOutboxTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, errors.New("unexpected Query")
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
