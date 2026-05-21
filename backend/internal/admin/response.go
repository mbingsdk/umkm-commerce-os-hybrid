package admin

import "github.com/google/uuid"

type MeResponse struct {
	User AdminUserResponse `json:"user"`
}

type AdminUserResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PlatformRole string    `json:"platform_role"`
}

func NewMeResponse(adminCtx Context) MeResponse {
	return MeResponse{
		User: AdminUserResponse{
			ID:           adminCtx.UserID,
			Name:         adminCtx.Name,
			Email:        adminCtx.Email,
			PlatformRole: adminCtx.PlatformRole,
		},
	}
}
