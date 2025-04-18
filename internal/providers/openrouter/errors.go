// Package openrouter provides the implementation of the OpenRouter LLM provider
package openrouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/phrazzld/architect/internal/llm"
)

// ErrorType represents different categories of errors that can occur when using the OpenRouter API
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

	// ErrorTypeInsufficientCredits represents insufficient credits or payment required errors
	ErrorTypeInsufficientCredits
)

// APIErrorResponse represents the error structure returned by the OpenRouter API
type APIErrorResponse struct {
	Error APIErrorDetail `json:"error"`
}

// APIErrorDetail contains the details of an API error returned by OpenRouter
type APIErrorDetail struct {
	Code    interface{} `json:"code"` // Can be string or int
	Message string      `json:"message"`
	Type    string      `json:"type,omitempty"`
	Param   string      `json:"param,omitempty"`
}

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

// Category implements the llm.CategorizedError interface, mapping OpenRouter error types
// to standardized error categories for provider-agnostic error handling.
func (e *APIError) Category() llm.ErrorCategory {
	switch e.Type {
	case ErrorTypeAuth:
		return llm.CategoryAuth
	case ErrorTypeRateLimit, ErrorTypeInsufficientCredits:
		return llm.CategoryRateLimit
	case ErrorTypeInvalidRequest:
		return llm.CategoryInvalidRequest
	case ErrorTypeNotFound:
		return llm.CategoryNotFound
	case ErrorTypeServer:
		return llm.CategoryServer
	case ErrorTypeNetwork:
		return llm.CategoryNetwork
	case ErrorTypeCancelled:
		return llm.CategoryCancelled
	case ErrorTypeInputLimit:
		return llm.CategoryInputLimit
	case ErrorTypeContentFiltered:
		return llm.CategoryContentFiltered
	default:
		return llm.CategoryUnknown
	}
}

// Ensure APIError implements the llm.CategorizedError interface (compile-time check)
var _ llm.CategorizedError = (*APIError)(nil)

