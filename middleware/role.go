package middleware

import (
	"inventory-system/model"
	"inventory-system/utils"
	"net/http"

	"go.uber.org/zap"
)

// RequireRole middleware untuk validasi role user
func RequireRole(allowedRoles ...model.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Ambil user dari context (setelah Auth middleware)
			user := GetUserFromContext(r.Context())
			if user == nil {
				utils.ResponseError(w, http.StatusUnauthorized,
					"Authentication required", nil)
				return
			}

			// Cek apakah role user diizinkan
			hasPermission := false
			for _, allowedRole := range allowedRoles {
				if user.Role == allowedRole {
					hasPermission = true
					break
				}
			}

			// Jika tidak diizinkan, return 403 Forbidden
			if !hasPermission {
				utils.Logger.Warn("Access denied",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("user_role", string(user.Role)),
				)

				utils.ResponseError(w, http.StatusForbidden,
					"Access denied",
					"Your role does not have permission to access this resource")
				return
			}

			// Lanjut ke handler jika authorized
			next.ServeHTTP(w, r)
		})
	}
}
