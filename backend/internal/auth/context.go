package auth

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

type AuthContext struct {
	UserID       uuid.UUID
	PlatformRole string
}

func WithContext(ctx context.Context, authCtx AuthContext) context.Context {
	return context.WithValue(ctx, contextKey{}, authCtx)
}

func FromContext(ctx context.Context) (AuthContext, bool) {
	authCtx, ok := ctx.Value(contextKey{}).(AuthContext)
	return authCtx, ok
}
