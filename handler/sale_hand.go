package handler

import (
	"encoding/json"
	"inventory-system/dto/sale"
	"inventory-system/middleware"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SaleHandler handles HTTP requests for sales
type SaleHandler struct {
	service *service.Service
	log     *zap.Logger
}

// NewSaleHandler creates new sale handler instance
func NewSaleHandler(service *service.Service, log *zap.Logger) *SaleHandler {
	return &SaleHandler{
		service: service,
		log:     log,
	}
}

// Create handles POST /api/sales - creates new sale
func (sh *SaleHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req sale.CreateSaleRequest

	// Parse JSON request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Validate request data
	if err := utils.ValidateStruct(req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Get authenticated user from context
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		utils.ResponseError(w, http.StatusUnauthorized, "Authentication required", nil)
		return
	}

	// Call service to create sale
	createdSale, err := sh.service.Sale.CreateSale(r.Context(), req, user.ID)
	if err != nil {
		sh.log.Error("Failed to create sale", zap.Error(err))

		// Determine appropriate HTTP status
		statusCode := http.StatusBadRequest
		if err.Error() == "insufficient stock" {
			statusCode = http.StatusConflict
		} else if err.Error() == "not found" {
			statusCode = http.StatusNotFound
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	// Return success response
	utils.ResponseSuccess(w, http.StatusCreated, "Sale created successfully", createdSale)
}

// FindByID handles GET /api/sales/{id} - gets sale by ID
func (sh *SaleHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	// Get sale ID from URL parameter
	saleIDStr := chi.URLParam(r, "id")
	saleID, err := uuid.Parse(saleIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid sale ID format", nil)
		return
	}

	// Call service to get sale
	saleData, err := sh.service.Sale.GetSaleByID(r.Context(), saleID)
	if err != nil {
		sh.log.Error("Failed to get sale", zap.Error(err))
		utils.ResponseError(w, http.StatusNotFound, "Sale not found", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Sale retrieved successfully", saleData)
}

// FindAll handles GET /api/sales - gets all sales with pagination
func (sh *SaleHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters from query string
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Set default values
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

	// Get authenticated user for filtering
	user := middleware.GetUserFromContext(r.Context())
	var userID *uuid.UUID

	// Determine if we need to filter by user
	// Staff users only see their own sales, admins see all
	if user != nil && !user.IsStaff() {
		// Admin or super admin - see all sales
		userID = nil
	} else if user != nil {
		// Staff user - only see own sales
		userID = &user.ID
	}

	// Call service to get sales
	sales, pagination, err := sh.service.Sale.GetAllSales(r.Context(), userID, page, limit)
	if err != nil {
		sh.log.Error("Failed to get sales", zap.Error(err))
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to retrieve sales", nil)
		return
	}

	// Prepare response with pagination metadata
	response := map[string]interface{}{
		"sales":      sales,
		"pagination": pagination,
	}

	utils.ResponseSuccess(w, http.StatusOK, "Sales retrieved successfully", response)
}

// UpdateStatus handles PUT /api/sales/{id}/status - updates sale status
func (sh *SaleHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// Get sale ID from URL
	saleIDStr := chi.URLParam(r, "id")
	saleID, err := uuid.Parse(saleIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid sale ID format", nil)
		return
	}

	// Parse request body
	var req sale.UpdateSaleStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Call service to update status
	updatedSale, err := sh.service.Sale.UpdateSaleStatus(r.Context(), saleID, req)
	if err != nil {
		sh.log.Error("Failed to update sale status", zap.Error(err))

		statusCode := http.StatusBadRequest
		if err.Error() == "not found" {
			statusCode = http.StatusNotFound
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Sale status updated successfully", updatedSale)
}
