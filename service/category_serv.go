package service

import (
	"context"
	"fmt"
	"inventory-system/dto/category"
	"inventory-system/model"
	"inventory-system/repository"
	"inventory-system/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CategoryService interface {
	Create(ctx context.Context, req category.CreateCategoryRequest) (*category.CategoryResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*category.CategoryResponse, error)
	FindAll(ctx context.Context, page int, limit int) ([]category.CategoryResponse, utils.Pagination, error)
	Update(ctx context.Context, id uuid.UUID, req category.UpdateCategoryRequest) (*category.CategoryResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewCategoryService(repo *repository.Repository, log *zap.Logger) CategoryService {
	return &categoryService{repo: repo, log: log}
}

func (cs *categoryService) Create(ctx context.Context, req category.CreateCategoryRequest) (*category.CategoryResponse, error) {
	// Validate input
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check name uniquess
	if existing, _ := cs.repo.Category.FindByName(ctx, req.Name); existing != nil {
		return nil, fmt.Errorf("name already exists")
	}

	// Prepare category object
	newCategory := &model.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	// Save to db
	if err := cs.repo.Category.Create(ctx, newCategory); err != nil {
		cs.log.Error("Failed to create category", zap.Error(err))
		return nil, fmt.Errorf("Failed to create category")
	}

	// Prepare response
	response := cs.convertToResponse(newCategory)

	cs.log.Info("Category created", zap.String("category_id", newCategory.ID.String()))
	return response, nil
}

func (cs *categoryService) FindByID(ctx context.Context, id uuid.UUID) (*category.CategoryResponse, error) {
	foundCategory, err := cs.repo.Category.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	return cs.convertToResponse(foundCategory), nil
}

func (cs *categoryService) FindAll(ctx context.Context, page int, limit int) ([]category.CategoryResponse, utils.Pagination, error) {
	// Setup pagination
	pagination := utils.NewPagination(page, limit)

	// Get data with pagination
	categories, err := cs.repo.Category.FindAll(ctx, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to get categories")
	}

	// Get total count
	total, err := cs.repo.Category.CountAll(ctx)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count categories")
	}

	// Set total in pagination
	pagination.SetTotal(total)

	// Convert to response
	responses := make([]category.CategoryResponse, 0, len(categories))
	for _, c := range categories {
		responses = append(responses, *cs.convertToResponse(&c))
	}

	cs.log.Info("Categories fetched with pagination",
		zap.Int("page", page),
		zap.Int("limit", limit),
		zap.Int("total", total))

	return responses, pagination, nil
}

func (cs *categoryService) Update(ctx context.Context, id uuid.UUID, req category.UpdateCategoryRequest) (*category.CategoryResponse, error) {
	categoryToUpdate, err := cs.repo.Category.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	updated := false

	// update fields if provided and different
	if req.Name != nil && *req.Name != categoryToUpdate.Name {
		categoryToUpdate.Name = *req.Name
		updated = true
	}

	if req.Description != nil && *req.Description != categoryToUpdate.Description {
		categoryToUpdate.Description = *req.Description
		updated = true
	}

	if updated {
		if err := cs.repo.Category.Update(ctx, categoryToUpdate); err != nil {
			return nil, fmt.Errorf("failed to update category")
		}
	}

	return cs.convertToResponse(categoryToUpdate), nil
}

func (cs *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := cs.repo.Category.FindByID(ctx, id); err != nil {
		return fmt.Errorf("category not found")
	}

	if err := cs.repo.Category.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to detele category")
	}

	cs.log.Info("Category deleted", zap.String("category_id", id.String()))
	return nil
}

func (cs *categoryService) convertToResponse(c *model.Category) *category.CategoryResponse {
	return &category.CategoryResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
