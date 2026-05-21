package admin

import (
	"time"

	"github.com/google/uuid"
)

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

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type AdminTenantListResponse struct {
	ID           string                     `json:"id"`
	Name         string                     `json:"name"`
	Slug         string                     `json:"slug"`
	Status       string                     `json:"status"`
	Plan         *AdminPlanResponse         `json:"plan,omitempty"`
	PrimaryStore *AdminStoreSummaryResponse `json:"primary_store,omitempty"`
	Owner        *AdminOwnerResponse        `json:"owner,omitempty"`
	Counts       AdminTenantCountsResponse  `json:"counts"`
	CreatedAt    string                     `json:"created_at"`
}

type AdminTenantDetailResponse struct {
	Tenant       AdminTenantBasicResponse    `json:"tenant"`
	Plan         *AdminPlanResponse          `json:"plan,omitempty"`
	PrimaryStore *AdminStoreSummaryResponse  `json:"primary_store,omitempty"`
	Owner        *AdminOwnerResponse         `json:"owner,omitempty"`
	Counts       AdminTenantCountsResponse   `json:"counts"`
	LatestAudits []AdminAuditSnippetResponse `json:"latest_audits,omitempty"`
}

type AdminTenantBasicResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type AdminPlanResponse struct {
	ID                 string `json:"id"`
	Code               string `json:"code"`
	Name               string `json:"name"`
	PriceMonthly       int64  `json:"price_monthly"`
	ProductLimit       int    `json:"product_limit"`
	StaffLimit         int    `json:"staff_limit"`
	CanUsePOS          bool   `json:"can_use_pos"`
	CanUseDiscovery    bool   `json:"can_use_discovery"`
	CanUseCourier      bool   `json:"can_use_courier"`
	CanUseCustomDomain bool   `json:"can_use_custom_domain"`
	IsActive           bool   `json:"is_active"`
}

type AdminStoreSummaryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Status    string `json:"status"`
	City      string `json:"city,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type AdminOwnerResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

type AdminTenantCountsResponse struct {
	Stores          int64 `json:"stores"`
	Products        int64 `json:"products"`
	Orders          int64 `json:"orders"`
	Users           int64 `json:"users"`
	POSTransactions int64 `json:"pos_transactions"`
}

type AdminAuditSnippetResponse struct {
	ID          string  `json:"id"`
	ActorUserID string  `json:"actor_user_id,omitempty"`
	ActorName   string  `json:"actor_name,omitempty"`
	Action      string  `json:"action"`
	TargetType  string  `json:"target_type,omitempty"`
	TargetID    *string `json:"target_id,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type AdminTenantMutationResponse struct {
	ID     string             `json:"id"`
	Name   string             `json:"name"`
	Slug   string             `json:"slug"`
	Status string             `json:"status"`
	Plan   *AdminPlanResponse `json:"plan,omitempty"`
}

func NewTenantListResponse(items []TenantListItem) []AdminTenantListResponse {
	response := make([]AdminTenantListResponse, 0, len(items))
	for _, item := range items {
		response = append(response, AdminTenantListResponse{
			ID:           item.Tenant.ID.String(),
			Name:         item.Tenant.Name,
			Slug:         item.Tenant.Slug,
			Status:       item.Tenant.Status,
			Plan:         newPlanResponse(item.Plan),
			PrimaryStore: newStoreSummaryResponse(item.PrimaryStore),
			Owner:        newOwnerResponse(item.Owner),
			Counts:       newTenantCountsResponse(item.Counts),
			CreatedAt:    formatTime(item.Tenant.CreatedAt),
		})
	}
	return response
}

func NewTenantDetailResponse(item TenantDetail) AdminTenantDetailResponse {
	return AdminTenantDetailResponse{
		Tenant:       newTenantBasicResponse(item.Tenant),
		Plan:         newPlanResponse(item.Plan),
		PrimaryStore: newStoreSummaryResponse(item.PrimaryStore),
		Owner:        newOwnerResponse(item.Owner),
		Counts:       newTenantCountsResponse(item.Counts),
		LatestAudits: newAuditSnippetResponses(item.LatestAudits),
	}
}

func NewTenantMutationResponse(tenant Tenant, plan *Plan) AdminTenantMutationResponse {
	return AdminTenantMutationResponse{
		ID:     tenant.ID.String(),
		Name:   tenant.Name,
		Slug:   tenant.Slug,
		Status: tenant.Status,
		Plan:   newPlanResponse(plan),
	}
}

func newTenantBasicResponse(tenant Tenant) AdminTenantBasicResponse {
	return AdminTenantBasicResponse{
		ID:        tenant.ID.String(),
		Name:      tenant.Name,
		Slug:      tenant.Slug,
		Status:    tenant.Status,
		CreatedAt: formatTime(tenant.CreatedAt),
		UpdatedAt: formatTime(tenant.UpdatedAt),
	}
}

func newPlanResponse(plan *Plan) *AdminPlanResponse {
	if plan == nil || plan.ID == uuid.Nil {
		return nil
	}
	return &AdminPlanResponse{
		ID:                 plan.ID.String(),
		Code:               plan.Code,
		Name:               plan.Name,
		PriceMonthly:       plan.PriceMonthly,
		ProductLimit:       plan.ProductLimit,
		StaffLimit:         plan.StaffLimit,
		CanUsePOS:          plan.CanUsePOS,
		CanUseDiscovery:    plan.CanUseDiscovery,
		CanUseCourier:      plan.CanUseCourier,
		CanUseCustomDomain: plan.CanUseCustomDomain,
		IsActive:           plan.IsActive,
	}
}

func newStoreSummaryResponse(store *StoreSummary) *AdminStoreSummaryResponse {
	if store == nil || store.ID == uuid.Nil {
		return nil
	}
	return &AdminStoreSummaryResponse{
		ID:        store.ID.String(),
		Name:      store.Name,
		Slug:      store.Slug,
		Status:    store.Status,
		City:      store.City,
		CreatedAt: formatTime(store.CreatedAt),
	}
}

func newOwnerResponse(owner *OwnerSummary) *AdminOwnerResponse {
	if owner == nil || owner.ID == uuid.Nil {
		return nil
	}
	return &AdminOwnerResponse{
		ID:     owner.ID.String(),
		Name:   owner.Name,
		Email:  owner.Email,
		Status: owner.Status,
	}
}

func newTenantCountsResponse(counts TenantCounts) AdminTenantCountsResponse {
	return AdminTenantCountsResponse{
		Stores:          counts.StoreCount,
		Products:        counts.ProductCount,
		Orders:          counts.OrderCount,
		Users:           counts.UserCount,
		POSTransactions: counts.POSTransactionCount,
	}
}

func newAuditSnippetResponses(items []AuditSnippet) []AdminAuditSnippetResponse {
	response := make([]AdminAuditSnippetResponse, 0, len(items))
	for _, item := range items {
		var targetID *string
		if item.TargetID != nil && *item.TargetID != uuid.Nil {
			value := item.TargetID.String()
			targetID = &value
		}
		actorUserID := ""
		if item.ActorUserID != uuid.Nil {
			actorUserID = item.ActorUserID.String()
		}
		response = append(response, AdminAuditSnippetResponse{
			ID:          item.ID.String(),
			ActorUserID: actorUserID,
			ActorName:   item.ActorName,
			Action:      item.Action,
			TargetType:  item.TargetType,
			TargetID:    targetID,
			CreatedAt:   formatTime(item.CreatedAt),
		})
	}
	return response
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}
