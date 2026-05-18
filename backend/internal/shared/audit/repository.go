package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

type Entry struct {
	TenantID    uuid.UUID
	StoreID     *uuid.UUID
	ActorUserID *uuid.UUID
	Action      string
	EntityType  string
	EntityID    *uuid.UUID
	BeforeData  any
	AfterData   any
	Reason      string
	IPAddress   string
	UserAgent   string
}

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, q db.Queryer, entry Entry) error {
	beforeData, err := marshalJSON(entry.BeforeData)
	if err != nil {
		return err
	}

	afterData, err := marshalJSON(entry.AfterData)
	if err != nil {
		return err
	}

	const query = `
		INSERT INTO tenant_audit_logs (
			tenant_id,
			store_id,
			actor_user_id,
			action,
			entity_type,
			entity_id,
			before_data,
			after_data,
			reason,
			ip_address,
			user_agent
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			NULLIF($5, ''),
			$6,
			$7,
			$8,
			NULLIF($9, ''),
			NULLIF($10, ''),
			NULLIF($11, '')
		)
	`

	_, err = q.Exec(
		ctx,
		query,
		entry.TenantID,
		entry.StoreID,
		entry.ActorUserID,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		beforeData,
		afterData,
		entry.Reason,
		entry.IPAddress,
		entry.UserAgent,
	)
	return err
}

func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}

	return json.Marshal(value)
}
