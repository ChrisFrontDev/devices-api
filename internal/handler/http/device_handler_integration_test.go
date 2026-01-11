package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	httphandler "devices-api/internal/handler/http"
	"devices-api/internal/handler/http/dto"
	"devices-api/internal/repository"
	"devices-api/internal/service"
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
	migrationsPath := filepath.Join("..", "..", "..", "migrations")
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

// setupTestRouter creates a test router with fresh database state
func setupTestRouter(t *testing.T) *httptest.Server {
	ctx := context.Background()
	err := pgContainer.Cleanup(ctx)
	require.NoError(t, err, "failed to cleanup database")

	pool := pgContainer.GetPool()
	repo := repository.NewPostgresDeviceRepository(pool)
	svc := service.NewDeviceService(repo)
	router := httphandler.SetupRouter(svc)

	return httptest.NewServer(router)
}

// ========== Health Check Tests ==========

func TestHealthCheck(t *testing.T) {
	server := setupTestRouter(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "ok", result["status"])
}

// ========== Create Device Tests ==========

func TestCreateDevice_Success(t *testing.T) {
	server := setupTestRouter(t)
	defer server.Close()

	payload := dto.CreateDeviceRequest{
		Name:  "iPhone 15",
		Brand: "Apple",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(
		server.URL+"/api/v1/devices",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result dto.DeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil.String(), result.ID)
	assert.Equal(t, "iPhone 15", result.Name)
	assert.Equal(t, "Apple", result.Brand)
	assert.Equal(t, "active", result.State)
	assert.NotEmpty(t, result.CreatedAt)
}

func TestCreateDevice_ValidationError_EmptyName(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	payload := dto.CreateDeviceRequest{
		Name:  "",
		Brand: "Apple",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(
		server.URL+"/api/v1/devices",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result dto.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "validation_error", result.Error)
}

func TestCreateDevice_ValidationError_ShortName(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	payload := dto.CreateDeviceRequest{
		Name:  "ab",
		Brand: "Apple",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(
		server.URL+"/api/v1/devices",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result dto.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "validation_error", result.Error)
}

func TestCreateDevice_InvalidJSON(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	resp, err := http.Post(
		server.URL+"/api/v1/devices",
		"application/json",
		bytes.NewBuffer([]byte("invalid json")),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ========== Get Device Tests ==========

func TestGetDevice_Success(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create a device first
	created := createTestDevice(t, server, "iPhone 15", "Apple")

	// Get the device
	resp, err := http.Get(server.URL + "/api/v1/devices/" + created.ID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.DeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "iPhone 15", result.Name)
	assert.Equal(t, "Apple", result.Brand)
}

func TestGetDevice_NotFound(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	nonExistentID := uuid.New().String()
	resp, err := http.Get(server.URL + "/api/v1/devices/" + nonExistentID)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	var result dto.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "not_found", result.Error)
}

func TestGetDevice_InvalidUUID(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/devices/invalid-uuid")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var result dto.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "invalid_id", result.Error)
}

// ========== List Devices Tests ==========

func TestListDevices_Empty(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/v1/devices")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.ListDevicesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Total)
	assert.Len(t, result.Devices, 0)
}

func TestListDevices_WithData(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create multiple devices
	createTestDevice(t, server, "iPhone 15", "Apple")
	createTestDevice(t, server, "Galaxy S24", "Samsung")
	createTestDevice(t, server, "Pixel 9", "Google")

	resp, err := http.Get(server.URL + "/api/v1/devices")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.ListDevicesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, 3, result.Total)
	assert.Len(t, result.Devices, 3)
}

func TestListDevices_WithPagination(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create 5 devices
	for i := 1; i <= 5; i++ {
		createTestDevice(t, server, "Device "+string(rune('A'+i-1)), "Brand")
	}

	// Get first page
	resp, err := http.Get(server.URL + "/api/v1/devices?limit=2&offset=0")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.ListDevicesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)
	assert.Equal(t, 2, result.Limit)
	assert.Equal(t, 0, result.Offset)
}

func TestListDevices_FilterByBrand(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create devices from different brands
	createTestDevice(t, server, "iPhone 15", "Apple")
	createTestDevice(t, server, "MacBook Pro", "Apple")
	createTestDevice(t, server, "Galaxy S24", "Samsung")

	resp, err := http.Get(server.URL + "/api/v1/devices?brand=Apple")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.ListDevicesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Total)

	for _, device := range result.Devices {
		assert.Equal(t, "Apple", device.Brand)
	}
}

func TestListDevices_FilterByState(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create devices and update their states
	device1 := createTestDevice(t, server, "Device 1", "Brand1")
	device2 := createTestDevice(t, server, "Device 2", "Brand2")

	// Update device2 to inactive
	updatePayload := dto.PartialUpdateDeviceRequest{
		State: stringPtr("inactive"),
	}
	updateTestDevice(t, server, device2.ID, updatePayload)

	// Filter by active state
	resp, err := http.Get(server.URL + "/api/v1/devices?state=active")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.ListDevicesResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Total)
	assert.Equal(t, device1.ID, result.Devices[0].ID)
}

