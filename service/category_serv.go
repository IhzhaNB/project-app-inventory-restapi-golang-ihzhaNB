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
	FindAll(ctx context.Context) ([]category.CategoryResponse, error)
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
	response := &category.CategoryResponse{
		ID:          newCategory.ID.String(),
		Name:        newCategory.Name,
		Description: newCategory.Description,
		CreatedAt:   newCategory.CreatedAt,
		UpdatedAt:   newCategory.UpdatedAt,
	}

	cs.log.Info("Category created", zap.String("category_id", newCategory.ID.String()))
	return response, nil
}

func (cs *categoryService) FindByID(ctx context.Context, id uuid.UUID) (*category.CategoryResponse, error) {
	foundCategory, err := cs.repo.Category.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("category not found")
	}

	return &category.CategoryResponse{
		ID:          foundCategory.ID.String(),
		Name:        foundCategory.Name,
		Description: foundCategory.Description,
		CreatedAt:   foundCategory.CreatedAt,
		UpdatedAt:   foundCategory.UpdatedAt,
	}, nil
}

func (cs *categoryService) FindAll(ctx context.Context) ([]category.CategoryResponse, error) {
	categories, err := cs.repo.Category.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories")
	}

	var responses []category.CategoryResponse
	for _, c := range categories { // c = category single
		responses = append(responses, category.CategoryResponse{
			ID:          c.ID.String(),
			Name:        c.Name,
			Description: c.Description,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		})
	}

	return responses, nil
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

	return &category.CategoryResponse{
		ID:          categoryToUpdate.ID.String(),
		Name:        categoryToUpdate.Name,
		Description: categoryToUpdate.Description,
		CreatedAt:   categoryToUpdate.CreatedAt,
		UpdatedAt:   categoryToUpdate.UpdatedAt,
	}, nil
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
