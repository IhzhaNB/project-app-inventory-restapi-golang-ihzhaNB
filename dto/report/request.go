package report

// ProductReportRequest - Get product inventory report
type ProductReportRequest struct {
	CategoryID      *string `json:"category_id,omitempty" validate:"omitempty,uuid4"`
	ShelfID         *string `json:"shelf_id,omitempty" validate:"omitempty,uuid4"`
	WarehouseID     *string `json:"warehouse_id,omitempty" validate:"omitempty,uuid4"`
	IncludeLowStock *bool   `json:"include_low_stock,omitempty"` // true = only low stock products
}

// SalesReportRequest - Get sales transaction report
type SalesReportRequest struct {
	StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02"`
}

// RevenueReportRequest - Get revenue analytics report
type RevenueReportRequest struct {
	StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02"`
	GroupBy   string `json:"group_by,omitempty" validate:"omitempty,oneof=day week month"`
}