// IsAPIError checks if an error is an APIError and returns it
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// GetErrorType determines the error type based on the error message, HTTP status code,
// and OpenRouter-specific codes and messages
func GetErrorType(err error, statusCode int, responseBody []byte) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errMsg := err.Error()

	// Try to parse the response body as an OpenRouter error response
	var apiErrorResp APIErrorResponse
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &apiErrorResp); err == nil {
			// If successfully parsed, use the API error details
			if apiErrorResp.Error.Message != "" {
				errMsg = apiErrorResp.Error.Message
			}

			// Check for specific OpenRouter error types if available
			if apiErrorResp.Error.Type != "" {
				errorType := apiErrorResp.Error.Type
				if strings.Contains(errorType, "auth") ||
					strings.Contains(errorType, "authentication") ||
					strings.Contains(errorType, "unauthorized") {
					return ErrorTypeAuth
				}

				if strings.Contains(errorType, "rate_limit") {
					return ErrorTypeRateLimit
				}

				if strings.Contains(errorType, "invalid_request") {
					return ErrorTypeInvalidRequest
				}

				if strings.Contains(errorType, "not_found") {
					return ErrorTypeNotFound
				}

				if strings.Contains(errorType, "server_error") {
					return ErrorTypeServer
				}
			}
		}
	}

	// Check status code-based error types
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrorTypeAuth
	case http.StatusPaymentRequired:
		return ErrorTypeInsufficientCredits
	case http.StatusTooManyRequests:
		return ErrorTypeRateLimit
	case http.StatusBadRequest:
		return ErrorTypeInvalidRequest
	case http.StatusNotFound:
		return ErrorTypeNotFound
	}

	// Check for server errors
	if statusCode >= 500 && statusCode < 600 {
		return ErrorTypeServer
	}

	// If we don't have a specific HTTP status code or it's not a standard error,
	// check the error message content

	// Check for authorization errors
	if strings.Contains(errMsg, "auth") ||
		strings.Contains(errMsg, "unauthorized") ||
		strings.Contains(errMsg, "invalid key") ||
		strings.Contains(errMsg, "API key") {
		return ErrorTypeAuth
	}

	// Check for insufficient credits or payment required
	if strings.Contains(errMsg, "credit") ||
		strings.Contains(errMsg, "payment") ||
		strings.Contains(errMsg, "billing") {
		return ErrorTypeInsufficientCredits
	}

	// Check for rate limit or quota errors
	if strings.Contains(errMsg, "rate limit") ||
		strings.Contains(errMsg, "quota") ||
		strings.Contains(errMsg, "too many requests") {
		return ErrorTypeRateLimit
	}

	// Check for content filtering
	if strings.Contains(errMsg, "safety") ||
		strings.Contains(errMsg, "content_filter") ||
		strings.Contains(errMsg, "blocked") ||
		strings.Contains(errMsg, "filtered") ||
		strings.Contains(errMsg, "moderation") {
		return ErrorTypeContentFiltered
	}

	// Check for token limit errors
	if strings.Contains(errMsg, "token limit") ||
		strings.Contains(errMsg, "tokens exceeds") ||
		strings.Contains(errMsg, "maximum context length") {
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
func FormatAPIError(err error, statusCode int, responseBody []byte) *APIError {
	if err == nil {
		return nil
	}

	// Check if it's already an APIError
	if apiErr, ok := IsAPIError(err); ok {
		return apiErr
	}

	// Parse error details from response body if available
	var errorDetails string
	var apiErrorResp APIErrorResponse
	if len(responseBody) > 0 {
		if jsonErr := json.Unmarshal(responseBody, &apiErrorResp); jsonErr == nil {
			if apiErrorResp.Error.Message != "" {
				errorDetails = fmt.Sprintf("API Error: %s", apiErrorResp.Error.Message)
				if apiErrorResp.Error.Type != "" {
					errorDetails += fmt.Sprintf(" (Type: %s)", apiErrorResp.Error.Type)
				}
				if apiErrorResp.Error.Param != "" {
					errorDetails += fmt.Sprintf(" (Param: %s)", apiErrorResp.Error.Param)
				}
			}
		}
	}

	// Determine error type
	errType := GetErrorType(err, statusCode, responseBody)

	// Create base API error
	apiErr := &APIError{
		Original:   err,
		Type:       errType,
		Message:    err.Error(),
		StatusCode: statusCode,
		Details:    errorDetails,
	}

	// Enhance error details based on type
	switch errType {
	case ErrorTypeAuth:
		apiErr.Message = "Authentication failed with the OpenRouter API"
		apiErr.Suggestion = "Check that your OpenRouter API key is valid and has not expired. Ensure OPENROUTER_API_KEY environment variable is set correctly."

	case ErrorTypeRateLimit:
		apiErr.Message = "Request rate limit exceeded on the OpenRouter API"
		apiErr.Suggestion = "Wait and try again later. Consider adjusting the --max-concurrent and --rate-limit flags to limit request rate."

	case ErrorTypeInsufficientCredits:
		apiErr.Message = "Insufficient credits or payment required on the OpenRouter API"
		apiErr.Suggestion = "Check your OpenRouter account balance and add credits if needed. Visit https://openrouter.ai/account for account details."

	case ErrorTypeInvalidRequest:
		apiErr.Message = "Invalid request sent to the OpenRouter API"
		apiErr.Suggestion = "Check the prompt format and parameters. Ensure they comply with the API requirements."

	case ErrorTypeNotFound:
		apiErr.Message = "The requested model or resource was not found"
		apiErr.Suggestion = "Verify that the model name is correct and uses the format 'provider/model' or 'provider/organization/model'."

	case ErrorTypeServer:
		apiErr.Message = "OpenRouter API server error occurred"
		apiErr.Suggestion = "This is typically a temporary issue with OpenRouter or the underlying model provider. Wait a few moments and try again."

	case ErrorTypeNetwork:
		apiErr.Message = "Network error while connecting to the OpenRouter API"
		apiErr.Suggestion = "Check your internet connection and try again. If persistent, there may be connectivity issues to OpenRouter's servers."

	case ErrorTypeCancelled:
		apiErr.Message = "Request to OpenRouter API was cancelled"
		apiErr.Suggestion = "The operation was interrupted. Try again with a longer timeout if needed."

	case ErrorTypeInputLimit:
		apiErr.Message = "Input token limit exceeded for the selected model"
		apiErr.Suggestion = "Reduce the input size by using --include, --exclude, or --exclude-names flags to filter the context."

	case ErrorTypeContentFiltered:
		apiErr.Message = "Content was filtered by safety settings"
		apiErr.Suggestion = "Your prompt or content may have triggered safety filters. Review and modify your input to comply with content policies."

	default:
		apiErr.Message = fmt.Sprintf("Error calling OpenRouter API: %v", err)
		apiErr.Suggestion = "Check the logs for more details or try again."
	}

	// If we have detailed error information from the API, include it in the error message
	if errorDetails != "" && !strings.Contains(apiErr.Message, errorDetails) {
		apiErr.Message = fmt.Sprintf("%s (%s)", apiErr.Message, errorDetails)
	}

	return apiErr
}
