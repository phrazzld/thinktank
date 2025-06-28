// Package cli provides the command-line interface logic for the thinktank tool
package cli

import (
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

func TestNewMissingInstructionsError(t *testing.T) {
	err := NewMissingInstructionsError()
	if err == nil {
		t.Fatal("NewMissingInstructionsError() returned nil")
	}

	cliErr, ok := IsCLIError(err)
	if !ok {
		t.Fatal("NewMissingInstructionsError() did not return a CLIError")
	}

	if cliErr.Type != CLIErrorMissingRequired {
		t.Errorf("Expected CLIErrorMissingRequired, got %v", cliErr.Type)
	}

	msg := cliErr.UserFacingMessage()
	if !strings.Contains(msg, "instructions file") {
		t.Errorf("User message should mention instructions file, got: %s", msg)
	}
}

func TestNewInvalidModelError(t *testing.T) {
	model := "invalid-model"
	err := NewInvalidModelError(model)
	if err == nil {
		t.Fatal("NewInvalidModelError() returned nil")
	}

	cliErr, ok := IsCLIError(err)
	if !ok {
		t.Fatal("NewInvalidModelError() did not return a CLIError")
	}

	if cliErr.Type != CLIErrorInvalidValue {
		t.Errorf("Expected CLIErrorInvalidValue, got %v", cliErr.Type)
	}

	msg := cliErr.UserFacingMessage()
	if !strings.Contains(msg, model) {
		t.Errorf("User message should mention model %s, got: %s", model, msg)
	}
}

func TestNewNoAPIKeyError(t *testing.T) {
	provider := "test-provider"
	err := NewNoAPIKeyError(provider)
	if err == nil {
		t.Fatal("NewNoAPIKeyError() returned nil")
	}

	cliErr, ok := IsCLIError(err)
	if !ok {
		t.Fatal("NewNoAPIKeyError() did not return a CLIError")
	}

	if cliErr.Type != CLIErrorAuthentication {
		t.Errorf("Expected CLIErrorAuthentication, got %v", cliErr.Type)
	}

	msg := cliErr.UserFacingMessage()
	if !strings.Contains(msg, provider) {
		t.Errorf("User message should mention provider %s, got: %s", provider, msg)
	}
}

func TestMissingTargetPathError(t *testing.T) {
	// Test the predefined error
	err := ErrMissingTargetPath
	if err == nil {
		t.Fatal("ErrMissingTargetPath is nil")
	}

	if err.Type != CLIErrorMissingRequired {
		t.Errorf("Expected CLIErrorMissingRequired, got %v", err.Type)
	}

	msg := err.UserFacingMessage()
	if !strings.Contains(msg, "target path") {
		t.Errorf("User message should mention target path, got: %s", msg)
	}
}

func TestNewCLIErrorWithContext(t *testing.T) {
	message := "test message"
	suggestion := "test suggestion"
	context := "test context"
	err := NewCLIErrorWithContext(CLIErrorInvalidValue, message, suggestion, context)

	if err == nil {
		t.Fatal("NewCLIErrorWithContext() returned nil")
	}

	if err.Type != CLIErrorInvalidValue {
		t.Errorf("Expected CLIErrorInvalidValue, got %v", err.Type)
	}

	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}

	if err.Context != context {
		t.Errorf("Expected context %s, got %s", context, err.Context)
	}
}

func TestNewInvalidPathError(t *testing.T) {
	path := "/invalid/path"
	reason := "file not found"
	err := NewInvalidPathError(path, reason)
	if err == nil {
		t.Fatal("NewInvalidPathError() returned nil")
	}

	cliErr, ok := IsCLIError(err)
	if !ok {
		t.Fatal("NewInvalidPathError() did not return a CLIError")
	}

	if cliErr.Type != CLIErrorFileAccess {
		t.Errorf("Expected CLIErrorFileAccess, got %v", cliErr.Type)
	}

	msg := cliErr.UserFacingMessage()
	if !strings.Contains(msg, path) {
		t.Errorf("User message should mention path %s, got: %s", path, msg)
	}
	if !strings.Contains(msg, reason) {
		t.Errorf("User message should mention reason %s, got: %s", reason, msg)
	}
}

func TestNewCLIErrorWithCorrelation(t *testing.T) {
	message := "test message"
	suggestion := "test suggestion"
	correlationID := "test-correlation-id"

	err := NewCLIErrorWithCorrelation(message, suggestion, correlationID)
	if err == nil {
		t.Fatal("NewCLIErrorWithCorrelation() returned nil")
	}

	// This function returns an LLM error, not a CLI error directly
	var llmErr *llm.LLMError
	if !errors.As(err, &llmErr) {
		t.Fatal("NewCLIErrorWithCorrelation() did not return an LLMError")
	}

	if llmErr.ErrorCategory != llm.CategoryInvalidRequest {
		t.Errorf("Expected CategoryInvalidRequest, got %v", llmErr.ErrorCategory)
	}
}

