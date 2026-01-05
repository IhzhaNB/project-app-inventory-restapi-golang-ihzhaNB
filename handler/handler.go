package handler

import (
	"inventory-system/service"

	"go.uber.org/zap"
)

type Handler struct {
	Auth      *AuthHandler
	User      *UserHandler
	Warehouse *WarehouseHandler
	Category  *CategoryHandler
	Shelf     *ShelfHandler
	Product   *ProductHandler
	Sale      *SaleHandler
}

func NewHandlers(svc *service.Service, log *zap.Logger) Handler {
	return Handler{
		Auth:      NewAuthHandler(svc, log),
		User:      NewUserHandler(svc, log),
		Warehouse: NewWarehouseHandler(svc, log),
		Category:  NewCategoryHandler(svc, log),
		Shelf:     NewShelfHandler(svc, log),
		Product:   NewProductHandler(svc, log),
		Sale:      NewSaleHandler(svc, log),
	}
}
