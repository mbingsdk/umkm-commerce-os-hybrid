package admin

import (
	"strings"
	"testing"

	"github.com/sdkdev/umkm-commerce-os/backend/internal/shared/querytext"
)

func TestNormalizeTenantListFiltersBoundsPaginationAndSearch(t *testing.T) {
	filters, err := normalizeTenantListFilters(TenantListFilters{
		Query: strings.Repeat("tenant", querytext.MaxSearchLength),
		Limit: maxTenantListLimit + 10,
	})
	if err != nil {
		t.Fatalf("normalize tenant filters: %v", err)
	}

	if filters.Limit != maxTenantListLimit {
		t.Fatalf("expected max limit %d, got %d", maxTenantListLimit, filters.Limit)
	}
	if len([]rune(filters.Query)) != querytext.MaxSearchLength {
		t.Fatalf("expected bounded query length %d, got %d", querytext.MaxSearchLength, len([]rune(filters.Query)))
	}

	defaulted, err := normalizeTenantListFilters(TenantListFilters{})
	if err != nil {
		t.Fatalf("normalize default tenant filters: %v", err)
	}
	if defaulted.Limit != defaultTenantListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultTenantListLimit, defaulted.Limit)
	}
}
