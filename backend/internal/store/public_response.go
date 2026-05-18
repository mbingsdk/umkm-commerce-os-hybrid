package store

import "github.com/google/uuid"

type PublicResponse struct {
	ID            uuid.UUID                  `json:"id"`
	Name          string                     `json:"name"`
	Slug          string                     `json:"slug"`
	Description   string                     `json:"description,omitempty"`
	LogoURL       string                     `json:"logo_url,omitempty"`
	BannerURL     string                     `json:"banner_url,omitempty"`
	Phone         string                     `json:"phone,omitempty"`
	Whatsapp      string                     `json:"whatsapp,omitempty"`
	City          string                     `json:"city,omitempty"`
	Province      string                     `json:"province,omitempty"`
	BusinessHours []BusinessHourResponseItem `json:"business_hours"`
}

func NewPublicResponse(item *Store, hours []BusinessHour) PublicResponse {
	responseHours := make([]BusinessHourResponseItem, 0, len(hours))
	for _, hour := range hours {
		responseHours = append(responseHours, BusinessHourResponseItem{
			DayOfWeek: hour.DayOfWeek,
			OpenTime:  hour.OpenTime,
			CloseTime: hour.CloseTime,
			IsClosed:  hour.IsClosed,
		})
	}

	return PublicResponse{
		ID:            item.ID,
		Name:          item.Name,
		Slug:          item.Slug,
		Description:   item.Description,
		LogoURL:       item.LogoURL,
		BannerURL:     item.BannerURL,
		Phone:         item.Phone,
		Whatsapp:      item.Whatsapp,
		City:          item.City,
		Province:      item.Province,
		BusinessHours: responseHours,
	}
}
