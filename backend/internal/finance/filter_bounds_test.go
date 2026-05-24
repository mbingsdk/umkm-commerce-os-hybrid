package finance

import (
	"strings"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

func TestNormalizeExpenseListFiltersBoundsPaginationAndSearch(t *testing.T) {
	filters := normalizeListFilters(ListExpenseFilters{
		Query: strings.Repeat("expense", querytext.MaxSearchLength),
		Limit: maxExpenseListLimit + 10,
	})

	if filters.Limit != maxExpenseListLimit {
		t.Fatalf("expected max limit %d, got %d", maxExpenseListLimit, filters.Limit)
	}
	if len([]rune(filters.Query)) != querytext.MaxSearchLength {
		t.Fatalf("expected bounded query length %d, got %d", querytext.MaxSearchLength, len([]rune(filters.Query)))
	}

	defaulted := normalizeListFilters(ListExpenseFilters{})
	if defaulted.Limit != defaultExpenseListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultExpenseListLimit, defaulted.Limit)
	}
}
