package testutil

import (
	"context"
	"errors"
	"testing"

	"github.com/phrazzld/thinktank/internal/llm"
)

// TestNewRegistryTestSetup verifies that the test setup is correctly initialized
func TestNewRegistryTestSetup(t *testing.T) {
	setup := NewRegistryTestSetup()

	// Check that components are not nil
	if setup.Registry == nil {
		t.Error("Expected Registry to be initialized, but it was nil")
	}
	if setup.Logger == nil {
		t.Error("Expected Logger to be initialized, but it was nil")
	}
	if setup.TestClient == nil {
		t.Error("Expected TestClient to be initialized, but it was nil")
	}
	if setup.FailureClient == nil {
		t.Error("Expected FailureClient to be initialized, but it was nil")
	}

	// Check that standard fields are set
	if setup.ModelName == "" {
		t.Error("Expected ModelName to be set, but it was empty")
	}
	if setup.ProviderName == "" {
		t.Error("Expected ProviderName to be set, but it was empty")
	}
	if setup.CorrelationID == "" {
		t.Error("Expected CorrelationID to be set, but it was empty")
	}
}

// TestSetupStandardRegistry verifies that standard registry setup works correctly
func TestSetupStandardRegistry(t *testing.T) {
	setup := NewRegistryTestSetup()
	setup.SetupStandardRegistry()

	// Verify that model was added
	ctx := context.Background()
	model, err := setup.Registry.GetModel(ctx, setup.ModelName)
	if err != nil {
		t.Errorf("Expected model %s to be added, but got error: %v", setup.ModelName, err)
	}
	if model == nil {
		t.Errorf("Expected model %s to be added, but it was not", setup.ModelName)
	} else if model.Provider != setup.ProviderName {
		t.Errorf("Expected model provider to be %s, but got %s", setup.ProviderName, model.Provider)
	}

	// Verify that provider was added
	provider, err := setup.Registry.GetProvider(ctx, setup.ProviderName)
	if err != nil {
		t.Errorf("Expected provider %s to be added, but got error: %v", setup.ProviderName, err)
	}
	if provider == nil {
		t.Errorf("Expected provider %s to be added, but it was not", setup.ProviderName)
	}

	// Verify that provider implementation was added
	impl, err := setup.Registry.GetProviderImplementation(ctx, setup.ProviderName)
	if err != nil {
		t.Errorf("Expected provider implementation for %s to be added, but got error: %v", setup.ProviderName, err)
	}
	if impl == nil {
		t.Errorf("Expected provider implementation for %s to be added, but it was not", setup.ProviderName)
	}
}

