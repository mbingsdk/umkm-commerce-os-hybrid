package courier

const timeFormatRFC3339 = "2006-01-02T15:04:05Z07:00"

type ZoneResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Rate        int64  `json:"rate"`
	IsActive    bool   `json:"is_active"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func NewZoneResponse(zone Zone) ZoneResponse {
	return ZoneResponse{
		ID:          zone.ID.String(),
		Name:        zone.Name,
		Description: zone.Description,
		Rate:        zone.Rate,
		IsActive:    zone.IsActive,
		SortOrder:   zone.SortOrder,
		CreatedAt:   zone.CreatedAt.Format(timeFormatRFC3339),
		UpdatedAt:   zone.UpdatedAt.Format(timeFormatRFC3339),
	}
}

func NewZoneResponses(zones []Zone) []ZoneResponse {
	response := make([]ZoneResponse, 0, len(zones))
	for _, zone := range zones {
		response = append(response, NewZoneResponse(zone))
	}
	return response
}
