package category

import "github.com/google/uuid"

type PublicResponse struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Slug     string    `json:"slug"`
	ImageURL string    `json:"image_url,omitempty"`
}

func NewPublicResponse(item *PublicCategory) PublicResponse {
	return PublicResponse{
		ID:       item.ID,
		Name:     item.Name,
		Slug:     item.Slug,
		ImageURL: item.ImageURL,
	}
}
