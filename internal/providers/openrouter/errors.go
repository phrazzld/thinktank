// Package openrouter provides the implementation of the OpenRouter LLM provider
package openrouter

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/phrazzld/thinktank/internal/llm"
)

// APIErrorResponse represents the error structure returned by the OpenRouter API
type APIErrorResponse struct {
	Error APIErrorDetail `json:"error"`
}

// APIErrorDetail contains the details of an API error returned by OpenRouter
type APIErrorDetail struct {
	Code     interface{}            `json:"code"` // Can be string or int
	Message  string                 `json:"message"`
	Type     string                 `json:"type,omitempty"`
	Param    string                 `json:"param,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IsOpenRouterError checks if an error is an llm.LLMError originating from OpenRouter
func IsOpenRouterError(err error) (*llm.LLMError, bool) {
	var llmErr *llm.LLMError
	if errors.As(err, &llmErr) && llmErr.Provider == "openrouter" {
		return llmErr, true
	}
	return nil, false
}

// ParseErrorResponse parses the OpenRouter API error response body
// This function returns the detailed error message extracted from the API response
func ParseErrorResponse(responseBody []byte) (string, string, string) {
	if len(responseBody) == 0 {
		return "", "", ""
	}

	var apiErrorResp APIErrorResponse
	if err := json.Unmarshal(responseBody, &apiErrorResp); err != nil {
		return "", "", ""
	}

	var errorMessage, errorType, errorParam string

	if apiErrorResp.Error.Message != "" {
		errorMessage = apiErrorResp.Error.Message
	}

	// Check if there's a raw error message in metadata (common for provider errors)
	if apiErrorResp.Error.Metadata != nil {
		if rawError, ok := apiErrorResp.Error.Metadata["raw"].(string); ok && rawError != "" {
			// Try to parse the raw error JSON to get the actual provider error
			var providerError struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
					Param   string `json:"param"`
				} `json:"error"`
			}
			if err := json.Unmarshal([]byte(rawError), &providerError); err == nil {
				if providerError.Error.Message != "" {
					// Use the provider's error message as it's more specific
					errorMessage = providerError.Error.Message
				}
				if providerError.Error.Type != "" && errorType == "" {
					errorType = providerError.Error.Type
				}
				if providerError.Error.Param != "" && errorParam == "" {
					errorParam = providerError.Error.Param
				}
			}
		}
	}

	if apiErrorResp.Error.Type != "" && errorType == "" {
		errorType = apiErrorResp.Error.Type
	}

	if apiErrorResp.Error.Param != "" && errorParam == "" {
		errorParam = apiErrorResp.Error.Param
	}

	return errorMessage, errorType, errorParam
}

// FormatErrorDetails formats detailed error information from OpenRouter API response
func FormatErrorDetails(errorMessage string, errorType string, errorParam string) string {
	if errorMessage == "" {
		return ""
	}

	details := fmt.Sprintf("API Error: %s", errorMessage)
	if errorType != "" {
		details += fmt.Sprintf(" (Type: %s)", errorType)
	}
	if errorParam != "" {
		details += fmt.Sprintf(" (Param: %s)", errorParam)
	}

	return details
}

// FormatAPIErrorFromResponse creates a standardized LLMError from an OpenRouter API error
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
	errorMessage, errorType, errorParam := ParseErrorResponse(responseBody)
	errorDetails := FormatErrorDetails(errorMessage, errorType, errorParam)

	// We'll use the shared library error categorization instead of determining a message
	// from the API response - the CreateStandardErrorWithMessage function handles this

	// Try to categorize error from specific OpenRouter error types
	category := llm.CategoryUnknown

	// Check for BYOK-specific errors or misleading streaming errors
	if errorType == "custom_auth_missing" ||
		strings.Contains(errorMessage, "requires you to use your own") ||
		(strings.Contains(errorMessage, "Organization not authorized to use streaming") ||
			strings.Contains(errorMessage, "organization must be verified to stream")) {
		// The streaming error is misleading - it's actually a BYOK authentication issue
		category = llm.CategoryInvalidRequest
	} else if errorType != "" {
		if strings.Contains(errorType, "auth") ||
			strings.Contains(errorType, "authentication") ||
			strings.Contains(errorType, "unauthorized") {
			category = llm.CategoryAuth
		} else if strings.Contains(errorType, "rate_limit") {
			category = llm.CategoryRateLimit
		} else if strings.Contains(errorType, "invalid_request") {
			category = llm.CategoryInvalidRequest
		} else if strings.Contains(errorType, "not_found") {
			category = llm.CategoryNotFound
		} else if strings.Contains(errorType, "server_error") {
			category = llm.CategoryServer
		}
	}

	// If we couldn't categorize based on OpenRouter specific info,
	// use the shared categorization logic
	if category == llm.CategoryUnknown {
		category = llm.DetectErrorCategory(err, statusCode)
	}

	// Create a formatted error message with OpenRouter specific suggestions
	llmError := llm.CreateStandardErrorWithMessage("openrouter", category, err, errorDetails)

	// Add OpenRouter-specific suggestions for certain error types
	switch category {
	case llm.CategoryAuth:
		llmError.Suggestion = "Check that your OpenRouter API key is valid, has not expired, and starts with 'sk-or'. Ensure OPENROUTER_API_KEY environment variable is set correctly and not confused with other provider keys."
	case llm.CategoryInsufficientCredits:
		llmError.Suggestion = "Check your OpenRouter account balance and add credits if needed. Visit https://openrouter.ai/account for account details."
	case llm.CategoryNotFound:
		llmError.Suggestion = "Verify that the model name is correct and uses the format 'provider/model' or 'provider/organization/model'."
	case llm.CategoryServer:
		llmError.Suggestion = "This is typically a temporary issue with OpenRouter or the underlying model provider. Wait a few moments and try again."
	case llm.CategoryNetwork:
		llmError.Suggestion = "Check your internet connection and try again. If persistent, there may be connectivity issues to OpenRouter's servers."
	case llm.CategoryInvalidRequest:
		// Check if this is a BYOK error or misleading streaming error
		if strings.Contains(errorMessage, "requires you to use your own") ||
			errorType == "custom_auth_missing" ||
			strings.Contains(errorMessage, "Organization not authorized to use streaming") ||
			strings.Contains(errorMessage, "organization must be verified to stream") {
			// Special handling for o3 organization verification errors
			if strings.Contains(errorMessage, "organization must be verified") {
				llmError.Suggestion = "The o3 model requires an OpenAI organization-verified account. Personal API keys cannot access o3. You need to: 1) Verify your OpenAI organization at https://platform.openai.com/settings/organization/general, or 2) Use o4-mini instead, which works with personal accounts."
			} else {
				llmError.Suggestion = "This model requires you to bring your own API key (BYOK). Please add your provider's API key at https://openrouter.ai/settings/integrations to use this model. Note: The 'streaming' error is misleading - the actual issue is missing BYOK configuration."
			}
		}
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

// CreateAPIError creates a new LLMError with OpenRouter-specific settings
func CreateAPIError(category llm.ErrorCategory, errMsg string, originalErr error, details string) *llm.LLMError {
	llmError := llm.New(
		"openrouter", // Provider
		"",           // Code
		0,            // StatusCode
		errMsg,       // Message
		"",           // RequestID
		originalErr,  // Original error
		category,     // Error category
	)

	// Add details if provided
	if details != "" {
		llmError.Details = details
	}

	// Add OpenRouter-specific suggestions
	switch category {
	case llm.CategoryAuth:
		llmError.Suggestion = "Check that your OpenRouter API key is valid, has not expired, and starts with 'sk-or'. Ensure OPENROUTER_API_KEY environment variable is set correctly and not confused with other provider keys."
	case llm.CategoryRateLimit:
		llmError.Suggestion = "Wait and try again later. Consider adjusting the --max-concurrent and --rate-limit flags to limit request rate."
	case llm.CategoryInsufficientCredits:
		llmError.Suggestion = "Check your OpenRouter account balance and add credits if needed. Visit https://openrouter.ai/account for account details."
	case llm.CategoryInvalidRequest:
		llmError.Suggestion = "Check the prompt format and parameters. Ensure they comply with the API requirements."
	case llm.CategoryNotFound:
		llmError.Suggestion = "Verify that the model name is correct and uses the format 'provider/model' or 'provider/organization/model'."
	case llm.CategoryServer:
		llmError.Suggestion = "This is typically a temporary issue with OpenRouter or the underlying model provider. Wait a few moments and try again."
	case llm.CategoryNetwork:
		llmError.Suggestion = "Check your internet connection and try again. If persistent, there may be connectivity issues to OpenRouter's servers."
	case llm.CategoryCancelled:
		llmError.Suggestion = "The operation was interrupted. Try again with a longer timeout if needed."
	case llm.CategoryInputLimit:
		llmError.Suggestion = "Reduce the input size by using --include, --exclude, or --exclude-names flags to filter the context."
	case llm.CategoryContentFiltered:
		llmError.Suggestion = "Your prompt or content may have triggered safety filters. Review and modify your input to comply with content policies."
	default:
		llmError.Suggestion = "Check the logs for more details or try again."
	}

	return llmError
}
