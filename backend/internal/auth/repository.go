package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailAlreadyInUse = errors.New("email already in use")
)

type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

func (r *UserRepository) Create(ctx context.Context, q db.Queryer, params CreateUserParams) (*User, error) {
	const query = `
		INSERT INTO users (name, email, phone, password_hash)
		VALUES ($1, $2, NULLIF($3, ''), $4)
		RETURNING id, name, email, COALESCE(phone, ''), password_hash, platform_role, status
	`

	var user User
	err := q.QueryRow(
		ctx,
		query,
		params.Name,
		params.Email,
		params.Phone,
		params.PasswordHash,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.PlatformRole,
		&user.Status,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrEmailAlreadyInUse
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, q db.Queryer, email string) (*User, error) {
	const query = `
		SELECT id, name, email, COALESCE(phone, ''), password_hash, platform_role, status
		FROM users
		WHERE email = $1
		  AND deleted_at IS NULL
	`

	return scanUser(q.QueryRow(ctx, query, email))
}

func (r *UserRepository) FindByID(ctx context.Context, q db.Queryer, userID uuid.UUID) (*User, error) {
	const query = `
		SELECT id, name, email, COALESCE(phone, ''), password_hash, platform_role, status
		FROM users
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	return scanUser(q.QueryRow(ctx, query, userID))
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, q db.Queryer, userID uuid.UUID, at time.Time) error {
	const query = `
		UPDATE users
		SET last_login_at = $2,
		    updated_at = $2
		WHERE id = $1
		  AND deleted_at IS NULL
	`

	_, err := q.Exec(ctx, query, userID, at)
	return err
}

func scanUser(row pgx.Row) (*User, error) {
	var user User
	err := row.Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.PlatformRole,
		&user.Status,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
