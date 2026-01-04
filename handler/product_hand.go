package handler

import (
	"encoding/json"
	"inventory-system/dto/product"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service *service.Service
	log     *zap.Logger
}

func NewProductHandler(service *service.Service, log *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		log:     log,
	}
}

// ========== CREATE PRODUCT ==========
func (ph *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req product.CreateProductRequest

	// Parse request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call service
	createdProduct, err := ph.service.Product.Create(r.Context(), req)
	if err != nil {
		ph.log.Error("Failed to create product", zap.Error(err))

		// Determine appropriate status code
		statusCode := http.StatusBadRequest
		if err.Error() == "category not found" || err.Error() == "shelf not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "validation failed" {
			statusCode = http.StatusUnprocessableEntity
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusCreated, "Product created successfully", createdProduct)
}

// ========== GET PRODUCT BY ID ==========
func (ph *ProductHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid product ID format", nil)
		return
	}

	// Call service
	productData, err := ph.service.Product.FindByID(r.Context(), productID)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "Product not found", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Product retrieved", productData)
}

// ========== GET ALL PRODUCTS (WITH PAGINATION) ==========
func (ph *ProductHandler) FindAll(w http.ResponseWriter, r *http.Request) {
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
	products, pagination, err := ph.service.Product.FindAll(r.Context(), page, limit)
	if err != nil {
		ph.log.Error("Failed to get products", zap.Error(err))
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to retrieve products", nil)
		return
	}

	// Response with pagination
	response := map[string]interface{}{
		"products":   products,
		"pagination": pagination,
	}

	utils.ResponseSuccess(w, http.StatusOK, "Products retrieved successfully", response)
}

// ========== GET LOW STOCK PRODUCTS ==========
func (ph *ProductHandler) FindLowStock(w http.ResponseWriter, r *http.Request) {
	// Call service (without threshold parameter)
	products, err := ph.service.Product.FindLowStock(r.Context())
	if err != nil {
		ph.log.Error("Failed to get low stock products", zap.Error(err))
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to retrieve low stock products", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Low stock products retrieved", products)
}

// ========== UPDATE PRODUCT ==========
func (ph *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid product ID format", nil)
		return
	}

	var req product.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call service
	updatedProduct, err := ph.service.Product.Update(r.Context(), productID, req)
	if err != nil {
		ph.log.Error("Failed to update product", zap.Error(err))

		statusCode := http.StatusBadRequest
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "category not found" || err.Error() == "shelf not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "validation failed" {
			statusCode = http.StatusUnprocessableEntity
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Product updated successfully", updatedProduct)
}

// ========== UPDATE PRODUCT STOCK ========== (UNTUK STAFF)
func (ph *ProductHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid product ID format", nil)
		return
	}

	var req product.UpdateStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Validate request menggunakan utils validator
	if err := utils.ValidateStruct(req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Call service
	updatedProduct, err := ph.service.Product.UpdateStock(r.Context(), productID, req)
	if err != nil {
		ph.log.Error("Failed to update product stock", zap.Error(err))

		statusCode := http.StatusBadRequest
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "stock quantity cannot be negative" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "validation failed" {
			statusCode = http.StatusUnprocessableEntity
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Product stock updated successfully", updatedProduct)
}

// ========== DELETE PRODUCT ==========
func (ph *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid product ID format", nil)
		return
	}

	err = ph.service.Product.Delete(r.Context(), productID)
	if err != nil {
		ph.log.Error("Failed to delete product", zap.Error(err))

		statusCode := http.StatusBadRequest
		if err.Error() == "product not found" {
			statusCode = http.StatusNotFound
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Product deleted successfully", nil)
}

// ========== GET PRODUCTS BY CATEGORY ==========
func (ph *ProductHandler) FindByCategoryID(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "category_id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid category ID format", nil)
		return
	}

	products, err := ph.service.Product.FindByCategoryID(r.Context(), categoryID)
	if err != nil {
		ph.log.Error("Failed to get products by category", zap.Error(err))

		statusCode := http.StatusBadRequest
		if err.Error() == "category not found" {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusInternalServerError
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Products by category retrieved", products)
}

// ========== GET PRODUCTS BY SHELF ==========
func (ph *ProductHandler) FindByShelfID(w http.ResponseWriter, r *http.Request) {
	shelfIDStr := chi.URLParam(r, "shelf_id")
	shelfID, err := uuid.Parse(shelfIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid shelf ID format", nil)
		return
	}

	products, err := ph.service.Product.FindByShelfID(r.Context(), shelfID)
	if err != nil {
		ph.log.Error("Failed to get products by shelf", zap.Error(err))

		statusCode := http.StatusBadRequest
		if err.Error() == "shelf not found" {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusInternalServerError
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Products by shelf retrieved", products)
}
