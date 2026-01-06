package middleware

import (
	"context"
	"inventory-system/model"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Context key type (gunakan yang sama dengan utils)
type contextKey string

const userContextKey contextKey = "user"

// Auth middleware untuk validasi token
func Auth(authService service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Ambil token dari header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.ResponseError(w, http.StatusUnauthorized, "Authentication required: No Authorization header", nil)
				return
			}

			// Parse format "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.ResponseError(w, http.StatusUnauthorized, "Invalid authorization format", nil)
				return
			}

			tokenString := parts[1]

			// Parse ke UUID
			token, err := uuid.Parse(tokenString)
			if err != nil {
				utils.ResponseError(w, http.StatusUnauthorized, "Invalid token format: Must be valid UUID", nil)
				return
			}

			// Validasi token
			user, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				utils.Logger.Warn("Invalid token",
					zap.String("token", tokenString),
					zap.Error(err),
				)
				utils.ResponseError(w, http.StatusUnauthorized, "Invalid or expired token", err.Error())
				return
			}

			// Simpan user di request context MENGGUNAKAN utils.SetUserToContext
			ctx := utils.SetUserToContext(r.Context(), user)

			// Lanjut ke handler dengan context baru
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext helper untuk ambil user dari context
// SEKARANG GUNAKAN utils.GetUserFromContext, tapi kita keep untuk backward compatibility
func GetUserFromContext(ctx context.Context) *model.User {
	return utils.GetUserFromContext(ctx)
}