// ========== Update Device Tests ==========

func TestUpdateDevice_Success(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create a device
	created := createTestDevice(t, server, "iPhone 14", "Apple")

	// Update the device
	payload := dto.UpdateDeviceRequest{
		Name:  "iPhone 15 Pro",
		Brand: "Apple",
		State: "in-use",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/devices/"+created.ID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.DeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, created.ID, result.ID)
	assert.Equal(t, "iPhone 15 Pro", result.Name)
	assert.Equal(t, "in-use", result.State)
}

func TestUpdateDevice_NotFound(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	payload := dto.UpdateDeviceRequest{
		Name:  "iPhone 15",
		Brand: "Apple",
		State: "active",
	}

	body, _ := json.Marshal(payload)
	nonExistentID := uuid.New().String()
	req, _ := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/devices/"+nonExistentID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestUpdateDevice_CannotUpdateInUseDevice(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create a device
	created := createTestDevice(t, server, "iPhone 14", "Apple")

	// Update to in-use state
	updateToInUse := dto.UpdateDeviceRequest{
		Name:  "iPhone 14",
		Brand: "Apple",
		State: "in-use",
	}
	body1, _ := json.Marshal(updateToInUse)
	req1, _ := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/devices/"+created.ID,
		bytes.NewBuffer(body1),
	)
	req1.Header.Set("Content-Type", "application/json")
	resp1, _ := http.DefaultClient.Do(req1)
	resp1.Body.Close()

	// Try to update again (should fail)
	updatePayload := dto.UpdateDeviceRequest{
		Name:  "iPhone 15",
		Brand: "Samsung",
		State: "in-use",
	}
	body2, _ := json.Marshal(updatePayload)
	req2, _ := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/devices/"+created.ID,
		bytes.NewBuffer(body2),
	)
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := http.DefaultClient.Do(req2)
	require.NoError(t, err)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp2.StatusCode)

	var result dto.ErrorResponse
	err = json.NewDecoder(resp2.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "business_rule_violation", result.Error)
}

// ========== Partial Update Device Tests ==========

func TestPartialUpdateDevice_UpdateName(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create a device
	created := createTestDevice(t, server, "iPhone 14", "Apple")

	// Partially update (only name)
	payload := dto.PartialUpdateDeviceRequest{
		Name: stringPtr("iPhone 15"),
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(
		http.MethodPatch,
		server.URL+"/api/v1/devices/"+created.ID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.DeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "iPhone 15", result.Name)
	assert.Equal(t, "Apple", result.Brand) // Brand unchanged
}

func TestPartialUpdateDevice_UpdateState(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create a device
	created := createTestDevice(t, server, "iPhone 15", "Apple")

	// Partially update (only state)
	payload := dto.PartialUpdateDeviceRequest{
		State: stringPtr("inactive"),
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(
		http.MethodPatch,
		server.URL+"/api/v1/devices/"+created.ID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result dto.DeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "inactive", result.State)
	assert.Equal(t, "iPhone 15", result.Name) // Name unchanged
}

// ========== Delete Device Tests ==========

func TestDeleteDevice_Success(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create a device
	created := createTestDevice(t, server, "iPhone 15", "Apple")

	// Delete the device
	req, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/devices/"+created.ID,
		nil,
	)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify device is deleted
	getResp, _ := http.Get(server.URL + "/api/v1/devices/" + created.ID)
	defer getResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
}

func TestDeleteDevice_NotFound(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	nonExistentID := uuid.New().String()
	req, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/devices/"+nonExistentID,
		nil,
	)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteDevice_CannotDeleteInUseDevice(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// Create a device
	created := createTestDevice(t, server, "iPhone 15", "Apple")

	// Update to in-use state
	updatePayload := dto.PartialUpdateDeviceRequest{
		State: stringPtr("in-use"),
	}
	updateTestDevice(t, server, created.ID, updatePayload)

	// Try to delete
	req, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/devices/"+created.ID,
		nil,
	)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	var result dto.ErrorResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "business_rule_violation", result.Error)
}