// TestSetupErrorCase verifies that error cases are configured correctly
func TestSetupErrorCase(t *testing.T) {
	testCases := []struct {
		name        string
		errorType   string
		methodCheck func(*testing.T, *RegistryTestSetup)
	}{
		{
			name:      "Model not found error",
			errorType: "model_not_found",
			methodCheck: func(t *testing.T, setup *RegistryTestSetup) {
				t.Helper()
				_, err := setup.Registry.GetModel("any-model")
				if err == nil {
					t.Error("Expected GetModel to return an error, but it did not")
				}
			},
		},
		{
			name:      "Provider not found error",
			errorType: "provider_not_found",
			methodCheck: func(t *testing.T, setup *RegistryTestSetup) {
				t.Helper()
				_, err := setup.Registry.GetProvider("any-provider")
				if err == nil {
					t.Error("Expected GetProvider to return an error, but it did not")
				}
			},
		},
		{
			name:      "Implementation not found error",
			errorType: "implementation_not_found",
			methodCheck: func(t *testing.T, setup *RegistryTestSetup) {
				t.Helper()
				_, err := setup.Registry.GetProviderImplementation("any-provider")
				if err == nil {
					t.Error("Expected GetProviderImplementation to return an error, but it did not")
				}
			},
		},
		{
			name:      "Client creation error",
			errorType: "client_creation_failed",
			methodCheck: func(t *testing.T, setup *RegistryTestSetup) {
				t.Helper()
				setup.SetupStandardRegistry() // We need a model and provider first

				// Then we override the provider implementation
				setup.SetupErrorCase("client_creation_failed")

				// Check if the error is configured correctly
				if setup.ConfiguredError != llm.ErrClientInitialization {
					t.Errorf("Expected ConfiguredError to be %v, but got %v", llm.ErrClientInitialization, setup.ConfiguredError)
				}

				// Get the provider implementation
				impl, err := setup.Registry.GetProviderImplementation(setup.ProviderName)
				if err != nil {
					t.Errorf("Expected GetProviderImplementation to succeed, but got error: %v", err)
					return
				}

				// Check if it's our mock provider with the error
				mockProvider, ok := impl.(*MockProvider)
				if !ok {
					t.Error("Expected implementation to be a MockProvider")
					return
				}

				if mockProvider.Error != llm.ErrClientInitialization {
					t.Errorf("Expected provider error to be %v, but got %v", llm.ErrClientInitialization, mockProvider.Error)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setup := NewRegistryTestSetup()
			setup.SetupErrorCase(tc.errorType)

			if setup.ConfiguredError == nil {
				t.Error("Expected ConfiguredError to be set, but it was nil")
			}

			tc.methodCheck(t, setup)
		})
	}
}

// TestMockLLMClient verifies that MockLLMClient works correctly
func TestMockLLMClient(t *testing.T) {
	response := BasicSuccessResponse
	client := NewMockLLMClient("test-model", response, nil)

	// Test GetModelName
	if client.GetModelName() != "test-model" {
		t.Errorf("Expected GetModelName to return 'test-model', but got '%s'", client.GetModelName())
	}

	// Test GenerateContent with success
	ctx := context.Background()
	result, err := client.GenerateContent(ctx, "test prompt", nil)
	if err != nil {
		t.Errorf("Expected GenerateContent to succeed, but got error: %v", err)
	}
	if result != response {
		t.Error("Expected GenerateContent to return the configured response")
	}

	// Check call recording
	if len(client.Calls) != 1 {
		t.Errorf("Expected 1 call to GenerateContent, but got %d", len(client.Calls))
	} else {
		call := client.Calls[0]
		if call.Prompt != "test prompt" {
			t.Errorf("Expected prompt to be 'test prompt', but got '%s'", call.Prompt)
		}
	}

	// Test with error
	errorClient := NewMockLLMClient("error-model", nil, errors.New("test error"))
	_, err = errorClient.GenerateContent(ctx, "test prompt", nil)
	if err == nil {
		t.Error("Expected GenerateContent to return an error, but it did not")
	}
}

// TestMockProvider verifies that MockProvider works correctly
func TestMockProvider(t *testing.T) {
	client := NewMockLLMClient("test-model", BasicSuccessResponse, nil)
	provider := NewMockProvider("test-provider", client, nil)

	// Test successful client creation
	ctx := context.Background()
	result, err := provider.CreateClient(ctx, "api-key", "model-id", "")
	if err != nil {
		t.Errorf("Expected CreateClient to succeed, but got error: %v", err)
	}
	if result != client {
		t.Error("Expected CreateClient to return the configured client")
	}

	// Check call recording
	if len(provider.Calls) != 1 {
		t.Errorf("Expected 1 call to CreateClient, but got %d", len(provider.Calls))
	} else {
		call := provider.Calls[0]
		if call.APIKey != "api-key" {
			t.Errorf("Expected APIKey to be 'api-key', but got '%s'", call.APIKey)
		}
		if call.ModelID != "model-id" {
			t.Errorf("Expected ModelID to be 'model-id', but got '%s'", call.ModelID)
		}
	}

	// Test with error
	errorProvider := NewMockProvider("error-provider", nil, errors.New("test error"))
	_, err = errorProvider.CreateClient(ctx, "api-key", "model-id", "")
	if err == nil {
		t.Error("Expected CreateClient to return an error, but it did not")
	}
}

// TestRegistryCallTracking verifies that registry method calls are properly tracked
func TestRegistryCallTracking(t *testing.T) {
	// Create a registry and call some methods
	registry := NewMockRegistry()
	registry.GetAllModelNames()                  // Call 1
	_, _ = registry.GetModel("test-model")       // Call 2, ignoring return values for test
	_, _ = registry.GetProvider("test-provider") // Call 3, ignoring return values for test

	// Check that the calls were recorded
	calls := registry.GetMethodCalls("GetAllModelNames")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call to GetAllModelNames, got %d", len(calls))
	}

	calls = registry.GetMethodCalls("GetModel")
	if len(calls) != 1 {
		t.Errorf("Expected 1 call to GetModel, got %d", len(calls))
	}

	// Check that the call sequence is correct
	sequence := registry.GetCallSequence()
	if len(sequence) != 3 {
		t.Errorf("Expected 3 calls in sequence, got %d", len(sequence))
	}

	// Check specific method order
	if len(sequence) >= 3 {
		if sequence[0] != "GetAllModelNames" {
			t.Errorf("Expected first call to be GetAllModelNames, got %s", sequence[0])
		}
		if sequence[1] != "GetModel" {
			t.Errorf("Expected second call to be GetModel, got %s", sequence[1])
		}
		if sequence[2] != "GetProvider" {
			t.Errorf("Expected third call to be GetProvider, got %s", sequence[2])
		}
	}

	// Test ClearMethodCalls
	registry.ClearMethodCalls()
	calls = registry.GetMethodCalls("GetAllModelNames")
	if len(calls) != 0 {
		t.Errorf("Expected 0 calls after clear, got %d", len(calls))
	}
}

