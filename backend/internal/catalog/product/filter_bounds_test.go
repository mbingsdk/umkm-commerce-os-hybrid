package product

import (
	"strings"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

func TestNormalizeListFiltersBoundsDashboardProductList(t *testing.T) {
	filters := normalizeListFilters(ListFilters{
		Query: strings.Repeat("b", querytext.MaxSearchLength+10),
		Limit: maxListLimit + 50,
	})

	if filters.Limit != maxListLimit {
		t.Fatalf("expected max limit %d, got %d", maxListLimit, filters.Limit)
	}
	if len([]rune(filters.Query)) != querytext.MaxSearchLength {
		t.Fatalf("expected bounded query length %d, got %d", querytext.MaxSearchLength, len([]rune(filters.Query)))
	}

	defaulted := normalizeListFilters(ListFilters{})
	if defaulted.Limit != defaultListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultListLimit, defaulted.Limit)
	}
}
