package storage

import (
	"net/http"
)

func validateImage(data []byte, maxBytes int64) (string, string, error) {
	if len(data) == 0 {
		return "", "", validation("file", "File is required")
	}
	if int64(len(data)) > maxBytes {
		return "", "", validation("file", "File size exceeds the configured limit")
	}

	sniffLen := len(data)
	if sniffLen > 512 {
		sniffLen = 512
	}

	switch mimeType := http.DetectContentType(data[:sniffLen]); mimeType {
	case MIMEJPEG:
		return mimeType, ".jpg", nil
	case MIMEPNG:
		return mimeType, ".png", nil
	case MIMEWEBP:
		return mimeType, ".webp", nil
	default:
		return "", "", validation("file", "File type must be image/jpeg, image/png, or image/webp")
	}
}
