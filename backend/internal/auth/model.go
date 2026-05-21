package auth

import (
	"time"

	"github.com/google/uuid"
)

const (
	PlatformRoleUser       = "user"
	PlatformRoleSuperAdmin = "super_admin"

	UserStatusActive = "active"
)

type User struct {
	ID           uuid.UUID
	Name         string
	Email        string
	Phone        string
	PasswordHash string
	PlatformRole string
	Status       string
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	Revoked   bool
}

type CreateUserParams struct {
	Name         string
	Email        string
	Phone        string
	PasswordHash string
}

type CreateRefreshTokenParams struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
}
