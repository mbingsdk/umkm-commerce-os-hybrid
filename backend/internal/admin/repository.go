package admin

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var ErrAdminUserNotFound = errors.New("admin user not found")

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) FindUserByID(ctx context.Context, q db.Queryer, userID uuid.UUID) (*User, error) {
	const query = `
		SELECT id, name, email, platform_role, status
		FROM users
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	var user User
	if err := q.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PlatformRole,
		&user.Status,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAdminUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *Repository) CreateAuditLog(ctx context.Context, q db.Queryer, entry AuditEntry) (*AuditLog, error) {
	beforeData, err := marshalOptionalJSON(entry.BeforeData)
	if err != nil {
		return nil, err
	}
	afterData, err := marshalOptionalJSON(entry.AfterData)
	if err != nil {
		return nil, err
	}

	const query = `
		INSERT INTO admin_audit_logs (
			actor_user_id,
			action,
			target_type,
			target_id,
			before_data,
			after_data,
			ip_address,
			user_agent
		)
		VALUES (
			$1,
			$2,
			NULLIF($3, ''),
			$4,
			$5,
			$6,
			NULLIF($7, ''),
			NULLIF($8, '')
		)
		RETURNING id, actor_user_id, action, COALESCE(target_type, ''), target_id, created_at
	`

	var log AuditLog
	if err := q.QueryRow(
		ctx,
		query,
		entry.ActorUserID,
		entry.Action,
		entry.TargetType,
		entry.TargetID,
		beforeData,
		afterData,
		entry.IPAddress,
		entry.UserAgent,
	).Scan(
		&log.ID,
		&log.ActorUserID,
		&log.Action,
		&log.TargetType,
		&log.TargetID,
		&log.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &log, nil
}

func marshalOptionalJSON(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	return json.Marshal(value)
}
