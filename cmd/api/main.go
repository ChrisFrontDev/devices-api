package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"devices-api/internal/config"

	"github.com/jackc/pgx/v5"
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

	// 3. Check Database Health
	logger.Info("Checking database health...")
	checkDatabaseHealth(ctx, cfg.Database.URL, logger)

	logger.Info("Server is running. Press Ctrl+C to stop.")

	// 4. Wait for termination signal
	<-ctx.Done()
	logger.Info("Shutdown signal received. Exiting...")
}

func checkDatabaseHealth(ctx context.Context, dbURL string, logger *slog.Logger) {
	// Retry loop to wait for DB to be ready
	var conn *pgx.Conn
	var err error
	maxRetries := 10

	for i := 0; i < maxRetries; i++ {
		// Use a short timeout context for connection attempts
		connCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		conn, err = pgx.Connect(connCtx, dbURL)
		cancel()

		if err == nil {
			break
		}
		logger.Warn("Failed to connect to database. Retrying in 2s...", "error", err, "attempt", i+1, "max_retries", maxRetries)

		select {
		case <-ctx.Done():
			logger.Info("Shutdown signal received during database connection retry")
			return // Exit if app is shutting down
		case <-time.After(2 * time.Second):
			continue
		}
	}

	if err != nil {
		logger.Error("Unable to connect to database after retries", "error", err, "max_retries", maxRetries)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// Execute a simple query to verify database is operational
	var greeting string
	err = conn.QueryRow(context.Background(), "SELECT 'Hello, world!'").Scan(&greeting)
	if err != nil {
		logger.Error("Database query failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Database health check passed", "message", greeting)
}
