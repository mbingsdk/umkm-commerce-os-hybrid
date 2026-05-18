package auth

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type database interface {
	db.Queryer
	WithTx(ctx context.Context, fn func(tx db.Tx) error) error
}

type userStore interface {
	Create(ctx context.Context, q db.Queryer, params CreateUserParams) (*User, error)
	FindByEmail(ctx context.Context, q db.Queryer, email string) (*User, error)
	FindByID(ctx context.Context, q db.Queryer, userID uuid.UUID) (*User, error)
	UpdateLastLogin(ctx context.Context, q db.Queryer, userID uuid.UUID, at time.Time) error
}

type refreshTokenStore interface {
	Create(ctx context.Context, q db.Queryer, params CreateRefreshTokenParams) (*RefreshToken, error)
	FindByHashForUpdate(ctx context.Context, q db.Queryer, tokenHash string) (*RefreshToken, error)
	RevokeAndReplace(ctx context.Context, q db.Queryer, currentTokenID uuid.UUID, nextTokenID uuid.UUID, revokedAt time.Time) error
	RevokeByHashAndUserID(ctx context.Context, q db.Queryer, tokenHash string, userID uuid.UUID, revokedAt time.Time) error
}

type passwordHasher interface {
	Hash(value string) (string, error)
	Compare(hash, value string) error
}

type accessTokenService interface {
	Generate(userID uuid.UUID, platformRole string) (string, error)
}

type refreshTokenService interface {
	Generate() (string, error)
	Hash(raw string) string
	TTL() time.Duration
}

type Service struct {
	db            database
	users         userStore
	refreshTokens refreshTokenStore
	passwords     passwordHasher
	accessTokens  accessTokenService
	refresh       refreshTokenService
	now           func() time.Time
}

