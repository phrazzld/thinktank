// Package auditlog provides structured logging for audit purposes
package auditlog

import "time"

// AuditEntry defines the structure for a single audit log record.
type AuditEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Operation   string                 `json:"operation"`             // e.g., "ExecuteStart", "GatherContext", "GenerateContent", "SaveOutput", "ExecuteEnd"
	Status      string                 `json:"status"`                // e.g., "Success", "Failure", "InProgress"
	DurationMs  *int64                 `json:"duration_ms,omitempty"` // Optional duration in milliseconds
	Inputs      map[string]interface{} `json:"inputs,omitempty"`      // CLI flags, file paths, etc.
	Outputs     map[string]interface{} `json:"outputs,omitempty"`     // Result details, file paths written
	TokenCounts *TokenCountInfo        `json:"token_counts,omitempty"`
	Error       *ErrorInfo             `json:"error,omitempty"`
	Message     string                 `json:"message,omitempty"` // Optional human-readable message
}

// TokenCountInfo holds token count details.
type TokenCountInfo struct {
	PromptTokens int32 `json:"prompt_tokens"`
	OutputTokens int32 `json:"output_tokens,omitempty"` // Only applicable for generation
	TotalTokens  int32 `json:"total_tokens"`
	Limit        int32 `json:"limit,omitempty"`
}

// ErrorInfo holds structured error details.
type ErrorInfo struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"` // e.g., "ValidationError", "APIError", "FileIOError"
	// Potentially add StackTrace if needed/configured
}
