package http

import (
	"devices-api/internal/domain"
	"devices-api/internal/handler/http/dto"
)

// MapDeviceToResponse converts a domain device to a response DTO
func MapDeviceToResponse(device *domain.Device) dto.DeviceResponse {
	return dto.DeviceResponse{
		ID:        device.ID,
		Name:      device.Name,
		Brand:     device.Brand,
		State:     string(device.State),
		CreatedAt: device.CreatedAt,
	}
}

// MapDevicesToResponse converts a list of domain devices to response DTOs
func MapDevicesToResponse(devices []*domain.Device) []dto.DeviceResponse {
	responses := make([]dto.DeviceResponse, len(devices))
	for i, device := range devices {
		responses[i] = MapDeviceToResponse(device)
	}
	return responses
}
