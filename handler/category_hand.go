package handler

import (
	"encoding/json"
	"inventory-system/dto/category"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	service *service.Service
	log     *zap.Logger
}

func NewCategoryHandler(service *service.Service, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{
		service: service,
		log:     log,
	}
}

func (ch *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req category.CreateCategoryRequest

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// call service
	createdCategory, err := ch.service.Category.Create(r.Context(), req)
	if err != nil {
		ch.log.Error("Failed to create category", zap.Error(err))

		statusCode := http.StatusBadRequest
		if strings.Contains(err.Error(), "name already exists") {
			statusCode = http.StatusConflict
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusCreated, "Category created successfully", createdCategory)
}

func (ch *CategoryHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	// Ambil id dari url param route
	categoryIDStr := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}

	// Call Service
	categoryData, err := ch.service.Category.FindByID(r.Context(), categoryID)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "category not found", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Category retrivied", categoryData)
}

func (ch *CategoryHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	categories, err := ch.service.Category.FindAll(r.Context())
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to get categoies", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Category retrivied", categories)
}

func (ch *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}

	var req category.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call service
	updatedCategory, err := ch.service.Category.Update(r.Context(), categoryID, req)
	if err != nil {
		ch.log.Error("Failed to update category", zap.Error(err))
		utils.ResponseError(w, http.StatusBadRequest, "Failed to update category", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Category updated successfully", updatedCategory)
}

func (ch *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	categoryIDStr := chi.URLParam(r, "id")
	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid category ID", nil)
		return
	}

	if err := ch.service.Category.Delete(r.Context(), categoryID); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Failed to delete category", nil)
	}

	utils.ResponseSuccess(w, http.StatusOK, "Category deleted successfully", nil)
}
