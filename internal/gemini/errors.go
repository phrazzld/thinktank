// Package gemini provides a client for interacting with Google's Gemini API
package gemini

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrorType represents different categories of errors that can occur when using the Gemini API
type ErrorType int

const (
	// ErrorTypeUnknown represents an unknown or uncategorized error
	ErrorTypeUnknown ErrorType = iota

	// ErrorTypeAuth represents authentication and authorization errors
	ErrorTypeAuth

	// ErrorTypeRateLimit represents rate limiting or quota errors
	ErrorTypeRateLimit

	// ErrorTypeInvalidRequest represents invalid request errors
	ErrorTypeInvalidRequest

	// ErrorTypeNotFound represents model not found errors
	ErrorTypeNotFound

	// ErrorTypeServer represents server errors
	ErrorTypeServer

	// ErrorTypeNetwork represents network connectivity errors
	ErrorTypeNetwork

	// ErrorTypeCancelled represents cancelled context errors
	ErrorTypeCancelled

	// ErrorTypeInputLimit represents input token limit exceeded errors
	ErrorTypeInputLimit

	// ErrorTypeContentFiltered represents content filtered by safety settings errors
	ErrorTypeContentFiltered
)

// APIError represents an enhanced error with detailed information
type APIError struct {
	// Original is the original error
	Original error

	// Type is the categorized error type
	Type ErrorType

	// Message is a user-friendly error message
	Message string

	// StatusCode is the HTTP status code (if applicable)
	StatusCode int

	// Suggestion is a helpful suggestion for resolving the error
	Suggestion string

	// Details contains additional error details
	Details string
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Original)
	}
	return e.Message
}

// Unwrap returns the original error
func (e *APIError) Unwrap() error {
	return e.Original
}

// UserFacingError returns a user-friendly error message with suggestions
func (e *APIError) UserFacingError() string {
	var sb strings.Builder

	// Start with the error message
	sb.WriteString(e.Message)

	// Add suggestions if available
	if e.Suggestion != "" {
		sb.WriteString("\n\nSuggestion: ")
		sb.WriteString(e.Suggestion)
	}

	return sb.String()
}

// DebugInfo returns detailed debugging information about the error
func (e *APIError) DebugInfo() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Error Type: %v\n", e.Type))
	sb.WriteString(fmt.Sprintf("Message: %s\n", e.Message))

	if e.StatusCode != 0 {
		sb.WriteString(fmt.Sprintf("Status Code: %d\n", e.StatusCode))
	}

	if e.Original != nil {
		sb.WriteString(fmt.Sprintf("Original Error: %v\n", e.Original))
	}

	if e.Details != "" {
		sb.WriteString(fmt.Sprintf("Details: %s\n", e.Details))
	}

	if e.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("Suggestion: %s\n", e.Suggestion))
	}

	return sb.String()
}

// IsAPIError checks if an error is an APIError and returns it
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// GetErrorType determines the error type based on the error message and status code
func GetErrorType(err error, statusCode int) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errMsg := err.Error()

	// Check for authorization errors
	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		return ErrorTypeAuth
	}

	// Check for rate limit or quota errors
	if statusCode == http.StatusTooManyRequests ||
		strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "quota") {
		return ErrorTypeRateLimit
	}

	// Check for invalid request errors
	if statusCode == http.StatusBadRequest {
		return ErrorTypeInvalidRequest
	}

	// Check for not found errors
	if statusCode == http.StatusNotFound {
		return ErrorTypeNotFound
	}

	// Check for server errors
	if statusCode >= 500 && statusCode < 600 {
		return ErrorTypeServer
	}

	// Check for content filtering
	if strings.Contains(errMsg, "safety") ||
		strings.Contains(errMsg, "blocked") ||
		strings.Contains(errMsg, "filtered") {
		return ErrorTypeContentFiltered
	}

	// Check for token limit errors
	if strings.Contains(errMsg, "token limit") ||
		strings.Contains(errMsg, "tokens exceeds") {
		return ErrorTypeInputLimit
	}

	// Check for network errors
	if strings.Contains(errMsg, "network") ||
		strings.Contains(errMsg, "connection") ||
		strings.Contains(errMsg, "timeout") {
		return ErrorTypeNetwork
	}

	// Check for cancellation
	if strings.Contains(errMsg, "canceled") ||
		strings.Contains(errMsg, "cancelled") ||
		strings.Contains(errMsg, "deadline exceeded") {
		return ErrorTypeCancelled
	}

	return ErrorTypeUnknown
}

// FormatAPIError creates a detailed API error from a standard error
func FormatAPIError(err error, statusCode int) *APIError {
	if err == nil {
		return nil
	}

	// Check if it's already an APIError
	if apiErr, ok := IsAPIError(err); ok {
		return apiErr
	}

	// Determine error type
	errType := GetErrorType(err, statusCode)

	// Create base API error
	apiErr := &APIError{
		Original:   err,
		Type:       errType,
		Message:    err.Error(),
		StatusCode: statusCode,
	}

	// Enhance error details based on type
	switch errType {
	case ErrorTypeAuth:
		apiErr.Message = "Authentication failed with the Gemini API"
		apiErr.Suggestion = "Check that your API key is valid and has not expired. Ensure environment variables are set correctly."

	case ErrorTypeRateLimit:
		apiErr.Message = "Request rate limit or quota exceeded on the Gemini API"
		apiErr.Suggestion = "Wait and try again later. Consider upgrading your API usage tier if this happens frequently."

	case ErrorTypeInvalidRequest:
		apiErr.Message = "Invalid request sent to the Gemini API"
		apiErr.Suggestion = "Check the prompt format and parameters. Ensure they comply with the API requirements."

	case ErrorTypeNotFound:
		apiErr.Message = "The requested model or resource was not found"
		apiErr.Suggestion = "Verify that the model name is correct and that the model is available in your region."

	case ErrorTypeServer:
		apiErr.Message = "Gemini API server error occurred"
		apiErr.Suggestion = "This is typically a temporary issue. Wait a few moments and try again."

	case ErrorTypeNetwork:
		apiErr.Message = "Network error while connecting to the Gemini API"
		apiErr.Suggestion = "Check your internet connection and try again. If persistent, there may be connectivity issues to Google's servers."

	case ErrorTypeCancelled:
		apiErr.Message = "Request to Gemini API was cancelled"
		apiErr.Suggestion = "The operation was interrupted. Try again with a longer timeout if needed."

	case ErrorTypeInputLimit:
		apiErr.Message = "Input token limit exceeded for the Gemini model"
		apiErr.Suggestion = "Reduce the input size by using --include, --exclude, or --exclude-names flags to filter the context."

	case ErrorTypeContentFiltered:
		apiErr.Message = "Content was filtered by Gemini API safety settings"
		apiErr.Suggestion = "Your prompt or content may have triggered safety filters. Review and modify your input to comply with content policies."

	default:
		apiErr.Message = fmt.Sprintf("Error calling Gemini API: %v", err)
		apiErr.Suggestion = "Check the logs for more details or try again."
	}

	return apiErr
}
