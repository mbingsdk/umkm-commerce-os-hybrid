package querytext

import "strings"

const MaxSearchLength = 100

func NormalizeSearch(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	runes := []rune(trimmed)
	if len(runes) <= MaxSearchLength {
		return trimmed
	}
	return string(runes[:MaxSearchLength])
}
