package model

import (
	"time"

	"github.com/google/uuid"
)

// SaleStatus represents possible states of a sale transaction
type SaleStatus string

const (
	SaleStatusPending   SaleStatus = "pending"
	SaleStatusCompleted SaleStatus = "completed"
	SaleStatusCancelled SaleStatus = "cancelled"
)

// Sale represents a sales transaction
type Sale struct {
	BaseModel
	InvoiceNumber string     `db:"invoice_number" json:"invoice_number"`
	UserID        uuid.UUID  `db:"user_id" json:"user_id"`
	TotalAmount   float64    `db:"total_amount" json:"total_amount"`
	Status        SaleStatus `db:"status" json:"status"`
}

// SaleItem represents individual product sold in a sale
type SaleItem struct {
	BaseModel
	SaleID     uuid.UUID `db:"sale_id" json:"sale_id"`
	ProductID  uuid.UUID `db:"product_id" json:"product_id"`
	Quantity   int       `db:"quantity" json:"quantity"`
	UnitPrice  float64   `db:"unit_price" json:"unit_price"`
	TotalPrice float64   `db:"total_price" json:"total_price"`
}

// SaleItemWithProduct combines sale item with product details for reporting
type SaleItemWithProduct struct {
	SaleItem
	ProductName string `db:"product_name" json:"product_name"`
}

// SalesReport contains aggregated sales data for reporting
type SalesReport struct {
	TotalSales     int       `json:"total_sales"`
	TotalRevenue   float64   `json:"total_revenue"`
	TotalItemsSold int       `json:"total_items_sold"`
	AverageSale    float64   `json:"average_sale"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
}
