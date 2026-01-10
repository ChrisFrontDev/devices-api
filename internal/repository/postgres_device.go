package repository

import (
	"context"
	"errors"
	"fmt"

	"devices-api/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDeviceRepository implements the domain.DeviceRepository interface
type PostgresDeviceRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresDeviceRepository creates a new PostgreSQL device repository
func NewPostgresDeviceRepository(pool *pgxpool.Pool) *PostgresDeviceRepository {
	return &PostgresDeviceRepository{
		pool: pool,
	}
}

// Create persists a new device
func (r *PostgresDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	query := `
		INSERT INTO devices (id, name, brand, state, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.pool.Exec(ctx, query,
		device.ID,
		device.Name,
		device.Brand,
		device.State,
		device.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

// GetByID retrieves a device by its unique identifier
func (r *PostgresDeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	query := `
		SELECT id, name, brand, state, created_at
		FROM devices
		WHERE id = $1
	`

	var device domain.Device
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&device.ID,
		&device.Name,
		&device.Brand,
		&device.State,
		&device.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDeviceNotFound
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	return &device, nil
}

// List retrieves all devices with optional pagination
func (r *PostgresDeviceRepository) List(ctx context.Context, limit, offset int) ([]*domain.Device, error) {
	query := `
		SELECT id, name, brand, state, created_at
		FROM devices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}
	defer rows.Close()

	return r.scanDevices(rows)
}

// ListByBrand retrieves devices filtered by brand
func (r *PostgresDeviceRepository) ListByBrand(ctx context.Context, brand string, limit, offset int) ([]*domain.Device, error) {
	query := `
		SELECT id, name, brand, state, created_at
		FROM devices
		WHERE brand = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, brand, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices by brand: %w", err)
	}
	defer rows.Close()

	return r.scanDevices(rows)
}

// ListByState retrieves devices filtered by state
func (r *PostgresDeviceRepository) ListByState(ctx context.Context, state domain.DeviceState, limit, offset int) ([]*domain.Device, error) {
	query := `
		SELECT id, name, brand, state, created_at
		FROM devices
		WHERE state = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, state, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices by state: %w", err)
	}
	defer rows.Close()

	return r.scanDevices(rows)
}

// Update modifies an existing device
func (r *PostgresDeviceRepository) Update(ctx context.Context, device *domain.Device) error {
	query := `
		UPDATE devices
		SET name = $2, brand = $3, state = $4
		WHERE id = $1
	`

	result, err := r.pool.Exec(ctx, query,
		device.ID,
		device.Name,
		device.Brand,
		device.State,
	)

	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrDeviceNotFound
	}

	return nil
}

// Delete removes a device by its unique identifier
func (r *PostgresDeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM devices WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrDeviceNotFound
	}

	return nil
}

// ExistsByID checks if a device exists
func (r *PostgresDeviceRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM devices WHERE id = $1)`

	var exists bool
	err := r.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check device existence: %w", err)
	}

	return exists, nil
}

// scanDevices is a helper function to scan multiple device rows
func (r *PostgresDeviceRepository) scanDevices(rows pgx.Rows) ([]*domain.Device, error) {
	var devices []*domain.Device

	for rows.Next() {
		var device domain.Device
		err := rows.Scan(
			&device.ID,
			&device.Name,
			&device.Brand,
			&device.State,
			&device.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, &device)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating devices: %w", err)
	}

	return devices, nil
}
