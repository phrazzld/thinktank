package cli

import (
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

func TestNewNoAPIKeyErrorAllProviders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		provider           string
		expectedMessage    string
		expectedLLMWrapped bool
	}{
		{
			name:               "openai provider",
			provider:           "openai",
			expectedMessage:    "Provider 'openai' is no longer supported. All models now use OpenRouter",
			expectedLLMWrapped: true,
		},
		{
			name:               "gemini provider",
			provider:           "gemini",
			expectedMessage:    "Provider 'gemini' is no longer supported. All models now use OpenRouter",
			expectedLLMWrapped: true,
		},
		{
			name:               "openrouter provider",
			provider:           "openrouter",
			expectedMessage:    "OPENROUTER_API_KEY",
			expectedLLMWrapped: true,
		},
		{
			name:               "unknown provider",
			provider:           "unknown-provider",
			expectedMessage:    "Provider 'unknown-provider' is no longer supported. All models now use OpenRouter",
			expectedLLMWrapped: true,
		},
		{
			name:               "empty provider",
			provider:           "",
			expectedMessage:    "Provider '' is no longer supported. All models now use OpenRouter",
			expectedLLMWrapped: true,
		},
		{
			name:               "special characters provider",
			provider:           "test@provider#123",
			expectedMessage:    "Provider 'test@provider#123' is no longer supported. All models now use OpenRouter",
			expectedLLMWrapped: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewNoAPIKeyError(tt.provider)
			if err == nil {
				t.Fatal("NewNoAPIKeyError() returned nil")
			}

			// Check if it's wrapped as an LLM error
			if tt.expectedLLMWrapped {
				_, isLLMErr := err.(*llm.LLMError)
				if !isLLMErr {
					t.Errorf("Expected error to be wrapped as LLMError, got %T", err)
				}
			}

			// Check error message contains expected content
			if !strings.Contains(err.Error(), tt.expectedMessage) {
				t.Errorf("Error message should contain %q, got: %s", tt.expectedMessage, err.Error())
			}

			// Verify the underlying CLI error
			cliErr, isCLIErr := IsCLIError(err)
			if !isCLIErr {
				t.Fatal("Expected to be able to extract CLIError from result")
			}

			if cliErr.Type != CLIErrorAuthentication {
				t.Errorf("Expected CLIErrorAuthentication, got %v", cliErr.Type)
			}

			msg := cliErr.UserFacingMessage()
			if !strings.Contains(msg, "OPENROUTER_API_KEY") {
				t.Errorf("User message should mention OPENROUTER_API_KEY, got: %s", msg)
			}
		})
	}
}

func TestMapCLIErrorToLLMCategoryComprehensive(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		cliErrorType CLIErrorType
		expected     llm.ErrorCategory
	}{
		{
			name:         "authentication error",
			cliErrorType: CLIErrorAuthentication,
			expected:     llm.CategoryAuth,
		},
		{
			name:         "invalid value error",
			cliErrorType: CLIErrorInvalidValue,
			expected:     llm.CategoryInvalidRequest,
		},
		{
			name:         "missing required error",
			cliErrorType: CLIErrorMissingRequired,
			expected:     llm.CategoryInvalidRequest,
		},
		{
			name:         "file access error",
			cliErrorType: CLIErrorFileAccess,
			expected:     llm.CategoryInputLimit,
		},
		{
			name:         "configuration error",
			cliErrorType: CLIErrorConfiguration,
			expected:     llm.CategoryInvalidRequest,
		},
		{
			name:         "undefined error type",
			cliErrorType: CLIErrorType(999), // Invalid error type
			expected:     llm.CategoryInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapCLIErrorToLLMCategory(tt.cliErrorType)
			if result != tt.expected {
				t.Errorf("mapCLIErrorToLLMCategory(%v) = %v, want %v", tt.cliErrorType, result, tt.expected)
			}
		})
	}
}

func TestWrapAsLLMErrorAdditional(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		inputErr       *CLIError
		expectLLMError bool
		expectCategory llm.ErrorCategory
	}{
		{
			name:           "CLI authentication error",
			inputErr:       &CLIError{Type: CLIErrorAuthentication, Message: "auth failed"},
			expectLLMError: true,
			expectCategory: llm.CategoryAuth,
		},
		{
			name:           "CLI invalid value error",
			inputErr:       &CLIError{Type: CLIErrorInvalidValue, Message: "invalid value"},
			expectLLMError: true,
			expectCategory: llm.CategoryInvalidRequest,
		},
		{
			name:           "CLI missing required error",
			inputErr:       &CLIError{Type: CLIErrorMissingRequired, Message: "missing required"},
			expectLLMError: true,
			expectCategory: llm.CategoryInvalidRequest,
		},
		{
			name:           "CLI file access error",
			inputErr:       &CLIError{Type: CLIErrorFileAccess, Message: "file access failed"},
			expectLLMError: true,
			expectCategory: llm.CategoryInputLimit,
		},
		{
			name:           "CLI configuration error",
			inputErr:       &CLIError{Type: CLIErrorConfiguration, Message: "config failed"},
			expectLLMError: true,
			expectCategory: llm.CategoryInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapAsLLMError(tt.inputErr)
			if result == nil {
				t.Fatal("WrapAsLLMError() returned nil")
			}

			if tt.expectLLMError {
				llmErr, ok := result.(*llm.LLMError)
				if !ok {
					t.Errorf("Expected result to be *llm.LLMError, got %T", result)
					return
				}

				if llmErr.ErrorCategory != tt.expectCategory {
					t.Errorf("Expected category %v, got %v", tt.expectCategory, llmErr.ErrorCategory)
				}

				// Verify original error is preserved
				if !strings.Contains(result.Error(), tt.inputErr.Error()) {
					t.Errorf("Wrapped error should contain original error message")
				}
			}
		})
	}
}

func TestCLIErrorCategoryAdditional(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		cliError    *CLIError
		expectedCat llm.ErrorCategory
	}{
		{
			name:        "auth error category",
			cliError:    &CLIError{Type: CLIErrorAuthentication},
			expectedCat: llm.CategoryAuth,
		},
		{
			name:        "invalid value error category",
			cliError:    &CLIError{Type: CLIErrorInvalidValue},
			expectedCat: llm.CategoryInvalidRequest,
		},
		{
			name:        "missing required error category",
			cliError:    &CLIError{Type: CLIErrorMissingRequired},
			expectedCat: llm.CategoryInvalidRequest,
		},
		{
			name:        "file access error category",
			cliError:    &CLIError{Type: CLIErrorFileAccess},
			expectedCat: llm.CategoryInputLimit,
		},
		{
			name:        "configuration error category",
			cliError:    &CLIError{Type: CLIErrorConfiguration},
			expectedCat: llm.CategoryInvalidRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cliError.Category()
			if result != tt.expectedCat {
				t.Errorf("CLIError.Category() = %v, want %v", result, tt.expectedCat)
			}
		})
	}
}
