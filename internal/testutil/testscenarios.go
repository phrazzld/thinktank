// Package testutil provides utilities for testing in the thinktank project
package testutil

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
)

// Common test scenarios and setups for registry API testing

// CreateStandardTestRegistry creates a registry with standard models and providers
func CreateStandardTestRegistry() *MockRegistry {
	registry := NewMockRegistry()

	// Add standard test model
	registry.AddModel(testModelDefinition{
		Name:       "test-model",
		Provider:   "test-provider",
		APIModelID: "test-model-id",
	})

	// Add standard test provider
	registry.AddProvider(testProviderDefinition{
		Name:    "test-provider",
		BaseURL: "https://test.example.com",
	})

	return registry
}

// SetupSuccessClient creates a registry with a model, provider, and a success client
func SetupSuccessClient(t testing.TB) (*MockRegistry, *MockLLMClient, *MockProvider) {
	t.Helper()

	// Create a registry with standard models and providers
	registry := CreateStandardTestRegistry()

	// Create a success response
	response := CreateSuccessResponse("Test successful response")

	// Create a success client
	client := NewMockLLMClient("test-model", response, nil)

	// Create a success provider
	provider := NewMockProvider("test-provider", client, nil)

	// Add the provider implementation
	registry.AddProviderImplementation("test-provider", provider)

	return registry, client, provider
}

// SetupErrorClient creates a registry with a model, provider, and an error client
func SetupErrorClient(t testing.TB, errorType string) (*MockRegistry, *MockLLMClient, *MockProvider, error) {
	t.Helper()

	// Create a registry with standard models and providers
	registry := CreateStandardTestRegistry()

	var client *MockLLMClient
	var provider *MockProvider
	var expectedError error

	switch errorType {
	case "auth_error":
		authError := CreateAuthError("test-provider").Original
		client = NewMockLLMClient("test-model", nil, authError)
		provider = NewMockProvider("test-provider", client, nil)
		expectedError = authError

	case "rate_limit":
		rateLimitError := CreateRateLimitError("test-provider").Original
		client = NewMockLLMClient("test-model", nil, rateLimitError)
		provider = NewMockProvider("test-provider", client, nil)
		expectedError = rateLimitError

	case "safety_blocked":
		client = NewMockLLMClient("test-model", nil, llm.ErrSafetyBlocked)
		provider = NewMockProvider("test-provider", client, nil)
		expectedError = llm.ErrSafetyBlocked

	case "client_creation_error":
		client = nil
		provider = NewMockProvider("test-provider", nil, llm.ErrClientInitialization)
		expectedError = llm.ErrClientInitialization

	case "model_not_found":
		client = nil
		provider = nil
		registry.SetGetModelError(errors.New("model not found"))
		expectedError = llm.ErrModelNotFound

	case "provider_not_found":
		client = nil
		provider = nil
		registry.SetGetProviderError(errors.New("provider not found"))
		expectedError = errors.New("provider not found")

	default:
		client = NewMockLLMClient("test-model", nil, errors.New("unknown error"))
		provider = NewMockProvider("test-provider", client, nil)
		expectedError = errors.New("unknown error")
	}

	// Add the provider implementation if applicable
	if provider != nil {
		registry.AddProviderImplementation("test-provider", provider)
	}

	return registry, client, provider, expectedError
}

// CreateContextWithTimeout creates a context with a timeout and correlation ID
// Note: Caller is responsible for calling the returned CancelFunc to avoid context leaks
func CreateContextWithTimeout(timeout time.Duration, correlationID string) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	return logutil.WithCustomCorrelationID(ctx, correlationID), cancel
}

// CommonTestParameters returns a map of common test parameters
func CommonTestParameters() map[string]interface{} {
	return map[string]interface{}{
		"temperature": 0.7,
		"top_p":       0.95,
		"max_tokens":  1024,
	}
}

// AssertErrorType checks if an error is of the expected type
func AssertErrorType(t testing.TB, err, expectedErr error) {
	t.Helper()

	if expectedErr == nil {
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		return
	}

	if err == nil {
		t.Errorf("Expected error %v, but got nil", expectedErr)
		return
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error %v, but got %v", expectedErr, err)
	}
}

