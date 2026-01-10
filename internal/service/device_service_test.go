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

// ========== GetDevice Tests ==========

// TestGetDevice_Success tests successful device retrieval
func TestGetDevice_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	expectedDevice, _ := domain.NewDevice("iPhone 15", "Apple")
	expectedDevice.ID = deviceID

	mockRepo.On("GetByID", ctx, deviceID).Return(expectedDevice, nil)

	// Act
	device, err := svc.GetDevice(ctx, deviceID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, deviceID, device.ID)
	assert.Equal(t, "iPhone 15", device.Name)
	assert.Equal(t, "Apple", device.Brand)
	mockRepo.AssertExpectations(t)
}

// TestGetDevice_NotFound tests device not found error
func TestGetDevice_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	mockRepo.On("GetByID", ctx, deviceID).Return(nil, domain.ErrDeviceNotFound)

	// Act
	device, err := svc.GetDevice(ctx, deviceID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.True(t, domain.IsNotFoundError(err))
	mockRepo.AssertExpectations(t)
}

// TestGetDevice_RepositoryError tests repository failure
func TestGetDevice_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	repoErr := errors.New("database connection failed")
	mockRepo.On("GetByID", ctx, deviceID).Return(nil, repoErr)

	// Act
	device, err := svc.GetDevice(ctx, deviceID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, device)
	mockRepo.AssertExpectations(t)
}

// ========== ListDevices Tests ==========

// TestListDevices_Success tests successful device listing
func TestListDevices_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	device1, _ := domain.NewDevice("iPhone 15", "Apple")
	device2, _ := domain.NewDevice("Galaxy S24", "Samsung")
	expectedDevices := []*domain.Device{device1, device2}

	mockRepo.On("List", ctx, 10, 0).Return(expectedDevices, nil)

	// Act
	devices, err := svc.ListDevices(ctx, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 2)
	assert.Equal(t, "iPhone 15", devices[0].Name)
	assert.Equal(t, "Galaxy S24", devices[1].Name)
	mockRepo.AssertExpectations(t)
}

// TestListDevices_EmptyList tests empty device list
func TestListDevices_EmptyList(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	mockRepo.On("List", ctx, 10, 0).Return([]*domain.Device{}, nil)

	// Act
	devices, err := svc.ListDevices(ctx, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 0)
	mockRepo.AssertExpectations(t)
}

// TestListDevices_WithPagination tests pagination parameters
func TestListDevices_WithPagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	device1, _ := domain.NewDevice("iPhone 15", "Apple")
	expectedDevices := []*domain.Device{device1}

	mockRepo.On("List", ctx, 5, 10).Return(expectedDevices, nil)

	// Act
	devices, err := svc.ListDevices(ctx, 5, 10)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 1)
	mockRepo.AssertExpectations(t)
}

// TestListDevices_DefaultPagination tests default pagination values
func TestListDevices_DefaultPagination(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	mockRepo.On("List", ctx, 10, 0).Return([]*domain.Device{}, nil)

	// Act - pass invalid values that should be normalized
	devices, err := svc.ListDevices(ctx, 0, -1)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	mockRepo.AssertExpectations(t)
}

// TestListDevices_RepositoryError tests repository failure
func TestListDevices_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	repoErr := errors.New("database connection failed")
	mockRepo.On("List", ctx, 10, 0).Return(nil, repoErr)

	// Act
	devices, err := svc.ListDevices(ctx, 10, 0)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, devices)
	assert.Contains(t, err.Error(), "failed to list devices")
	mockRepo.AssertExpectations(t)
}

// ========== ListDevicesByBrand Tests ==========

// TestListDevicesByBrand_Success tests successful filtering by brand
func TestListDevicesByBrand_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	device1, _ := domain.NewDevice("iPhone 15", "Apple")
	device2, _ := domain.NewDevice("MacBook Pro", "Apple")
	expectedDevices := []*domain.Device{device1, device2}

	mockRepo.On("ListByBrand", ctx, "Apple", 10, 0).Return(expectedDevices, nil)

	// Act
	devices, err := svc.ListDevicesByBrand(ctx, "Apple", 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 2)
	mockRepo.AssertExpectations(t)
}

// TestListDevicesByBrand_EmptyBrand tests empty brand validation
func TestListDevicesByBrand_EmptyBrand(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	// Act
	devices, err := svc.ListDevicesByBrand(ctx, "", 10, 0)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, devices)
	assert.True(t, domain.IsValidationError(err))
}

// ========== ListDevicesByState Tests ==========

// TestListDevicesByState_Success tests successful filtering by state
func TestListDevicesByState_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	device1, _ := domain.NewDevice("iPhone 15", "Apple")
	expectedDevices := []*domain.Device{device1}

	mockRepo.On("ListByState", ctx, domain.DeviceStateActive, 10, 0).Return(expectedDevices, nil)

	// Act
	devices, err := svc.ListDevicesByState(ctx, domain.DeviceStateActive, 10, 0)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, devices)
	assert.Len(t, devices, 1)
	mockRepo.AssertExpectations(t)
}

// TestListDevicesByState_InvalidState tests invalid state validation
func TestListDevicesByState_InvalidState(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	// Act
	devices, err := svc.ListDevicesByState(ctx, domain.DeviceState("invalid-state"), 10, 0)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, devices)
	assert.True(t, domain.IsValidationError(err))
}

// ========== UpdateDevice Tests ==========

// TestUpdateDevice_Success tests successful full device update
func TestUpdateDevice_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	existingDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	existingDevice.ID = deviceID

	mockRepo.On("GetByID", ctx, deviceID).Return(existingDevice, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Device")).Return(nil)

	// Act
	device, err := svc.UpdateDevice(ctx, deviceID, "iPhone 15 Pro", "Apple", domain.DeviceStateActive)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "iPhone 15 Pro", device.Name)
	assert.Equal(t, "Apple", device.Brand)
	assert.Equal(t, domain.DeviceStateActive, device.State)
	mockRepo.AssertExpectations(t)
}

