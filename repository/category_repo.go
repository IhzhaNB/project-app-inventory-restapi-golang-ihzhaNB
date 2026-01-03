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

type CategoryRepo interface {
	Create(ctx context.Context, category *model.Category) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error)
	FindByName(ctx context.Context, code string) (*model.Category, error)
	FindAll(ctx context.Context) ([]model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewCategoryRepo(db database.PgxIface, log *zap.Logger) CategoryRepo {
	return &categoryRepo{db: db, log: log}
}

func (cr *categoryRepo) Create(ctx context.Context, category *model.Category) error {
	query := `
		INSERT INTO categories (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	// Generate metadata sebelum insert
	now := time.Now()
	category.ID = uuid.New()
	category.CreatedAt = now
	category.UpdatedAt = now

	// Execute INSERT statement
	_, err := cr.db.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Description,
		category.CreatedAt,
		category.UpdatedAt,
	)
	if err != nil {
		cr.log.Error("Failed to create category",
			zap.Error(err),
			zap.String("name", category.Name),
		)
		return fmt.Errorf("Create category failed: %w", err)
	}

	// Log success untuk audit trail
	cr.log.Info("Warehouse Created",
		zap.String("id", category.ID.String()),
		zap.String("name", category.Name),
	)

	return nil
}

func (cr *categoryRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, deleted_at
		FROM categories WHERE id = $1 AND deleted_at IS NULL
	`

	var category model.Category

	// Query single row berdasarkan ID
	if err := cr.db.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	); err != nil {
		cr.log.Warn("Category not found",
			zap.String("id", id.String()),
			zap.Error(err),
		)
		return nil, fmt.Errorf("Category not found: %w", err)
	}

	return &category, nil
}

func (cr *categoryRepo) FindByName(ctx context.Context, name string) (*model.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, deleted_at
		FROM categories WHERE name = $1 AND deleted_at IS NULL
	`

	var category model.Category

	// Query single row berdasarkan Name
	if err := cr.db.QueryRow(ctx, query, name).Scan(
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
		&category.DeletedAt,
	); err != nil {
		cr.log.Warn("Category not found",
			zap.String("name", name),
			zap.Error(err),
		)
		return nil, fmt.Errorf("Category not found: %w", err)
	}

	return &category, nil
}

func (cr *categoryRepo) FindAll(ctx context.Context) ([]model.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at, deleted_at
		FROM categories WHERE deleted_at IS NULL
		ORDER BY created_at DESC
	`

	// Query semua category
	rows, err := cr.db.Query(ctx, query)
	if err != nil {
		cr.log.Error("Failed to query category", zap.Error(err))
		return nil, fmt.Errorf("query category failed: %w", err)
	}
	defer rows.Close()

	// Iterate hasil query
	var categories []model.Category
	for rows.Next() {
		var category model.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
			&category.DeletedAt,
		); err != nil {
			cr.log.Error("Failed to scan category", zap.Error(err))
			return nil, fmt.Errorf("scan category failed: %w", err)
		}

		categories = append(categories, category)
	}

	// Cek error dari rows
	if err = rows.Err(); err != nil {
		cr.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	cr.log.Info("Fetched all categories", zap.Int("total_categories", len(categories)))
	return categories, nil
}

func (cr *categoryRepo) Update(ctx context.Context, category *model.Category) error {
	query := `
		UPDATE categories
		SET name = $1, description = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	category.UpdatedAt = time.Now()

	result, err := cr.db.Exec(ctx, query,
		category.Name,
		category.Description,
		category.UpdatedAt,
		category.ID,
	)
	if err != nil {
		cr.log.Error("Failed to update category",
			zap.Error(err),
			zap.String("id", category.ID.String()),
		)
		return fmt.Errorf("update category failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("warehouse not found")
	}

	cr.log.Info("Category updated", zap.String("id", category.ID.String()))
	return nil
}

func (cr *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE categories SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	now := time.Now()

	result, err := cr.db.Exec(ctx, query, now, id)
	if err != nil {
		cr.log.Error("Failed to delete category",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return fmt.Errorf("delete category failed")
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	cr.log.Info("Category deleted", zap.String("id", id.String()))
	return nil
}
