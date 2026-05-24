package order

import (
	"strings"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

func TestNormalizeOrderListFiltersBoundsPaginationAndSearch(t *testing.T) {
	filters, err := normalizeListFilters(ListFilters{
		Query: strings.Repeat("order", querytext.MaxSearchLength),
		Limit: maxListLimit + 25,
	})
	if err != nil {
		t.Fatalf("normalize list filters: %v", err)
	}

	if filters.Limit != maxListLimit {
		t.Fatalf("expected max limit %d, got %d", maxListLimit, filters.Limit)
	}
	if len([]rune(filters.Query)) != querytext.MaxSearchLength {
		t.Fatalf("expected bounded query length %d, got %d", querytext.MaxSearchLength, len([]rune(filters.Query)))
	}

	defaulted, err := normalizeListFilters(ListFilters{})
	if err != nil {
		t.Fatalf("normalize default filters: %v", err)
	}
	if defaulted.Limit != defaultListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultListLimit, defaulted.Limit)
	}
}
