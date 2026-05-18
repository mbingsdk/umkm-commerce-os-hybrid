package category

import "github.com/google/uuid"

type PublicCategory struct {
	ID       uuid.UUID
	Name     string
	Slug     string
	ImageURL string
}
