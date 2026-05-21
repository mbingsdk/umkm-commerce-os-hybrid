package admin

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"
)

type UpdateTenantStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type UpdateTenantPlanRequest struct {
	PlanID string `json:"plan_id"`
	Reason string `json:"reason,omitempty"`
}

type CreatePlanRequest struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	PriceMonthly    int64  `json:"price_monthly"`
	ProductLimit    *int   `json:"product_limit"`
	StaffLimit      *int   `json:"staff_limit"`
	CanUsePOS       *bool  `json:"can_use_pos"`
	CanUseDiscovery *bool  `json:"can_use_discovery"`
	CanUseCourier   *bool  `json:"can_use_courier"`
	IsActive        *bool  `json:"is_active"`
}

type UpdatePlanRequest struct {
	Code            *string     `json:"code"`
	Name            *string     `json:"name"`
	Description     *string     `json:"description"`
	PriceMonthly    *int64      `json:"price_monthly"`
	ProductLimit    NullableInt `json:"product_limit"`
	StaffLimit      NullableInt `json:"staff_limit"`
	CanUsePOS       *bool       `json:"can_use_pos"`
	CanUseDiscovery *bool       `json:"can_use_discovery"`
	CanUseCourier   *bool       `json:"can_use_courier"`
	IsActive        *bool       `json:"is_active"`
}

type CreateFeaturedRequest struct {
	ItemType  string     `json:"item_type"`
	TenantID  string     `json:"tenant_id"`
	StoreID   string     `json:"store_id,omitempty"`
	ProductID string     `json:"product_id,omitempty"`
	Placement string     `json:"placement"`
	SortOrder int        `json:"sort_order"`
	StartsAt  *time.Time `json:"starts_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
	IsActive  *bool      `json:"is_active"`
}

type UpdateFeaturedRequest struct {
	ItemType  *string    `json:"item_type"`
	TenantID  *string    `json:"tenant_id"`
	StoreID   *string    `json:"store_id"`
	ProductID *string    `json:"product_id"`
	Placement *string    `json:"placement"`
	SortOrder *int       `json:"sort_order"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	IsActive  *bool      `json:"is_active"`
}

type NullableInt struct {
	Set   bool
	Value *int
}

func (n *NullableInt) UnmarshalJSON(data []byte) error {
	n.Set = true
	if strings.EqualFold(strings.TrimSpace(string(data)), "null") {
		n.Value = nil
		return nil
	}

	var value int
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	n.Value = &value
	return nil
}

type AuditInput struct {
	ActorUserID string
	Action      string
	TargetType  string
	TargetID    string
	BeforeData  any
	AfterData   any
	IPAddress   string
	UserAgent   string
}

func RequestMetadata(r *http.Request) (ipAddress string, userAgent string) {
	userAgent = strings.TrimSpace(r.UserAgent())

	host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
	if err == nil {
		return host, userAgent
	}

	return strings.TrimSpace(r.RemoteAddr), userAgent
}