func TestMapCLIErrorToLLMCategory(t *testing.T) {
	tests := []struct {
		name     string
		cliType  CLIErrorType
		expected llm.ErrorCategory
	}{
		{
			name:     "authentication error",
			cliType:  CLIErrorAuthentication,
			expected: llm.CategoryAuth,
		},
		{
			name:     "invalid value error",
			cliType:  CLIErrorInvalidValue,
			expected: llm.CategoryInvalidRequest,
		},
		{
			name:     "missing required error",
			cliType:  CLIErrorMissingRequired,
			expected: llm.CategoryInvalidRequest,
		},
		{
			name:     "file access error",
			cliType:  CLIErrorFileAccess,
			expected: llm.CategoryInputLimit,
		},
		{
			name:     "configuration error",
			cliType:  CLIErrorConfiguration,
			expected: llm.CategoryInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapCLIErrorToLLMCategory(tt.cliType)
			if result != tt.expected {
				t.Errorf("mapCLIErrorToLLMCategory(%v) = %v, want %v", tt.cliType, result, tt.expected)
			}
		})
	}
}

func TestWrapAsLLMError(t *testing.T) {
	cliErr := ErrMissingInstructions
	llmErr := WrapAsLLMError(cliErr)

	if llmErr == nil {
		t.Fatal("WrapAsLLMError() returned nil")
	}

	var llmErrType *llm.LLMError
	if !errors.As(llmErr, &llmErrType) {
		t.Fatal("WrapAsLLMError() did not return an LLMError")
	}

	if llmErrType.ErrorCategory != llm.CategoryInvalidRequest {
		t.Errorf("Expected CategoryInvalidRequest, got %v", llmErrType.ErrorCategory)
	}
}

func TestCLIError_Error(t *testing.T) {
	err := &CLIError{
		Type:    CLIErrorMissingRequired,
		Message: "test message",
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "test message") {
		t.Errorf("Error() should contain message, got: %s", errMsg)
	}
}

func TestCLIError_UserFacingMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		expected string
	}{
		{
			name: "error with suggestion",
			err: &CLIError{
				Type:       CLIErrorMissingRequired,
				Message:    "test message",
				Suggestion: "test suggestion",
			},
			expected: "test message. test suggestion",
		},
		{
			name: "error without suggestion",
			err: &CLIError{
				Type:    CLIErrorInvalidValue,
				Message: "test message",
			},
			expected: "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.err.UserFacingMessage()
			if msg != tt.expected {
				t.Errorf("UserFacingMessage() = %s, want %s", msg, tt.expected)
			}
		})
	}
}

func TestIsCLIError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "CLI error",
			err:      NewMissingInstructionsError(),
			expected: true,
		},
		{
			name:     "wrapped CLI error",
			err:      errors.New("wrapped: " + NewMissingInstructionsError().Error()),
			expected: false, // Not a direct CLI error
		},
		{
			name:     "non-CLI error",
			err:      errors.New("generic error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := IsCLIError(tt.err)
			if ok != tt.expected {
				t.Errorf("IsCLIError(%v) = %v, want %v", tt.err, ok, tt.expected)
			}
		})
	}
}

func TestCLIErrorWithContext(t *testing.T) {
	// Test creating a CLI error and verifying it maintains all properties
	err := &CLIError{
		Type:       CLIErrorAuthentication,
		Message:    "API key not found",
		Suggestion: "Set the OPENAI_API_KEY environment variable",
		Context:    "test-context",
	}

	// Test that it implements error interface
	var genericErr error = err
	if genericErr.Error() == "" {
		t.Error("CLIError should implement error interface")
	}

	// Test that IsCLIError works correctly
	if cliErr, ok := IsCLIError(err); !ok {
		t.Error("IsCLIError should return true for CLIError")
	} else {
		if cliErr.Type != CLIErrorAuthentication {
			t.Errorf("Expected CLIErrorAuthentication, got %v", cliErr.Type)
		}
		if cliErr.Context != "test-context" {
			t.Errorf("Expected context 'test-context', got %s", cliErr.Context)
		}
	}
}

func TestWrapCLIError(t *testing.T) {
	originalErr := errors.New("original error")
	message := "wrapped message"
	suggestion := "try this"
	context := "test context"

	wrappedErr := WrapCLIError(originalErr, CLIErrorFileAccess, message, suggestion, context)

	if wrappedErr == nil {
		t.Fatal("WrapCLIError() returned nil")
	}

	if wrappedErr.Type != CLIErrorFileAccess {
		t.Errorf("Expected CLIErrorFileAccess, got %v", wrappedErr.Type)
	}

	if wrappedErr.Message != message {
		t.Errorf("Expected message %s, got %s", message, wrappedErr.Message)
	}

	if wrappedErr.OriginalErr != originalErr {
		t.Errorf("Expected original error to be wrapped")
	}

	// Test unwrapping
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("WrapCLIError should allow unwrapping to original error")
	}
}

func TestCLIError_Category(t *testing.T) {
	err := &CLIError{
		Type: CLIErrorAuthentication,
	}

	category := err.Category()
	if category != llm.CategoryAuth {
		t.Errorf("Expected CategoryAuth, got %v", category)
	}
}

func TestCLIError_ErrorWithContext(t *testing.T) {
	tests := []struct {
		name     string
		err      *CLIError
		expected string
	}{
		{
			name: "error with context",
			err: &CLIError{
				Message: "test message",
				Context: "test context",
			},
			expected: "test context: test message",
		},
		{
			name: "error without context",
			err: &CLIError{
				Message: "test message",
			},
			expected: "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %s, want %s", result, tt.expected)
			}
		})
	}
}
