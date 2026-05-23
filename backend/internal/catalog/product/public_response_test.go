package product

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

func TestPublicProductResponseDoesNotExposeCostPrice(t *testing.T) {
	response := NewPublicDetailResponse(&PublicProduct{
		ID:          uuid.New(),
		Name:        "Bouquet Mawar",
		Slug:        "bouquet-mawar",
		Description: "Produk publik",
		Price:       50000,
		Stock: Stock{
			QuantityAvailable: 10,
			LowStockThreshold: 5,
		},
	}, store.PublicContext{
		Store: store.Store{
			Name: "Toko Bunga",
			Slug: "toko-bunga",
			City: "Makassar",
		},
	})

	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal public response: %v", err)
	}
	if strings.Contains(string(payload), "cost_price") {
		t.Fatalf("public product response leaked cost_price: %s", payload)
	}
}
