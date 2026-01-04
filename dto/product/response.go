package product

import (
	"time"
)

type ProductResponse struct {
	ID            string    `json:"id"`
	CategoryID    string    `json:"category_id"`
	ShelfID       string    `json:"shelf_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	UnitPrice     float64   `json:"unit_price"`
	CostPrice     float64   `json:"cost_price"`
	StockQuantity int       `json:"stock_quantity"`
	MinStockLevel int       `json:"min_stock_level"`
	IsLowStock    bool      `json:"is_low_stock"` // calculated field
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type LowStockProductResponse struct {
	ProductResponse
	StockDeficit int `json:"stock_deficit"` // berapa kekurangan dari min_stock_level
}

type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}
