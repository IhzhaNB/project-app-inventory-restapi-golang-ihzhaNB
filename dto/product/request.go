package product

// CreateProductRequest - untuk create product baru
type CreateProductRequest struct {
	CategoryID    string  `json:"category_id" validate:"required,uuid4"`
	ShelfID       string  `json:"shelf_id" validate:"required,uuid4"`
	Name          string  `json:"name" validate:"required,min=3,max=200"`
	Description   string  `json:"description,omitempty" validate:"max=1000"`
	UnitPrice     float64 `json:"unit_price" validate:"required,min=0"`
	CostPrice     float64 `json:"cost_price" validate:"required,min=0"`
	StockQuantity int     `json:"stock_quantity" validate:"min=0"`
	MinStockLevel int     `json:"min_stock_level" validate:"min=0"`
}

// UpdateProductRequest - untuk update product (semua field optional)
type UpdateProductRequest struct {
	CategoryID    *string  `json:"category_id,omitempty" validate:"omitempty,uuid4"`
	ShelfID       *string  `json:"shelf_id,omitempty" validate:"omitempty,uuid4"`
	Name          *string  `json:"name,omitempty" validate:"omitempty,min=3,max=200"`
	Description   *string  `json:"description,omitempty" validate:"omitempty,max=1000"`
	UnitPrice     *float64 `json:"unit_price,omitempty" validate:"omitempty,min=0"`
	CostPrice     *float64 `json:"cost_price,omitempty" validate:"omitempty,min=0"`
	StockQuantity *int     `json:"stock_quantity,omitempty" validate:"omitempty,min=0"`
	MinStockLevel *int     `json:"min_stock_level,omitempty" validate:"omitempty,min=0"`
}

// UpdateStockRequest - khusus untuk update stock quantity saja
type UpdateStockRequest struct {
	Quantity int    `json:"quantity" validate:"required,min=0"`
	Notes    string `json:"notes,omitempty" validate:"max=500"` // catatan kenapa update stock
}
