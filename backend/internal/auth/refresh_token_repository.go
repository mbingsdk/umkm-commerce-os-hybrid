package auth

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
)

var ErrRefreshTokenNotFound = errors.New("refresh token not found")

type RefreshTokenRepository struct{}

func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{}
}

func (r *RefreshTokenRepository) Create(
	ctx context.Context,
	q db.Queryer,
	params CreateRefreshTokenParams,
) (*RefreshToken, error) {
	const query = `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at,
			ip_address,
			user_agent
		)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''))
		RETURNING id, user_id, expires_at, revoked_at IS NOT NULL
	`

	var refreshToken RefreshToken
	err := q.QueryRow(
		ctx,
		query,
		params.UserID,
		params.TokenHash,
		params.ExpiresAt,
		params.IPAddress,
		params.UserAgent,
	).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.ExpiresAt,
		&refreshToken.Revoked,
	)
	if err != nil {
		return nil, err
	}

	return &refreshToken, nil
}

func (r *RefreshTokenRepository) FindByHashForUpdate(
	ctx context.Context,
	q db.Queryer,
	tokenHash string,
) (*RefreshToken, error) {
	const query = `
		SELECT id, user_id, expires_at, revoked_at IS NOT NULL
		FROM refresh_tokens
		WHERE token_hash = $1
		FOR UPDATE
	`

	var refreshToken RefreshToken
	err := q.QueryRow(ctx, query, tokenHash).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.ExpiresAt,
		&refreshToken.Revoked,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	return &refreshToken, nil
}

func (r *RefreshTokenRepository) RevokeAndReplace(
	ctx context.Context,
	q db.Queryer,
	currentTokenID uuid.UUID,
	nextTokenID uuid.UUID,
	revokedAt time.Time,
) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = $2,
		    replaced_by_token_id = $3
		WHERE id = $1
		  AND revoked_at IS NULL
	`

	_, err := q.Exec(ctx, query, currentTokenID, revokedAt, nextTokenID)
	return err
}

func (r *RefreshTokenRepository) RevokeByHashAndUserID(
	ctx context.Context,
	q db.Queryer,
	tokenHash string,
	userID uuid.UUID,
	revokedAt time.Time,
) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = $3
		WHERE token_hash = $1
		  AND user_id = $2
		  AND revoked_at IS NULL
	`

	_, err := q.Exec(ctx, query, tokenHash, userID, revokedAt)
	return err
}
