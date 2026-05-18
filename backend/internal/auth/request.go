package auth

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RegisterInput struct {
	Name      string
	Email     string
	Phone     string
	Password  string
	IPAddress string
	UserAgent string
}

type LoginInput struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}
