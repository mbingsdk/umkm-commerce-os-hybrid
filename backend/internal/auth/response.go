package auth

import "github.com/google/uuid"

type UserResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone,omitempty"`
	PlatformRole string    `json:"platform_role"`
}

type TenantResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Slug   string    `json:"slug"`
	Role   string    `json:"role"`
	Status string    `json:"status"`
}

type AuthResponse struct {
	User         UserResponse     `json:"user"`
	Tenants      []TenantResponse `json:"tenants,omitempty"`
	AccessToken  string           `json:"access_token"`
	RefreshToken string           `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type MeResponse struct {
	User    UserResponse     `json:"user"`
	Tenants []TenantResponse `json:"tenants"`
}

func NewUserResponse(user *User) UserResponse {
	return UserResponse{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		Phone:        user.Phone,
		PlatformRole: user.PlatformRole,
	}
}
