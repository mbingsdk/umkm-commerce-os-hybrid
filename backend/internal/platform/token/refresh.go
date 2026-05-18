package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"time"
)

type RefreshTokenService struct {
	ttl time.Duration
}

func NewRefreshTokenService(ttl time.Duration) RefreshTokenService {
	return RefreshTokenService{ttl: ttl}
}

func (s RefreshTokenService) Generate() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func (s RefreshTokenService) Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s RefreshTokenService) TTL() time.Duration {
	return s.ttl
}
