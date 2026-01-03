package repository

import (
	"context"
	"fmt"
	"inventory-system/database"
	"inventory-system/model"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type WarehouseRepo interface {
	Create(ctx context.Context, warehouse *model.Warehouse) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Warehouse, error)
	FindByCode(ctx context.Context, code string) (*model.Warehouse, error)
	FindAll(ctx context.Context) ([]model.Warehouse, error)
	Update(ctx context.Context, warehouse *model.Warehouse) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type warehouseRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewWarehouseRepo(db database.PgxIface, log *zap.Logger) WarehouseRepo {
	return &warehouseRepo{db: db, log: log}
}

func (wr *warehouseRepo) Create(ctx context.Context, warehouse *model.Warehouse) error {
	query := `
		INSERT INTO warehouses (id, code, name, address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Generate metadata sebelum insert
	now := time.Now()
	warehouse.ID = uuid.New()
	warehouse.CreatedAt = now
	warehouse.UpdatedAt = now

	// Execute INSERT statement
	_, err := wr.db.Exec(ctx, query,
		warehouse.ID,
		warehouse.Code,
		warehouse.Name,
		warehouse.Address,
		warehouse.CreatedAt,
		warehouse.UpdatedAt,
	)
	if err != nil {
		wr.log.Error("Failed to create warehouse",
			zap.Error(err),
			zap.String("name", warehouse.Name),
		)
		return fmt.Errorf("Create warehouse failed: %w", err)
	}

	// Log success untuk audit trail
	wr.log.Info("Warehouse Created",
		zap.String("id", warehouse.ID.String()),
		zap.String("name", warehouse.Name),
	)

	return nil
}

func (wr *warehouseRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Warehouse, error) {
	query := `
		SELECT id, code, name, address, created_at, updated_at, deleted_at
		FROM warehouses WHERE id = $1 AND deleted_at IS NULL
	`

	var warehouse model.Warehouse

	// Query single row berdasarkan ID
	err := wr.db.QueryRow(ctx, query, id).Scan(
		&warehouse.ID,
		&warehouse.Code,
		&warehouse.Name,
		&warehouse.Address,
		&warehouse.CreatedAt,
		&warehouse.UpdatedAt,
		&warehouse.DeletedAt,
	)
	if err != nil {
		// Warehouse tidak ditemukan
		wr.log.Warn("Warehouse not found",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("Warehouse not found: %w", err)
	}

	return &warehouse, nil
}

func (wr *warehouseRepo) FindByCode(ctx context.Context, code string) (*model.Warehouse, error) {
	query := `
		SELECT id, code, name, address, created_at, updated_at, deleted_at
		FROM warehouses WHERE code = $1 AND deleted_at IS NULL
	`

	var warehouse model.Warehouse

	// Query single row berdasarkan Code
	err := wr.db.QueryRow(ctx, query, code).Scan(
		&warehouse.ID,
		&warehouse.Code,
		&warehouse.Name,
		&warehouse.Address,
		&warehouse.CreatedAt,
		&warehouse.UpdatedAt,
		&warehouse.DeletedAt,
	)
	if err != nil {
		// Warehouse tidak ditemukan
		wr.log.Warn("Warehouse not found",
			zap.String("code", code),
			zap.Error(err),
		)
		return nil, fmt.Errorf("Warehouse not found: %w", err)
	}

	return &warehouse, nil
}

func (wr *warehouseRepo) FindAll(ctx context.Context) ([]model.Warehouse, error) {
	query := `
		SELECT id, code, name, address, created_at, updated_at, deleted_at
		FROM warehouses WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	// Query semua warehouse
	rows, err := wr.db.Query(ctx, query)
	if err != nil {
		wr.log.Error("Failed to query warehouse", zap.Error(err))
		return nil, fmt.Errorf("query warehouse failed: %w", err)
	}
	defer rows.Close()

	// Iterate hasil query
	var warehouses []model.Warehouse
	for rows.Next() {
		var warehouse model.Warehouse
		err := rows.Scan(
			&warehouse.ID,
			&warehouse.Code,
			&warehouse.Name,
			&warehouse.Address,
			&warehouse.CreatedAt,
			&warehouse.UpdatedAt,
			&warehouse.DeletedAt,
		)
		if err != nil {
			wr.log.Error("Failed to scan warehouse", zap.Error(err))
			return nil, fmt.Errorf("scan warehouse failed: %w", err)
		}

		warehouses = append(warehouses, warehouse)
	}

	// Cek error dari rows
	if err = rows.Err(); err != nil {
		wr.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	wr.log.Info("Fetched all warehouse", zap.Int("total_warehouses", len(warehouses)))
	return warehouses, nil
}

func (wr *warehouseRepo) Update(ctx context.Context, warehouse *model.Warehouse) error {
	query := `
		UPDATE warehouses
		SET code = $1, name = $2, address = $3, updated_at = $4
		WHERE id = $5 AND deleted_at IS NULL
	`

	// Update timestamp
	warehouse.UpdatedAt = time.Now()

	// Execute UPDATE statement
	result, err := wr.db.Exec(ctx, query,
		warehouse.Code,
		warehouse.Name,
		warehouse.Address,
		warehouse.UpdatedAt,
		warehouse.ID,
	)
	if err != nil {
		wr.log.Error("Failed to update warehouse",
			zap.Error(err),
			zap.String("id", warehouse.ID.String()),
		)
		return fmt.Errorf("update warehouse failed: %w", err)
	}

	// Cek jika warehouse benar2 terupdate
	if result.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	wr.log.Info("Warehouse updated", zap.String("id", warehouse.ID.String()))
	return nil
}

func (wr *warehouseRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE warehouses SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	now := time.Now()

	// Execute delete
	result, err := wr.db.Exec(ctx, query, now, id)
	if err != nil {
		wr.log.Error("Failed to delete warehouse",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return fmt.Errorf("delete warehouse failed: %w", err)
	}

	// Validasi warehouse ditemukan
	if result.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	wr.log.Info("Warehouse deleted", zap.String("id", id.String()))
	return nil
}
