package service

import (
	"context"
	"fmt"
	"inventory-system/dto/report"
	"inventory-system/repository"
	"inventory-system/utils"
	"time"

	"go.uber.org/zap"
)

type ReportService interface {
	// 1. Product inventory report (total barang)
	GetProductReport(ctx context.Context) (*report.ProductReportResponse, error)

	// 2. Sales report (penjualan)
	GetSalesReport(ctx context.Context, req report.SalesReportRequest) (*report.SalesReportResponse, error)

	// 3. Revenue report (pendapatan) - untuk admin/super_admin saja
	GetRevenueReport(ctx context.Context, req report.RevenueReportRequest) (*report.RevenueReportResponse, error)
}

type reportService struct {
	repo *repository.Repository
	log  *zap.Logger
}

func NewReportService(repo *repository.Repository, log *zap.Logger) ReportService {
	return &reportService{
		repo: repo,
		log:  log,
	}
}

// ========== 1. PRODUCT INVENTORY REPORT ==========
func (rs *reportService) GetProductReport(ctx context.Context) (*report.ProductReportResponse, error) {
	// Langsung panggil repository
	reportData, err := rs.repo.Report.GetProductInventoryReport(ctx)
	if err != nil {
		rs.log.Error("Failed to get product report", zap.Error(err))
		return nil, fmt.Errorf("failed to get product report")
	}

	rs.log.Info("Product report generated")
	return reportData, nil
}

// ========== 2. SALES REPORT ==========
func (rs *reportService) GetSalesReport(ctx context.Context, req report.SalesReportRequest) (*report.SalesReportResponse, error) {
	// Validasi input
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse tanggal
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format. Use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format. Use YYYY-MM-DD")
	}

	// Validasi range tanggal
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}

	// Batasi max range (opsional: 1 tahun)
	maxRange := 365 * 24 * time.Hour
	if endDate.Sub(startDate) > maxRange {
		return nil, fmt.Errorf("date range cannot exceed 1 year")
	}

	// Panggil repository
	reportData, err := rs.repo.Report.GetSalesReport(ctx, startDate, endDate)
	if err != nil {
		rs.log.Error("Failed to get sales report", zap.Error(err))
		return nil, fmt.Errorf("failed to get sales report")
	}

	rs.log.Info("Sales report generated",
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
		zap.Int("total_sales", reportData.TotalSales))

	return reportData, nil
}

// ========== 3. REVENUE REPORT ==========
func (rs *reportService) GetRevenueReport(ctx context.Context, req report.RevenueReportRequest) (*report.RevenueReportResponse, error) {
	// Validasi input
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse tanggal
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format. Use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format. Use YYYY-MM-DD")
	}

	// Validasi range tanggal
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}

	// Panggil repository
	reportData, err := rs.repo.Report.GetRevenueReport(ctx, startDate, endDate, req.GroupBy)
	if err != nil {
		rs.log.Error("Failed to get revenue report", zap.Error(err))
		return nil, fmt.Errorf("failed to get revenue report")
	}

	rs.log.Info("Revenue report generated",
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
		zap.String("group_by", req.GroupBy))

	return reportData, nil
}
