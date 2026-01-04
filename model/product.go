package model

import "github.com/google/uuid"

type Product struct {
	BaseModel
	CategoryID    uuid.UUID `db:"category_id" json:"category_id"`
	ShelfID       uuid.UUID `db:"shelf_id" json:"shelf_id"`
	Name          string    `db:"name" json:"name"`
	Description   string    `db:"description" json:"description,omitempty"`
	UnitPrice     float64   `db:"unit_price" json:"unit_price"`
	CostPrice     float64   `db:"cost_price" json:"cost_price"`
	StockQuantity int       `db:"stock_quantity" json:"stock_quantity"`
	MinStockLevel int       `db:"min_stock_level" json:"min_stock_level"`
}
