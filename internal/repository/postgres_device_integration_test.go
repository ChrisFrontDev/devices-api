package repository_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"devices-api/internal/domain"
	"devices-api/internal/repository"
	"devices-api/internal/testhelper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	pgContainer *testhelper.PostgresContainer
)

// TestMain runs once before all tests
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Setup
	var err error
	pgContainer, err = testhelper.NewPostgresContainer(ctx)
	if err != nil {
		panic("failed to start postgres container: " + err.Error())
	}

	// Apply migrations
	migrationsPath := filepath.Join("..", "..", "migrations")
	err = pgContainer.ApplyMigrations(ctx, migrationsPath)
	if err != nil {
		pgContainer.Close(ctx)
		panic("failed to apply migrations: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Teardown
	if err := pgContainer.Close(ctx); err != nil {
		panic("failed to close postgres container: " + err.Error())
	}

	os.Exit(code)
}

// setupTest runs before each test
func setupTest(t *testing.T) *repository.PostgresDeviceRepository {
	ctx := context.Background()
	err := pgContainer.Cleanup(ctx)
	require.NoError(t, err, "failed to cleanup database")

	pool := pgContainer.GetPool()
	return repository.NewPostgresDeviceRepository(pool)
}

// ========== Create Tests ==========

func TestPostgresDeviceRepository_Create_Success(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device, err := domain.NewDevice("iPhone 15", "Apple")
	require.NoError(t, err)

	err = repo.Create(ctx, device)
	assert.NoError(t, err)

	// Verify device was created
	retrieved, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, device.ID, retrieved.ID)
	assert.Equal(t, device.Name, retrieved.Name)
	assert.Equal(t, device.Brand, retrieved.Brand)
	assert.Equal(t, device.State, retrieved.State)
}

func TestPostgresDeviceRepository_Create_MultipleDevices(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	devices := []struct {
		name  string
		brand string
	}{
		{"iPhone 15", "Apple"},
		{"Galaxy S24", "Samsung"},
		{"Pixel 9", "Google"},
	}

	for _, d := range devices {
		device, err := domain.NewDevice(d.name, d.brand)
		require.NoError(t, err)

		err = repo.Create(ctx, device)
		assert.NoError(t, err)
	}

	// Verify all devices were created
	list, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, list, 3)
}

// ========== GetByID Tests ==========

func TestPostgresDeviceRepository_GetByID_Success(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device, err := domain.NewDevice("iPhone 15", "Apple")
	require.NoError(t, err)

	err = repo.Create(ctx, device)
	require.NoError(t, err)

	retrieved, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, device.ID, retrieved.ID)
	assert.Equal(t, "iPhone 15", retrieved.Name)
	assert.Equal(t, "Apple", retrieved.Brand)
	assert.Equal(t, domain.DeviceStateActive, retrieved.State)
}

func TestPostgresDeviceRepository_GetByID_NotFound(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	nonExistentID := uuid.New()
	device, err := repo.GetByID(ctx, nonExistentID)

	assert.Error(t, err)
	assert.Nil(t, device)
	assert.True(t, domain.IsNotFoundError(err))
}

// ========== List Tests ==========

func TestPostgresDeviceRepository_List_Success(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	// Create test devices
	devices := []struct {
		name  string
		brand string
	}{
		{"iPhone 15", "Apple"},
		{"Galaxy S24", "Samsung"},
		{"Pixel 9", "Google"},
	}

	for _, d := range devices {
		device, err := domain.NewDevice(d.name, d.brand)
		require.NoError(t, err)
		err = repo.Create(ctx, device)
		require.NoError(t, err)

		// Add small delay to ensure different created_at timestamps
		time.Sleep(10 * time.Millisecond)
	}

	list, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, list, 3)

	// Verify ordering (most recent first)
	assert.Equal(t, "Pixel 9", list[0].Name)
	assert.Equal(t, "Galaxy S24", list[1].Name)
	assert.Equal(t, "iPhone 15", list[2].Name)
}

