package service

import (
	"context"
	"fmt"
	"inventory-system/dto/warehouse"
	"inventory-system/model"
	"inventory-system/repository"
	"inventory-system/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WarehouseService interface {
	Create(ctx context.Context, req warehouse.CreateWarehouseRequest) (*warehouse.WarehouseResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*warehouse.WarehouseResponse, error)
	FindAll(ctx context.Context) ([]warehouse.WarehouseResponse, error)
	Update(ctx context.Context, id uuid.UUID, req warehouse.UpdateWarehouseRequest) (*warehouse.WarehouseResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type warehouseService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewWarehouseService(repo *repository.Repository, log *zap.Logger) WarehouseService {
	return &warehouseService{repo: repo, log: log}
}

func (ws *warehouseService) Create(ctx context.Context, req warehouse.CreateWarehouseRequest) (*warehouse.WarehouseResponse, error) {
	// Validate input
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// prepare warehouse object
	newWarehouse := &model.Warehouse{
		Name:    req.Name,
		Address: req.Address,
	}

	// Save to database
	if err := ws.repo.Warehouse.Create(ctx, newWarehouse); err != nil {
		ws.log.Error("Failed to create warehouse", zap.Error(err))
		return nil, fmt.Errorf("failed to create warehouse")
	}

	// prepare response
	response := &warehouse.WarehouseResponse{
		ID:        newWarehouse.ID.String(),
		Name:      newWarehouse.Name,
		Address:   newWarehouse.Address,
		CreatedAt: newWarehouse.CreatedAt,
		UpdatedAt: newWarehouse.UpdatedAt,
	}

	ws.log.Info("Warehouse created", zap.String("warehouse_id", newWarehouse.ID.String()))
	return response, nil
}

func (ws *warehouseService) FindByID(ctx context.Context, id uuid.UUID) (*warehouse.WarehouseResponse, error) {
	foundWarehouse, err := ws.repo.Warehouse.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	return &warehouse.WarehouseResponse{
		ID:        foundWarehouse.ID.String(),
		Name:      foundWarehouse.Name,
		Address:   foundWarehouse.Address,
		CreatedAt: foundWarehouse.CreatedAt,
		UpdatedAt: foundWarehouse.UpdatedAt,
	}, nil
}

func (ws *warehouseService) FindAll(ctx context.Context) ([]warehouse.WarehouseResponse, error) {
	warehouses, err := ws.repo.Warehouse.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get warehouses")
	}

	var responses []warehouse.WarehouseResponse
	for _, w := range warehouses { // w = warehouse (single)
		responses = append(responses, warehouse.WarehouseResponse{
			ID:        w.ID.String(),
			Name:      w.Name,
			Address:   w.Address,
			CreatedAt: w.CreatedAt,
			UpdatedAt: w.UpdatedAt,
		})
	}

	return responses, nil
}

func (ws *warehouseService) Update(ctx context.Context, id uuid.UUID, req warehouse.UpdateWarehouseRequest) (*warehouse.WarehouseResponse, error) {
	warehouseToUpdate, err := ws.repo.Warehouse.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("warehouse not found")
	}

	updated := false

	// update fields if provided and different
	if req.Name != nil && *req.Name != warehouseToUpdate.Name {
		warehouseToUpdate.Name = *req.Name
		updated = true
	}

	if req.Address != nil && *req.Address != warehouseToUpdate.Address {
		warehouseToUpdate.Address = *req.Address
		updated = true
	}

	// Save if change were made
	if updated {
		if err := ws.repo.Warehouse.Update(ctx, warehouseToUpdate); err != nil {
			return nil, fmt.Errorf("failed to update warehouse")
		}
	}

	return &warehouse.WarehouseResponse{
		ID:        warehouseToUpdate.ID.String(),
		Name:      warehouseToUpdate.Name,
		Address:   warehouseToUpdate.Address,
		CreatedAt: warehouseToUpdate.CreatedAt,
		UpdatedAt: warehouseToUpdate.UpdatedAt,
	}, nil
}

func (ws *warehouseService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := ws.repo.Warehouse.FindByID(ctx, id); err != nil {
		return fmt.Errorf("warehouse not found")
	}

	if err := ws.repo.Warehouse.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete warehouse")
	}

	ws.log.Info("Warehouse deleted", zap.String("warehouse_id", id.String()))
	return nil
}
