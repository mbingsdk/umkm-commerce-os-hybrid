package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestJWTServiceGenerateAndParse(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 18, 10, 0, 0, 0, time.UTC)
	userID := uuid.New()
	service := NewJWTService("test-secret", 15*time.Minute).WithClock(func() time.Time {
		return now
	})

	raw, err := service.Generate(userID, "user")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	claims, err := service.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if claims.Subject != userID.String() {
		t.Fatalf("Subject = %q, want %q", claims.Subject, userID.String())
	}
	if claims.UserID != userID.String() {
		t.Fatalf("UserID = %q, want %q", claims.UserID, userID.String())
	}
	if claims.PlatformRole != "user" {
		t.Fatalf("PlatformRole = %q, want user", claims.PlatformRole)
	}
	if got := claims.ExpiresAt.Time; !got.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("ExpiresAt = %v, want %v", got, now.Add(15*time.Minute))
	}
}
