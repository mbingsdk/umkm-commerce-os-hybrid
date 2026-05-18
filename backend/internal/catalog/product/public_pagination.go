package product

import (
	"encoding/base64"
	"encoding/json"
)

func EncodePublicCursor(item PublicListItem) (string, error) {
	payload, err := json.Marshal(PublicCursor{
		CreatedAt: item.CreatedAt,
		ID:        item.ID,
	})
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func DecodePublicCursor(raw string) (*PublicCursor, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var cursor PublicCursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, err
	}

	return &cursor, nil
}