// ========== End-to-End Workflow Tests ==========

func TestEndToEndWorkflow(t *testing.T) {

	server := setupTestRouter(t)
	defer server.Close()

	// 1. Create a device
	created := createTestDevice(t, server, "iPhone 15", "Apple")
	assert.NotEmpty(t, created.ID)

	// 2. Get the device
	getResp, _ := http.Get(server.URL + "/api/v1/devices/" + created.ID)
	assert.Equal(t, http.StatusOK, getResp.StatusCode)
	getResp.Body.Close()

	// 3. Update the device
	updatePayload := dto.UpdateDeviceRequest{
		Name:  "iPhone 15 Pro",
		Brand: "Apple",
		State: "in-use",
	}
	body, _ := json.Marshal(updatePayload)
	updateReq, _ := http.NewRequest(
		http.MethodPut,
		server.URL+"/api/v1/devices/"+created.ID,
		bytes.NewBuffer(body),
	)
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, _ := http.DefaultClient.Do(updateReq)
	assert.Equal(t, http.StatusOK, updateResp.StatusCode)
	updateResp.Body.Close()

	// 4. List devices
	listResp, _ := http.Get(server.URL + "/api/v1/devices")
	var listResult dto.ListDevicesResponse
	json.NewDecoder(listResp.Body).Decode(&listResult)
	assert.Equal(t, 1, listResult.Total)
	listResp.Body.Close()

	// 5. Update to inactive state
	patchPayload := dto.PartialUpdateDeviceRequest{
		State: stringPtr("inactive"),
	}
	patchBody, _ := json.Marshal(patchPayload)
	patchReq, _ := http.NewRequest(
		http.MethodPatch,
		server.URL+"/api/v1/devices/"+created.ID,
		bytes.NewBuffer(patchBody),
	)
	patchReq.Header.Set("Content-Type", "application/json")
	patchResp, _ := http.DefaultClient.Do(patchReq)
	assert.Equal(t, http.StatusOK, patchResp.StatusCode)
	patchResp.Body.Close()

	// 6. Delete the device
	deleteReq, _ := http.NewRequest(
		http.MethodDelete,
		server.URL+"/api/v1/devices/"+created.ID,
		nil,
	)
	deleteResp, _ := http.DefaultClient.Do(deleteReq)
	assert.Equal(t, http.StatusNoContent, deleteResp.StatusCode)
	deleteResp.Body.Close()

	// 7. Verify device is deleted
	verifyResp, _ := http.Get(server.URL + "/api/v1/devices/" + created.ID)
	assert.Equal(t, http.StatusNotFound, verifyResp.StatusCode)
	verifyResp.Body.Close()
}

// ========== Helper Functions ==========

func createTestDevice(t *testing.T, server *httptest.Server, name, brand string) dto.DeviceResponse {
	payload := dto.CreateDeviceRequest{
		Name:  name,
		Brand: brand,
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(
		server.URL+"/api/v1/devices",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result dto.DeviceResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}

func updateTestDevice(t *testing.T, server *httptest.Server, deviceID string, payload dto.PartialUpdateDeviceRequest) {
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(
		http.MethodPatch,
		server.URL+"/api/v1/devices/"+deviceID,
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
}

func stringPtr(s string) *string {
	return &s
}
