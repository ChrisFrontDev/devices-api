package http

import (
	"errors"
	"fmt"
	"net/http"

	"devices-api/internal/domain"
	"devices-api/internal/handler/http/dto"
	"devices-api/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DeviceHandler handles HTTP requests for devices
type DeviceHandler struct {
	service *service.DeviceService
}

// NewDeviceHandler creates a new device handler
func NewDeviceHandler(service *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		service: service,
	}
}

// CreateDevice godoc
// @Summary Create a new device
// @Description Create a new device with name and brand
// @Tags devices
// @Accept json
// @Produce json
// @Param device body dto.CreateDeviceRequest true "Device data"
// @Success 201 {object} dto.DeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices [post]
func (h *DeviceHandler) CreateDevice(c *gin.Context) {
	var req dto.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	device, err := h.service.CreateDevice(c.Request.Context(), req.Name, req.Brand)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, MapDeviceToResponse(device))
}

// GetDevice godoc
// @Summary Get a device by ID
// @Description Get a single device by its ID
// @Tags devices
// @Produce json
// @Param id path string true "Device ID (UUID)"
// @Success 200 {object} dto.DeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices/{id} [get]
func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	device, err := h.service.GetDevice(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, MapDeviceToResponse(device))
}

// ListDevices godoc
// @Summary List all devices
// @Description Get all devices with optional pagination, brand, or state filters
// @Tags devices
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Param brand query string false "Filter by brand"
// @Param state query string false "Filter by state (active, in-use, inactive)"
// @Success 200 {object} dto.ListDevicesResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices [get]
func (h *DeviceHandler) ListDevices(c *gin.Context) {
	// Parse query parameters
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := parsePositiveInt(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if parsed, err := parsePositiveInt(o); err == nil {
			offset = parsed
		}
	}

	brand := c.Query("brand")
	stateStr := c.Query("state")

	var devices []*domain.Device
	var err error

	// Handle filtering
	if brand != "" {
		devices, err = h.service.ListDevicesByBrand(c.Request.Context(), brand, limit, offset)
	} else if stateStr != "" {
		state := domain.DeviceState(stateStr)
		devices, err = h.service.ListDevicesByState(c.Request.Context(), state, limit, offset)
	} else {
		devices, err = h.service.ListDevices(c.Request.Context(), limit, offset)
	}

	if err != nil {
		h.handleError(c, err)
		return
	}

	response := dto.ListDevicesResponse{
		Devices: MapDevicesToResponse(devices),
		Total:   len(devices),
		Limit:   limit,
		Offset:  offset,
	}

	c.JSON(http.StatusOK, response)
}

// UpdateDevice godoc
// @Summary Fully update a device
// @Description Fully update an existing device (all fields required)
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID (UUID)"
// @Param device body dto.UpdateDeviceRequest true "Device data"
// @Success 200 {object} dto.DeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices/{id} [put]
func (h *DeviceHandler) UpdateDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	state := domain.DeviceState(req.State)
	device, err := h.service.UpdateDevice(c.Request.Context(), id, req.Name, req.Brand, state)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, MapDeviceToResponse(device))
}

// PartialUpdateDevice godoc
// @Summary Partially update a device
// @Description Partially update an existing device (only provided fields are updated)
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID (UUID)"
// @Param device body dto.PartialUpdateDeviceRequest true "Device data"
// @Success 200 {object} dto.DeviceResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices/{id} [patch]
func (h *DeviceHandler) PartialUpdateDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	var req dto.PartialUpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	var state *domain.DeviceState
	if req.State != nil {
		s := domain.DeviceState(*req.State)
		state = &s
	}

	device, err := h.service.PartialUpdateDevice(c.Request.Context(), id, req.Name, req.Brand, state)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, MapDeviceToResponse(device))
}

// DeleteDevice godoc
// @Summary Delete a device
// @Description Delete an existing device
// @Tags devices
// @Param id path string true "Device ID (UUID)"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 422 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /devices/{id} [delete]
func (h *DeviceHandler) DeleteDevice(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_id",
			Message: "Invalid UUID format",
		})
		return
	}

	if err := h.service.DeleteDevice(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// handleError maps domain errors to appropriate HTTP responses
func (h *DeviceHandler) handleError(c *gin.Context, err error) {
	if domain.IsNotFoundError(err) {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
		return
	}

	if domain.IsValidationError(err) {
		var validationErr *domain.ValidationError
		if errors.As(err, &validationErr) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Error:   "validation_error",
				Message: validationErr.Message,
				Field:   validationErr.Field,
			})
			return
		}
	}

	if domain.IsBusinessRuleError(err) {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{
			Error:   "business_rule_violation",
			Message: err.Error(),
		})
		return
	}

	// Internal server error
	c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error:   "internal_error",
		Message: "An unexpected error occurred",
	})
}

// parsePositiveInt is a helper to parse positive integers
func parsePositiveInt(s string) (int, error) {
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, errors.New("value must be non-negative")
	}
	return i, nil
}
