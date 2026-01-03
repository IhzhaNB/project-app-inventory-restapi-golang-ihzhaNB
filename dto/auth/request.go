package auth

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LogoutRequest struct {
	Token string `json:"-"` // Dari header Authorization
}
