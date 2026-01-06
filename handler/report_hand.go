package handler

import (
	"inventory-system/dto/report"
	"inventory-system/service"
	"inventory-system/utils"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type ReportHandler struct {
	service *service.Service
	log     *zap.Logger
}

func NewReportHandler(service *service.Service, log *zap.Logger) *ReportHandler {
	return &ReportHandler{
		service: service,
		log:     log,
	}
}

// ========== 1. GET PRODUCT INVENTORY REPORT ==========
// GET /api/reports/products
// Semua authenticated user bisa akses
func (rh *ReportHandler) GetProductReport(w http.ResponseWriter, r *http.Request) {
	// Panggil service
	reportData, err := rh.service.Report.GetProductReport(r.Context())
	if err != nil {
		rh.log.Error("Failed to get product report", zap.Error(err))

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "validation") {
			statusCode = http.StatusBadRequest
		}

		utils.ResponseError(w, statusCode, "Failed to get product report", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Product report retrieved", reportData)
}

// ========== 2. GET SALES REPORT ==========
// GET /api/reports/sales?start_date=2024-01-01&end_date=2024-12-31
// Semua authenticated user bisa akses
func (rh *ReportHandler) GetSalesReport(w http.ResponseWriter, r *http.Request) {
	// Ambil query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Validasi required parameters
	if startDate == "" || endDate == "" {
		utils.ResponseError(w, http.StatusBadRequest,
			"start_date and end_date are required", nil)
		return
	}

	// Buat request DTO
	req := report.SalesReportRequest{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Panggil service
	reportData, err := rh.service.Report.GetSalesReport(r.Context(), req)
	if err != nil {
		rh.log.Error("Failed to get sales report", zap.Error(err))

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "validation") ||
			strings.Contains(err.Error(), "invalid date") ||
			strings.Contains(err.Error(), "date range") {
			statusCode = http.StatusBadRequest
		}

		utils.ResponseError(w, statusCode, "Failed to get sales report", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Sales report retrieved", reportData)
}

// ========== 3. GET REVENUE REPORT ==========
// GET /api/admin/reports/revenue?start_date=2024-01-01&end_date=2024-12-31&group_by=month
// Hanya admin & super_admin bisa akses (diatur di middleware router)
func (rh *ReportHandler) GetRevenueReport(w http.ResponseWriter, r *http.Request) {
	// Ambil query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	groupBy := r.URL.Query().Get("group_by") // optional: day, week, month

	// Validasi required parameters
	if startDate == "" || endDate == "" {
		utils.ResponseError(w, http.StatusBadRequest,
			"start_date and end_date are required", nil)
		return
	}

	// Buat request DTO
	req := report.RevenueReportRequest{
		StartDate: startDate,
		EndDate:   endDate,
		GroupBy:   groupBy,
	}

	// Panggil service
	reportData, err := rh.service.Report.GetRevenueReport(r.Context(), req)
	if err != nil {
		rh.log.Error("Failed to get revenue report", zap.Error(err))

		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "validation") ||
			strings.Contains(err.Error(), "invalid date") {
			statusCode = http.StatusBadRequest
		}

		utils.ResponseError(w, statusCode, "Failed to get revenue report", err.Error())
		return
	}

	utils.ResponseSuccess(w, http.StatusOK, "Revenue report retrieved", reportData)
}
