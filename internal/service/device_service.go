package service

import (
	"context"
	"fmt"

	"devices-api/internal/domain"
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
