package tenant

type CreateStoreRequest struct {
	TenantName string             `json:"tenant_name"`
	TenantSlug string             `json:"tenant_slug"`
	Store      StoreCreateRequest `json:"store"`
}

type StoreCreateRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Phone       string `json:"phone"`
	Whatsapp    string `json:"whatsapp"`
	Email       string `json:"email"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Province    string `json:"province"`
	PostalCode  string `json:"postal_code"`
}
