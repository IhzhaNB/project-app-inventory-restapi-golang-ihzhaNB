package repository

import (
	"context"
	"fmt"
	"inventory-system/database"
	"inventory-system/model"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SaleRepo defines database operations for sales
type SaleRepo interface {
	// Sale operations
	CreateSale(ctx context.Context, sale *model.Sale) error
	FindSaleByID(ctx context.Context, id uuid.UUID) (*model.Sale, error)
	FindAllSales(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Sale, error)
	CountAllSales(ctx context.Context, userID *uuid.UUID) (int, error)
	UpdateSaleStatus(ctx context.Context, id uuid.UUID, status model.SaleStatus) error

	// Sale items operations
	CreateSaleItems(ctx context.Context, items []model.SaleItem) error
	FindSaleItems(ctx context.Context, saleID uuid.UUID) ([]model.SaleItem, error)
	FindSaleItemsWithProduct(ctx context.Context, saleID uuid.UUID) ([]model.SaleItemWithProduct, error)

	// Report operations
	GetSalesReport(ctx context.Context, startDate, endDate time.Time) (*model.SalesReport, error)
}

type saleRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

// NewSaleRepo creates new sale repository instance
func NewSaleRepo(db database.PgxIface, log *zap.Logger) SaleRepo {
	return &saleRepo{db: db, log: log}
}

// CreateSale inserts new sale record
func (sr *saleRepo) CreateSale(ctx context.Context, sale *model.Sale) error {
	query := `
		INSERT INTO sales (id, invoice_number, user_id, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	// Generate sale metadata
	now := time.Now()
	sale.ID = uuid.New()
	sale.CreatedAt = now
	sale.UpdatedAt = now

	// Set default status if empty
	if sale.Status == "" {
		sale.Status = model.SaleStatusCompleted
	}

	_, err := sr.db.Exec(ctx, query,
		sale.ID, sale.InvoiceNumber, sale.UserID, sale.TotalAmount,
		sale.Status, sale.CreatedAt, sale.UpdatedAt,
	)
	if err != nil {
		sr.log.Error("Failed to create sale", zap.Error(err))
		return fmt.Errorf("create sale failed: %w", err)
	}

	sr.log.Info("Sale created", zap.String("invoice", sale.InvoiceNumber))
	return nil
}

// CreateSaleItems inserts multiple sale items in batch
func (sr *saleRepo) CreateSaleItems(ctx context.Context, items []model.SaleItem) error {
	if len(items) == 0 {
		return fmt.Errorf("no items to insert")
	}

	// Build batch insert query
	query := `
		INSERT INTO sale_items (id, sale_id, product_id, quantity, unit_price, total_price, created_at)
		VALUES `

	args := make([]interface{}, 0)
	valueStrings := make([]string, 0)

	// Prepare values for each item
	for i, item := range items {
		// Generate metadata for each item
		now := time.Now()
		item.ID = uuid.New()
		item.CreatedAt = now
		item.UpdatedAt = now

		// Build position parameters
		pos := i * 7
		valueStrings = append(valueStrings,
			fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
				pos+1, pos+2, pos+3, pos+4, pos+5, pos+6, pos+7))

		// Add values to args slice
		args = append(args,
			item.ID, item.SaleID, item.ProductID, item.Quantity,
			item.UnitPrice, item.TotalPrice, item.CreatedAt)
	}

	// Combine all value strings
	query += strings.Join(valueStrings, ", ")

	// Execute batch insert
	_, err := sr.db.Exec(ctx, query, args...)
	if err != nil {
		sr.log.Error("Failed to create sale items", zap.Error(err))
		return fmt.Errorf("create sale items failed: %w", err)
	}

	sr.log.Info("Sale items created", zap.Int("count", len(items)))
	return nil
}

// FindSaleByID retrieves sale by ID
func (sr *saleRepo) FindSaleByID(ctx context.Context, id uuid.UUID) (*model.Sale, error) {
	query := `
		SELECT id, invoice_number, user_id, total_amount, status, created_at, updated_at, deleted_at
		FROM sales WHERE id = $1 AND deleted_at IS NULL
	`

	var sale model.Sale
	err := sr.db.QueryRow(ctx, query, id).Scan(
		&sale.ID, &sale.InvoiceNumber, &sale.UserID, &sale.TotalAmount,
		&sale.Status, &sale.CreatedAt, &sale.UpdatedAt, &sale.DeletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("sale not found: %w", err)
	}

	return &sale, nil
}

// FindSaleItems retrieves all items for a sale
func (sr *saleRepo) FindSaleItems(ctx context.Context, saleID uuid.UUID) ([]model.SaleItem, error) {
	query := `
		SELECT id, sale_id, product_id, quantity, unit_price, total_price, created_at, updated_at
		FROM sale_items WHERE sale_id = $1 ORDER BY created_at
	`

	rows, err := sr.db.Query(ctx, query, saleID)
	if err != nil {
		sr.log.Error("Failed to query sale items", zap.Error(err))
		return nil, fmt.Errorf("query sale items failed: %w", err)
	}
	defer rows.Close()

	var items []model.SaleItem
	for rows.Next() {
		var item model.SaleItem
		err := rows.Scan(
			&item.ID, &item.SaleID, &item.ProductID, &item.Quantity,
			&item.UnitPrice, &item.TotalPrice, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			sr.log.Error("Failed to scan sale item", zap.Error(err))
			return nil, fmt.Errorf("scan sale item failed: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// FindSaleItemsWithProduct retrieves sale items with product names
func (sr *saleRepo) FindSaleItemsWithProduct(ctx context.Context, saleID uuid.UUID) ([]model.SaleItemWithProduct, error) {
	query := `
		SELECT si.id, si.sale_id, si.product_id, si.quantity, si.unit_price, 
		       si.total_price, si.created_at, si.updated_at, p.name as product_name
		FROM sale_items si
		JOIN products p ON si.product_id = p.id
		WHERE si.sale_id = $1
		ORDER BY si.created_at
	`

	rows, err := sr.db.Query(ctx, query, saleID)
	if err != nil {
		sr.log.Error("Failed to query sale items with product", zap.Error(err))
		return nil, fmt.Errorf("query sale items failed: %w", err)
	}
	defer rows.Close()

	var items []model.SaleItemWithProduct
	for rows.Next() {
		var item model.SaleItemWithProduct
		err := rows.Scan(
			&item.ID, &item.SaleID, &item.ProductID, &item.Quantity,
			&item.UnitPrice, &item.TotalPrice, &item.CreatedAt, &item.UpdatedAt,
			&item.ProductName,
		)
		if err != nil {
			sr.log.Error("Failed to scan sale item", zap.Error(err))
			return nil, fmt.Errorf("scan sale item failed: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// FindAllSales retrieves sales with optional user filter and pagination
func (sr *saleRepo) FindAllSales(ctx context.Context, userID *uuid.UUID, limit, offset int) ([]model.Sale, error) {
	var query string
	var args []interface{}

	// Build query based on user filter
	if userID != nil {
		// Filter by specific user
		query = `
			SELECT id, invoice_number, user_id, total_amount, status, created_at, updated_at, deleted_at
			FROM sales WHERE user_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`
		args = []interface{}{*userID, limit, offset}
	} else {
		// Get all sales
		query = `
			SELECT id, invoice_number, user_id, total_amount, status, created_at, updated_at, deleted_at
			FROM sales WHERE deleted_at IS NULL
			ORDER BY created_at DESC LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := sr.db.Query(ctx, query, args...)
	if err != nil {
		sr.log.Error("Failed to query sales", zap.Error(err))
		return nil, fmt.Errorf("query sales failed: %w", err)
	}
	defer rows.Close()

	var sales []model.Sale
	for rows.Next() {
		var sale model.Sale
		err := rows.Scan(
			&sale.ID, &sale.InvoiceNumber, &sale.UserID, &sale.TotalAmount,
			&sale.Status, &sale.CreatedAt, &sale.UpdatedAt, &sale.DeletedAt,
		)
		if err != nil {
			sr.log.Error("Failed to scan sale", zap.Error(err))
			return nil, fmt.Errorf("scan sale failed: %w", err)
		}
		sales = append(sales, sale)
	}

	sr.log.Info("Fetched sales", zap.Int("count", len(sales)))
	return sales, nil
}

// CountAllSales counts total sales with optional user filter
func (sr *saleRepo) CountAllSales(ctx context.Context, userID *uuid.UUID) (int, error) {
	var query string
	var args []interface{}

	if userID != nil {
		query = `SELECT COUNT(*) FROM sales WHERE user_id = $1 AND deleted_at IS NULL`
		args = []interface{}{*userID}
	} else {
		query = `SELECT COUNT(*) FROM sales WHERE deleted_at IS NULL`
		args = []interface{}{}
	}

	var count int
	err := sr.db.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		sr.log.Error("Failed to count sales", zap.Error(err))
		return 0, fmt.Errorf("count sales failed: %w", err)
	}

	return count, nil
}

