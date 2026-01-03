package service

import (
	"inventory-system/repository"

	"go.uber.org/zap"
)

type Service struct {
	Auth      AuthService
	User      UserService
	Warehouse WarehouseService
	Category  CategoryService
}

func NewService(repo *repository.Repository, log *zap.Logger) *Service {
	return &Service{
		Auth:      NewAuthService(repo, log),
		User:      NewUserService(repo, log),
		Warehouse: NewWarehouseService(repo, log),
		Category:  NewCategoryService(repo, log),
	}
}
