package service

import (
	"context"
	"fmt"
	"inventory-system/dto/product"
	"inventory-system/model"
	"inventory-system/repository"
	"inventory-system/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type ProductService interface {
	Create(ctx context.Context, req product.CreateProductRequest) (*product.ProductResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*product.ProductResponse, error)
	FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]product.ProductResponse, error)
	FindByShelfID(ctx context.Context, shelfID uuid.UUID) ([]product.ProductResponse, error)
	FindAll(ctx context.Context, page int, limit int) ([]product.ProductResponse, utils.Pagination, error)
	FindLowStock(ctx context.Context) ([]product.ProductResponse, error)
	Update(ctx context.Context, id uuid.UUID, req product.UpdateProductRequest) (*product.ProductResponse, error)
	UpdateStock(ctx context.Context, id uuid.UUID, req product.UpdateStockRequest) (*product.ProductResponse, error)
	CheckStock(ctx context.Context, id uuid.UUID, requiredQuantity int) (*model.Product, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type productService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewProductService(repo *repository.Repository, log *zap.Logger) ProductService {
	return &productService{repo: repo, log: log}
}

// ========== CREATE ==========
func (ps *productService) Create(ctx context.Context, req product.CreateProductRequest) (*product.ProductResponse, error) {
	// Validate input
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check Category ID
	categoryID, err := uuid.Parse(req.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category ID format")
	}
	if _, err := ps.repo.Category.FindByID(ctx, categoryID); err != nil {
		return nil, fmt.Errorf("category not found")
	}

	// Check Shelf ID
	shelfID, err := uuid.Parse(req.ShelfID)
	if err != nil {
		return nil, fmt.Errorf("invalid shelf ID format")
	}
	if _, err := ps.repo.Shelf.FindByID(ctx, shelfID); err != nil {
		return nil, fmt.Errorf("shelf not found")
	}

	// Prepare product object
	newProduct := &model.Product{
		CategoryID:    categoryID,
		ShelfID:       shelfID,
		Name:          req.Name,
		Description:   req.Description,
		UnitPrice:     req.UnitPrice,
		CostPrice:     req.CostPrice,
		StockQuantity: req.StockQuantity,
		MinStockLevel: req.MinStockLevel,
	}

	// Set default min stock level
	if newProduct.MinStockLevel == 0 {
		newProduct.MinStockLevel = 5 // DEFAULT sesuai requirement
	}

	// Save to db
	if err := ps.repo.Product.Create(ctx, newProduct); err != nil {
		ps.log.Error("Failed to create product", zap.Error(err))
		return nil, fmt.Errorf("failed to create product")
	}

	// Response
	response := ps.convertToResponse(newProduct)

	ps.log.Info("Product created",
		zap.String("product_id", newProduct.ID.String()),
		zap.String("category_id", newProduct.CategoryID.String()),
		zap.String("shelf_id", newProduct.ShelfID.String()),
	)
	return response, nil
}

// ========== FIND BY ID ==========
func (ps *productService) FindByID(ctx context.Context, id uuid.UUID) (*product.ProductResponse, error) {
	foundProduct, err := ps.repo.Product.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	return ps.convertToResponse(foundProduct), nil
}

// ========== FIND BY CATEGORY ==========
func (ps *productService) FindByCategoryID(ctx context.Context, categoryID uuid.UUID) ([]product.ProductResponse, error) {
	// Validate category exists
	if _, err := ps.repo.Category.FindByID(ctx, categoryID); err != nil {
		return nil, fmt.Errorf("category not found")
	}

	// Get products by category
	products, err := ps.repo.Product.FindByCategoryID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by category")
	}

	// Convert to response
	responses := make([]product.ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, *ps.convertToResponse(&p))
	}

	return responses, nil
}

// ========== FIND BY SHELF ==========
func (ps *productService) FindByShelfID(ctx context.Context, shelfID uuid.UUID) ([]product.ProductResponse, error) {
	// Validate shelf exists
	if _, err := ps.repo.Shelf.FindByID(ctx, shelfID); err != nil {
		return nil, fmt.Errorf("shelf not found")
	}

	// Get products by shelf
	products, err := ps.repo.Product.FindByShelfID(ctx, shelfID)
	if err != nil {
		return nil, fmt.Errorf("failed to get products by shelf")
	}

	// Convert to response
	responses := make([]product.ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, *ps.convertToResponse(&p))
	}

	return responses, nil
}

// ========== FIND ALL WITH PAGINATION ==========
func (ps *productService) FindAll(ctx context.Context, page int, limit int) ([]product.ProductResponse, utils.Pagination, error) {
	// Setup pagination
	pagination := utils.NewPagination(page, limit)

	// Get data with pagination
	products, err := ps.repo.Product.FindAll(ctx, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to get products")
	}

	// Get total count
	total, err := ps.repo.Product.CountAll(ctx)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count products")
	}

	// Set total in pagination
	pagination.SetTotal(total)

	// Convert to response
	responses := make([]product.ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, *ps.convertToResponse(&p))
	}

	return responses, pagination, nil
}

// ========== FIND LOW STOCK ==========
func (ps *productService) FindLowStock(ctx context.Context) ([]product.ProductResponse, error) {
	products, err := ps.repo.Product.FindLowStock(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock products")
	}

	// Convert to response
	responses := make([]product.ProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, *ps.convertToResponse(&p))
	}

	ps.log.Info("Low stock products fetched", zap.Int("count", len(responses)))
	return responses, nil
}

