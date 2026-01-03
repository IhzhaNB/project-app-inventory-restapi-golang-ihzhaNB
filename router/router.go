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
			// Middleware: AllowSelfOrAdmin (bisa lihat diri sendiri atau admin)
			r.With(middleware.AllowSelfOrAdmin).Get("/{id}", hdl.User.FindByID)
			r.With(middleware.AllowSelfOrAdmin).Put("/{id}", hdl.User.Update)
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

		// ========== PRODUCT ROUTES (nanti) ==========
		// r.Route("/api/products", func(r chi.Router) {
		//     r.Get("/", hdl.Product.FindAll)       // semua bisa lihat
		//     r.Get("/{id}", hdl.Product.FindByID)  // semua bisa lihat detail
		// })

		// ========== SALE ROUTES (nanti) ==========
		// r.Route("/api/sales", func(r chi.Router) {
		//     r.Get("/", hdl.Sale.FindAll)         // staff hanya lihat punya sendiri
		//     r.Post("/", hdl.Sale.Create)         // semua bisa create sale
		//     r.Get("/{id}", hdl.Sale.FindByID)    // sesuai permission
		// })
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
