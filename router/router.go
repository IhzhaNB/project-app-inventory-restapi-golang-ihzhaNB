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

func SetupRouter(svc *service.Service, hdl handler.Handler) *chi.Mux {
	router := chi.NewRouter()

	// ==================== GLOBAL MIDDLEWARE ====================
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(chimiddleware.Recoverer)
	router.Use(middleware.Logger)

	// ==================== PUBLIC ROUTES ====================
	router.Group(func(r chi.Router) {
		r.Post("/api/auth/login", hdl.Auth.Login)
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Inventory Management System API v1.0"))
		})
	})

	// ==================== AUTHENTICATED ROUTES ====================
	// Semua user yang sudah login (staff, admin, super_admin)
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(svc.Auth))

		// ========== LOGOUT ==========
		r.Post("/api/auth/logout", hdl.Auth.Logout)

		// ========== USER ROUTES ==========
		r.Route("/api/users", func(r chi.Router) {
			// Middleware: AllowSelfOrAdmin (bisa manage diri sendiri atau admin)
			r.With(middleware.AllowSelfOrAdmin).Group(func(r chi.Router) {
				r.Get("/{id}", hdl.User.FindByID)
				r.Put("/{id}", hdl.User.Update)
			})
		})

		// ========== WAREHOUSE ROUTES ==========
		r.Route("/api/warehouses", func(r chi.Router) {
			r.Get("/", hdl.Warehouse.FindAll)
			r.Get("/{id}", hdl.Warehouse.FindByID)
		})

		// ========== CATEGORY ROUTES ==========
		r.Route("/api/categories", func(r chi.Router) {
			r.Get("/", hdl.Category.FindAll)
			r.Get("/{id}", hdl.Category.FindByID)
		})

		// ========== SHELF ROUTES ==========
		r.Route("/api/shelves", func(r chi.Router) {
			r.Get("/", hdl.Shelf.FindAll)
			r.Get("/{id}", hdl.Shelf.FindByID)

			r.Get("/warehouse/{warehouse_id}", hdl.Shelf.FindByWarehouseID)
		})

		// ========== PRODUCT ROUTES ==========
		r.Route("/api/products", func(r chi.Router) {
			r.Get("/", hdl.Product.FindAll)
			r.Get("/{id}", hdl.Product.FindByID)
			r.Get("/low-stock", hdl.Product.FindLowStock)

			r.Get("/category/{category_id}", hdl.Product.FindByCategoryID)
			r.Get("/shelf/{shelf_id}", hdl.Product.FindByShelfID)

			r.Put("/{id}/stock", hdl.Product.UpdateStock)
		})

		// ========== SALE ROUTES ==========
		r.Route("/api/sales", func(r chi.Router) {
			// Staff can create sales and list their own sales
			r.Get("/", hdl.Sale.FindAll) // GET /api/sales?page=1&limit=10
			r.Post("/", hdl.Sale.Create) // POST /api/sales

			// Sales endpoints with ownership checking
			// Staff can only access their own sales, admins can access all
			r.With(middleware.AllowSelfOrAdmin).Group(func(r chi.Router) {
				r.Get("/{id}", hdl.Sale.FindByID)            // GET /api/sales/{id}
				r.Put("/{id}/status", hdl.Sale.UpdateStatus) // PUT /api/sales/{id}/status
			})
		})
	})

	// ==================== ADMIN ROUTES ====================
	// Hanya Admin & Super Admin
	router.Group(func(r chi.Router) {
		r.Use(middleware.Auth(svc.Auth))
		r.Use(middleware.RequireRole(model.RoleAdmin, model.RoleSuperAdmin))

		// ========== USER ROUTES ==========
		r.Route("/api/admin/users", func(r chi.Router) {
			r.Get("/", hdl.User.FindAll)
			r.Post("/", hdl.User.Create)
			r.Delete("/{id}", hdl.User.Delete)
		})

		// ========== WAREHOUSE ROUTES ==========
		r.Route("/api/admin/warehouses", func(r chi.Router) {
			r.Post("/", hdl.Warehouse.Create)
			r.Put("/{id}", hdl.Warehouse.Update)
			r.Delete("/{id}", hdl.Warehouse.Delete)
		})

		// ========== CATEGORY ROUTES ==========
		r.Route("/api/admin/categories", func(r chi.Router) {
			r.Post("/", hdl.Category.Create)
			r.Put("/{id}", hdl.Category.Update)
			r.Delete("/{id}", hdl.Category.Delete)
		})

		// ========== SHELF ROUTES ==========
		r.Route("/api/admin/shelves", func(r chi.Router) {
			r.Post("/", hdl.Shelf.Create)
			r.Put("/{id}", hdl.Shelf.Update)
			r.Delete("/{id}", hdl.Shelf.Delete)
		})

		// ========== PRODUCT ROUTES ==========
		r.Route("/api/admin/products", func(r chi.Router) {
			r.Post("/", hdl.Product.Create)
			r.Put("/{id}", hdl.Product.Update)
			r.Delete("/{id}", hdl.Product.Delete)
		})

		// ========== SALE ADMIN ROUTES ==========
		r.Route("/api/admin/sales", func(r chi.Router) {
			// Admin can see all sales without ownership filter
			r.Get("/", hdl.Sale.FindAll) // GET /api/admin/sales

			// Admin-only report endpoint
			r.Get("/report", hdl.Sale.GetSalesReport) // GET /api/admin/sales/report?start_date=...&end_date=...
		})

		// ========== REPORT ROUTES (nanti) ==========
		// r.Route("/api/admin/reports", func(r chi.Router) {
		//     r.Get("/products", hdl.Report.Products)
		//     r.Get("/sales", hdl.Report.Sales)
		//     r.Get("/revenue", hdl.Report.Revenue) // hanya admin
		// })
	})

	// ==================== ERROR HANDLERS ====================
	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 - Not Found", http.StatusNotFound)
	})

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "405 - Method Not Allowed", http.StatusMethodNotAllowed)
	})

	return router
}
