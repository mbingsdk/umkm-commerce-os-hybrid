package admin

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	PlatformRole string
	Status       string
}

type AuditEntry struct {
	ActorUserID uuid.UUID
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	BeforeData  any
	AfterData   any
	IPAddress   string
	UserAgent   string
}

type AuditLog struct {
	ID          uuid.UUID
	ActorUserID uuid.UUID
	Action      string
	TargetType  string
	TargetID    *uuid.UUID
	CreatedAt   time.Time
}
