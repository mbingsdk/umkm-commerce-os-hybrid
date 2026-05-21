package discovery

type PaginationMeta struct {
	Pagination Pagination `json:"pagination"`
}

type Pagination struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor,omitempty"`
	HasMore    bool    `json:"has_more"`
}

type HomeResponse struct {
	FeaturedStores    []StoreResponse    `json:"featured_stores"`
	FeaturedProducts  []ProductResponse  `json:"featured_products"`
	LatestStores      []StoreResponse    `json:"latest_stores"`
	LatestProducts    []ProductResponse  `json:"latest_products"`
	PopularCategories []CategoryResponse `json:"popular_categories,omitempty"`
	PopularCities     []CityResponse     `json:"popular_cities,omitempty"`
}

type SearchResponse struct {
	Stores   []StoreResponse   `json:"stores,omitempty"`
	Products []ProductResponse `json:"products,omitempty"`
}

type StoreResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
	BannerURL   string `json:"banner_url,omitempty"`
	City        string `json:"city,omitempty"`
	Province    string `json:"province,omitempty"`
	StoreURL    string `json:"store_url"`
}

type ProductResponse struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Slug            string        `json:"slug"`
	Description     string        `json:"description,omitempty"`
	Price           int64         `json:"price"`
	PrimaryImageURL string        `json:"primary_image_url,omitempty"`
	Category        *CategoryMini `json:"category,omitempty"`
	Store           StoreMini     `json:"store"`
	StoreURL        string        `json:"store_url"`
	ProductURL      string        `json:"product_url"`
}

type StoreMini struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	City     string `json:"city,omitempty"`
	Province string `json:"province,omitempty"`
}

type CategoryMini struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CategoryResponse struct {
	Name  string `json:"name"`
	Slug  string `json:"slug"`
	Count int    `json:"count"`
}

type CityResponse struct {
	City  string `json:"city"`
	Count int    `json:"count"`
}

func NewStoreResponse(store Store) StoreResponse {
	return StoreResponse{
		ID:          store.ID.String(),
		Name:        store.Name,
		Slug:        store.Slug,
		Description: store.Description,
		LogoURL:     store.LogoURL,
		BannerURL:   store.BannerURL,
		City:        store.City,
		Province:    store.Province,
		StoreURL:    storeURL(store.Slug),
	}
}

func NewStoreResponses(stores []Store) []StoreResponse {
	response := make([]StoreResponse, 0, len(stores))
	for _, store := range stores {
		response = append(response, NewStoreResponse(store))
	}
	return response
}

func NewProductResponse(product Product) ProductResponse {
	var category *CategoryMini
	if product.CategorySlug != "" || product.CategoryName != "" {
		category = &CategoryMini{
			Name: product.CategoryName,
			Slug: product.CategorySlug,
		}
	}

	return ProductResponse{
		ID:              product.ID.String(),
		Name:            product.Name,
		Slug:            product.Slug,
		Description:     product.Description,
		Price:           product.Price,
		PrimaryImageURL: product.PrimaryImageURL,
		Category:        category,
		Store: StoreMini{
			ID:       product.StoreID.String(),
			Name:     product.StoreName,
			Slug:     product.StoreSlug,
			City:     product.StoreCity,
			Province: product.StoreProvince,
		},
		StoreURL:   storeURL(product.StoreSlug),
		ProductURL: productURL(product.StoreSlug, product.Slug),
	}
}

func NewProductResponses(products []Product) []ProductResponse {
	response := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		response = append(response, NewProductResponse(product))
	}
	return response
}

func NewCategoryResponses(categories []CategoryAggregate) []CategoryResponse {
	response := make([]CategoryResponse, 0, len(categories))
	for _, category := range categories {
		response = append(response, CategoryResponse(category))
	}
	return response
}

func NewCityResponses(cities []CityAggregate) []CityResponse {
	response := make([]CityResponse, 0, len(cities))
	for _, city := range cities {
		response = append(response, CityResponse(city))
	}
	return response
}

func storeURL(storeSlug string) string {
	return "/s/" + storeSlug
}

func productURL(storeSlug string, productSlug string) string {
	return "/s/" + storeSlug + "/products/" + productSlug
}
