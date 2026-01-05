package service

import (
	"context"
	"fmt"
	"inventory-system/dto/sale"
	"inventory-system/model"
	"inventory-system/repository"
	"inventory-system/utils"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SaleService defines business logic for sales
type SaleService interface {
	// Sale operations
	CreateSale(ctx context.Context, req sale.CreateSaleRequest, userID uuid.UUID) (*sale.SaleResponse, error)
	GetSaleByID(ctx context.Context, id uuid.UUID) (*sale.SaleResponse, error)
	GetAllSales(ctx context.Context, userID *uuid.UUID, page, limit int) ([]sale.SaleResponse, utils.Pagination, error)
	UpdateSaleStatus(ctx context.Context, id uuid.UUID, req sale.UpdateSaleStatusRequest) (*sale.SaleResponse, error)

	// Report operations
	GetSalesReport(ctx context.Context, req sale.SalesReportRequest) (*sale.SalesReportResponse, error)
}

type saleService struct {
	repo *repository.Repository
	log  *zap.Logger
}

// NewSaleService creates new sale service instance
func NewSaleService(repo *repository.Repository, log *zap.Logger) SaleService {
	return &saleService{repo: repo, log: log}
}

// CreateSale processes new sale transaction
func (ss *saleService) CreateSale(ctx context.Context, req sale.CreateSaleRequest, userID uuid.UUID) (*sale.SaleResponse, error) {
	// Validate request structure
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Ensure at least one item
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("sale must have at least one item")
	}

	// Process each sale item
	var totalAmount float64 = 0
	var saleItems []model.SaleItem

	for _, itemReq := range req.Items {
		// Convert product ID string to UUID
		productID, err := uuid.Parse(itemReq.ProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid product ID format: %s", itemReq.ProductID)
		}

		// Check if product has sufficient stock
		product, err := ss.repo.Product.CheckStock(ctx, productID, itemReq.Quantity)
		if err != nil {
			return nil, fmt.Errorf("insufficient stock for product %s: %w", itemReq.ProductID, err)
		}

		// Calculate item total
		itemTotal := product.UnitPrice * float64(itemReq.Quantity)
		totalAmount += itemTotal

		// Prepare sale item
		saleItem := model.SaleItem{
			ProductID:  productID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  product.UnitPrice,
			TotalPrice: itemTotal,
		}
		saleItems = append(saleItems, saleItem)
	}

	// Generate unique invoice number
	invoiceNumber := generateInvoiceNumber()

	// Create sale record
	newSale := &model.Sale{
		InvoiceNumber: invoiceNumber,
		UserID:        userID,
		TotalAmount:   totalAmount,
		Status:        model.SaleStatusCompleted,
	}

	// Save sale to database
	if err := ss.repo.Sale.CreateSale(ctx, newSale); err != nil {
		return nil, fmt.Errorf("failed to create sale: %w", err)
	}

	// Link sale ID to all items
	for i := range saleItems {
		saleItems[i].SaleID = newSale.ID
	}

	// Save sale items
	if err := ss.repo.Sale.CreateSaleItems(ctx, saleItems); err != nil {
		return nil, fmt.Errorf("failed to create sale items: %w", err)
	}

	// Update product stock (deduct sold quantities)
	for _, item := range saleItems {
		product, err := ss.repo.Product.FindByID(ctx, item.ProductID)
		if err != nil {
			ss.log.Error("Failed to get product for stock update", zap.Error(err))
			continue
		}

		// Calculate new stock quantity
		newStock := product.StockQuantity - item.Quantity
		if newStock < 0 {
			newStock = 0
		}

		// Update product stock
		if err := ss.repo.Product.UpdateStock(ctx, item.ProductID, newStock); err != nil {
			ss.log.Error("Failed to update product stock", zap.Error(err))
		}
	}

	// Get complete sale details for response
	saleWithItems, err := ss.getSaleWithItems(ctx, newSale.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sale details: %w", err)
	}

	ss.log.Info("Sale created",
		zap.String("invoice", newSale.InvoiceNumber),
		zap.Float64("total", newSale.TotalAmount))

	return saleWithItems, nil
}

