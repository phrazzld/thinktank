// Package gemini provides the implementation of the Google Gemini AI provider
package gemini

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/phrazzld/thinktank/internal/llm"
)

// APIErrorResponse represents the error structure returned by the Gemini API
type APIErrorResponse struct {
	Error struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Status  string      `json:"status"`
		Details interface{} `json:"details"`
	} `json:"error"`
}

// IsGeminiError checks if an error is an llm.LLMError originating from Gemini
func IsGeminiError(err error) (*llm.LLMError, bool) {
	var llmErr *llm.LLMError
	if errors.As(err, &llmErr) && llmErr.Provider == "gemini" {
		return llmErr, true
	}
	return nil, false
}

// ParseErrorResponse parses the Gemini API error response body
// This function returns the detailed error message extracted from the API response
func ParseErrorResponse(responseBody []byte) (string, string, int) {
	if len(responseBody) == 0 {
		return "", "", 0
	}

	var apiErrorResp APIErrorResponse
	if err := json.Unmarshal(responseBody, &apiErrorResp); err != nil {
		return "", "", 0
	}

	return apiErrorResp.Error.Message, apiErrorResp.Error.Status, apiErrorResp.Error.Code
}

// FormatErrorDetails formats detailed error information from Gemini API response
func FormatErrorDetails(errorMessage string, errorStatus string, errorCode int) string {
	if errorMessage == "" {
		return ""
	}

	details := fmt.Sprintf("API Error: %s", errorMessage)
	if errorStatus != "" {
		details += fmt.Sprintf(" (Status: %s)", errorStatus)
	}
	if errorCode != 0 {
		details += fmt.Sprintf(" (Code: %d)", errorCode)
	}

	return details
}

// MapGeminiErrorToCategory maps Gemini-specific error status/codes to our error categories
func MapGeminiErrorToCategory(errorStatus string, errorCode int) llm.ErrorCategory {
	// Map based on status
	switch errorStatus {
	case "UNAUTHENTICATED", "PERMISSION_DENIED":
		return llm.CategoryAuth
	case "RESOURCE_EXHAUSTED":
		return llm.CategoryRateLimit
	case "INVALID_ARGUMENT":
		return llm.CategoryInvalidRequest
	case "NOT_FOUND":
		return llm.CategoryNotFound
	case "UNAVAILABLE", "INTERNAL":
		return llm.CategoryServer
	case "DEADLINE_EXCEEDED":
		return llm.CategoryCancelled
	case "OUT_OF_RANGE":
		return llm.CategoryInputLimit
	}

	// Map common error codes
	switch errorCode {
	case 401, 403:
		return llm.CategoryAuth
	case 404:
		return llm.CategoryNotFound
	case 429:
		return llm.CategoryRateLimit
	case 400:
		return llm.CategoryInvalidRequest
	case 500, 502, 503:
		return llm.CategoryServer
	}

	return llm.CategoryUnknown
}

// SafetyFilter maps to specific content filter methods in Gemini
func IsSafetyFilter(errorMessage string) bool {
	lowerMsg := strings.ToLower(errorMessage)
	return strings.Contains(lowerMsg, "safety") ||
		strings.Contains(lowerMsg, "blocked") ||
		strings.Contains(lowerMsg, "content filter") ||
		strings.Contains(lowerMsg, "content policy")
}