// AssertClientCalled checks if a client was called with the expected prompt
func AssertClientCalled(t testing.TB, client *MockLLMClient, expectedPrompt string) {
	t.Helper()

	if client == nil {
		t.Error("Client is nil, cannot check call")
		return
	}

	if len(client.Calls) == 0 {
		t.Error("Client was not called")
		return
	}

	// Check if any call had the expected prompt
	found := false
	for _, call := range client.Calls {
		if call.Prompt == expectedPrompt {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected client to be called with prompt %q, but it was not", expectedPrompt)
	}
}

// RunCommonParameterTests runs common parameter validation tests
func RunCommonParameterTests(t *testing.T, validateFunc func(string, string, interface{}) (bool, error)) {
	t.Helper()

	// Temperature tests
	t.Run("Temperature", func(t *testing.T) {
		validTemps := []float64{0.0, 0.5, 1.0}
		for _, temp := range validTemps {
			valid, err := validateFunc("test-model", "temperature", temp)
			if err != nil {
				t.Errorf("Unexpected error for temperature %v: %v", temp, err)
			}
			if !valid {
				t.Errorf("Expected temperature %v to be valid, but it was not", temp)
			}
		}

		invalidTemps := []float64{-0.1, 1.1, 2.0}
		for _, temp := range invalidTemps {
			valid, err := validateFunc("test-model", "temperature", temp)
			if err != nil {
				t.Errorf("Unexpected error for temperature %v: %v", temp, err)
			}
			if valid {
				t.Errorf("Expected temperature %v to be invalid, but it was marked valid", temp)
			}
		}

		// Test invalid type
		_, err := validateFunc("test-model", "temperature", "not a number")
		if err == nil {
			t.Error("Expected error for non-numeric temperature, but got nil")
		}
	})

	// Max tokens tests
	t.Run("MaxTokens", func(t *testing.T) {
		validTokens := []int{1, 100, 1024, 8192}
		for _, tokens := range validTokens {
			valid, err := validateFunc("test-model", "max_tokens", tokens)
			if err != nil {
				t.Errorf("Unexpected error for max_tokens %v: %v", tokens, err)
			}
			if !valid {
				t.Errorf("Expected max_tokens %v to be valid, but it was not", tokens)
			}
		}

		invalidTokens := []int{0, -1, 9000}
		for _, tokens := range invalidTokens {
			valid, err := validateFunc("test-model", "max_tokens", tokens)
			if err != nil {
				t.Errorf("Unexpected error for max_tokens %v: %v", tokens, err)
			}
			if valid {
				t.Errorf("Expected max_tokens %v to be invalid, but it was marked valid", tokens)
			}
		}
	})
}

// RunCommonModelTests runs common model-related tests
func RunCommonModelTests(t testing.TB, registry *MockRegistry) {
	t.Helper()

	// Test getting standard models
	standardModels := []string{"test-model"}
	ctx := context.Background()
	for _, modelName := range standardModels {
		model, err := registry.GetModel(ctx, modelName)
		if err != nil {
			t.Errorf("Failed to get model %s: %v", modelName, err)
			continue
		}

		if model == nil {
			t.Errorf("Model %s was nil", modelName)
			continue
		}

		t.Logf("Successfully retrieved model %s", modelName)
	}

	// Test non-existent model
	_, err := registry.GetModel(ctx, "non-existent-model")
	if err == nil {
		t.Error("Expected error when getting non-existent model, but got nil")
	}
}

// RunCommonProviderTests runs common provider-related tests
func RunCommonProviderTests(t testing.TB, registry *MockRegistry) {
	t.Helper()

	// Test getting standard providers
	standardProviders := []string{"test-provider"}
	ctx := context.Background()
	for _, providerName := range standardProviders {
		provider, err := registry.GetProvider(ctx, providerName)
		if err != nil {
			t.Errorf("Failed to get provider %s: %v", providerName, err)
			continue
		}

		if provider == nil {
			t.Errorf("Provider %s was nil", providerName)
			continue
		}

		t.Logf("Successfully retrieved provider %s", providerName)
	}

	// Test non-existent provider
	_, err := registry.GetProvider(ctx, "non-existent-provider")
	if err == nil {
		t.Error("Expected error when getting non-existent provider, but got nil")
	}
}