// GetSaleByID retrieves sale with all items
func (ss *saleService) GetSaleByID(ctx context.Context, id uuid.UUID) (*sale.SaleResponse, error) {
	// Get sale from repository
	saleData, err := ss.repo.Sale.FindSaleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sale not found")
	}

	// Get sale with items
	return ss.getSaleWithItems(ctx, saleData.ID)
}

// GetAllSales retrieves sales list with pagination
func (ss *saleService) GetAllSales(ctx context.Context, userID *uuid.UUID, page, limit int) ([]sale.SaleResponse, utils.Pagination, error) {
	// Initialize pagination
	pagination := utils.NewPagination(page, limit)

	// Get sales from repository
	sales, err := ss.repo.Sale.FindAllSales(ctx, userID, pagination.Limit, pagination.Offset())
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to get sales: %w", err)
	}

	// Get total count for pagination
	total, err := ss.repo.Sale.CountAllSales(ctx, userID)
	if err != nil {
		return nil, pagination, fmt.Errorf("failed to count sales: %w", err)
	}

	// Update pagination with total count
	pagination.SetTotal(total)

	// Convert sales to response format
	responses := make([]sale.SaleResponse, 0, len(sales))
	for _, s := range sales {
		response := sale.SaleResponse{
			ID:            s.ID.String(),
			InvoiceNumber: s.InvoiceNumber,
			UserID:        s.UserID.String(),
			TotalAmount:   s.TotalAmount,
			Status:        string(s.Status),
			CreatedAt:     s.CreatedAt,
			UpdatedAt:     s.UpdatedAt,
		}
		responses = append(responses, response)
	}

	return responses, pagination, nil
}

// UpdateSaleStatus changes sale status and handles stock restoration if cancelled
func (ss *saleService) UpdateSaleStatus(ctx context.Context, id uuid.UUID, req sale.UpdateSaleStatusRequest) (*sale.SaleResponse, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get existing sale to check current status
	existingSale, err := ss.repo.Sale.FindSaleByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("sale not found")
	}

	// Convert string status to model type
	var newStatus model.SaleStatus
	switch req.Status {
	case "pending":
		newStatus = model.SaleStatusPending
	case "completed":
		newStatus = model.SaleStatusCompleted
	case "cancelled":
		newStatus = model.SaleStatusCancelled
	default:
		return nil, fmt.Errorf("invalid status: %s", req.Status)
	}

	// Update status in database
	if err := ss.repo.Sale.UpdateSaleStatus(ctx, id, newStatus); err != nil {
		return nil, fmt.Errorf("failed to update sale status: %w", err)
	}

	// If cancelling a completed sale, restore product stock
	if existingSale.Status == model.SaleStatusCompleted && newStatus == model.SaleStatusCancelled {
		if err := ss.restoreProductStock(ctx, id); err != nil {
			ss.log.Error("Failed to restore stock after cancellation", zap.Error(err))
		}
	}

	// Get updated sale with items
	return ss.getSaleWithItems(ctx, id)
}

// GetSalesReport generates sales report for given date range
func (ss *saleService) GetSalesReport(ctx context.Context, req sale.SalesReportRequest) (*sale.SalesReportResponse, error) {
	// Validate request
	if err := utils.ValidateStruct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Parse date strings to time.Time
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format. Use YYYY-MM-DD")
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format. Use YYYY-MM-DD")
	}

	// Validate date range
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start date cannot be after end date")
	}

	// Get report from repository
	report, err := ss.repo.Sale.GetSalesReport(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sales report: %w", err)
	}

	// Convert to response DTO
	response := &sale.SalesReportResponse{
		TotalSales:     report.TotalSales,
		TotalRevenue:   report.TotalRevenue,
		TotalItemsSold: report.TotalItemsSold,
		AverageSale:    report.AverageSale,
		StartDate:      report.StartDate,
		EndDate:        report.EndDate,
	}

	ss.log.Info("Sales report generated",
		zap.Time("start", startDate),
		zap.Time("end", endDate),
		zap.Int("sales", report.TotalSales))

	return response, nil
}