func TestPostgresDeviceRepository_List_WithPagination(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	// Create 5 devices
	for i := 1; i <= 5; i++ {
		device, err := domain.NewDevice("Device "+string(rune('A'+i-1)), "Brand")
		require.NoError(t, err)
		err = repo.Create(ctx, device)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Test first page
	page1, err := repo.List(ctx, 2, 0)
	assert.NoError(t, err)
	assert.Len(t, page1, 2)

	// Test second page
	page2, err := repo.List(ctx, 2, 2)
	assert.NoError(t, err)
	assert.Len(t, page2, 2)

	// Verify different results
	assert.NotEqual(t, page1[0].ID, page2[0].ID)
}

func TestPostgresDeviceRepository_List_EmptyDatabase(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	list, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, list, 0)
}

// ========== ListByBrand Tests ==========

func TestPostgresDeviceRepository_ListByBrand_Success(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	// Create devices from different brands
	devices := []struct {
		name  string
		brand string
	}{
		{"iPhone 15", "Apple"},
		{"MacBook Pro", "Apple"},
		{"Galaxy S24", "Samsung"},
		{"Pixel 9", "Google"},
	}

	for _, d := range devices {
		device, err := domain.NewDevice(d.name, d.brand)
		require.NoError(t, err)
		err = repo.Create(ctx, device)
		require.NoError(t, err)
	}

	// Filter by Apple brand
	appleDevices, err := repo.ListByBrand(ctx, "Apple", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, appleDevices, 2)

	for _, device := range appleDevices {
		assert.Equal(t, "Apple", device.Brand)
	}
}

func TestPostgresDeviceRepository_ListByBrand_NoneFound(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device, err := domain.NewDevice("iPhone 15", "Apple")
	require.NoError(t, err)
	err = repo.Create(ctx, device)
	require.NoError(t, err)

	// Search for non-existent brand
	devices, err := repo.ListByBrand(ctx, "Microsoft", 10, 0)
	assert.NoError(t, err)
	assert.Len(t, devices, 0)
}

// ========== ListByState Tests ==========

func TestPostgresDeviceRepository_ListByState_Success(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	// Create devices with different states
	device1, _ := domain.NewDevice("Device 1", "Brand1")
	device1.State = domain.DeviceStateActive
	require.NoError(t, repo.Create(ctx, device1))

	device2, _ := domain.NewDevice("Device 2", "Brand2")
	device2.State = domain.DeviceStateInUse
	require.NoError(t, repo.Create(ctx, device2))

	device3, _ := domain.NewDevice("Device 3", "Brand3")
	device3.State = domain.DeviceStateActive
	require.NoError(t, repo.Create(ctx, device3))

	// Filter by active state
	activeDevices, err := repo.ListByState(ctx, domain.DeviceStateActive, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, activeDevices, 2)

	// Filter by in-use state
	inUseDevices, err := repo.ListByState(ctx, domain.DeviceStateInUse, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, inUseDevices, 1)
}

// ========== Update Tests ==========

func TestPostgresDeviceRepository_Update_Success(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	// Create initial device
	device, err := domain.NewDevice("iPhone 14", "Apple")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, device))

	// Update device
	device.Name = "iPhone 15 Pro"
	device.State = domain.DeviceStateInUse
	err = repo.Update(ctx, device)
	assert.NoError(t, err)

	// Verify update
	updated, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.Equal(t, "iPhone 15 Pro", updated.Name)
	assert.Equal(t, domain.DeviceStateInUse, updated.State)
}

func TestPostgresDeviceRepository_Update_NotFound(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device, err := domain.NewDevice("iPhone 15", "Apple")
	require.NoError(t, err)

	// Try to update non-existent device
	err = repo.Update(ctx, device)
	assert.Error(t, err)
	assert.True(t, domain.IsNotFoundError(err))
}

func TestPostgresDeviceRepository_Update_AllFields(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	// Create initial device
	device, err := domain.NewDevice("iPhone 14", "Apple")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, device))

	// Update all mutable fields
	device.Name = "Galaxy S24"
	device.Brand = "Samsung"
	device.State = domain.DeviceStateInactive

	err = repo.Update(ctx, device)
	assert.NoError(t, err)

	// Verify all fields updated
	updated, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Galaxy S24", updated.Name)
	assert.Equal(t, "Samsung", updated.Brand)
	assert.Equal(t, domain.DeviceStateInactive, updated.State)
}

