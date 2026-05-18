package storage

import "github.com/google/uuid"

const (
	MIMEJPEG = "image/jpeg"
	MIMEPNG  = "image/png"
	MIMEWEBP = "image/webp"
)

type Asset struct {
	Key      string
	URL      string
	MIMEType string
	Size     int64
}

type StoreInput struct {
	TenantID uuid.UUID
	Folder   string
	Data     []byte
}
