// Package auditlog provides structured logging capabilities for the architect tool.
package auditlog

import "time"

// ErrorDetails provides structured error information.
// It includes the error message and optional type and details fields.
type ErrorDetails struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`    // e.g., "APIError", "FileError"
	Details string `json:"details,omitempty"` // e.g., stack trace or additional context
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
