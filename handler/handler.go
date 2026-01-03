package handler

import (
	"inventory-system/service"

	"go.uber.org/zap"
)

type Handler struct {
	Auth      *AuthHandler
	User      *UserHandler
	Warehouse *WarehouseHandler
}

func NewHandlers(svc *service.Service, log *zap.Logger) Handler {
	return Handler{
		Auth:      NewAuthHandler(svc, log),
		User:      NewUserHandler(svc, log),
		Warehouse: NewWarehouseHandler(svc, log),
	}
}
