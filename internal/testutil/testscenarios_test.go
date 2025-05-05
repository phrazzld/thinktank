// Package testutil provides utilities for testing in the thinktank project
package testutil

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/logutil"
)

// TestCreateStandardTestRegistry tests the creation of a registry with standard models and providers
func TestCreateStandardTestRegistry(t *testing.T) {
	registry := CreateStandardTestRegistry()

	// Verify that the registry contains the expected models and providers
	model, err := registry.GetModel("test-model")
	if err != nil {
		t.Errorf("Failed to get test model: %v", err)
	}
	if model == nil {
		t.Error("Expected test model to be in registry, but got nil")
	}

	provider, err := registry.GetProvider("test-provider")
	if err != nil {
		t.Errorf("Failed to get test provider: %v", err)
	}
	if provider == nil {
		t.Error("Expected test provider to be in registry, but got nil")
	}
}

// TestSetupSuccessClient tests the creation of a success client
func TestSetupSuccessClient(t *testing.T) {
	registry, client, provider := SetupSuccessClient(t)

	// Check that all components were created correctly
	if registry == nil {
		t.Fatal("Expected registry to be created, but got nil")
	}
	if client == nil {
		t.Fatal("Expected client to be created, but got nil")
	}
	if provider == nil {
		t.Fatal("Expected provider to be created, but got nil")
	}

	// Test that client returns a successful response
	response, err := client.GenerateContent(context.Background(), "test prompt", nil)
	if err != nil {
		t.Errorf("Expected client.GenerateContent to succeed, but got error: %v", err)
	}
	if response == nil {
		t.Error("Expected client to return a response, but got nil")
	} else if response.Content != "Test successful response" {
		t.Errorf("Expected client to return 'Test successful response', but got '%s'", response.Content)
	}

	// Test that the provider returns the client
	impl, err := provider.CreateClient(context.Background(), "api-key", "model-id", "")
	if err != nil {
		t.Errorf("Expected provider.CreateClient to succeed, but got error: %v", err)
	}
	if impl != client {
		t.Error("Expected provider to return the same client that was created")
	}
}

// TestSetupErrorClient tests the creation of error clients
func TestSetupErrorClient(t *testing.T) {
	testCases := []struct {
		name      string
		errorType string
	}{
		{"Auth error", "auth_error"},
		{"Rate limit error", "rate_limit"},
		{"Safety blocked", "safety_blocked"},
		{"Client creation error", "client_creation_error"},
		{"Model not found", "model_not_found"},
		{"Provider not found", "provider_not_found"},
		{"Unknown error", "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry, client, provider, expectedError := SetupErrorClient(t, tc.errorType)

			// Check that registry was created
			if registry == nil {
				t.Fatal("Expected registry to be created, but got nil")
			}

			// Verify expected behavior based on error type
			switch tc.errorType {
			case "auth_error", "rate_limit", "safety_blocked":
				// These should have a client that returns an error
				if client == nil {
					t.Error("Expected client to be created")
				} else {
					_, err := client.GenerateContent(context.Background(), "test", nil)
					if err == nil {
						t.Error("Expected client.GenerateContent to return an error, but got nil")
					} else if !errors.Is(err, expectedError) {
						t.Errorf("Expected error to be %v, but got %v", expectedError, err)
					}
				}

			case "client_creation_error":
				// Client should be nil, but provider should return an error
				if client != nil {
					t.Error("Expected client to be nil")
				}
				if provider == nil {
					t.Error("Expected provider to be created")
				} else {
					_, err := provider.CreateClient(context.Background(), "api-key", "model-id", "")
					if err == nil {
						t.Error("Expected provider.CreateClient to return an error, but got nil")
					} else if !errors.Is(err, expectedError) {
						t.Errorf("Expected error to be %v, but got %v", expectedError, err)
					}
				}

			case "model_not_found":
				// Registry should return an error for GetModel
				_, err := registry.GetModel("test-model")
				if err == nil {
					t.Error("Expected registry.GetModel to return an error, but got nil")
				}

			case "provider_not_found":
				// Registry should return an error for GetProvider
				_, err := registry.GetProvider("test-provider")
				if err == nil {
					t.Error("Expected registry.GetProvider to return an error, but got nil")
				}
			}
		})
	}
}

// TestCreateContextWithTimeout tests context creation with timeout
func TestCreateContextWithTimeout(t *testing.T) {
	correlationID := "test-correlation-id"
	timeout := 100 * time.Millisecond

	ctx, cancel := CreateContextWithTimeout(timeout, correlationID)
	defer cancel() // Make sure to cancel the context to prevent leaks

	// Check that the correlation ID was set
	actualCorrelationID := logutil.GetCorrelationID(ctx)
	if actualCorrelationID != correlationID {
		t.Errorf("Expected correlation ID to be '%s', but got '%s'", correlationID, actualCorrelationID)
	}

	// Check that the timeout was set by waiting for the context to expire
	select {
	case <-ctx.Done():
		// Expected behavior
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected context to expire after timeout, but it did not")
	}
}

// TestCommonTestParameters tests getting common test parameters
func TestCommonTestParameters(t *testing.T) {
	params := CommonTestParameters()

	// Check that expected parameters are present
	expectedKeys := []string{"temperature", "top_p", "max_tokens"}
	for _, key := range expectedKeys {
		if _, ok := params[key]; !ok {
			t.Errorf("Expected parameter '%s' to be present, but it was not", key)
		}
	}

	// Check specific parameter values
	if temp, ok := params["temperature"].(float64); !ok || temp != 0.7 {
		t.Errorf("Expected temperature to be 0.7, but got %v", params["temperature"])
	}
	if topP, ok := params["top_p"].(float64); !ok || topP != 0.95 {
		t.Errorf("Expected top_p to be 0.95, but got %v", params["top_p"])
	}
	if maxTokens, ok := params["max_tokens"].(int); !ok || maxTokens != 1024 {
		t.Errorf("Expected max_tokens to be 1024, but got %v", params["max_tokens"])
	}
}

// TestAssertErrorType tests the error assertion helper
func TestAssertErrorType(t *testing.T) {
	// Test directly without a mock reporter
	// Test case 1: No error expected, no error provided - should not fail
	AssertErrorType(t, nil, nil)

	// We can't test the failure cases directly as they would fail the test
	// In a real test, AssertErrorType will correctly report test failures
}

// TestAssertClientCalled tests the client call assertion helper
func TestAssertClientCalled(t *testing.T) {
	// Create a client and make a call
	client := NewMockLLMClient("test-model", BasicSuccessResponse, nil)
	_, _ = client.GenerateContent(context.Background(), "test prompt", nil)

	// Test case: Verify that a client with the right call doesn't fail
	AssertClientCalled(t, client, "test prompt")

	// We can't test the failure cases directly as they would fail the test
	// In a real test, AssertClientCalled will correctly report test failures
}

// We don't need the mock test reporter anymore
