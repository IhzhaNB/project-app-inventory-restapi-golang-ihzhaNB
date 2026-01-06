package handler

import (
	"encoding/json"
	"inventory-system/dto/user"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UserHandler struct {
	service *service.Service
	log     *zap.Logger
}

func NewUserHandler(service *service.Service, log *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		log:     log,
	}
}

// CREATE USER HANDLER
// POST /api/admin/users (Admin & Super Admin only)
// NOTE: Authorization sudah dicek di middleware router
func (uh *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req user.CreateUserRequest

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call service
	createdUser, err := uh.service.User.Create(r.Context(), req)
	if err != nil {
		uh.log.Error("Failed to create user", zap.Error(err))

		statusCode := http.StatusBadRequest
		if strings.Contains(err.Error(), "email already exists") {
			statusCode = http.StatusConflict
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusCreated, "User created successfully", createdUser)
}

// FIND USER BY ID HANDLER
// GET /api/users/{id} (All authenticated users)
// NOTE: Middleware sudah memastikan user hanya bisa akses diri sendiri
func (uh *UserHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	// Call service
	userData, err := uh.service.User.FindByID(r.Context(), userID)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "User not found", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "User retrieved", userData)
}

// FIND ALL USERS HANDLER
// GET /api/admin/users (Admin & Super Admin only)
func (uh *UserHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters from query string
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Default values
	page := 1
	limit := 10

	// Parse page parameter
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		} else {
			utils.ResponseError(w, http.StatusBadRequest, "Invalid page parameter", nil)
			return
		}
	}

	// Parse limit parameter
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		} else {
			utils.ResponseError(w, http.StatusBadRequest, "Invalid limit parameter (max 100)", nil)
			return
		}
	}

	// Call service
	users, pagination, err := uh.service.User.FindAll(r.Context(), page, limit)
	if err != nil {
		uh.log.Error("Failed to get users", zap.Error(err))
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to retrieve users", nil)
		return
	}

	// Response with pagination
	response := map[string]interface{}{
		"users":      users,
		"pagination": pagination,
	}

	utils.ResponseSuccess(w, http.StatusOK, "Users retrieved successfully", response)
}

// UPDATE USER HANDLER
// PUT /api/users/{id} (All authenticated users)
func (uh *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	var req user.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call service
	updatedUser, err := uh.service.User.Update(r.Context(), userID, req)
	if err != nil {
		uh.log.Error("Failed to update user", zap.Error(err))
		utils.ResponseError(w, http.StatusBadRequest, "Failed to update user", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "User updated successfully", updatedUser)
}

// DELETE USER HANDLER
// DELETE /api/admin/users/{id} (Admin & Super Admin only)
func (uh *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid user ID", nil)
		return
	}

	err = uh.service.User.Delete(r.Context(), userID)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Failed to delete user", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "User deleted successfully", nil)
}
