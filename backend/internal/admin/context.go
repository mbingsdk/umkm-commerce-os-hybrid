package admin

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

type Context struct {
	UserID       uuid.UUID
	Name         string
	Email        string
	PlatformRole string
}

func withContext(ctx context.Context, adminCtx Context) context.Context {
	return context.WithValue(ctx, contextKey{}, adminCtx)
}

func FromContext(ctx context.Context) (Context, bool) {
	adminCtx, ok := ctx.Value(contextKey{}).(Context)
	return adminCtx, ok
}
