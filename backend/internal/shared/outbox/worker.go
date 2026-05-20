package outbox

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

const (
	defaultPollInterval = 5 * time.Second
	defaultBatchSize    = 10
	defaultMaxAttempts  = 5
	defaultRetryDelay   = 30 * time.Second
)

type pollingStore interface {
	FetchPending(context.Context, db.Queryer, int, int) ([]Event, error)
	MarkSucceeded(context.Context, db.Queryer, uuid.UUID) (*Event, error)
	MarkFailed(context.Context, db.Queryer, uuid.UUID, int, time.Time) (*Event, error)
}

type WorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int
	MaxAttempts  int
	RetryDelay   time.Duration
}

type Worker struct {
	db         db.Queryer
	events     pollingStore
	dispatcher *Dispatcher
	logger     *slog.Logger
	config     WorkerConfig
	now        func() time.Time
}

func NewWorker(database db.Queryer, events pollingStore, dispatcher *Dispatcher, logger *slog.Logger, config WorkerConfig) *Worker {
	if dispatcher == nil {
		dispatcher = NewDefaultDispatcher()
	}
	if logger == nil {
		logger = slog.Default()
	}

	return &Worker{
		db:         database,
		events:     events,
		dispatcher: dispatcher,
		logger:     logger,
		config:     normalizeWorkerConfig(config),
		now:        time.Now,
	}
}

func (w *Worker) Run(ctx context.Context) error {
	w.logger.Info("outbox worker started",
		"poll_interval_ms", w.config.PollInterval.Milliseconds(),
		"batch_size", w.config.BatchSize,
		"max_attempts", w.config.MaxAttempts,
	)

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		if err := w.ProcessOnce(ctx); err != nil {
			w.logger.Error("outbox poll failed", "error", err)
		}

		select {
		case <-ctx.Done():
			w.logger.Info("outbox worker stopped")
			return nil
		case <-ticker.C:
		}
	}
}

func (w *Worker) ProcessOnce(ctx context.Context) error {
	events, err := w.events.FetchPending(ctx, w.db, w.config.BatchSize, w.config.MaxAttempts)
	if err != nil {
		return err
	}

	for _, event := range events {
		if ctx.Err() != nil {
			return nil
		}
		if err := w.processEvent(ctx, event); err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) processEvent(ctx context.Context, event Event) error {
	attrs := eventLogAttrs(event)
	w.logger.Info("outbox event processing", attrs...)

	if err := w.dispatcher.Handle(ctx, event); err != nil {
		retryAt := w.now().UTC().Add(w.config.RetryDelay)
		updated, markErr := w.events.MarkFailed(ctx, w.db, event.ID, w.config.MaxAttempts, retryAt)
		if markErr != nil {
			return markErr
		}

		levelAttrs := append(attrs,
			"status", updated.Status,
			"attempts", updated.Attempts,
			"error", err,
		)
		if updated.Status == StatusFailed {
			w.logger.Error("outbox event failed permanently", levelAttrs...)
		} else {
			w.logger.Warn("outbox event scheduled for retry", append(levelAttrs, "retry_at", retryAt.Format(time.RFC3339))...)
		}
		return nil
	}

	updated, err := w.events.MarkSucceeded(ctx, w.db, event.ID)
	if err != nil {
		return err
	}
	w.logger.Info("outbox event processed", append(attrs, "status", updated.Status)...)
	return nil
}

func eventLogAttrs(event Event) []any {
	return []any{
		"event_id", event.ID.String(),
		"event_type", event.EventType,
		"tenant_id", event.TenantID.String(),
		"aggregate_type", event.AggregateType,
		"aggregate_id", event.AggregateID.String(),
		"attempts", event.Attempts,
	}
}

func normalizeWorkerConfig(config WorkerConfig) WorkerConfig {
	if config.PollInterval <= 0 {
		config.PollInterval = defaultPollInterval
	}
	if config.BatchSize <= 0 {
		config.BatchSize = defaultBatchSize
	}
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = defaultMaxAttempts
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = defaultRetryDelay
	}
	return config
}