// ========== Delete Tests ==========

func TestPostgresDeviceRepository_Delete_Success(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device, err := domain.NewDevice("iPhone 15", "Apple")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, device))

	// Delete device
	err = repo.Delete(ctx, device.ID)
	assert.NoError(t, err)

	// Verify deletion
	deleted, err := repo.GetByID(ctx, device.ID)
	assert.Error(t, err)
	assert.Nil(t, deleted)
	assert.True(t, domain.IsNotFoundError(err))
}

func TestPostgresDeviceRepository_Delete_NotFound(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	nonExistentID := uuid.New()
	err := repo.Delete(ctx, nonExistentID)

	assert.Error(t, err)
	assert.True(t, domain.IsNotFoundError(err))
}

func TestPostgresDeviceRepository_Delete_DoesNotAffectOthers(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	// Create multiple devices
	device1, _ := domain.NewDevice("Device 1", "Brand1")
	device2, _ := domain.NewDevice("Device 2", "Brand2")
	require.NoError(t, repo.Create(ctx, device1))
	require.NoError(t, repo.Create(ctx, device2))

	// Delete one device
	err := repo.Delete(ctx, device1.ID)
	assert.NoError(t, err)

	// Verify other device still exists
	existing, err := repo.GetByID(ctx, device2.ID)
	assert.NoError(t, err)
	assert.NotNil(t, existing)
	assert.Equal(t, device2.ID, existing.ID)
}

// ========== ExistsByID Tests ==========

func TestPostgresDeviceRepository_ExistsByID_True(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device, err := domain.NewDevice("iPhone 15", "Apple")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, device))

	exists, err := repo.ExistsByID(ctx, device.ID)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestPostgresDeviceRepository_ExistsByID_False(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	nonExistentID := uuid.New()
	exists, err := repo.ExistsByID(ctx, nonExistentID)
	assert.NoError(t, err)
	assert.False(t, exists)
}

// ========== Concurrent Operations Tests ==========

func TestPostgresDeviceRepository_ConcurrentCreates(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	const numDevices = 10
	done := make(chan bool, numDevices)
	errors := make(chan error, numDevices)

	for i := 0; i < numDevices; i++ {
		go func(index int) {
			device, err := domain.NewDevice("Device "+string(rune('A'+index)), "Brand")
			if err != nil {
				errors <- err
				done <- false
				return
			}

			err = repo.Create(ctx, device)
			if err != nil {
				errors <- err
				done <- false
				return
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numDevices; i++ {
		<-done
	}
	close(errors)

	// Check for errors
	for err := range errors {
		assert.NoError(t, err)
	}

	// Verify all devices created
	list, err := repo.List(ctx, 100, 0)
	assert.NoError(t, err)
	assert.Len(t, list, numDevices)
}

// ========== Data Integrity Tests ==========

func TestPostgresDeviceRepository_TimestampPersistence(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device, err := domain.NewDevice("iPhone 15", "Apple")
	require.NoError(t, err)

	originalTime := device.CreatedAt
	require.NoError(t, repo.Create(ctx, device))

	// Retrieve and verify timestamp
	retrieved, err := repo.GetByID(ctx, device.ID)
	assert.NoError(t, err)

	// Timestamps should match within a second (accounting for precision)
	timeDiff := retrieved.CreatedAt.Sub(originalTime).Abs()
	assert.True(t, timeDiff < time.Second)
}

func TestPostgresDeviceRepository_UUIDUniqueness(t *testing.T) {
	repo := setupTest(t)
	ctx := context.Background()

	device1, err := domain.NewDevice("Device 1", "Brand1")
	require.NoError(t, err)
	require.NoError(t, repo.Create(ctx, device1))

	// Try to create device with same UUID (simulate duplicate)
	device2, err := domain.NewDevice("Device 2", "Brand2")
	require.NoError(t, err)
	device2.ID = device1.ID

	err = repo.Create(ctx, device2)
	assert.Error(t, err) // Should fail due to unique constraint
}
