// Package openai provides the implementation of the OpenAI LLM provider
package openai

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/phrazzld/thinktank/internal/llm"
)

// APIErrorResponse represents the error structure returned by the OpenAI API
type APIErrorResponse struct {
	Error APIErrorDetail `json:"error"`
}

// APIErrorDetail contains the details of an API error returned by OpenAI
type APIErrorDetail struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
}

// IsOpenAIError checks if an error is an llm.LLMError originating from OpenAI
func IsOpenAIError(err error) (*llm.LLMError, bool) {
	var llmErr *llm.LLMError
	if errors.As(err, &llmErr) && llmErr.Provider == "openai" {
		return llmErr, true
	}
	return nil, false
}

// ParseErrorResponse parses the OpenAI API error response body
// This function returns the detailed error message extracted from the API response
func ParseErrorResponse(responseBody []byte) (string, string, string, string) {
	if len(responseBody) == 0 {
		return "", "", "", ""
	}

	var apiErrorResp APIErrorResponse
	if err := json.Unmarshal(responseBody, &apiErrorResp); err != nil {
		return "", "", "", ""
	}

	var errorMessage, errorType, errorCode, errorParam string

	if apiErrorResp.Error.Message != "" {
		errorMessage = apiErrorResp.Error.Message
	}

	if apiErrorResp.Error.Type != "" {
		errorType = apiErrorResp.Error.Type
	}

	if apiErrorResp.Error.Code != "" {
		errorCode = apiErrorResp.Error.Code
	}

	if apiErrorResp.Error.Param != "" {
		errorParam = apiErrorResp.Error.Param
	}

	return errorMessage, errorType, errorCode, errorParam
}

// FormatErrorDetails formats detailed error information from OpenAI API response
func FormatErrorDetails(errorMessage, errorType, errorCode, errorParam string) string {
	if errorMessage == "" {
		return ""
	}

	details := fmt.Sprintf("API Error: %s", errorMessage)
	if errorType != "" {
		details += fmt.Sprintf(" (Type: %s)", errorType)
	}
	if errorCode != "" {
		details += fmt.Sprintf(" (Code: %s)", errorCode)
	}
	if errorParam != "" {
		details += fmt.Sprintf(" (Param: %s)", errorParam)
	}

	return details
}

// MapOpenAIErrorToCategory maps OpenAI-specific error types/codes to our error categories
func MapOpenAIErrorToCategory(errorType, errorCode string) llm.ErrorCategory {
	// First check error type
	switch errorType {
	case "authentication_error":
		return llm.CategoryAuth
	case "invalid_request_error":
		return llm.CategoryInvalidRequest
	case "rate_limit_error":
		return llm.CategoryRateLimit
	case "server_error":
		return llm.CategoryServer
	}

	// Then check error code
	switch errorCode {
	case "context_length_exceeded":
		return llm.CategoryInputLimit
	case "model_not_found":
		return llm.CategoryNotFound
	case "insufficient_quota":
		return llm.CategoryInsufficientCredits
	case "content_filter":
		return llm.CategoryContentFiltered
	}

	// Default to unknown
	return llm.CategoryUnknown
}

// FormatAPIErrorFromResponse creates a standardized LLMError from an OpenAI API error
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
	errorMessage, errorType, errorCode, errorParam := ParseErrorResponse(responseBody)
	errorDetails := FormatErrorDetails(errorMessage, errorType, errorCode, errorParam)

	// Try to categorize error from specific OpenAI error types/codes
	category := MapOpenAIErrorToCategory(errorType, errorCode)

	// If we couldn't categorize based on OpenAI specific info,
	// use the shared categorization logic
	if category == llm.CategoryUnknown {
		category = llm.DetectErrorCategory(err, statusCode)
	}

	// Create a formatted error message with OpenAI specific suggestions
	llmError := llm.CreateStandardErrorWithMessage("openai", category, err, errorDetails)

	// Add OpenAI-specific suggestions for certain error types
	switch category {
	case llm.CategoryAuth:
		llmError.Suggestion = "Check that your OpenAI API key is valid and has not expired. Ensure OPENAI_API_KEY environment variable is set correctly."
	case llm.CategoryInsufficientCredits:
		llmError.Suggestion = "Check your OpenAI account balance and add credits if needed. Visit https://platform.openai.com/account/billing for account details."
	case llm.CategoryNotFound:
		llmError.Suggestion = "Verify that the model ID is correct and that you have access to the requested model."
	case llm.CategoryServer:
		llmError.Suggestion = "This is typically a temporary issue with OpenAI's servers. Wait a few moments and try again."
	case llm.CategoryInputLimit:
		llmError.Suggestion = "Your input is too long for the model's context window. Reduce the input size or use a model with a larger context window."
	case llm.CategoryContentFiltered:
		llmError.Suggestion = "Your prompt or request was flagged by OpenAI's content filters. Modify your prompt to comply with OpenAI's usage policies."
	}

	// If we have the original error code from OpenAI, add it to our error
	if errorCode != "" {
		llmError.Code = errorCode
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

// CreateAPIError creates a new LLMError with OpenAI-specific settings
func CreateAPIError(category llm.ErrorCategory, errMsg string, originalErr error, details string) *llm.LLMError {
	llmError := llm.New(
		"openai",    // Provider
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

	// Add OpenAI-specific suggestions
	switch category {
	case llm.CategoryAuth:
		llmError.Suggestion = "Check that your OpenAI API key is valid and has not expired. Ensure OPENAI_API_KEY environment variable is set correctly."
	case llm.CategoryRateLimit:
		llmError.Suggestion = "Wait and try again later. OpenAI has rate limits based on your account tier. Consider adding a delay between requests."
	case llm.CategoryInsufficientCredits:
		llmError.Suggestion = "Check your OpenAI account balance and add credits if needed. Visit https://platform.openai.com/account/billing for account details."
	case llm.CategoryInvalidRequest:
		llmError.Suggestion = "Check the prompt format and parameters. Ensure they comply with the OpenAI API requirements."
	case llm.CategoryNotFound:
		llmError.Suggestion = "Verify that the model ID is correct and that you have access to the requested model."
	case llm.CategoryServer:
		llmError.Suggestion = "This is typically a temporary issue with OpenAI's servers. Wait a few moments and try again."
	case llm.CategoryInputLimit:
		llmError.Suggestion = "Your input is too long for the model's context window. Reduce the input size or use a model with a larger context window."
	case llm.CategoryContentFiltered:
		llmError.Suggestion = "Your prompt or request was flagged by OpenAI's content filters. Modify your prompt to comply with OpenAI's usage policies."
	default:
		llmError.Suggestion = "Check the logs for more details or try again."
	}

	return llmError
}
