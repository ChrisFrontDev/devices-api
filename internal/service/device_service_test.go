package service_test

import (
	"context"
	"errors"
	"testing"

	"devices-api/internal/domain"
	"devices-api/internal/service"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeviceRepository is a mock implementation of domain.DeviceRepository
type MockDeviceRepository struct {
	mock.Mock
}

func (m *MockDeviceRepository) Create(ctx context.Context, device *domain.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) List(ctx context.Context, limit, offset int) ([]*domain.Device, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) ListByBrand(ctx context.Context, brand string, limit, offset int) ([]*domain.Device, error) {
	args := m.Called(ctx, brand, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) ListByState(ctx context.Context, state domain.DeviceState, limit, offset int) ([]*domain.Device, error) {
	args := m.Called(ctx, state, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Device), args.Error(1)
}

func (m *MockDeviceRepository) Update(ctx context.Context, device *domain.Device) error {
	args := m.Called(ctx, device)
	return args.Error(0)
}

func (m *MockDeviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDeviceRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

// TestCreateDevice_Success tests successful device creation
func TestCreateDevice_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Device")).Return(nil)

	// Act
	device, err := svc.CreateDevice(ctx, "iPhone 15", "Apple")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "iPhone 15", device.Name)
	assert.Equal(t, "Apple", device.Brand)
	assert.Equal(t, domain.DeviceStateActive, device.State)
	assert.NotEqual(t, uuid.Nil, device.ID)
	assert.False(t, device.CreatedAt.IsZero())
	mockRepo.AssertExpectations(t)
}

// TestCreateDevice_ValidationError tests validation failure
func TestCreateDevice_ValidationError(t *testing.T) {
	tests := []struct {
		name        string
		deviceName  string
		deviceBrand string
		expectError bool
	}{
		{
			name:        "empty name",
			deviceName:  "",
			deviceBrand: "Apple",
			expectError: true,
		},
		{
			name:        "name too short",
			deviceName:  "ab",
			deviceBrand: "Apple",
			expectError: true,
		},
		{
			name:        "name too long",
			deviceName:  string(make([]byte, 101)),
			deviceBrand: "Apple",
			expectError: true,
		},
		{
			name:        "empty brand",
			deviceName:  "iPhone 15",
			deviceBrand: "",
			expectError: true,
		},
		{
			name:        "brand too short",
			deviceName:  "iPhone 15",
			deviceBrand: "A",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockRepo := new(MockDeviceRepository)
			svc := service.NewDeviceService(mockRepo)
			ctx := context.Background()

			// Act
			device, err := svc.CreateDevice(ctx, tt.deviceName, tt.deviceBrand)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, device)
				assert.True(t, domain.IsValidationError(err))
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, device)
			}
		})
	}
}

// TestCreateDevice_RepositoryError tests repository failure
func TestCreateDevice_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	repoErr := errors.New("database connection failed")
	mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Device")).Return(repoErr)

	// Act
	device, err := svc.CreateDevice(ctx, "iPhone 15", "Apple")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.Contains(t, err.Error(), "failed to save device")
	mockRepo.AssertExpectations(t)
}
