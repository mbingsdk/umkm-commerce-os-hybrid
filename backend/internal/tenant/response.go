package tenant

import (
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/permission"
)

type ListItemResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Slug        string               `json:"slug"`
	Role        string               `json:"role"`
	Status      string               `json:"status"`
	Store       StoreSummaryResponse `json:"store"`
	Permissions []string             `json:"permissions"`
}

type StoreSummaryResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Slug   string    `json:"slug"`
	Status string    `json:"status"`
}

type OnboardingResponse struct {
	Tenant     TenantResponse       `json:"tenant"`
	Store      StoreSummaryResponse `json:"store"`
	Membership MembershipResponse   `json:"membership"`
}

type TenantResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Slug   string    `json:"slug"`
	Status string    `json:"status"`
}

type MembershipResponse struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}

func NewListItemResponse(item TenantListItem) ListItemResponse {
	return ListItemResponse{
		ID:     item.ID,
		Name:   item.Name,
		Slug:   item.Slug,
		Role:   item.Role,
		Status: item.Status,
		Store: StoreSummaryResponse{
			ID:     item.Store.ID,
			Name:   item.Store.Name,
			Slug:   item.Store.Slug,
			Status: item.Store.Status,
		},
		Permissions: permission.ListForRole(item.Role),
	}
}
