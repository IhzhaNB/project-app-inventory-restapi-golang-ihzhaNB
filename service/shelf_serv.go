package service

import (
	"context"
	"fmt"
	"inventory-system/dto/shelf"
	"inventory-system/model"
	"inventory-system/repository"
	"inventory-system/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ShelfService interface {
	Create(ctx context.Context, req shelf.CreateShelfRequest) (*shelf.ShelfResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*shelf.ShelfResponse, error)
	FindAll(ctx context.Context) ([]shelf.ShelfResponse, error)
	FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]shelf.ShelfResponse, error)
	Update(ctx context.Context, id uuid.UUID, req shelf.UpdateShelfRequest) (*shelf.ShelfResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type shelfService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewShelfService(repo *repository.Repository, log *zap.Logger) ShelfService {
	return &shelfService{repo: repo, log: log}
}

func (ss *shelfService) Create(ctx context.Context, req shelf.CreateShelfRequest) (*shelf.ShelfResponse, error) {
	// Validate input
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		return nil, fmt.Errorf("invalid warehouse ID format")
	}

	// Check Warehouse ID
	if _, err := ss.repo.Warehouse.FindByID(ctx, warehouseID); err != nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	// prepare warehouse object
	newShelf := &model.Shelf{
		WarehouseID: warehouseID,
		Name:        req.Name,
	}

	// Save to database
	if err := ss.repo.Shelf.Create(ctx, newShelf); err != nil {
		ss.log.Error("Failed to create shelf", zap.Error(err))
		return nil, fmt.Errorf("failed to create shelf")
	}

	// prepare response
	response := &shelf.ShelfResponse{
		ID:          newShelf.ID.String(),
		WarehouseID: newShelf.WarehouseID.String(),
		Name:        newShelf.Name,
		CreatedAt:   newShelf.CreatedAt,
		UpdatedAt:   newShelf.UpdatedAt,
	}

	ss.log.Info("Shelf created",
		zap.String("shelf_id", newShelf.ID.String()),
		zap.String("warehouse_id", newShelf.WarehouseID.String()))
	return response, nil
}

func (ss *shelfService) FindByID(ctx context.Context, id uuid.UUID) (*shelf.ShelfResponse, error) {
	foundShelf, err := ss.repo.Shelf.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("shelf not found")
	}

	return &shelf.ShelfResponse{
		ID:          foundShelf.ID.String(),
		WarehouseID: foundShelf.WarehouseID.String(),
		Name:        foundShelf.Name,
		CreatedAt:   foundShelf.CreatedAt,
		UpdatedAt:   foundShelf.UpdatedAt,
	}, nil
}

func (ss *shelfService) FindAll(ctx context.Context) ([]shelf.ShelfResponse, error) {
	shelves, err := ss.repo.Shelf.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelf")
	}

	var responses []shelf.ShelfResponse
	for _, s := range shelves { // s = shelf (single)
		responses = append(responses, shelf.ShelfResponse{
			ID:          s.ID.String(),
			WarehouseID: s.WarehouseID.String(),
			Name:        s.Name,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		})
	}

	return responses, nil
}

func (ss *shelfService) FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]shelf.ShelfResponse, error) {
	// Check warehouse exists
	if _, err := ss.repo.Warehouse.FindByID(ctx, warehouseID); err != nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	// Get shelves by warehouse ID
	shelves, err := ss.repo.Shelf.FindByWarehouseID(ctx, warehouseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shelves")
	}

	// Convert to response
	var responses []shelf.ShelfResponse
	for _, s := range shelves {
		responses = append(responses, shelf.ShelfResponse{
			ID:          s.ID.String(),
			WarehouseID: s.WarehouseID.String(),
			Name:        s.Name,
			CreatedAt:   s.CreatedAt,
			UpdatedAt:   s.UpdatedAt,
		})
	}

	return responses, nil
}

func (ss *shelfService) Update(ctx context.Context, id uuid.UUID, req shelf.UpdateShelfRequest) (*shelf.ShelfResponse, error) {
	shelfToUpdate, err := ss.repo.Shelf.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("shelf not found")
	}

	updated := false

	// Check and update warehouse ID if provided
	if req.WarehouseID != nil {
		warehouseIDStr, err := uuid.Parse(*req.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("invalid warehouse ID format")
		}

		// Check warehouse exists
		if _, err := ss.repo.Warehouse.FindByID(ctx, warehouseIDStr); err != nil {
			return nil, fmt.Errorf("warehouse not found")
		}

		if warehouseIDStr != shelfToUpdate.WarehouseID {
			shelfToUpdate.WarehouseID = warehouseIDStr
			updated = true
		}
	}

	if req.Name != nil && *req.Name != shelfToUpdate.Name {
		shelfToUpdate.Name = *req.Name
		updated = true
	}

	// Save if change were made
	if updated {
		if err := ss.repo.Shelf.Update(ctx, shelfToUpdate); err != nil {
			return nil, fmt.Errorf("failed to update warehouse")
		}
	}

	return &shelf.ShelfResponse{
		ID:          shelfToUpdate.ID.String(),
		WarehouseID: shelfToUpdate.WarehouseID.String(),
		Name:        shelfToUpdate.Name,
		CreatedAt:   shelfToUpdate.CreatedAt,
		UpdatedAt:   shelfToUpdate.UpdatedAt,
	}, nil
}

func (ss *shelfService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := ss.repo.Shelf.FindByID(ctx, id); err != nil {
		return fmt.Errorf("Shelf not found")
	}

	if err := ss.repo.Shelf.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete shelf")
	}

	ss.log.Info("Shelf deleted", zap.String("shelf_id", id.String()))
	return nil
}
