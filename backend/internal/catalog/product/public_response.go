package product

import (
	"github.com/google/uuid"
	"github.com/sdkdev/umkm-commerce-os/backend/internal/store"
)

const (
	StockStatusInStock    = "in_stock"
	StockStatusLowStock   = "low_stock"
	StockStatusOutOfStock = "out_of_stock"
)

type PublicPaginationMeta struct {
	Pagination PublicPagination `json:"pagination"`
}

type PublicPagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

type PublicListItemResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Slug            string    `json:"slug"`
	Price           int64     `json:"price"`
	CompareAtPrice  *int64    `json:"compare_at_price,omitempty"`
	PrimaryImageURL string    `json:"primary_image_url,omitempty"`
	StockStatus     string    `json:"stock_status"`
}

type PublicImageResponse struct {
	URL     string `json:"url"`
	AltText string `json:"alt_text,omitempty"`
}

type PublicCategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type PublicStockResponse struct {
	StockStatus       string `json:"stock_status"`
	QuantityAvailable int    `json:"quantity_available"`
}

type PublicStoreSummaryResponse struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	City string `json:"city,omitempty"`
}

type PublicDetailResponse struct {
	ID             uuid.UUID                  `json:"id"`
	Name           string                     `json:"name"`
	Slug           string                     `json:"slug"`
	Description    string                     `json:"description,omitempty"`
	Price          int64                      `json:"price"`
	CompareAtPrice *int64                     `json:"compare_at_price,omitempty"`
	WeightGram     int                        `json:"weight_gram"`
	Images         []PublicImageResponse      `json:"images"`
	Category       *PublicCategoryResponse    `json:"category,omitempty"`
	Stock          PublicStockResponse        `json:"stock"`
	Store          PublicStoreSummaryResponse `json:"store"`
}

func NewPublicListItemResponse(item *PublicListItem) PublicListItemResponse {
	return PublicListItemResponse{
		ID:              item.ID,
		Name:            item.Name,
		Slug:            item.Slug,
		Price:           item.Price,
		CompareAtPrice:  item.CompareAtPrice,
		PrimaryImageURL: item.PrimaryImageURL,
		StockStatus:     publicStockStatus(item.Stock),
	}
}

func NewPublicDetailResponse(item *PublicProduct, currentStore store.PublicContext) PublicDetailResponse {
	images := make([]PublicImageResponse, 0, len(item.Images))
	for _, image := range item.Images {
		images = append(images, PublicImageResponse{
			URL:     image.URL,
			AltText: image.AltText,
		})
	}

	var category *PublicCategoryResponse
	if item.Category != nil {
		category = &PublicCategoryResponse{
			ID:   item.Category.ID,
			Name: item.Category.Name,
			Slug: item.Category.Slug,
		}
	}

	return PublicDetailResponse{
		ID:             item.ID,
		Name:           item.Name,
		Slug:           item.Slug,
		Description:    item.Description,
		Price:          item.Price,
		CompareAtPrice: item.CompareAtPrice,
		WeightGram:     item.WeightGram,
		Images:         images,
		Category:       category,
		Stock: PublicStockResponse{
			StockStatus:       publicStockStatus(item.Stock),
			QuantityAvailable: item.Stock.QuantityAvailable,
		},
		Store: PublicStoreSummaryResponse{
			Name: currentStore.Store.Name,
			Slug: currentStore.Store.Slug,
			City: currentStore.Store.City,
		},
	}
}

func publicStockStatus(stock Stock) string {
	switch {
	case stock.QuantityAvailable <= 0:
		return StockStatusOutOfStock
	case stock.QuantityAvailable <= stock.LowStockThreshold:
		return StockStatusLowStock
	default:
		return StockStatusInStock
	}
}
