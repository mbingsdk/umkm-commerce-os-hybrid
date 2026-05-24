package inventory

import (
	"strings"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

func TestNormalizeStockFiltersBoundsPaginationAndSearch(t *testing.T) {
	filters := normalizeStockFilters(ListStockFilters{
		Query: strings.Repeat("sku", querytext.MaxSearchLength),
		Limit: maxListLimit + 10,
	})

	if filters.Limit != maxListLimit {
		t.Fatalf("expected max limit %d, got %d", maxListLimit, filters.Limit)
	}
	if len([]rune(filters.Query)) != querytext.MaxSearchLength {
		t.Fatalf("expected bounded query length %d, got %d", querytext.MaxSearchLength, len([]rune(filters.Query)))
	}

	defaulted := normalizeStockFilters(ListStockFilters{})
	if defaulted.Limit != defaultListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultListLimit, defaulted.Limit)
	}
}
