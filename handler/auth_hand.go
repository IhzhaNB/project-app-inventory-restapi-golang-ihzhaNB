package handler

import (
	"encoding/json"
	"inventory-system/dto/auth"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ============================================
// AUTH HANDLER STRUCT
// ============================================
type AuthHandler struct {
	authService *service.Service
	log         *zap.Logger
}

func NewAuthHandler(authService *service.Service, log *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		log:         log,
	}
}

// ============================================
// LOGIN HANDLER
// ============================================
// POST /api/auth/login
// Public endpoint untuk user authentication
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest

	// 1. Parse JSON request body
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid JSON format", nil)
		return
	}
	defer r.Body.Close()

	// 2. Call auth service untuk proses login
	resp, err := ah.authService.Auth.Login(r.Context(), req)
	if err != nil {
		// Determine appropriate status code
		statusCode := http.StatusUnauthorized
		if strings.Contains(err.Error(), "validation") {
			statusCode = http.StatusBadRequest // Validation error
		}

		utils.ResponseError(w, statusCode, "Login failed", err.Error())
		return
	}

	// 3. Log success dan return response
	ah.log.Info("User logged in", zap.String("email", req.Email))
	utils.ResponseSuccess(w, http.StatusOK, "Login successful", resp)
}

// ============================================
// LOGOUT HANDLER
// ============================================
// POST /api/auth/logout
// Protected endpoint (butuh Authorization header)
func (ah *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// 1. Extract token dari Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		utils.ResponseError(w, http.StatusBadRequest, "No token provided", nil)
		return
	}

	// 2. Parse "Bearer <token>" format
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	token, err := uuid.Parse(tokenStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid token format", err.Error())
		return
	}

	// 3. Call auth service untuk invalidate session
	err = ah.authService.Auth.Logout(r.Context(), token)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Logout failed", err.Error())
		return
	}

	// 4. Return success response
	ah.log.Info("User logged out", zap.String("token", token.String()))
	utils.ResponseSuccess(w, http.StatusOK, "Logout successful", nil)
}
