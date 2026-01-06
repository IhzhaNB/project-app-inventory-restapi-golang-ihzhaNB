package report

import "time"

// ========== PRODUCT REPORT ==========
// Product inventory summary
type ProductReportResponse struct {
	TotalProducts      int     `json:"total_products"`     // Active products count
	TotalValue         float64 `json:"total_value"`        // Inventory total value (cost * stock)
	TotalStock         int     `json:"total_stock"`        // Total items in stock
	LowStockCount      int     `json:"low_stock_count"`    // Products with stock <= min_stock_level
	OutOfStockCount    int     `json:"out_of_stock_count"` // Products with zero stock
	AvgStockPerProduct float64 `json:"avg_stock_per_product"`
}

// ========== SALES REPORT ==========
// Sales transaction summary (moved from sale)
type SalesReportResponse struct {
	TotalSales     int       `json:"total_sales"`      // Number of completed sales
	TotalRevenue   float64   `json:"total_revenue"`    // Total income from sales
	TotalItemsSold int       `json:"total_items_sold"` // Total products sold
	AverageSale    float64   `json:"average_sale"`     // Average per transaction
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
}

// ========== REVENUE REPORT ==========
// Revenue data for time period
type TimePeriodRevenue struct {
	Period     string  `json:"period"`      // "2024-01", "2024-01-15", "Week 1"
	Date       string  `json:"date"`        // For sorting (YYYY-MM-DD)
	Revenue    float64 `json:"revenue"`     // Revenue in period
	SalesCount int     `json:"sales_count"` // Transactions count
}

// Detailed revenue analytics
type RevenueReportResponse struct {
	TotalRevenue float64   `json:"total_revenue"`
	TotalSales   int       `json:"total_sales"`
	AverageSale  float64   `json:"average_sale"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`

	// Grouped data based on request
	DailyRevenue   []TimePeriodRevenue `json:"daily_revenue,omitempty"`   // When group_by=day
	WeeklyRevenue  []TimePeriodRevenue `json:"weekly_revenue,omitempty"`  // When group_by=week
	MonthlyRevenue []TimePeriodRevenue `json:"monthly_revenue,omitempty"` // When group_by=month
}
