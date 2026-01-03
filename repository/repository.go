package repository

import (
	"inventory-system/database"

	"go.uber.org/zap"
)

type Repository struct {
	Session   SessionRepo
	User      UserRepo
	Warehouse WarehouseRepo
	Category  CategoryRepo
}

func NewRepository(db database.PgxIface, log *zap.Logger) *Repository {
	return &Repository{
		Session:   NewSessionRepo(db, log),
		User:      NewUserRepo(db, log),
		Warehouse: NewWarehouseRepo(db, log),
		Category:  NewCategoryRepo(db, log),
	}
}
