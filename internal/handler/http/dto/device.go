package dto

import (
	"time"
)

// CreateDeviceRequest represents the request to create a device
type CreateDeviceRequest struct {
	Name  string `json:"name" binding:"required,min=3,max=100"`
	Brand string `json:"brand" binding:"required,min=2,max=50"`
}

// UpdateDeviceRequest represents the request to fully update a device
type UpdateDeviceRequest struct {
	Name  string `json:"name" binding:"required,min=3,max=100"`
	Brand string `json:"brand" binding:"required,min=2,max=50"`
	State string `json:"state" binding:"required,oneof=active in-use inactive"`
}

// PartialUpdateDeviceRequest represents the request to partially update a device
type PartialUpdateDeviceRequest struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,min=3,max=100"`
	Brand *string `json:"brand,omitempty" binding:"omitempty,min=2,max=50"`
	State *string `json:"state,omitempty" binding:"omitempty,oneof=active in-use inactive"`
}

// DeviceResponse represents a device in the API response
type DeviceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Brand     string    `json:"brand"`
	State     string    `json:"state"`
	CreatedAt time.Time `json:"created_at"`
}

// ListDevicesResponse represents a list of devices response
type ListDevicesResponse struct {
	Devices []DeviceResponse `json:"devices"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Field   string `json:"field,omitempty"`
}
