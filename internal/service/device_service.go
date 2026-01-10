package service

import (
	"context"
	"fmt"

	"devices-api/internal/domain"

	"github.com/google/uuid"
)

// DeviceService handles business logic for device operations
type DeviceService struct {
	repo domain.DeviceRepository
}

// NewDeviceService creates a new device service
func NewDeviceService(repo domain.DeviceRepository) *DeviceService {
	return &DeviceService{
		repo: repo,
	}
}

// CreateDevice creates a new device
func (s *DeviceService) CreateDevice(ctx context.Context, name, brand string) (*domain.Device, error) {
	// Create device with domain validation
	device, err := domain.NewDevice(name, brand)
	if err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	// Persist device
	if err := s.repo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to save device: %w", err)
	}

	return device, nil
}

// GetDevice retrieves a device by ID
func (s *DeviceService) GetDevice(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	device, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return device, nil
}

// ListDevices retrieves all devices with pagination
func (s *DeviceService) ListDevices(ctx context.Context, limit, offset int) ([]*domain.Device, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	devices, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return devices, nil
}

// ListDevicesByBrand retrieves devices filtered by brand
func (s *DeviceService) ListDevicesByBrand(ctx context.Context, brand string, limit, offset int) ([]*domain.Device, error) {
	if brand == "" {
		return nil, domain.NewValidationError("brand", "cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	devices, err := s.repo.ListByBrand(ctx, brand, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices by brand: %w", err)
	}

	return devices, nil
}

// ListDevicesByState retrieves devices filtered by state
func (s *DeviceService) ListDevicesByState(ctx context.Context, state domain.DeviceState, limit, offset int) ([]*domain.Device, error) {
	// Validate state
	tempDevice := &domain.Device{
		ID:    uuid.New(),
		Name:  "temp",
		Brand: "temp",
		State: state,
	}
	if err := tempDevice.ValidateState(); err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	devices, err := s.repo.ListByState(ctx, state, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices by state: %w", err)
	}

	return devices, nil
}
