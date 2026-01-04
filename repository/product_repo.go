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

type ProductRepo interface {
	Create(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error)
	FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.Product, error)
	FindByShelfID(ctx context.Context, shelfID uuid.UUID) ([]model.Product, error)
	FindAll(ctx context.Context, limit int, offset int) ([]model.Product, error)
	CountAll(ctx context.Context) (int, error)
	FindLowStock(ctx context.Context) ([]model.Product, error)
	Update(ctx context.Context, product *model.Product) error
	UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error
	CheckStock(ctx context.Context, id uuid.UUID, requiredQuantity int) (*model.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type productRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewProductRepo(db database.PgxIface, log *zap.Logger) ProductRepo {
	return &productRepo{db: db, log: log}
}

func (pr *productRepo) Create(ctx context.Context, product *model.Product) error {
	query := `
		INSERT INTO products (
    		id, category_id, shelf_id, name, description, 
    		unit_price, cost_price, stock_quantity, min_stock_level,
    		created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	// Generate metadata sebelum insert
	now := time.Now()
	product.ID = uuid.New()
	product.CreatedAt = now
	product.UpdatedAt = now

	// Execute INSERT statement
	_, err := pr.db.Exec(ctx, query,
		product.ID, product.CategoryID, product.ShelfID, product.Name,
		product.Description, product.UnitPrice, product.CostPrice, product.StockQuantity,
		product.MinStockLevel, product.CreatedAt, product.UpdatedAt,
	)
	if err != nil {
		pr.log.Error("Failed to create product", zap.Error(err),
			zap.String("name", product.Name),
		)
		return fmt.Errorf("Create product failed: %w", err)
	}

	pr.log.Info("Product created",
		zap.String("id", product.ID.String()),
		zap.String("name", product.Name),
	)
	return nil
}

func (pr *productRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	query := `
		SELECT 
			id, category_id, shelf_id, name, description,
			unit_price, cost_price, stock_quantity, min_stock_level,
			created_at, updated_at, deleted_at
		FROM products 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var product model.Product

	err := pr.db.QueryRow(ctx, query, id).Scan(
		&product.ID, &product.CategoryID, &product.ShelfID, &product.Name,
		&product.Description, &product.UnitPrice, &product.CostPrice, &product.StockQuantity,
		&product.MinStockLevel, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("Product not found: %w", err)
	}

	return &product, nil
}

func (pr *productRepo) FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]model.Product, error) {
	query := `
        SELECT 
            id, category_id, shelf_id, name, description,
            unit_price, cost_price, stock_quantity, min_stock_level,
            created_at, updated_at, deleted_at
        FROM products 
        WHERE category_id = $1 AND deleted_at IS NULL
        ORDER BY name
    `

	rows, err := pr.db.Query(ctx, query, categoryID)
	if err != nil {
		pr.log.Error("Failed to query products by category", zap.Error(err),
			zap.String("category_id", categoryID.String()))
		return nil, fmt.Errorf("query products by category failed: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		err := rows.Scan(
			&product.ID, &product.CategoryID, &product.ShelfID, &product.Name,
			&product.Description, &product.UnitPrice, &product.CostPrice, &product.StockQuantity,
			&product.MinStockLevel, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		)
		if err != nil {
			pr.log.Error("Failed to scan product", zap.Error(err))
			return nil, fmt.Errorf("scan product failed: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		pr.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	pr.log.Info("Fetched products by category",
		zap.String("category_id", categoryID.String()),
		zap.Int("count", len(products)))

	return products, nil
}

func (pr *productRepo) FindByShelfID(ctx context.Context, shelfID uuid.UUID) ([]model.Product, error) {
	query := `
        SELECT 
            id, category_id, shelf_id, name, description,
            unit_price, cost_price, stock_quantity, min_stock_level,
            created_at, updated_at, deleted_at
        FROM products 
        WHERE shelf_id = $1 AND deleted_at IS NULL
        ORDER BY name
    `

	rows, err := pr.db.Query(ctx, query, shelfID)
	if err != nil {
		pr.log.Error("Failed to query products by shelf", zap.Error(err),
			zap.String("shelf_id", shelfID.String()))
		return nil, fmt.Errorf("query products by shelf failed: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		err := rows.Scan(
			&product.ID, &product.CategoryID, &product.ShelfID, &product.Name,
			&product.Description, &product.UnitPrice, &product.CostPrice, &product.StockQuantity,
			&product.MinStockLevel, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		)
		if err != nil {
			pr.log.Error("Failed to scan product", zap.Error(err))
			return nil, fmt.Errorf("scan product failed: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		pr.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	pr.log.Info("Fetched products by shelf",
		zap.String("shelf_id", shelfID.String()),
		zap.Int("count", len(products)))

	return products, nil
}

func (pr *productRepo) FindAll(ctx context.Context, limit int, offset int) ([]model.Product, error) {
	query := `
        SELECT 
            id, category_id, shelf_id, name, description,
            unit_price, cost_price, stock_quantity, min_stock_level,
            created_at, updated_at, deleted_at
        FROM products 
        WHERE deleted_at IS NULL
        ORDER BY created_at DESC
        LIMIT $1 OFFSET $2
    `

	rows, err := pr.db.Query(ctx, query, limit, offset)
	if err != nil {
		pr.log.Error("Failed to query products", zap.Error(err))
		return nil, fmt.Errorf("query products failed: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		err := rows.Scan(
			&product.ID, &product.CategoryID, &product.ShelfID, &product.Name,
			&product.Description, &product.UnitPrice, &product.CostPrice, &product.StockQuantity,
			&product.MinStockLevel, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		)
		if err != nil {
			pr.log.Error("Failed to scan product", zap.Error(err))
			return nil, fmt.Errorf("scan product failed: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		pr.log.Error("Rows iteration error", zap.Error(err))
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	pr.log.Info("Fetched products with pagination",
		zap.Int("limit", limit),
		zap.Int("offset", offset),
		zap.Int("count", len(products)))

	return products, nil
}

func (pr *productRepo) CountAll(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM products WHERE deleted_at IS NULL`

	var count int
	err := pr.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		pr.log.Error("Failed to count products", zap.Error(err))
		return 0, fmt.Errorf("count products failed: %w", err)
	}

	return count, nil
}

func (pr *productRepo) FindLowStock(ctx context.Context) ([]model.Product, error) {
	query := `
		SELECT 
			id, category_id, shelf_id, name, description,
			unit_price, cost_price, stock_quantity, min_stock_level,
			created_at, updated_at, deleted_at
		FROM products 
		WHERE deleted_at IS NULL 
			AND stock_quantity <= min_stock_level
			AND stock_quantity > 0
		ORDER BY stock_quantity ASC
	`

	rows, err := pr.db.Query(ctx, query)
	if err != nil {
		pr.log.Error("Failed to query low stock products", zap.Error(err))
		return nil, fmt.Errorf("query low stock products failed: %w", err)
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var product model.Product
		if err := rows.Scan(
			&product.ID, &product.CategoryID, &product.ShelfID, &product.Name,
			&product.Description, &product.UnitPrice, &product.CostPrice, &product.StockQuantity,
			&product.MinStockLevel, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt,
		); err != nil {
			pr.log.Error("Failed to scan product", zap.Error(err))
			return nil, fmt.Errorf("scan product failed: %w", err)
		}
		products = append(products, product)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration failed: %w", err)
	}

	pr.log.Info("Low stock products fetched", zap.Int("count", len(products)))
	return products, nil
}

func (pr *productRepo) Update(ctx context.Context, product *model.Product) error {
	query := `
		UPDATE products 
		SET 
			category_id = $1,
			shelf_id = $2,
			name = $3,
			description = $4,
			unit_price = $5,
			cost_price = $6,
			stock_quantity = $7,
			min_stock_level = $8,
			updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
	`

	// Update timestamp
	product.UpdatedAt = time.Now()

	result, err := pr.db.Exec(ctx, query,
		product.CategoryID,
		product.ShelfID,
		product.Name,
		product.Description,
		product.UnitPrice,
		product.CostPrice,
		product.StockQuantity,
		product.MinStockLevel,
		product.UpdatedAt,
		product.ID,
	)
	if err != nil {
		pr.log.Error("Failed to update product",
			zap.Error(err),
			zap.String("id", product.ID.String()),
		)
		return fmt.Errorf("update product failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	pr.log.Info("product updated", zap.String("id", product.ID.String()))
	return nil
}

func (pr *productRepo) UpdateStock(ctx context.Context, id uuid.UUID, quantity int) error {
	// Validasi stok tidak negatif
	if quantity < 0 {
		return fmt.Errorf("stock quantity cannot be negative")
	}

	query := `
		UPDATE products 
		SET 
			stock_quantity = $1,
			updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := pr.db.Exec(ctx, query, quantity, time.Now(), id)
	if err != nil {
		pr.log.Error("Failed to update stock product", zap.Error(err),
			zap.String("id", id.String()),
		)
		return fmt.Errorf("update product stock failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	pr.log.Info("product stock updated", zap.String("id", id.String()))
	return nil
}

func (pr *productRepo) CheckStock(ctx context.Context, id uuid.UUID, requiredQuantity int) (*model.Product, error) {
	query := `
        SELECT 
            id, name, stock_quantity, unit_price, min_stock_level
        FROM products 
        WHERE id = $1 
            AND deleted_at IS NULL 
            AND stock_quantity >= $2
    `

	var product model.Product
	err := pr.db.QueryRow(ctx, query, id, requiredQuantity).Scan(
		&product.ID,
		&product.Name,
		&product.StockQuantity,
		&product.UnitPrice,
		&product.MinStockLevel,
	)

	if err != nil {
		pr.log.Error("Insufficient stock or product not found",
			zap.Error(err),
			zap.String("product_id", id.String()),
			zap.Int("required", requiredQuantity))
		return nil, fmt.Errorf("insufficient stock or product not found")
	}

	return &product, nil
}

func (pr *productRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE products 
		SET deleted_at = $1  -- FIX: HILANGKAN KOMA
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := pr.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		pr.log.Error("Failed to delete product",
			zap.Error(err),
			zap.String("id", id.String()),
		)
		return fmt.Errorf("delete product failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("product not found")
	}

	pr.log.Info("Product deleted", zap.String("id", id.String()))
	return nil
}
