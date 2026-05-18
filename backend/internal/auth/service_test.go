package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/platform/db"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/apperror"
)

type fakeDatabase struct{}

func (fakeDatabase) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (fakeDatabase) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (fakeDatabase) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

func (fakeDatabase) WithTx(ctx context.Context, fn func(tx db.Tx) error) error {
	return fn(fakeTx{})
}

type fakeTx struct{}

func (fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, nil
}

func (fakeTx) QueryRow(context.Context, string, ...any) pgx.Row {
	return nil
}

type fakeUsers struct {
	user       *User
	findErr    error
	createCall int
}

func (f *fakeUsers) Create(context.Context, db.Queryer, CreateUserParams) (*User, error) {
	f.createCall++
	if f.user != nil {
		return f.user, nil
	}
	return &User{
		ID:           uuid.New(),
		Name:         "Owner",
		Email:        "owner@example.com",
		PasswordHash: "hashed",
		PlatformRole: PlatformRoleUser,
		Status:       UserStatusActive,
	}, nil
}

func (f *fakeUsers) FindByEmail(context.Context, db.Queryer, string) (*User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.user, nil
}

func (f *fakeUsers) FindByID(context.Context, db.Queryer, uuid.UUID) (*User, error) {
	if f.findErr != nil {
		return nil, f.findErr
	}
	return f.user, nil
}

func (*fakeUsers) UpdateLastLogin(context.Context, db.Queryer, uuid.UUID, time.Time) error {
	return nil
}

type fakeRefreshTokens struct{}

func (fakeRefreshTokens) Create(context.Context, db.Queryer, CreateRefreshTokenParams) (*RefreshToken, error) {
	return &RefreshToken{ID: uuid.New()}, nil
}

func (fakeRefreshTokens) FindByHashForUpdate(context.Context, db.Queryer, string) (*RefreshToken, error) {
	return nil, ErrRefreshTokenNotFound
}

func (fakeRefreshTokens) RevokeAndReplace(context.Context, db.Queryer, uuid.UUID, uuid.UUID, time.Time) error {
	return nil
}

func (fakeRefreshTokens) RevokeByHashAndUserID(context.Context, db.Queryer, string, uuid.UUID, time.Time) error {
	return nil
}

type fakePasswords struct {
	compareErr error
}

func (fakePasswords) Hash(string) (string, error) {
	return "hashed", nil
}

func (f fakePasswords) Compare(string, string) error {
	return f.compareErr
}

type fakeAccessTokens struct{}

func (fakeAccessTokens) Generate(uuid.UUID, string) (string, error) {
	return "access", nil
}

type fakeRefreshTokenService struct{}

func (fakeRefreshTokenService) Generate() (string, error) {
	return "refresh", nil
}

func (fakeRefreshTokenService) Hash(raw string) string {
	return "hash:" + raw
}

func (fakeRefreshTokenService) TTL() time.Duration {
	return 30 * 24 * time.Hour
}

func TestRegisterRejectsWeakPasswordBeforePersistence(t *testing.T) {
	t.Parallel()

	users := &fakeUsers{}
	service := NewService(fakeDatabase{}, users, fakeRefreshTokens{}, fakePasswords{}, fakeAccessTokens{}, fakeRefreshTokenService{})

	_, err := service.Register(context.Background(), RegisterInput{
		Name:     "Owner",
		Email:    "owner@example.com",
		Password: "short",
	})
	if err == nil {
		t.Fatal("Register() expected validation error")
	}

	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.CodeValidation {
		t.Fatalf("Register() error = %#v, want validation app error", err)
	}
	if users.createCall != 0 {
		t.Fatalf("Create() calls = %d, want 0", users.createCall)
	}
}

func TestLoginUsesGenericMessageForUnknownEmailAndWrongPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		users     *fakeUsers
		passwords fakePasswords
	}{
		{
			name:      "unknown email",
			users:     &fakeUsers{findErr: ErrUserNotFound},
			passwords: fakePasswords{},
		},
		{
			name: "wrong password",
			users: &fakeUsers{user: &User{
				ID:           uuid.New(),
				Email:        "owner@example.com",
				PasswordHash: "hashed",
				PlatformRole: PlatformRoleUser,
				Status:       UserStatusActive,
			}},
			passwords: fakePasswords{compareErr: errors.New("wrong password")},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewService(fakeDatabase{}, tt.users, fakeRefreshTokens{}, tt.passwords, fakeAccessTokens{}, fakeRefreshTokenService{})

			_, err := service.Login(context.Background(), LoginInput{
				Email:    "owner@example.com",
				Password: "password123",
			})
			if err == nil {
				t.Fatal("Login() expected unauthorized error")
			}

			var appErr *apperror.AppError
			if !errors.As(err, &appErr) {
				t.Fatalf("Login() error = %#v, want app error", err)
			}
			if appErr.Code != apperror.CodeUnauthorized {
				t.Fatalf("Login() code = %s, want %s", appErr.Code, apperror.CodeUnauthorized)
			}
			if appErr.Message != "Email atau password salah." {
				t.Fatalf("Login() message = %q, want generic login failure", appErr.Message)
			}
		})
	}
}
