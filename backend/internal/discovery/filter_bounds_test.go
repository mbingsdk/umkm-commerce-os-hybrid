package discovery

import (
	"strings"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

func TestNormalizeDiscoveryFiltersBoundsSearchAndLimit(t *testing.T) {
	stores := normalizeStoreFilters(ListStoresFilters{
		Query: strings.Repeat("toko", querytext.MaxSearchLength),
		Limit: maxListLimit + 1,
	})
	if stores.Limit != maxListLimit {
		t.Fatalf("expected max store limit %d, got %d", maxListLimit, stores.Limit)
	}
	if len([]rune(stores.Query)) != querytext.MaxSearchLength {
		t.Fatalf("expected bounded store query length %d, got %d", querytext.MaxSearchLength, len([]rune(stores.Query)))
	}

	search, err := normalizeSearchFilters(SearchFilters{Query: "  bunga  "})
	if err != nil {
		t.Fatalf("normalize search: %v", err)
	}
	if search.Limit != defaultListLimit {
		t.Fatalf("expected default search limit %d, got %d", defaultListLimit, search.Limit)
	}
	if search.Query != "bunga" {
		t.Fatalf("expected trimmed search query, got %q", search.Query)
	}
}

func TestDiscoveryInvalidCursorRejected(t *testing.T) {
	if _, err := DecodeCursor("not-a-valid-cursor"); err == nil {
		t.Fatal("expected invalid discovery cursor to be rejected")
	}
}
