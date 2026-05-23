package discovery

import (
	"context"
	"testing"
)

func TestPublicDiscoverySafetyHidesSuspendedAndNonDiscoverableRecords(t *testing.T) {
	service := newDiscoveryTestService()

	result, err := service.Search(context.Background(), SearchFilters{Query: "Rahasia", Type: SearchTypeAll, Limit: 10})
	if err != nil {
		t.Fatalf("Search error = %v", err)
	}
	if len(result.Stores) != 0 || len(result.Products) != 0 {
		t.Fatalf("public discovery returned hidden records: %#v", result)
	}

	products, _, err := service.ListProducts(context.Background(), ListProductsFilters{Limit: 20})
	if err != nil {
		t.Fatalf("ListProducts error = %v", err)
	}
	for _, product := range products {
		if product.ID == discoveryProductB.String() || product.ID == discoveryProductC.String() {
			t.Fatalf("public discovery leaked non-discoverable or inactive product: %#v", product)
		}
	}
}
