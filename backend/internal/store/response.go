package store

import (
	"time"

	"github.com/google/uuid"
)

type Response struct {
	ID             uuid.UUID  `json:"id"`
	TenantID       uuid.UUID  `json:"tenant_id,omitempty"`
	Name           string     `json:"name"`
	Slug           string     `json:"slug,omitempty"`
	Description    string     `json:"description,omitempty"`
	LogoURL        string     `json:"logo_url,omitempty"`
	BannerURL      string     `json:"banner_url,omitempty"`
	Phone          string     `json:"phone,omitempty"`
	Whatsapp       string     `json:"whatsapp,omitempty"`
	Email          string     `json:"email,omitempty"`
	Address        string     `json:"address,omitempty"`
	City           string     `json:"city,omitempty"`
	Province       string     `json:"province,omitempty"`
	PostalCode     string     `json:"postal_code,omitempty"`
	Status         string     `json:"status"`
	IsDiscoverable bool       `json:"is_discoverable"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
}

type BusinessHoursResponse struct {
	Items []BusinessHourResponseItem `json:"items"`
}

type BusinessHourResponseItem struct {
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time,omitempty"`
	CloseTime string `json:"close_time,omitempty"`
	IsClosed  bool   `json:"is_closed"`
}

func NewResponse(store *Store) Response {
	return Response{
		ID:             store.ID,
		TenantID:       store.TenantID,
		Name:           store.Name,
		Slug:           store.Slug,
		Description:    store.Description,
		LogoURL:        store.LogoURL,
		BannerURL:      store.BannerURL,
		Phone:          store.Phone,
		Whatsapp:       store.Whatsapp,
		Email:          store.Email,
		Address:        store.Address,
		City:           store.City,
		Province:       store.Province,
		PostalCode:     store.PostalCode,
		Status:         store.Status,
		IsDiscoverable: store.IsDiscoverable,
		PublishedAt:    store.PublishedAt,
	}
}
