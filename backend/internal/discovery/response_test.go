package discovery

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestDiscoveryProductResponseDoesNotExposeCostPrice(t *testing.T) {
	response := NewProductResponse(Product{
		ID:              uuid.New(),
		Name:            "Bouquet Mawar",
		Slug:            "bouquet-mawar",
		Description:     "Produk discovery",
		Price:           50000,
		PrimaryImageURL: "https://cdn.example.test/product.jpg",
		StoreID:         uuid.New(),
		StoreName:       "Toko Bunga",
		StoreSlug:       "toko-bunga",
		StoreCity:       "Makassar",
	})

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal discovery response: %v", err)
	}
	if strings.Contains(string(payload), "cost_price") {
		t.Fatalf("discovery product response leaked cost_price: %s", payload)
	}
}
