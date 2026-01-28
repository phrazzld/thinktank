// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"errors"
	"fmt"

	"github.com/misty-step/thinktank/internal/llm"
)

// CLI error constants - these are sentinel errors for common CLI scenarios
// Following Go's error handling best practices with pre-defined errors
var (
	// ErrMissingInstructions indicates that the required instructions file is missing
	ErrMissingInstructions = NewCLIError(CLIErrorMissingRequired, "instructions file required", "specify a .txt or .md file with analysis instructions")

	// ErrInvalidModel indicates that the specified model is not supported
	ErrInvalidModel = NewCLIError(CLIErrorInvalidValue, "invalid model specified", "use a supported model name from the available providers")

	// ErrNoAPIKey indicates that the required API key is not available
	ErrNoAPIKey = NewCLIError(CLIErrorAuthentication, "API key not set", "set the appropriate environment variable for your model provider")

	// ErrMissingTargetPath indicates that the target path is missing
	ErrMissingTargetPath = NewCLIError(CLIErrorMissingRequired, "target path required", "specify a file or directory to analyze")

	// ErrInvalidPath indicates that a specified path is invalid or inaccessible
	ErrInvalidPath = NewCLIError(CLIErrorFileAccess, "invalid path specified", "ensure the path exists and is accessible")
)

// CLIErrorType categorizes different types of CLI errors for appropriate exit code mapping
type CLIErrorType int

const (
	// CLIErrorMissingRequired indicates missing required arguments or flags
	CLIErrorMissingRequired CLIErrorType = iota

	// CLIErrorInvalidValue indicates invalid flag or argument values
	CLIErrorInvalidValue

	// CLIErrorFileAccess indicates file or directory access issues
	CLIErrorFileAccess

	// CLIErrorConfiguration indicates configuration-related problems
	CLIErrorConfiguration

	// CLIErrorAuthentication indicates API key or authentication issues
	CLIErrorAuthentication
)

// CLIError represents a CLI-specific error with context and user-friendly messaging
type CLIError struct {
	Type        CLIErrorType
	Message     string
	Suggestion  string
	Context     string
	OriginalErr error
}

// Error implements the error interface
func (e *CLIError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s: %s", e.Context, e.Message)
	}
	return e.Message
}

// Unwrap implements error unwrapping for Go 1.13+ error chains
func (e *CLIError) Unwrap() error {
	return e.OriginalErr
}

// UserFacingMessage returns a user-friendly error message with actionable suggestions
func (e *CLIError) UserFacingMessage() string {
	if e.Suggestion != "" {
		return fmt.Sprintf("%s. %s", e.Message, e.Suggestion)
	}
	return e.Message
}

// NewCLIError creates a new CLI error with specified type, message, and suggestion
func NewCLIError(errType CLIErrorType, message, suggestion string) *CLIError {
	return &CLIError{
		Type:       errType,
		Message:    message,
		Suggestion: suggestion,
	}
}

// NewCLIErrorWithContext creates a CLI error with additional context
func NewCLIErrorWithContext(errType CLIErrorType, message, suggestion, context string) *CLIError {
	return &CLIError{
		Type:       errType,
		Message:    message,
		Suggestion: suggestion,
		Context:    context,
	}
}

// WrapCLIError wraps an existing error with CLI error context
func WrapCLIError(originalErr error, errType CLIErrorType, message, suggestion, context string) *CLIError {
	return &CLIError{
		Type:        errType,
		Message:     message,
		Suggestion:  suggestion,
		Context:     context,
		OriginalErr: originalErr,
	}
}

// NewCLIErrorWithCorrelation creates a CLI error with correlation ID for debugging
func NewCLIErrorWithCorrelation(message, suggestion, correlationID string) error {
	// Create CLI error first
	cliErr := NewCLIError(CLIErrorInvalidValue, message, suggestion)

	// Wrap as LLM error to integrate with existing infrastructure and add correlation ID
	return llm.WrapWithCorrelationID(cliErr, "cli", message, llm.CategoryInvalidRequest, correlationID)
}

// WrapAsLLMError converts a CLI error to an LLM error for integration with existing error handling
func WrapAsLLMError(cliErr *CLIError) error {
	// Map CLI error types to LLM categories
	category := mapCLIErrorToLLMCategory(cliErr.Type)

	// Create user-friendly message
	message := cliErr.UserFacingMessage()

	// Wrap as LLM error with CLI provider
	return llm.Wrap(cliErr, "cli", message, category)
}

// IsCLIError checks if an error is a CLI error and returns it
func IsCLIError(err error) (*CLIError, bool) {
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr, true
	}
	return nil, false
}

// Category returns the LLM error category for this CLI error
func (e *CLIError) Category() llm.ErrorCategory {
	return mapCLIErrorToLLMCategory(e.Type)
}

// mapCLIErrorToLLMCategory maps CLI error types to LLM error categories for exit code determination
func mapCLIErrorToLLMCategory(errType CLIErrorType) llm.ErrorCategory {
	switch errType {
	case CLIErrorAuthentication:
		return llm.CategoryAuth
	case CLIErrorMissingRequired, CLIErrorInvalidValue:
		return llm.CategoryInvalidRequest
	case CLIErrorFileAccess:
		return llm.CategoryInputLimit
	case CLIErrorConfiguration:
		return llm.CategoryInvalidRequest
	default:
		return llm.CategoryInvalidRequest
	}
}

// Helper functions for common CLI error scenarios

// NewMissingInstructionsError creates an error for missing instructions file
func NewMissingInstructionsError() error {
	return WrapAsLLMError(ErrMissingInstructions)
}

// NewInvalidModelError creates an error for invalid model specification
func NewInvalidModelError(modelName string) error {
	cliErr := NewCLIErrorWithContext(
		CLIErrorInvalidValue,
		fmt.Sprintf("invalid model: %s", modelName),
		"use a supported model name from the available providers",
		"model validation",
	)
	return WrapAsLLMError(cliErr)
}

// NewNoAPIKeyError creates an error for missing API key with provider-specific guidance
func NewNoAPIKeyError(provider string) error {
	var envVar string
	var message string

	switch provider {
	case "openrouter":
		envVar = "OPENROUTER_API_KEY"
		message = "API key not set"
	case "test":
		// Test provider doesn't require an API key - this shouldn't happen
		message = "test provider doesn't require an API key"
		envVar = ""
	default:
		// Obsolete providers (openai, gemini) - provide helpful migration message
		envVar = "OPENROUTER_API_KEY"
		message = fmt.Sprintf("Provider '%s' is no longer supported. All models now use OpenRouter.", provider)
	}

	var suggestion string
	if envVar != "" {
		suggestion = fmt.Sprintf("Set the %s environment variable. Get your key at: https://openrouter.ai/keys", envVar)
	} else {
		suggestion = "This should not happen - please report this issue"
	}

	cliErr := NewCLIErrorWithContext(
		CLIErrorAuthentication,
		message,
		suggestion,
		"API key validation",
	)
	return WrapAsLLMError(cliErr)
}

// NewInvalidPathError creates an error for invalid file/directory paths
func NewInvalidPathError(path, reason string) error {
	cliErr := NewCLIErrorWithContext(
		CLIErrorFileAccess,
		fmt.Sprintf("invalid path '%s': %s", path, reason),
		"ensure the path exists and you have appropriate permissions",
		"path validation",
	)
	return WrapAsLLMError(cliErr)
}
