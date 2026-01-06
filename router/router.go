package router

import (
	"inventory-system/handler"
	"inventory-system/middleware"
	"inventory-system/model"
	"inventory-system/service"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// SetupRouter configures all HTTP routes with proper middleware and authorization
// Routes are organized by access level: Public → Authenticated → Admin-only
func SetupRouter(svc *service.Service, hdl handler.Handler) *chi.Mux {
	router := chi.NewRouter()

	// ==================== GLOBAL MIDDLEWARE (Applied to all routes) ====================
	router.Use(chimiddleware.RequestID) // Adds unique ID to each request for tracing
	router.Use(chimiddleware.RealIP)    // Gets real client IP behind proxies
	router.Use(chimiddleware.Recoverer) // Recovers from panics and returns 500
	router.Use(middleware.Logger)       // Logs all HTTP requests with Zap logger

	// ==================== PUBLIC ROUTES (No authentication required) ====================
	router.Group(func(r chi.Router) {
		// POST /api/auth/login - User authentication endpoint
		// Returns: JWT token, user info, and token expiry
		r.Post("/api/auth/login", hdl.Auth.Login)

		// GET / - API root endpoint (health check/info)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Inventory Management System API v1.0"))
		})

		// Optional: Add health check endpoint
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		})
	})

	// ==================== AUTHENTICATED ROUTES (Requires valid Bearer token) ====================
	// Accessible to: staff, admin, super_admin (all logged-in users)
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(svc.Auth)) // Validates Authorization: Bearer <token>

		// ========== AUTH MANAGEMENT ==========
		// POST /api/auth/logout - Invalidates current session token
		r.Post("/api/auth/logout", hdl.Auth.Logout)

		// ========== USER PROFILE ROUTES ==========
		// Users can manage their own profile (staff), admins can manage any user
		r.Route("/api/users", func(r chi.Router) {
			// Apply ownership check: users can only access their own data unless they're admin
			r.With(middleware.AllowSelfOrAdmin).Group(func(r chi.Router) {
				// GET /api/users/{id} - Get user details by ID
				r.Get("/{id}", hdl.User.FindByID)

				// PUT /api/users/{id} - Update user profile
				r.Put("/{id}", hdl.User.Update)
			})
		})

		// ========== WAREHOUSE READ ROUTES ==========
		// All authenticated users can view warehouse information
		r.Route("/api/warehouses", func(r chi.Router) {
			// GET /api/warehouses - List all warehouses with pagination
			// Query params: ?page=1&limit=10
			r.Get("/", hdl.Warehouse.FindAll)

			// GET /api/warehouses/{id} - Get specific warehouse details
			r.Get("/{id}", hdl.Warehouse.FindByID)
		})

		// ========== CATEGORY READ ROUTES ==========
		// All authenticated users can view product categories
		r.Route("/api/categories", func(r chi.Router) {
			// GET /api/categories - List all categories with pagination
			// Query params: ?page=1&limit=10
			r.Get("/", hdl.Category.FindAll)

			// GET /api/categories/{id} - Get specific category details
			r.Get("/{id}", hdl.Category.FindByID)
		})

		// ========== SHELF READ ROUTES ==========
		// All authenticated users can view storage shelves
		r.Route("/api/shelves", func(r chi.Router) {
			// GET /api/shelves - List all shelves with pagination
			// Query params: ?page=1&limit=10
			r.Get("/", hdl.Shelf.FindAll)

			// GET /api/shelves/{id} - Get specific shelf details
			r.Get("/{id}", hdl.Shelf.FindByID)

			// GET /api/shelves/warehouse/{warehouse_id} - List shelves by warehouse
			r.Get("/warehouse/{warehouse_id}", hdl.Shelf.FindByWarehouseID)
		})

		// ========== PRODUCT ROUTES ==========
		// Product viewing and stock management (staff can update stock)
		r.Route("/api/products", func(r chi.Router) {
			// GET /api/products - List all products with pagination
			// Query params: ?page=1&limit=10&category_id=xxx&shelf_id=xxx
			r.Get("/", hdl.Product.FindAll)

			// GET /api/products/{id} - Get specific product details
			r.Get("/{id}", hdl.Product.FindByID)

			// GET /api/products/low-stock - Get products below minimum stock level
			// FEATURE REQUIREMENT: Check minimum stock (threshold: 5)
			r.Get("/low-stock", hdl.Product.FindLowStock)

			// GET /api/products/category/{category_id} - Filter products by category
			r.Get("/category/{category_id}", hdl.Product.FindByCategoryID)

			// GET /api/products/shelf/{shelf_id} - Filter products by shelf
			r.Get("/shelf/{shelf_id}", hdl.Product.FindByShelfID)

			// PUT /api/products/{id}/stock - Update product stock quantity
			// Staff permission: Can update stock (restock/adjustment)
			// Request body: { "quantity": 50, "notes": "restock from supplier" }
			r.Put("/{id}/stock", hdl.Product.UpdateStock)
		})

		// ========== SALE TRANSACTION ROUTES ==========
		// Sales management: staff can create/view their own sales
		r.Route("/api/sales", func(r chi.Router) {
			// GET /api/sales - List sales with pagination
			// Staff: only their own sales, Admin: all sales (filtered in handler)
			// Query params: ?page=1&limit=10
			r.Get("/", hdl.Sale.FindAll)

			// POST /api/sales - Create new sale transaction
			// Validates stock availability, updates inventory, generates invoice
			// Request body: { "items": [{"product_id": "uuid", "quantity": 2}] }
			r.Post("/", hdl.Sale.Create)

			// Protected endpoints with ownership checking
			// Staff can only access their own sales, admins can access any
			r.With(middleware.AllowSelfOrAdmin).Group(func(r chi.Router) {
				// GET /api/sales/{id} - Get sale details with items
				r.Get("/{id}", hdl.Sale.FindByID)

				// PUT /api/sales/{id}/status - Update sale status
				// Allowed statuses: pending, completed, cancelled
				// Cancelling a completed sale restores product stock
				r.Put("/{id}/status", hdl.Sale.UpdateStatus)
			})
		})
	})

	// ==================== ADMIN ROUTES (Admin & Super Admin only) ====================
	// Accessible to: admin, super_admin (requires elevated privileges)
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(svc.Auth))                                     // Requires authentication
		r.Use(middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin)) // Role check

		// ========== USER MANAGEMENT ROUTES ==========
		// Full CRUD operations for user management
		r.Route("/api/admin/users", func(r chi.Router) {
			// GET /api/admin/users - List all users with pagination
			// Query params: ?page=1&limit=10
			r.Get("/", hdl.User.FindAll)

			// POST /api/admin/users - Create new user account
			// Admin can create admin/staff, Super Admin can create any role
			// Request body includes: username, email, password, role, etc.
			r.Post("/", hdl.User.Create)

			// DELETE /api/admin/users/{id} - Soft delete user account
			r.Delete("/{id}", hdl.User.Delete)
		})

		// ========== WAREHOUSE MANAGEMENT ROUTES ==========
		// Full CRUD for warehouse master data
		r.Route("/api/admin/warehouses", func(r chi.Router) {
			// POST /api/admin/warehouses - Create new warehouse
			r.Post("/", hdl.Warehouse.Create)

			// PUT /api/admin/warehouses/{id} - Update warehouse details
			r.Put("/{id}", hdl.Warehouse.Update)

			// DELETE /api/admin/warehouses/{id} - Delete warehouse (soft delete)
			r.Delete("/{id}", hdl.Warehouse.Delete)
		})

		// ========== CATEGORY MANAGEMENT ROUTES ==========
		// Full CRUD for product categories
		r.Route("/api/admin/categories", func(r chi.Router) {
			// POST /api/admin/categories - Create new category
			r.Post("/", hdl.Category.Create)

			// PUT /api/admin/categories/{id} - Update category details
			r.Put("/{id}", hdl.Category.Update)

			// DELETE /api/admin/categories/{id} - Delete category (soft delete)
			r.Delete("/{id}", hdl.Category.Delete)
		})

		// ========== SHELF MANAGEMENT ROUTES ==========
		// Full CRUD for storage shelves
		r.Route("/api/admin/shelves", func(r chi.Router) {
			// POST /api/admin/shelves - Create new shelf
			r.Post("/", hdl.Shelf.Create)

			// PUT /api/admin/shelves/{id} - Update shelf details
			r.Put("/{id}", hdl.Shelf.Update)

			// DELETE /api/admin/shelves/{id} - Delete shelf (soft delete)
			r.Delete("/{id}", hdl.Shelf.Delete)
		})

		// ========== PRODUCT MANAGEMENT ROUTES ==========
		// Full CRUD for product master data (staff can only update stock)
		r.Route("/api/admin/products", func(r chi.Router) {
			// POST /api/admin/products - Create new product
			// Requires: category_id, shelf_id, name, prices, stock info
			r.Post("/", hdl.Product.Create)

			// PUT /api/admin/products/{id} - Update product details
			// Staff cannot access this - only product stock update
			r.Put("/{id}", hdl.Product.Update)

			// DELETE /api/admin/products/{id} - Delete product (soft delete)
			// Staff cannot delete master data (requirement)
			r.Delete("/{id}", hdl.Product.Delete)
		})

		// ========== SALE ADMINISTRATION ROUTES ==========
		// Admin-only sale features (view all sales, reports)
		r.Route("/api/admin/sales", func(r chi.Router) {
			// GET /api/admin/sales - View ALL sales (no ownership filter)
			// Admin can see sales from all users, not just their own
			// Query params: ?page=1&limit=10
			r.Get("/", hdl.Sale.FindAll)

			// GET /api/admin/sales/report - Generate sales report
			// FEATURE REQUIREMENT: Sales and revenue reporting
			// Query params: ?start_date=2024-01-01&end_date=2024-12-31
			// Returns: total sales, revenue, items sold, average sale
			r.Get("/report", hdl.Sale.GetSalesReport)
		})

		// ========== REPORT ROUTES (For Future Implementation) ==========
		// Uncomment when report service is implemented
		/*
			r.Route("/api/admin/reports", func(r chi.Router) {
				// GET /api/admin/reports/products - Product inventory report
				r.Get("/products", hdl.Report.Products)

				// GET /api/admin/reports/sales - Sales analytics report
				r.Get("/sales", hdl.Report.Sales)

				// GET /api/admin/reports/revenue - Revenue report (admin only)
				// Staff cannot access revenue reports (requirement)
				r.Get("/revenue", hdl.Report.Revenue)
			})
		*/
	})

	// ==================== ERROR HANDLERS ====================
	// Handle non-existent routes
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 - Not Found", http.StatusNotFound)
	})

	// Handle unsupported HTTP methods
	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "405 - Method Not Allowed", http.StatusMethodNotAllowed)
	})

	return router
}