// getSaleWithItems helper: retrieves sale with all items and product details
func (ss *saleService) getSaleWithItems(ctx context.Context, saleID uuid.UUID) (*sale.SaleResponse, error) {
	// Get sale details
	saleData, err := ss.repo.Sale.FindSaleByID(ctx, saleID)
	if err != nil {
		return nil, err
	}

	// Try to get items with product names
	itemsWithProduct, err := ss.repo.Sale.FindSaleItemsWithProduct(ctx, saleID)
	if err != nil {
		// Fallback to items without product names
		items, err := ss.repo.Sale.FindSaleItems(ctx, saleID)
		if err != nil {
			return nil, fmt.Errorf("failed to get sale items: %w", err)
		}

		// Convert items to response without product names
		itemResponses := make([]sale.SaleItemResponse, 0, len(items))
		for _, item := range items {
			itemResponses = append(itemResponses, sale.SaleItemResponse{
				ID:         item.ID.String(),
				ProductID:  item.ProductID.String(),
				Quantity:   item.Quantity,
				UnitPrice:  item.UnitPrice,
				TotalPrice: item.TotalPrice,
				CreatedAt:  item.CreatedAt,
			})
		}

		return &sale.SaleResponse{
			ID:            saleData.ID.String(),
			InvoiceNumber: saleData.InvoiceNumber,
			UserID:        saleData.UserID.String(),
			TotalAmount:   saleData.TotalAmount,
			Status:        string(saleData.Status),
			CreatedAt:     saleData.CreatedAt,
			UpdatedAt:     saleData.UpdatedAt,
			Items:         itemResponses,
		}, nil
	}

	// Convert items with product names to response
	itemResponses := make([]sale.SaleItemResponse, 0, len(itemsWithProduct))
	for _, item := range itemsWithProduct {
		itemResponses = append(itemResponses, sale.SaleItemResponse{
			ID:          item.ID.String(),
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
			CreatedAt:   item.CreatedAt,
		})
	}

	return &sale.SaleResponse{
		ID:            saleData.ID.String(),
		InvoiceNumber: saleData.InvoiceNumber,
		UserID:        saleData.UserID.String(),
		TotalAmount:   saleData.TotalAmount,
		Status:        string(saleData.Status),
		CreatedAt:     saleData.CreatedAt,
		UpdatedAt:     saleData.UpdatedAt,
		Items:         itemResponses,
	}, nil
}

// restoreProductStock helper: restores product stock when sale is cancelled
func (ss *saleService) restoreProductStock(ctx context.Context, saleID uuid.UUID) error {
	// Get all items from cancelled sale
	items, err := ss.repo.Sale.FindSaleItems(ctx, saleID)
	if err != nil {
		return fmt.Errorf("failed to get sale items: %w", err)
	}

	// Restore stock for each product
	for _, item := range items {
		product, err := ss.repo.Product.FindByID(ctx, item.ProductID)
		if err != nil {
			ss.log.Error("Failed to get product", zap.Error(err))
			continue
		}

		// Calculate restored stock
		newStock := product.StockQuantity + item.Quantity

		// Update product stock
		if err := ss.repo.Product.UpdateStock(ctx, item.ProductID, newStock); err != nil {
			ss.log.Error("Failed to restore stock", zap.Error(err))
		}
	}

	return nil
}

// generateInvoiceNumber helper: creates unique invoice number
func generateInvoiceNumber() string {
	datePart := time.Now().Format("20060102")
	randomPart := fmt.Sprintf("%04d", time.Now().Nanosecond()%10000)
	return fmt.Sprintf("INV-%s-%s", datePart, randomPart)
}
