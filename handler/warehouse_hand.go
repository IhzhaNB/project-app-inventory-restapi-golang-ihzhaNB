package handler

import (
	"encoding/json"
	"inventory-system/dto/warehouse"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WarehouseHandler struct {
	service *service.Service
	log     *zap.Logger
}

func NewWarehouseHandler(service *service.Service, log *zap.Logger) *WarehouseHandler {
	return &WarehouseHandler{
		service: service,
		log:     log,
	}
}

func (wh *WarehouseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req warehouse.CreateWarehouseRequest

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call Service
	createdWarehouse, err := wh.service.Warehouse.Create(r.Context(), req)
	if err != nil {
		wh.log.Error("Failed to create warehouse", zap.Error(err))

		statusCode := http.StatusBadRequest
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusCreated, "Warehouse created successfully", createdWarehouse)
}

func (wh *WarehouseHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	warehouseIDStr := chi.URLParam(r, "id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid warehouse ID", nil)
		return
	}

	// Call service
	warehouseData, err := wh.service.Warehouse.FindByID(r.Context(), warehouseID)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "warehouse not found", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Warehouse retrieved", warehouseData)
}

func (wh *WarehouseHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	// Get pagination parameters
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Default values
	page := 1
	limit := 10

	// Parse page
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		} else {
			utils.ResponseError(w, http.StatusBadRequest, "Invalid page parameter", nil)
			return
		}
	}

	// Parse limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		} else {
			utils.ResponseError(w, http.StatusBadRequest, "Invalid limit parameter (max 100)", nil)
			return
		}
	}

	// Call service
	warehouses, pagination, err := wh.service.Warehouse.FindAll(r.Context(), page, limit)
	if err != nil {
		wh.log.Error("Failed to get warehouses", zap.Error(err))
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to retrieve warehouses", nil)
		return
	}

	// Response with pagination
	response := map[string]interface{}{
		"warehouses": warehouses,
		"pagination": pagination,
	}

	utils.ResponseSuccess(w, http.StatusOK, "Warehouses retrieved successfully", response)
}

func (wh *WarehouseHandler) Update(w http.ResponseWriter, r *http.Request) {
	warehouseIDStr := chi.URLParam(r, "id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid warehouse ID", nil)
		return
	}

	var req warehouse.UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call Service
	updatedWarehouse, err := wh.service.Warehouse.Update(r.Context(), warehouseID, req)
	if err != nil {
		wh.log.Error("Failed to update warehouse", zap.Error(err))
		utils.ResponseError(w, http.StatusBadRequest, "Failed to update warehouse", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Warehouse updated successfully", updatedWarehouse)
}

func (wh *WarehouseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	warehouseIDStr := chi.URLParam(r, "id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid warehouse ID", nil)
		return
	}

	err = wh.service.Warehouse.Delete(r.Context(), warehouseID)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Failed to delete warehouse", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Warehouse deleted successfully", nil)
}
