package discovery

type HomeRequest struct {
	Limit int
}

type ListStoresRequest struct {
	Query    string
	City     string
	Category string
	Limit    int
	Cursor   string
}

type ListProductsRequest struct {
	Query    string
	City     string
	Category string
	PriceMin string
	PriceMax string
	Limit    int
	Cursor   string
}

type SearchRequest struct {
	Query    string
	Type     string
	City     string
	Category string
	Limit    int
}
