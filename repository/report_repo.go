package repository

import (
	"context"
	"fmt"
	"inventory-system/database"
	"inventory-system/dto/report"
	"time"

	"go.uber.org/zap"
)

// ReportRepo - hanya 3 method sesuai requirement
type ReportRepo interface {
	// 1. Product inventory report (total barang)
	GetProductInventoryReport(ctx context.Context) (*report.ProductReportResponse, error)

	// 2. Sales report (penjualan)
	GetSalesReport(ctx context.Context, startDate, endDate time.Time) (*report.SalesReportResponse, error)

	// 3. Revenue report (pendapatan) - untuk admin/super_admin saja
	GetRevenueReport(ctx context.Context, startDate, endDate time.Time, groupBy string) (*report.RevenueReportResponse, error)
}

type reportRepo struct {
	db  database.PgxIface
	log *zap.Logger
}

func NewReportRepo(db database.PgxIface, log *zap.Logger) ReportRepo {
	return &reportRepo{db: db, log: log}
}

// ========== 1. PRODUCT INVENTORY REPORT ==========
func (rr *reportRepo) GetProductInventoryReport(ctx context.Context) (*report.ProductReportResponse, error) {
	query := `
		SELECT 
			COUNT(*) as total_products,
			COALESCE(SUM(cost_price * stock_quantity), 0) as total_value,
			COALESCE(SUM(stock_quantity), 0) as total_stock,
			COUNT(CASE WHEN stock_quantity <= min_stock_level AND stock_quantity > 0 THEN 1 END) as low_stock_count,
			COUNT(CASE WHEN stock_quantity = 0 THEN 1 END) as out_of_stock_count
		FROM products 
		WHERE deleted_at IS NULL
	`

	var result report.ProductReportResponse
	err := rr.db.QueryRow(ctx, query).Scan(
		&result.TotalProducts,
		&result.TotalValue,
		&result.TotalStock,
		&result.LowStockCount,
		&result.OutOfStockCount,
	)

	if err != nil {
		rr.log.Error("Failed to get product inventory report", zap.Error(err))
		return nil, fmt.Errorf("failed to get product report: %w", err)
	}

	// Calculate average
	if result.TotalProducts > 0 {
		result.AvgStockPerProduct = float64(result.TotalStock) / float64(result.TotalProducts)
	}

	return &result, nil
}

// ========== 2. SALES REPORT ==========
// (SAMA dengan yang di sale_repo.go, kita pindahkan ke sini)
func (rr *reportRepo) GetSalesReport(ctx context.Context, startDate, endDate time.Time) (*report.SalesReportResponse, error) {
	query := `
		SELECT 
			COUNT(*) as total_sales,
			COALESCE(SUM(total_amount), 0) as total_revenue,
			COALESCE(SUM(
				(SELECT SUM(quantity) FROM sale_items WHERE sale_id = sales.id)
			), 0) as total_items_sold,
			CASE 
				WHEN COUNT(*) > 0 THEN COALESCE(SUM(total_amount), 0) / COUNT(*) 
				ELSE 0 
			END as average_sale
		FROM sales 
		WHERE deleted_at IS NULL 
			AND status = 'completed'
			AND created_at BETWEEN $1 AND $2
	`

	var result report.SalesReportResponse
	err := rr.db.QueryRow(ctx, query, startDate, endDate).Scan(
		&result.TotalSales,
		&result.TotalRevenue,
		&result.TotalItemsSold,
		&result.AverageSale,
	)

	if err != nil {
		rr.log.Error("Failed to get sales report", zap.Error(err))
		return nil, fmt.Errorf("failed to get sales report: %w", err)
	}

	result.StartDate = startDate
	result.EndDate = endDate

	return &result, nil
}

// ========== 3. REVENUE REPORT ==========
func (rr *reportRepo) GetRevenueReport(ctx context.Context, startDate, endDate time.Time, groupBy string) (*report.RevenueReportResponse, error) {
	// Get total summary
	summary, err := rr.GetSalesReport(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}

	response := &report.RevenueReportResponse{
		TotalRevenue: summary.TotalRevenue,
		TotalSales:   summary.TotalSales,
		AverageSale:  summary.AverageSale,
		StartDate:    startDate,
		EndDate:      endDate,
	}

	// Jika groupBy diminta, ambil data per period
	if groupBy != "" {
		var periodQuery string
		switch groupBy {
		case "day":
			periodQuery = `
				SELECT 
					DATE(created_at) as period_date,
					COUNT(*) as sales_count,
					COALESCE(SUM(total_amount), 0) as revenue
				FROM sales 
				WHERE deleted_at IS NULL 
					AND status = 'completed'
					AND created_at BETWEEN $1 AND $2
				GROUP BY DATE(created_at)
				ORDER BY period_date ASC
			`
		case "month":
			periodQuery = `
				SELECT 
					TO_CHAR(created_at, 'YYYY-MM') as period_month,
					TO_CHAR(created_at, 'Month YYYY') as period_name,
					COUNT(*) as sales_count,
					COALESCE(SUM(total_amount), 0) as revenue
				FROM sales 
				WHERE deleted_at IS NULL 
					AND status = 'completed'
					AND created_at BETWEEN $1 AND $2
				GROUP BY TO_CHAR(created_at, 'YYYY-MM'), TO_CHAR(created_at, 'Month YYYY')
				ORDER BY period_month ASC
			`
		default:
			return response, nil // return summary saja tanpa grouping
		}

		rows, err := rr.db.Query(ctx, periodQuery, startDate, endDate)
		if err != nil {
			rr.log.Warn("Failed to get grouped revenue", zap.Error(err))
			return response, nil // return summary meskipun grouping gagal
		}
		defer rows.Close()

		for rows.Next() {
			var period report.TimePeriodRevenue
			if groupBy == "day" {
				var date time.Time
				err := rows.Scan(&date, &period.SalesCount, &period.Revenue)
				if err == nil {
					period.Period = date.Format("2006-01-02")
					period.Date = date.Format("2006-01-02")
					response.DailyRevenue = append(response.DailyRevenue, period)
				}
			} else if groupBy == "month" {
				var monthStr, monthName string
				err := rows.Scan(&monthStr, &monthName, &period.SalesCount, &period.Revenue)
				if err == nil {
					period.Period = monthName
					period.Date = monthStr + "-01" // untuk sorting
					response.MonthlyRevenue = append(response.MonthlyRevenue, period)
				}
			}
		}
	}

	return response, nil
}
