package domain

import (
	"errors"
	"fmt"
)

// Common domain errors
var (
	ErrDeviceNotFound      = errors.New("device not found")
	ErrDeviceAlreadyExists = errors.New("device already exists")
	ErrInvalidInput        = errors.New("invalid input")
	ErrBusinessRule        = errors.New("business rule violation")
)

// ValidationError represents a validation error for a specific field
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) error {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

// BusinessRuleError represents a business rule violation
type BusinessRuleError struct {
	Message string
}

func (e *BusinessRuleError) Error() string {
	return fmt.Sprintf("business rule violation: %s", e.Message)
}

// NewBusinessRuleError creates a new business rule error
func NewBusinessRuleError(message string) error {
	return &BusinessRuleError{
		Message: message,
	}
}

// IsBusinessRuleError checks if an error is a BusinessRuleError
func IsBusinessRuleError(err error) bool {
	var businessRuleErr *BusinessRuleError
	return errors.As(err, &businessRuleErr)
}

// IsNotFoundError checks if an error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrDeviceNotFound)
}

// IsAlreadyExistsError checks if an error is an already exists error
func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrDeviceAlreadyExists)
}
