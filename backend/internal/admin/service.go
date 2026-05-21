package admin

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/auth"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type database interface {
	db.Queryer
}

type store interface {
	FindUserByID(ctx context.Context, q db.Queryer, userID uuid.UUID) (*User, error)
	CreateAuditLog(ctx context.Context, q db.Queryer, entry AuditEntry) (*AuditLog, error)
}

type Service struct {
	db   database
	repo store
}

func NewService(database database, repo store) *Service {
	return &Service{db: database, repo: repo}
}

func (s *Service) ValidateSuperAdmin(ctx context.Context, userID uuid.UUID) (Context, error) {
	user, err := s.repo.FindUserByID(ctx, s.db, userID)
	if err != nil {
		if errors.Is(err, ErrAdminUserNotFound) {
			return Context{}, apperror.Unauthorized("Invalid access token")
		}
		return Context{}, apperror.Internal(err)
	}

	if user.Status != auth.UserStatusActive {
		return Context{}, apperror.Forbidden("User account is not active")
	}
	if user.PlatformRole != auth.PlatformRoleSuperAdmin {
		return Context{}, apperror.Forbidden("Super admin access required")
	}

	return Context{
		UserID:       user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PlatformRole: user.PlatformRole,
	}, nil
}

func (s *Service) RecordAudit(ctx context.Context, entry AuditEntry) (*AuditLog, error) {
	entry.Action = strings.TrimSpace(entry.Action)
	entry.TargetType = strings.TrimSpace(entry.TargetType)
	entry.IPAddress = strings.TrimSpace(entry.IPAddress)
	entry.UserAgent = strings.TrimSpace(entry.UserAgent)

	if entry.ActorUserID == uuid.Nil {
		return nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "actor_user_id", "message": "actor_user_id is required"},
		})
	}
	if entry.Action == "" {
		return nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "action", "message": "action is required"},
		})
	}

	log, err := s.repo.CreateAuditLog(ctx, s.db, entry)
	if err != nil {
		return nil, apperror.Internal(err)
	}
	return log, nil
}
