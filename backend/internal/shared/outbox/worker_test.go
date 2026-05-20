package outbox

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

func TestUnknownEventDoesNotCrashWorker(t *testing.T) {
	t.Parallel()

	event := Event{
		ID:            uuid.New(),
		TenantID:      uuid.New(),
		EventType:     "MysteryEvent",
		AggregateType: "order",
		AggregateID:   uuid.New(),
		Status:        StatusProcessing,
		Attempts:      0,
		AvailableAt:   time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
		CreatedAt:     time.Date(2026, 5, 21, 10, 0, 0, 0, time.UTC),
	}
	store := &fakeWorkerOutboxStore{pending: []Event{event}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	worker := NewWorker(nil, store, NewDefaultDispatcher(), logger, WorkerConfig{
		BatchSize:    1,
		MaxAttempts:  3,
		RetryDelay:   time.Minute,
		PollInterval: time.Second,
	})
	worker.now = func() time.Time {
		return time.Date(2026, 5, 21, 10, 1, 0, 0, time.UTC)
	}

	if err := worker.ProcessOnce(context.Background()); err != nil {
		t.Fatalf("ProcessOnce error = %v", err)
	}
	if len(store.failed) != 1 {
		t.Fatalf("failed len = %d, want 1", len(store.failed))
	}
	if len(store.succeeded) != 0 {
		t.Fatalf("succeeded len = %d, want 0", len(store.succeeded))
	}
	if store.failed[0].eventID != event.ID || store.failed[0].maxAttempts != 3 {
		t.Fatalf("failed mark = %#v", store.failed[0])
	}
}

type fakeWorkerOutboxStore struct {
	pending   []Event
	succeeded []uuid.UUID
	failed    []failedMark
}

type failedMark struct {
	eventID     uuid.UUID
	maxAttempts int
	retryAt     time.Time
}

func (f *fakeWorkerOutboxStore) FetchPending(
	context.Context,
	db.Queryer,
	int,
	int,
) ([]Event, error) {
	events := append([]Event(nil), f.pending...)
	f.pending = nil
	return events, nil
}

func (f *fakeWorkerOutboxStore) MarkSucceeded(
	_ context.Context,
	_ db.Queryer,
	eventID uuid.UUID,
) (*Event, error) {
	f.succeeded = append(f.succeeded, eventID)
	return &Event{ID: eventID, Status: StatusProcessed}, nil
}

func (f *fakeWorkerOutboxStore) MarkFailed(
	_ context.Context,
	_ db.Queryer,
	eventID uuid.UUID,
	maxAttempts int,
	retryAt time.Time,
) (*Event, error) {
	f.failed = append(f.failed, failedMark{eventID: eventID, maxAttempts: maxAttempts, retryAt: retryAt})
	return &Event{ID: eventID, Status: StatusPending, Attempts: 1}, nil
}
