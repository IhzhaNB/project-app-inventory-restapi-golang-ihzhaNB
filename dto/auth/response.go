package auth

import "time"

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}