// ========== UPDATE ==========
func (ps *productService) Update(ctx context.Context, id uuid.UUID, req product.UpdateProductRequest) (*product.ProductResponse, error) {
	// Validate input
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing product
	productToUpdate, err := ps.repo.Product.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	updated := false

	// Check and update category ID if provided
	if req.CategoryID != nil {
		categoryID, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("invalid category ID format")
		}
		// Validate category exists
		if _, err := ps.repo.Category.FindByID(ctx, categoryID); err != nil {
			return nil, fmt.Errorf("category not found")
		}
		if categoryID != productToUpdate.CategoryID {
			productToUpdate.CategoryID = categoryID
			updated = true
		}
	}

	// Check and update shelf ID if provided
	if req.ShelfID != nil {
		shelfID, err := uuid.Parse(*req.ShelfID)
		if err != nil {
			return nil, fmt.Errorf("invalid shelf ID format")
		}
		// Validate shelf exists
		if _, err := ps.repo.Shelf.FindByID(ctx, shelfID); err != nil {
			return nil, fmt.Errorf("shelf not found")
		}
		if shelfID != productToUpdate.ShelfID {
			productToUpdate.ShelfID = shelfID
			updated = true
		}
	}

	// Update other fields
	if req.Name != nil && *req.Name != productToUpdate.Name {
		productToUpdate.Name = *req.Name
		updated = true
	}
	if req.Description != nil && *req.Description != productToUpdate.Description {
		productToUpdate.Description = *req.Description
		updated = true
	}
	if req.UnitPrice != nil && *req.UnitPrice != productToUpdate.UnitPrice {
		productToUpdate.UnitPrice = *req.UnitPrice
		updated = true
	}
	if req.CostPrice != nil && *req.CostPrice != productToUpdate.CostPrice {
		productToUpdate.CostPrice = *req.CostPrice
		updated = true
	}
	if req.StockQuantity != nil && *req.StockQuantity != productToUpdate.StockQuantity {
		productToUpdate.StockQuantity = *req.StockQuantity
		updated = true
	}
	if req.MinStockLevel != nil && *req.MinStockLevel != productToUpdate.MinStockLevel {
		productToUpdate.MinStockLevel = *req.MinStockLevel
		updated = true
	}

	// Save if changes were made
	if updated {
		if err := ps.repo.Product.Update(ctx, productToUpdate); err != nil {
			return nil, fmt.Errorf("failed to update product")
		}
	}

	return ps.convertToResponse(productToUpdate), nil
}

// ========== UPDATE STOCK ========== (UNTUK STAFF)
func (ps *productService) UpdateStock(ctx context.Context, id uuid.UUID, req product.UpdateStockRequest) (*product.ProductResponse, error) {
	// Validate DTO
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing product first
	existingProduct, err := ps.repo.Product.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("product not found")
	}

	// Update stock in database
	if err := ps.repo.Product.UpdateStock(ctx, id, req.Quantity); err != nil {
		return nil, fmt.Errorf("failed to update stock")
	}

	// Get updated product
	updatedProduct, err := ps.repo.Product.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated product")
	}

	// Log stock change for audit trail
	change := req.Quantity - existingProduct.StockQuantity
	ps.log.Info("Product stock updated",
		zap.String("product_id", id.String()),
		zap.String("product_name", existingProduct.Name),
		zap.Int("old_stock", existingProduct.StockQuantity),
		zap.Int("new_stock", updatedProduct.StockQuantity),
		zap.Int("change", change),
		zap.String("notes", req.Notes))

	return ps.convertToResponse(updatedProduct), nil
}

// ========== CHECK STOCK ========== (UNTUK SALE VALIDATION)
func (ps *productService) CheckStock(ctx context.Context, id uuid.UUID, requiredQuantity int) (*model.Product, error) {
	if requiredQuantity <= 0 {
		return nil, fmt.Errorf("required quantity must be positive")
	}

	product, err := ps.repo.Product.CheckStock(ctx, id, requiredQuantity)
	if err != nil {
		return nil, fmt.Errorf("stock check failed: %w", err)
	}

	return product, nil
}

// ========== DELETE ==========
func (ps *productService) Delete(ctx context.Context, id uuid.UUID) error {
	// Check if product exists
	if _, err := ps.repo.Product.FindByID(ctx, id); err != nil {
		return fmt.Errorf("product not found")
	}

	// Delete product
	if err := ps.repo.Product.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete product")
	}

	ps.log.Info("Product deleted", zap.String("product_id", id.String()))
	return nil
}

// ========== HELPER: CONVERT TO RESPONSE ==========
func (ps *productService) convertToResponse(p *model.Product) *product.ProductResponse {
	// Calculate if low stock
	isLowStock := p.StockQuantity <= p.MinStockLevel

	return &product.ProductResponse{
		ID:            p.ID.String(),
		CategoryID:    p.CategoryID.String(),
		ShelfID:       p.ShelfID.String(),
		Name:          p.Name,
		Description:   p.Description,
		UnitPrice:     p.UnitPrice,
		CostPrice:     p.CostPrice,
		StockQuantity: p.StockQuantity,
		MinStockLevel: p.MinStockLevel,
		IsLowStock:    isLowStock, // Calculated field
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}