// UpdateSaleStatus changes sale status
func (sr *saleRepo) UpdateSaleStatus(ctx context.Context, id uuid.UUID, status model.SaleStatus) error {
	query := `UPDATE sales SET status = $1, updated_at = $2 WHERE id = $3 AND deleted_at IS NULL`

	result, err := sr.db.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		sr.log.Error("Failed to update sale status", zap.Error(err))
		return fmt.Errorf("update sale status failed: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("sale not found")
	}

	sr.log.Info("Sale status updated", zap.String("status", string(status)))
	return nil
}

// GetSalesReport generates sales report for date range
func (sr *saleRepo) GetSalesReport(ctx context.Context, startDate, endDate time.Time) (*model.SalesReport, error) {
	query := `
		SELECT 
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(
				(SELECT SUM(quantity) FROM sale_items WHERE sale_id = sales.id)
			), 0) as total_items_sold,
			CASE WHEN COUNT(*) > 0 THEN COALESCE(SUM(total_amount), 0) / COUNT(*) ELSE 0 END as average_sale
		FROM sales 
		WHERE deleted_at IS NULL 
			AND status = 'completed'
			AND created_at BETWEEN $1 AND $2
	`

	var report model.SalesReport
	err := sr.db.QueryRow(ctx, query, startDate, endDate).Scan(
		&report.TotalSales, &report.TotalRevenue, &report.TotalItemsSold, &report.AverageSale,
	)
	if err != nil {
		sr.log.Error("Failed to get sales report", zap.Error(err))
		return nil, fmt.Errorf("get sales report failed: %w", err)
	}

	// Add date range to report
	report.StartDate = startDate
	report.EndDate = endDate

	return &report, nil
}
