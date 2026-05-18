package upload

type Response struct {
	URL      string `json:"url"`
	MIMEType string `json:"mime_type"`
	Size     int64  `json:"size"`
}