// TestCreatePreConfiguredMockRegistry tests the CreatePreConfiguredMockRegistry helper function
func TestCreatePreConfiguredMockRegistry(t *testing.T) {
	registry := CreatePreConfiguredMockRegistry()

	// Check for some standard models
	models := []string{"gpt-4", "gemini-1.5-pro", "claude-3-opus"}
	for _, modelName := range models {
		model, err := registry.GetModel(modelName)
		if err != nil {
			t.Errorf("Expected model %s to be in pre-configured registry, but got error: %v", modelName, err)
		}
		if model == nil {
			t.Errorf("Expected model %s to be in pre-configured registry, but it was nil", modelName)
		}
	}

	// Check for some standard providers
	providers := []string{"openai", "gemini", "openrouter"}
	for _, providerName := range providers {
		provider, err := registry.GetProvider(providerName)
		if err != nil {
			t.Errorf("Expected provider %s to be in pre-configured registry, but got error: %v", providerName, err)
		}
		if provider == nil {
			t.Errorf("Expected provider %s to be in pre-configured registry, but it was nil", providerName)
		}
	}
}

// TestGetTestProviderMap tests the GetTestProviderMap helper function
func TestGetTestProviderMap(t *testing.T) {
	providerMap := GetTestProviderMap()

	expectedProviders := []string{"success", "empty", "safety_blocked", "client_error", "rate_limited"}
	for _, name := range expectedProviders {
		provider, ok := providerMap[name]
		if !ok {
			t.Errorf("Expected provider %s to be in test provider map, but it was not", name)
			continue
		}

		if provider.Name != name {
			t.Errorf("Expected provider name to be %s, but got %s", name, provider.Name)
		}
	}

	// Check specific provider behaviors
	ctx := context.Background()

	// Success provider should return a success response
	successProvider := providerMap["success"]
	client, err := successProvider.CreateClient(ctx, "api-key", "model-id", "")
	if err != nil {
		t.Errorf("Expected success provider to succeed, but got error: %v", err)
	}
	if client == nil {
		t.Error("Expected success provider to return a client, but got nil")
	}

	// Client error provider should return an error
	clientErrorProvider := providerMap["client_error"]
	_, err = clientErrorProvider.CreateClient(ctx, "api-key", "model-id", "")
	if err == nil {
		t.Error("Expected client_error provider to return an error, but it did not")
	}
}