// FormatAPIErrorFromResponse creates a standardized LLMError from a Gemini API error
// and detailed response information
func FormatAPIErrorFromResponse(err error, statusCode int, responseBody []byte) *llm.LLMError {
	if err == nil {
		return nil
	}

	// Check if it's already an LLMError
	var llmErr *llm.LLMError
	if errors.As(err, &llmErr) {
		return llmErr
	}

	// Parse error details from response body if available
	errorMessage, errorStatus, errorCode := ParseErrorResponse(responseBody)
	errorDetails := FormatErrorDetails(errorMessage, errorStatus, errorCode)

	// Try to categorize error from specific Gemini error types/codes
	category := MapGeminiErrorToCategory(errorStatus, errorCode)

	// Special case for safety filters
	if IsSafetyFilter(errorMessage) {
		category = llm.CategoryContentFiltered
	}

	// If we couldn't categorize based on Gemini specific info,
	// use the shared categorization logic
	if category == llm.CategoryUnknown {
		category = llm.DetectErrorCategory(err, statusCode)
	}

	// Create a formatted error message with Gemini specific suggestions
	llmError := llm.CreateStandardErrorWithMessage("gemini", category, err, errorDetails)

	// Add Gemini-specific suggestions for certain error types
	switch category {
	case llm.CategoryAuth:
		llmError.Suggestion = "Check that your Google API key is valid and has not expired. Ensure GOOGLE_API_KEY environment variable is set correctly."
	case llm.CategoryInsufficientCredits:
		llmError.Suggestion = "Check your Google Cloud billing account and ensure it's active. Visit https://console.cloud.google.com/billing for account details."
	case llm.CategoryNotFound:
		llmError.Suggestion = "Verify that the model ID is correct and that Gemini API is available in your region."
	case llm.CategoryServer:
		llmError.Suggestion = "This is typically a temporary issue with Google's servers. Wait a few moments and try again."
	case llm.CategoryInputLimit:
		llmError.Suggestion = "Your input is too long for the model's context window. Reduce the input size to fit within Gemini's token limits."
	case llm.CategoryContentFiltered:
		llmError.Suggestion = "Your prompt or request was flagged by Gemini's content filters. Modify your prompt to comply with Google's content policy."
	}

	// If we have the original error code from Gemini, add it to our error
	if errorCode != 0 {
		llmError.Code = fmt.Sprintf("%d", errorCode)
	}

	// If we have a status code, add it
	if statusCode != 0 {
		llmError.StatusCode = statusCode
	}

	return llmError
}

// FormatAPIError implements the standard provider error handling interface.
// It creates a standardized LLMError from any error, wrapping it with
// provider-specific context.
func FormatAPIError(rawError error, providerName string) error {
	if rawError == nil {
		return nil
	}

	// Check if it's already an LLMError
	var llmErr *llm.LLMError
	if errors.As(rawError, &llmErr) {
		// If it's already a properly formatted LLMError from this provider, just return it
		if llmErr.Provider == providerName {
			return llmErr
		}
		// Otherwise, wrap it with this provider's name
		return llm.Wrap(rawError, providerName, llmErr.Message, llmErr.ErrorCategory)
	}

	// Determine error category from message
	category := llm.GetErrorCategoryFromMessage(rawError.Error())

	// Create error message
	message := fmt.Sprintf("Error from %s provider: %v", providerName, rawError)

	// Return wrapped error
	return llm.Wrap(rawError, providerName, message, category)
}

// CreateAPIError creates a new LLMError with Gemini-specific settings
func CreateAPIError(category llm.ErrorCategory, errMsg string, originalErr error, details string) *llm.LLMError {
	llmError := llm.New(
		"gemini",    // Provider
		"",          // Code
		0,           // StatusCode
		errMsg,      // Message
		"",          // RequestID
		originalErr, // Original error
		category,    // Error category
	)

	// Add details if provided
	if details != "" {
		llmError.Details = details
	}

	// Add Gemini-specific suggestions
	switch category {
	case llm.CategoryAuth:
		llmError.Suggestion = "Check that your Google API key is valid and has not expired. Ensure GOOGLE_API_KEY environment variable is set correctly."
	case llm.CategoryRateLimit:
		llmError.Suggestion = "Wait and try again later. Google has rate limits for the Gemini API based on your account tier."
	case llm.CategoryInsufficientCredits:
		llmError.Suggestion = "Check your Google Cloud billing account and ensure it's active. Visit https://console.cloud.google.com/billing for account details."
	case llm.CategoryInvalidRequest:
		llmError.Suggestion = "Check the prompt format and parameters. Ensure they comply with the Gemini API requirements."
	case llm.CategoryNotFound:
		llmError.Suggestion = "Verify that the model ID is correct and that Gemini API is available in your region."
	case llm.CategoryServer:
		llmError.Suggestion = "This is typically a temporary issue with Google's servers. Wait a few moments and try again."
	case llm.CategoryInputLimit:
		llmError.Suggestion = "Your input is too long for the model's context window. Reduce the input size to fit within Gemini's token limits."
	case llm.CategoryContentFiltered:
		llmError.Suggestion = "Your prompt or request was flagged by Gemini's content filters. Modify your prompt to comply with Google's content policy."
	default:
		llmError.Suggestion = "Check the logs for more details or try again."
	}

	return llmError
}
