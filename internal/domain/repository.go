package domain

import (
	"context"

	"github.com/google/uuid"
)

// DeviceRepository defines the interface for device persistence operations.
// This interface is defined in the domain layer (Dependency Inversion Principle).
// The actual implementation will be in the repository layer.
type DeviceRepository interface {
	// Create persists a new device
	Create(ctx context.Context, device *Device) error

	// GetByID retrieves a device by its unique identifier
	GetByID(ctx context.Context, id uuid.UUID) (*Device, error)

	// List retrieves all devices with optional pagination
	List(ctx context.Context, limit, offset int) ([]*Device, error)

	// ListByBrand retrieves devices filtered by brand
	ListByBrand(ctx context.Context, brand string, limit, offset int) ([]*Device, error)

	// ListByState retrieves devices filtered by state
	ListByState(ctx context.Context, state DeviceState, limit, offset int) ([]*Device, error)

	// Update modifies an existing device
	Update(ctx context.Context, device *Device) error

	// Delete removes a device by its unique identifier
	Delete(ctx context.Context, id uuid.UUID) error

	// ExistsByID checks if a device exists
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
