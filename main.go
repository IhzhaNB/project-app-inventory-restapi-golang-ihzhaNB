package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"inventory-system/database"
	"inventory-system/handler"
	"inventory-system/repository"
	"inventory-system/router"
	"inventory-system/service"
	"inventory-system/utils"

	"go.uber.org/zap"
)

func main() {
	// Load configuration
	config, err := utils.ReadConfiguration()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize logger
	logger, err := utils.InitLogger(config.PathLogging, config.Debug)
	if err != nil {
		log.Fatal("Failed to init logger:", err)
	}
	defer logger.Sync()

	// Set global logger
	utils.Logger = logger

	// Connect to database
	pool, err := database.InitDB(config.DB)
	if err != nil {
		logger.Fatal("Failed to connect database", zap.Error(err))
	}
	defer pool.Close()

	logger.Info("Database connected",
		zap.String("host", config.DB.Host),
		zap.String("database", config.DB.Name),
	)

	// Initialize repository, service, & handler
	repo := repository.NewRepository(pool, logger)
	svc := service.NewService(repo, logger)
	hdl := handler.NewHandlers(svc, logger)

	// Setup router
	r := router.SetupRouter(svc, hdl)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + config.Port,
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting",
			zap.String("port", config.Port),
			zap.String("app", config.AppName),
			zap.Bool("debug", config.Debug),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}
