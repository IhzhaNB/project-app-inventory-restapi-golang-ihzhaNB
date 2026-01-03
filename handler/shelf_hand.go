package handler

import (
	"encoding/json"
	"inventory-system/dto/shelf"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ShelfHandler struct {
	service *service.Service
	log     *zap.Logger
}

func NewShelfHandler(service *service.Service, log *zap.Logger) *ShelfHandler {
	return &ShelfHandler{
		service: service,
		log:     log,
	}
}

func (sh *ShelfHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req shelf.CreateShelfRequest

	// Parse request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call Service
	createdShelf, err := sh.service.Shelf.Create(r.Context(), req)
	if err != nil {
		sh.log.Error("Failed to create shelf", zap.Error(err))

		statusCode := http.StatusBadRequest
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}

		utils.ResponseError(w, statusCode, err.Error(), nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusCreated, "Shelf created successfully", createdShelf)
}

func (sh *ShelfHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	shelfIDStr := chi.URLParam(r, "id")
	shelfID, err := uuid.Parse(shelfIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid shelf ID", nil)
		return
	}

	// Call service
	shelfData, err := sh.service.Shelf.FindByID(r.Context(), shelfID)
	if err != nil {
		utils.ResponseError(w, http.StatusNotFound, "shelf not found", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Shelf retrieved", shelfData)
}

func (sh *ShelfHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	shelves, err := sh.service.Shelf.FindAll(r.Context())
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to get shelves", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Shelves retrivied", shelves)
}

func (sh *ShelfHandler) FindByWarehouseID(w http.ResponseWriter, r *http.Request) {
	warehouseIDStr := chi.URLParam(r, "warehouse_id")
	warehouseID, err := uuid.Parse(warehouseIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid warehouse ID", nil)
		return
	}

	shelves, err := sh.service.Shelf.FindByWarehouseID(r.Context(), warehouseID)
	if err != nil {
		utils.ResponseError(w, http.StatusInternalServerError, "Failed to get shelves", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Shelves retrivied", shelves)
}

func (sh *ShelfHandler) Update(w http.ResponseWriter, r *http.Request) {
	shelfIDStr := chi.URLParam(r, "id")
	shelfID, err := uuid.Parse(shelfIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid shelf ID", nil)
		return
	}

	var req shelf.UpdateShelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}
	defer r.Body.Close()

	// Call Service
	updatedShelf, err := sh.service.Shelf.Update(r.Context(), shelfID, req)
	if err != nil {
		sh.log.Error("Failed to update shelf", zap.Error(err))
		utils.ResponseError(w, http.StatusBadRequest, "Failed to update shelf", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "shelf updated successfully", updatedShelf)
}

func (sh *ShelfHandler) Delete(w http.ResponseWriter, r *http.Request) {
	shelfIDStr := chi.URLParam(r, "id")
	shelfID, err := uuid.Parse(shelfIDStr)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Invalid shelf ID", nil)
		return
	}

	err = sh.service.Shelf.Delete(r.Context(), shelfID)
	if err != nil {
		utils.ResponseError(w, http.StatusBadRequest, "Failed to delete shelf", nil)
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Shelf deleted successfully", nil)
}
