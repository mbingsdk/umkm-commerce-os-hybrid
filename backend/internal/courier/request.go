package courier

type CreateZoneRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Rate        int64  `json:"rate"`
	IsActive    *bool  `json:"is_active"`
	SortOrder   *int   `json:"sort_order"`
}

type UpdateZoneRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Rate        *int64  `json:"rate"`
	IsActive    *bool   `json:"is_active"`
	SortOrder   *int    `json:"sort_order"`
}
