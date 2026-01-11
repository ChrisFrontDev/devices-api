package testhelper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PostgresContainer wraps testcontainers postgres instance
type PostgresContainer struct {
	container *postgres.PostgresContainer
	pool      *pgxpool.Pool
	connStr   string
}

// NewPostgresContainer creates and starts a PostgreSQL testcontainer
func NewPostgresContainer(ctx context.Context) (*PostgresContainer, error) {
	container, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("devices_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		_ = container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresContainer{
		container: container,
		pool:      pool,
		connStr:   connStr,
	}, nil
}

// GetPool returns the database connection pool
func (pc *PostgresContainer) GetPool() *pgxpool.Pool {
	return pc.pool
}

// GetConnectionString returns the database connection string
func (pc *PostgresContainer) GetConnectionString() string {
	return pc.connStr
}

// ApplyMigrations applies SQL migration files to the test database
func (pc *PostgresContainer) ApplyMigrations(ctx context.Context, migrationsPath string) error {
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	// Sort files to ensure migrations run in order
	sort.Strings(files)

	for _, file := range files {
		// #nosec G304 - file path is from controlled migrations directory
		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		if _, err := pc.pool.Exec(ctx, string(sql)); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}
	}

	return nil
}

// Cleanup cleans up the database by truncating all tables
func (pc *PostgresContainer) Cleanup(ctx context.Context) error {
	_, err := pc.pool.Exec(ctx, "TRUNCATE TABLE devices CASCADE")
	return err
}

// Close closes the database connection pool and terminates the container
func (pc *PostgresContainer) Close(ctx context.Context) error {
	if pc.pool != nil {
		pc.pool.Close()
	}

	if pc.container != nil {
		return pc.container.Terminate(ctx)
	}

	return nil
}
