package pos

import (
	"strings"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

func TestNormalizePOSProductFiltersBoundsPaginationAndSearch(t *testing.T) {
	filters := normalizeProductFilters(ProductSearchFilters{
		Query: strings.Repeat("pos", querytext.MaxSearchLength),
		Limit: maxPOSListLimit + 1,
	})

	if filters.Limit != maxPOSListLimit {
		t.Fatalf("expected max limit %d, got %d", maxPOSListLimit, filters.Limit)
	}
	if len([]rune(filters.Query)) != querytext.MaxSearchLength {
		t.Fatalf("expected bounded query length %d, got %d", querytext.MaxSearchLength, len([]rune(filters.Query)))
	}

	defaulted := normalizeProductFilters(ProductSearchFilters{})
	if defaulted.Limit != defaultPOSListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultPOSListLimit, defaulted.Limit)
	}
}