// TestUpdateDevice_NotFound tests updating non-existent device
func TestUpdateDevice_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	mockRepo.On("GetByID", ctx, deviceID).Return(nil, domain.ErrDeviceNotFound)

	// Act
	device, err := svc.UpdateDevice(ctx, deviceID, "iPhone 15", "Apple", domain.DeviceStateActive)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.True(t, domain.IsNotFoundError(err))
	mockRepo.AssertExpectations(t)
}

// TestUpdateDevice_InUseDevice tests business rule: cannot update in-use device
func TestUpdateDevice_InUseDevice(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	inUseDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	inUseDevice.ID = deviceID
	inUseDevice.State = domain.DeviceStateInUse

	mockRepo.On("GetByID", ctx, deviceID).Return(inUseDevice, nil)

	// Act
	device, err := svc.UpdateDevice(ctx, deviceID, "iPhone 15", "Samsung", domain.DeviceStateInUse)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.True(t, domain.IsBusinessRuleError(err))
	mockRepo.AssertExpectations(t)
}

// TestUpdateDevice_ValidationError tests validation during update
func TestUpdateDevice_ValidationError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	existingDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	existingDevice.ID = deviceID

	mockRepo.On("GetByID", ctx, deviceID).Return(existingDevice, nil)

	// Act - try to update with invalid name (too short)
	device, err := svc.UpdateDevice(ctx, deviceID, "ab", "Apple", domain.DeviceStateActive)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.True(t, domain.IsValidationError(err))
	mockRepo.AssertExpectations(t)
}

// TestUpdateDevice_RepositoryError tests repository failure
func TestUpdateDevice_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	existingDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	existingDevice.ID = deviceID

	mockRepo.On("GetByID", ctx, deviceID).Return(existingDevice, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Device")).Return(errors.New("database error"))

	// Act
	device, err := svc.UpdateDevice(ctx, deviceID, "iPhone 15", "Apple", domain.DeviceStateActive)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, device)
	assert.Contains(t, err.Error(), "failed to update device")
	mockRepo.AssertExpectations(t)
}

// ========== PartialUpdateDevice Tests ==========

// TestPartialUpdateDevice_Success tests partial device update
func TestPartialUpdateDevice_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	existingDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	existingDevice.ID = deviceID

	newName := "iPhone 15"
	mockRepo.On("GetByID", ctx, deviceID).Return(existingDevice, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Device")).Return(nil)

	// Act - only update name
	device, err := svc.PartialUpdateDevice(ctx, deviceID, &newName, nil, nil)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, "iPhone 15", device.Name)
	assert.Equal(t, "Apple", device.Brand) // Brand unchanged
	mockRepo.AssertExpectations(t)
}

// TestPartialUpdateDevice_UpdateState tests updating only state
func TestPartialUpdateDevice_UpdateState(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	existingDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	existingDevice.ID = deviceID

	newState := domain.DeviceStateInactive
	mockRepo.On("GetByID", ctx, deviceID).Return(existingDevice, nil)
	mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Device")).Return(nil)

	// Act - only update state
	device, err := svc.PartialUpdateDevice(ctx, deviceID, nil, nil, &newState)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, device)
	assert.Equal(t, domain.DeviceStateInactive, device.State)
	assert.Equal(t, "iPhone 14", device.Name) // Name unchanged
	mockRepo.AssertExpectations(t)
}

// ========== DeleteDevice Tests ==========

// TestDeleteDevice_Success tests successful device deletion
func TestDeleteDevice_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	existingDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	existingDevice.ID = deviceID
	existingDevice.State = domain.DeviceStateActive

	mockRepo.On("GetByID", ctx, deviceID).Return(existingDevice, nil)
	mockRepo.On("Delete", ctx, deviceID).Return(nil)

	// Act
	err := svc.DeleteDevice(ctx, deviceID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

// TestDeleteDevice_NotFound tests deleting non-existent device
func TestDeleteDevice_NotFound(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	mockRepo.On("GetByID", ctx, deviceID).Return(nil, domain.ErrDeviceNotFound)

	// Act
	err := svc.DeleteDevice(ctx, deviceID)

	// Assert
	assert.Error(t, err)
	assert.True(t, domain.IsNotFoundError(err))
	mockRepo.AssertExpectations(t)
}

// TestDeleteDevice_InUseDevice tests business rule: cannot delete in-use device
func TestDeleteDevice_InUseDevice(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	inUseDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	inUseDevice.ID = deviceID
	inUseDevice.State = domain.DeviceStateInUse

	mockRepo.On("GetByID", ctx, deviceID).Return(inUseDevice, nil)

	// Act
	err := svc.DeleteDevice(ctx, deviceID)

	// Assert
	assert.Error(t, err)
	assert.True(t, domain.IsBusinessRuleError(err))
	assert.Contains(t, err.Error(), "in-use")
	mockRepo.AssertExpectations(t)
}

// TestDeleteDevice_RepositoryError tests repository failure
func TestDeleteDevice_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDeviceRepository)
	svc := service.NewDeviceService(mockRepo)
	ctx := context.Background()

	deviceID := uuid.New()
	existingDevice, _ := domain.NewDevice("iPhone 14", "Apple")
	existingDevice.ID = deviceID

	mockRepo.On("GetByID", ctx, deviceID).Return(existingDevice, nil)
	mockRepo.On("Delete", ctx, deviceID).Return(errors.New("database error"))

	// Act
	err := svc.DeleteDevice(ctx, deviceID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete device")
	mockRepo.AssertExpectations(t)
}
