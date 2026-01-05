package sale

import (
	"time"
)

// SaleResponse represents sale data returned to client
type SaleResponse struct {
	ID            string             `json:"id"`
	InvoiceNumber string             `json:"invoice_number"`
	UserID        string             `json:"user_id"`
	TotalAmount   float64            `json:"total_amount"`
	Status        string             `json:"status"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	Items         []SaleItemResponse `json:"items,omitempty"`
}

// SaleItemResponse represents sale item data for response
type SaleItemResponse struct {
	ID          string    `json:"id"`
	ProductID   string    `json:"product_id"`
	ProductName string    `json:"product_name,omitempty"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TotalPrice  float64   `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
}

// SaleListResponse includes pagination metadata
type SaleListResponse struct {
	Sales      []SaleResponse `json:"sales"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

// SalesReportResponse contains report data
type SalesReportResponse struct {
	TotalSales     int       `json:"total_sales"`
	TotalRevenue   float64   `json:"total_revenue"`
	TotalItemsSold int       `json:"total_items_sold"`
	AverageSale    float64   `json:"average_sale"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
}
