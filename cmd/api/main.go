package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devices-api/internal/config"
	httphandler "devices-api/internal/handler/http"
	"devices-api/internal/repository"
	"devices-api/internal/service"
	"devices-api/pkg/database"
)

func main() {
	// 0. Setup Logger (slog)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// 1. Setup graceful shutdown handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("Starting Devices API...")

	// 2. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("Failed to load config", "error", err)
		os.Exit(1)
	}
	logger.Info("Config loaded", "http_port", cfg.Server.HTTPPort, "grpc_port", cfg.Server.GRPCPort)

	// 3. Initialize Database Connection Pool
	logger.Info("Connecting to database...")
	dbPool, err := database.NewPostgresPool(ctx, cfg.Database.URL)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	logger.Info("Database connection established")

	// 4. Initialize Layers (Dependency Injection)
	deviceRepo := repository.NewPostgresDeviceRepository(dbPool)
	deviceService := service.NewDeviceService(deviceRepo)

	// 5. Setup HTTP Server
	router := httphandler.SetupRouter(deviceService)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTPPort),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 6. Start HTTP Server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", "port", cfg.Server.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	logger.Info("Server is running. Press Ctrl+C to stop.")

	// 7. Wait for termination signal
	<-ctx.Done()
	logger.Info("Shutdown signal received. Initiating graceful shutdown...")

	// 8. Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	logger.Info("Server stopped gracefully")
}