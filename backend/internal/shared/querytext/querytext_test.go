package querytext

import (
	"strings"
	"testing"
)

func TestNormalizeSearchTrimsAndBoundsLength(t *testing.T) {
	raw := "  " + strings.Repeat("a", MaxSearchLength+25) + "  "

	got := NormalizeSearch(raw)

	if len([]rune(got)) != MaxSearchLength {
		t.Fatalf("expected bounded search length %d, got %d", MaxSearchLength, len([]rune(got)))
	}
	if strings.HasPrefix(got, " ") || strings.HasSuffix(got, " ") {
		t.Fatalf("expected search value to be trimmed, got %q", got)
	}
}
