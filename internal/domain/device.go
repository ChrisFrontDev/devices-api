package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// DeviceState represents the operational state of a device
type DeviceState string

const (
	DeviceStateActive  DeviceState = "active"
	DeviceStateInUse   DeviceState = "in-use"
	DeviceStateInactive DeviceState = "inactive"
)

// Device represents a hardware device in the system
type Device struct {
	ID        uuid.UUID
	Name      string
	Brand     string
	CreatedAt time.Time
	State     DeviceState
}

// NewDevice creates a new device with validation
func NewDevice(name, brand string) (*Device, error) {
	device := &Device{
		ID:        uuid.New(),
		Name:      name,
		Brand:     brand,
		CreatedAt: time.Now().UTC(),
		State:     DeviceStateActive,
	}

	if err := device.Validate(); err != nil {
		return nil, err
	}

	return device, nil
}

// Validate checks if the device has valid data
func (d *Device) Validate() error {
	if d.ID == uuid.Nil {
		return NewValidationError("id", "cannot be empty")
	}

	if err := d.ValidateName(); err != nil {
		return err
	}

	if err := d.ValidateBrand(); err != nil {
		return err
	}

	if d.CreatedAt.IsZero() {
		return NewValidationError("created_at", "cannot be empty")
	}

	if err := d.ValidateState(); err != nil {
		return err
	}

	return nil
}

// ValidateName validates the device name
func (d *Device) ValidateName() error {
	name := strings.TrimSpace(d.Name)
	if name == "" {
		return NewValidationError("name", "cannot be empty")
	}
	if len(name) < 3 {
		return NewValidationError("name", "must be at least 3 characters")
	}
	if len(name) > 100 {
		return NewValidationError("name", "must not exceed 100 characters")
	}
	return nil
}

// ValidateBrand validates the device brand
func (d *Device) ValidateBrand() error {
	brand := strings.TrimSpace(d.Brand)
	if brand == "" {
		return NewValidationError("brand", "cannot be empty")
	}
	if len(brand) < 2 {
		return NewValidationError("brand", "must be at least 2 characters")
	}
	if len(brand) > 50 {
		return NewValidationError("brand", "must not exceed 50 characters")
	}
	return nil
}

// ValidateState validates the device state
func (d *Device) ValidateState() error {
	switch d.State {
	case DeviceStateActive, DeviceStateInUse, DeviceStateInactive:
		return nil
	default:
		return NewValidationError("state", fmt.Sprintf("invalid state: %s (must be: active, in-use, or inactive)", d.State))
	}
}

// CanUpdate checks if the device can be updated based on business rules
// Devices in "in-use" state cannot have their name or brand updated
func (d *Device) CanUpdate() error {
	if d.State == DeviceStateInUse {
		return NewBusinessRuleError("cannot update device in 'in-use' state")
	}
	return nil
}

// CanDelete checks if the device can be deleted based on business rules
// Devices in "in-use" state cannot be deleted
func (d *Device) CanDelete() error {
	if d.State == DeviceStateInUse {
		return NewBusinessRuleError("cannot delete device in 'in-use' state")
	}
	return nil
}

// Update updates the device fields with validation
func (d *Device) Update(name, brand string, state DeviceState) error {
	if err := d.CanUpdate(); err != nil {
		return err
	}

	// Create temporary device to validate new values
	temp := &Device{
		ID:        d.ID,
		Name:      name,
		Brand:     brand,
		CreatedAt: d.CreatedAt,
		State:     state,
	}

	if err := temp.Validate(); err != nil {
		return err
	}

	// Apply updates
	d.Name = name
	d.Brand = brand
	d.State = state

	return nil
}

// SetState changes the device state
func (d *Device) SetState(state DeviceState) error {
	d.State = state
	return d.ValidateState()
}
