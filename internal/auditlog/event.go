// Package auditlog provides structured logging capabilities for the architect tool.
package auditlog

import (
	"reflect"
	"time"
)

// ErrorDetails provides structured error information.
// It includes the error message and optional type and details fields.
type ErrorDetails struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`    // e.g., "APIError", "FileError"
	Details string `json:"details,omitempty"` // e.g., stack trace or additional context
}

// NewErrorDetails creates a new ErrorDetails with the given message and optional type and details.
// Only the message is required; type and details may be empty strings.
func NewErrorDetails(message string, typeStr ...string) ErrorDetails {
	details := ErrorDetails{
		Message: message,
	}

	if len(typeStr) > 0 && typeStr[0] != "" {
		details.Type = typeStr[0]
	}

	if len(typeStr) > 1 && typeStr[1] != "" {
		details.Details = typeStr[1]
	}

	return details
}

// AuditEvent represents a structured log entry.
// It contains information about operations, inputs, outputs, and errors
// in a format suitable for machine parsing.
type AuditEvent struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`     // e.g., "INFO", "ERROR"
	Operation string                 `json:"operation"` // e.g., "GeneratePlan"
	Message   string                 `json:"message"`   // Human-readable summary
	Inputs    map[string]interface{} `json:"inputs,omitempty"`
	Outputs   map[string]interface{} `json:"outputs,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     *ErrorDetails          `json:"error,omitempty"`
}

// NewAuditEvent creates a new AuditEvent with the given level, operation, and message.
// The timestamp is automatically set to the current UTC time.
// The returned event can be further customized using the With* methods.
func NewAuditEvent(level, operation, message string) AuditEvent {
	return AuditEvent{
		Timestamp: time.Now().UTC(),
		Level:     level,
		Operation: operation,
		Message:   message,
	}
}

// WithInput adds an input key-value pair to the AuditEvent.
// If the Inputs map is nil, it will be initialized.
// Returns the modified AuditEvent to allow method chaining.
func (e AuditEvent) WithInput(key string, value interface{}) AuditEvent {
	if e.Inputs == nil {
		e.Inputs = make(map[string]interface{})
	}
	e.Inputs[key] = value
	return e
}

// WithOutput adds an output key-value pair to the AuditEvent.
// If the Outputs map is nil, it will be initialized.
// Returns the modified AuditEvent to allow method chaining.
func (e AuditEvent) WithOutput(key string, value interface{}) AuditEvent {
	if e.Outputs == nil {
		e.Outputs = make(map[string]interface{})
	}
	e.Outputs[key] = value
	return e
}

// WithMetadata adds a metadata key-value pair to the AuditEvent.
// If the Metadata map is nil, it will be initialized.
// Returns the modified AuditEvent to allow method chaining.
func (e AuditEvent) WithMetadata(key string, value interface{}) AuditEvent {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// WithError sets the Error field of the AuditEvent.
// Returns the modified AuditEvent to allow method chaining.
func (e AuditEvent) WithError(err ErrorDetails) AuditEvent {
	e.Error = &err
	return e
}

// WithErrorFromGoError creates an ErrorDetails from a standard Go error and sets it on the AuditEvent.
// The error type is derived from the error's type name.
// Returns the modified AuditEvent to allow method chaining.
func (e AuditEvent) WithErrorFromGoError(err error) AuditEvent {
	if err == nil {
		return e
	}

	// Extract the type name using reflection
	t := reflect.TypeOf(err)
	typeName := t.String()

	// If it's a pointer, get the element type
	if t.Kind() == reflect.Ptr {
		typeName = t.Elem().Name()
	}

	return e.WithError(NewErrorDetails(err.Error(), typeName))
}
