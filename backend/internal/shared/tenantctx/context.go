package tenantctx

import (
	"context"

	"github.com/google/uuid"
)

type contextKey struct{}

type TenantContext struct {
	TenantID uuid.UUID
	StoreID  uuid.UUID
	UserID   uuid.UUID
	Role     string
}

func WithContext(ctx context.Context, tenantCtx TenantContext) context.Context {
	return context.WithValue(ctx, contextKey{}, tenantCtx)
}

func FromContext(ctx context.Context) (TenantContext, bool) {
	tenantCtx, ok := ctx.Value(contextKey{}).(TenantContext)
	return tenantCtx, ok
}
