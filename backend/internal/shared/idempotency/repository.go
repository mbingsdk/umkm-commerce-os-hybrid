package idempotency

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Begin(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	scope string,
	key string,
	requestHash string,
	lockedUntil time.Time,
) (*State, error) {
	record, created, err := r.CreateProcessing(ctx, q, tenantID, scope, key, requestHash, lockedUntil)
	if err != nil {
		return nil, err
	}
	if created {
		return &State{
			Record:       record,
			Created:      true,
			IsProcessing: true,
			LockedUntil:  record.LockedUntil,
		}, nil
	}

	record, err = r.Find(ctx, q, tenantID, scope, key)
	if err != nil {
		return nil, err
	}

	return ResolveExisting(*record, requestHash)
}

func (r *Repository) Find(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	scope string,
	key string,
) (*Record, error) {
	const query = `
		SELECT
			id,
			tenant_id,
			scope,
			idempotency_key,
			request_hash,
			response_body,
			status_code,
			locked_until,
			created_at,
			updated_at
		FROM idempotency_keys
		WHERE tenant_id = $1
		  AND scope = $2
		  AND idempotency_key = $3
		FOR UPDATE
	`

	record, err := scanRecord(q.QueryRow(ctx, query, tenantID, scope, key))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	return record, nil
}

func (r *Repository) CreateProcessing(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	scope string,
	key string,
	requestHash string,
	lockedUntil time.Time,
) (*Record, bool, error) {
	const query = `
		INSERT INTO idempotency_keys (
			tenant_id,
			scope,
			idempotency_key,
			request_hash,
			locked_until
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, scope, idempotency_key) DO NOTHING
		RETURNING
			id,
			tenant_id,
			scope,
			idempotency_key,
			request_hash,
			response_body,
			status_code,
			locked_until,
			created_at,
			updated_at
	`

	record, err := scanRecord(q.QueryRow(ctx, query, tenantID, scope, key, requestHash, lockedUntil))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return record, true, nil
}

func (r *Repository) SaveCompletedResponse(
	ctx context.Context,
	q db.Queryer,
	tenantID uuid.UUID,
	scope string,
	key string,
	statusCode int,
	responseBody json.RawMessage,
) error {
	const query = `
		UPDATE idempotency_keys
		SET response_body = $4::jsonb,
		    status_code = $5,
		    locked_until = NULL,
		    updated_at = now()
		WHERE tenant_id = $1
		  AND scope = $2
		  AND idempotency_key = $3
	`

	tag, err := q.Exec(ctx, query, tenantID, scope, key, string(responseBody), statusCode)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrKeyNotFound
	}
	return nil
}

func ResolveExisting(record Record, requestHash string) (*State, error) {
	if record.RequestHash != requestHash {
		return nil, apperror.IdempotencyConflict("Idempotency key was already used with a different request")
	}

	if record.IsCompleted() {
		return &State{
			Record:       &record,
			CanReplay:    true,
			ResponseBody: record.ResponseBody,
			StatusCode:   *record.StatusCode,
		}, nil
	}

	return &State{
		Record:       &record,
		IsProcessing: true,
		LockedUntil:  record.LockedUntil,
	}, nil
}

func scanRecord(row pgx.Row) (*Record, error) {
	var record Record
	if err := row.Scan(
		&record.ID,
		&record.TenantID,
		&record.Scope,
		&record.Key,
		&record.RequestHash,
		&record.ResponseBody,
		&record.StatusCode,
		&record.LockedUntil,
		&record.CreatedAt,
		&record.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &record, nil
}
