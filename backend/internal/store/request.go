package store

type UpdateProfileRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Phone          string `json:"phone"`
	Whatsapp       string `json:"whatsapp"`
	Email          string `json:"email"`
	Address        string `json:"address"`
	City           string `json:"city"`
	Province       string `json:"province"`
	PostalCode     string `json:"postal_code"`
	IsDiscoverable bool   `json:"is_discoverable"`
}

type BusinessHoursRequest struct {
	Items []BusinessHourRequestItem `json:"items"`
}

type BusinessHourRequestItem struct {
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	IsClosed  bool   `json:"is_closed"`
}