func NewService(
	database database,
	users userStore,
	refreshTokens refreshTokenStore,
	passwords passwordHasher,
	accessTokens accessTokenService,
	refresh refreshTokenService,
) *Service {
	return &Service{
		db:            database,
		users:         users,
		refreshTokens: refreshTokens,
		passwords:     passwords,
		accessTokens:  accessTokens,
		refresh:       refresh,
		now:           time.Now,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*AuthResponse, error) {
	normalized, err := validateRegister(input)
	if err != nil {
		return nil, err
	}

	passwordHash, err := s.passwords.Hash(normalized.Password)
	if err != nil {
		return nil, apperror.Internal(err)
	}

	var response *AuthResponse
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		user, err := s.users.Create(ctx, tx, CreateUserParams{
			Name:         normalized.Name,
			Email:        normalized.Email,
			Phone:        normalized.Phone,
			PasswordHash: passwordHash,
		})
		if err != nil {
			if errors.Is(err, ErrEmailAlreadyInUse) {
				return apperror.Validation("Validation failed", []map[string]string{
					{"field": "email", "message": "Email is already registered"},
				})
			}
			return apperror.Internal(err)
		}

		accessToken, refreshToken, err := s.issueTokens(ctx, tx, user, normalized.IPAddress, normalized.UserAgent)
		if err != nil {
			return err
		}

		response = &AuthResponse{
			User:         NewUserResponse(user),
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthResponse, error) {
	normalized, err := validateLogin(input)
	if err != nil {
		return nil, err
	}

	user, err := s.users.FindByEmail(ctx, s.db, normalized.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, invalidCredentials()
		}
		return nil, apperror.Internal(err)
	}

	if user.Status != UserStatusActive {
		return nil, apperror.Forbidden("User account is not active")
	}

	if err := s.passwords.Compare(user.PasswordHash, normalized.Password); err != nil {
		return nil, invalidCredentials()
	}

	var response *AuthResponse
	err = s.db.WithTx(ctx, func(tx db.Tx) error {
		now := s.now().UTC()
		if err := s.users.UpdateLastLogin(ctx, tx, user.ID, now); err != nil {
			return apperror.Internal(err)
		}

		accessToken, refreshToken, err := s.issueTokens(ctx, tx, user, normalized.IPAddress, normalized.UserAgent)
		if err != nil {
			return err
		}

		response = &AuthResponse{
			User:         NewUserResponse(user),
			Tenants:      []TenantResponse{},
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*TokenResponse, error) {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil, apperror.Validation("Validation failed", []map[string]string{
			{"field": "refresh_token", "message": "Refresh token is required"},
		})
	}

	tokenHash := s.refresh.Hash(rawRefreshToken)
	var response *TokenResponse

	err := s.db.WithTx(ctx, func(tx db.Tx) error {
		currentToken, err := s.refreshTokens.FindByHashForUpdate(ctx, tx, tokenHash)
		if err != nil {
			if errors.Is(err, ErrRefreshTokenNotFound) {
				return invalidRefreshToken()
			}
			return apperror.Internal(err)
		}

		now := s.now().UTC()
		if currentToken.Revoked || !now.Before(currentToken.ExpiresAt) {
			return invalidRefreshToken()
		}

		user, err := s.users.FindByID(ctx, tx, currentToken.UserID)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				return invalidRefreshToken()
			}
			return apperror.Internal(err)
		}
		if user.Status != UserStatusActive {
			return apperror.Forbidden("User account is not active")
		}

		accessToken, err := s.accessTokens.Generate(user.ID, user.PlatformRole)
		if err != nil {
			return apperror.Internal(err)
		}

		nextRawRefreshToken, err := s.refresh.Generate()
		if err != nil {
			return apperror.Internal(err)
		}

		nextToken, err := s.refreshTokens.Create(ctx, tx, CreateRefreshTokenParams{
			UserID:    user.ID,
			TokenHash: s.refresh.Hash(nextRawRefreshToken),
			ExpiresAt: now.Add(s.refresh.TTL()),
		})
		if err != nil {
			return apperror.Internal(err)
		}

		if err := s.refreshTokens.RevokeAndReplace(ctx, tx, currentToken.ID, nextToken.ID, now); err != nil {
			return apperror.Internal(err)
		}

		response = &TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: nextRawRefreshToken,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (s *Service) Logout(ctx context.Context, userID uuid.UUID, rawRefreshToken string) error {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return apperror.Validation("Validation failed", []map[string]string{
			{"field": "refresh_token", "message": "Refresh token is required"},
		})
	}

	tokenHash := s.refresh.Hash(rawRefreshToken)
	return s.db.WithTx(ctx, func(tx db.Tx) error {
		if err := s.refreshTokens.RevokeByHashAndUserID(ctx, tx, tokenHash, userID, s.now().UTC()); err != nil {
			return apperror.Internal(err)
		}

		return nil
	})
}

func (s *Service) Me(ctx context.Context, userID uuid.UUID) (*MeResponse, error) {
	user, err := s.users.FindByID(ctx, s.db, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, apperror.Unauthorized("Invalid access token")
		}
		return nil, apperror.Internal(err)
	}
	if user.Status != UserStatusActive {
		return nil, apperror.Forbidden("User account is not active")
	}

	return &MeResponse{
		User:    NewUserResponse(user),
		Tenants: []TenantResponse{},
	}, nil
}

func (s *Service) issueTokens(
	ctx context.Context,
	q db.Queryer,
	user *User,
	ipAddress string,
	userAgent string,
) (string, string, error) {
	accessToken, err := s.accessTokens.Generate(user.ID, user.PlatformRole)
	if err != nil {
		return "", "", apperror.Internal(err)
	}

	rawRefreshToken, err := s.refresh.Generate()
	if err != nil {
		return "", "", apperror.Internal(err)
	}

	if _, err := s.refreshTokens.Create(ctx, q, CreateRefreshTokenParams{
		UserID:    user.ID,
		TokenHash: s.refresh.Hash(rawRefreshToken),
		ExpiresAt: s.now().UTC().Add(s.refresh.TTL()),
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}); err != nil {
		return "", "", apperror.Internal(err)
	}

	return accessToken, rawRefreshToken, nil
}

func validateRegister(input RegisterInput) (RegisterInput, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = normalizeEmail(input.Email)
	input.Phone = normalizePhone(input.Phone)
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)

	var details []map[string]string
	if input.Name == "" {
		details = append(details, map[string]string{"field": "name", "message": "Name is required"})
	}
	if !isValidEmail(input.Email) {
		details = append(details, map[string]string{"field": "email", "message": "Email is invalid"})
	}
	if len(input.Password) < 8 {
		details = append(details, map[string]string{"field": "password", "message": "Password must be at least 8 characters"})
	}

	if len(details) > 0 {
		return RegisterInput{}, apperror.Validation("Validation failed", details)
	}

	return input, nil
}

func validateLogin(input LoginInput) (LoginInput, error) {
	input.Email = normalizeEmail(input.Email)
	input.IPAddress = strings.TrimSpace(input.IPAddress)
	input.UserAgent = strings.TrimSpace(input.UserAgent)

	var details []map[string]string
	if !isValidEmail(input.Email) {
		details = append(details, map[string]string{"field": "email", "message": "Email is invalid"})
	}
	if strings.TrimSpace(input.Password) == "" {
		details = append(details, map[string]string{"field": "password", "message": "Password is required"})
	}

	if len(details) > 0 {
		return LoginInput{}, apperror.Validation("Validation failed", details)
	}

	return input, nil
}

func normalizeEmail(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizePhone(value string) string {
	value = strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(value))
	switch {
	case strings.HasPrefix(value, "+62"):
		return strings.TrimPrefix(value, "+")
	case strings.HasPrefix(value, "0"):
		return "62" + strings.TrimPrefix(value, "0")
	default:
		return value
	}
}

func isValidEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value
}

func invalidCredentials() error {
	return apperror.Unauthorized("Email atau password salah.")
}

func invalidRefreshToken() error {
	return apperror.Unauthorized("Invalid refresh token")
}
