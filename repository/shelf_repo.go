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

type ShelfRepo interface {
	Create(ctx context.Context, shelf *model.Shelf) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Shelf, error)
	FindAll(ctx context.Context) ([]model.Shelf, error)
	FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]model.Shelf, error)
	Update(ctx context.Context, shelf *model.Shelf) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type shelfRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewShelfRepo(db database.PgxIface, log *zap.Logger) ShelfRepo {
	return &shelfRepo{db: db, log: log}
}

func (sr *shelfRepo) Create(ctx context.Context, shelf *model.Shelf) error {
	query := `
		INSERT INTO shelves (id, warehouse_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	// Generate metadata sebelum insert
	now := time.Now()
	shelf.ID = uuid.New()
	shelf.CreatedAt = now
	shelf.UpdatedAt = now

	// Execute INSERT statement
	_, err := sr.db.Exec(ctx, query,
		shelf.ID,
		shelf.WarehouseID,
		shelf.Name,
		shelf.CreatedAt,
		shelf.UpdatedAt,
	)
	if err != nil {
		sr.log.Error("Failed to create shelf",
			zap.Error(err),
			zap.String("name", shelf.Name),
		)
		return fmt.Errorf("Create shelf failed: %w", err)
	}

	// Log success untuk audit trail
	sr.log.Info("Shelf Created",
		zap.String("id", shelf.ID.String()),
		zap.String("name", shelf.Name),
	)

	return nil
}

func (sr *shelfRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Shelf, error) {
	query := `
		SELECT id, warehouse_id, name, created_at, updated_at, deleted_at
		FROM shelves WHERE id = $1 AND deleted_at IS NULL
	`

	var shelf model.Shelf

	// Query single row berdasarkan ID
	err := sr.db.QueryRow(ctx, query, id).Scan(
		&shelf.ID,
		&shelf.WarehouseID,
		&shelf.Name,
		&shelf.CreatedAt,
		&shelf.UpdatedAt,
		&shelf.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Shelf not found: %w", err)
	}

	return &shelf, nil
}

func (sr *shelfRepo) FindAll(ctx context.Context) ([]model.Shelf, error) {
	query := `
		SELECT id, warehouse_id, name, created_at, updated_at, deleted_at
		FROM shelves WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	// Query semua warehouse
	rows, err := sr.db.Query(ctx, query)
	if err != nil {
		sr.log.Error("Failed to query shelf", zap.Error(err))
		return nil, fmt.Errorf("query shelf failed: %w", err)
	}
	defer rows.Close()

	// Iterate hasil query
	var shelves []model.Shelf
	for rows.Next() {
		var shelf model.Shelf
		err := rows.Scan(
			&shelf.ID,
			&shelf.WarehouseID,
			&shelf.Name,
			&shelf.CreatedAt,
			&shelf.UpdatedAt,
			&shelf.DeletedAt,
		)
		if err != nil {
			sr.log.Error("Failed to scan shelves", zap.Error(err))
			return nil, fmt.Errorf("scan shelves failed: %w", err)
		}

		shelves = append(shelves, shelf)
	}

	// Cek error dari rows
	if err = rows.Err(); err != nil {
		sr.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	sr.log.Info("Fetched all shelves", zap.Int("total_shelves", len(shelves)))
	return shelves, nil
}

func (sr *shelfRepo) FindByWarehouseID(ctx context.Context, warehouseID uuid.UUID) ([]model.Shelf, error) {
	query := `
		SELECT id, warehouse_id, name, created_at, updated_at, deleted_at
		FROM shelves WHERE warehouse_id = $1 AND deleted_at IS NULL
		ORDER BY code
	`

	rows, err := sr.db.Query(ctx, query, warehouseID)
	if err != nil {
		sr.log.Error("Failed to query shelves by warehouse",
			zap.Error(err),
		)
		return nil, fmt.Errorf("query shelves failed: %w", err)
	}
	defer rows.Close()

	var shelves []model.Shelf
	for rows.Next() {
		var shelf model.Shelf
		err := rows.Scan(
			&shelf.ID,
			&shelf.WarehouseID,
			&shelf.Name,
			&shelf.CreatedAt,
			&shelf.UpdatedAt,
			&shelf.DeletedAt,
		)
		if err != nil {
			sr.log.Error("Failed to scan shelf", zap.Error(err))
			return nil, fmt.Errorf("scan shelf failed: %w", err)
		}
		shelves = append(shelves, shelf)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	return shelves, nil
}

func (sr *shelfRepo) Update(ctx context.Context, shelf *model.Shelf) error {
	query := `
		UPDATE shelves
		SET warehouse_id = $1, name = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	// Update timestamp
	shelf.UpdatedAt = time.Now()

	// Execute UPDATE statement
	result, err := sr.db.Exec(ctx, query,
		shelf.WarehouseID,
		shelf.Name,
		shelf.UpdatedAt,
		shelf.ID,
	)
	if err != nil {
		sr.log.Error("Failed to update shelf",
			zap.Error(err),
			zap.String("id", shelf.ID.String()),
		)
		return fmt.Errorf("update shelf failed: %w", err)
	}

	// Cek jika warehouse benar2 terupdate
	if result.RowsAffected() == 0 {
		return fmt.Errorf("shelf not found")
	}

	sr.log.Info("shelf updated", zap.String("id", shelf.ID.String()))
	return nil
}

func (sr *shelfRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE shelves SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL`

	// Execute delete
	result, err := sr.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		sr.log.Error("Failed to delete shelf",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return fmt.Errorf("delete shelf failed: %w", err)
	}

	// Validasi shelf ditemukan
	if result.RowsAffected() == 0 {
		return fmt.Errorf("shelf not found")
	}

	sr.log.Info("Shelf deleted", zap.String("id", id.String()))
	return nil
}
