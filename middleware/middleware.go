package middleware

import (
	"inventory-system/utils"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AllowSelfOrAdmin - user boleh akses jika:
// 1. Mengakses data diri sendiri, ATAU
// 2. Role adalah admin/super_admin
func AllowSelfOrAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ambil user dari context MENGGUNAKAN utils.GetUserFromContext
		currentUser := utils.GetUserFromContext(r.Context())
		if currentUser == nil {
			utils.ResponseError(w, http.StatusUnauthorized, "Authentication required", nil)
			return
		}

		// Get user ID from URL
		userIDStr := chi.URLParam(r, "id")
		requestedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			utils.ResponseError(w, http.StatusBadRequest, "Invalid user ID", nil)
			return
		}

		// Check: accessing own data OR is admin
		if currentUser.ID != requestedUserID && !currentUser.CanManageUsers() {
			utils.Logger.Warn("Access denied to user data",
				zap.String("current_user", currentUser.ID.String()),
				zap.String("requested_user", requestedUserID.String()),
				zap.String("role", string(currentUser.Role)),
			)
			utils.ResponseError(w, http.StatusForbidden,
				"Cannot access other user's data", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}
