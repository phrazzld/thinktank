// Package providers defines provider-related interfaces and implementations
package providers

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/misty-step/thinktank/internal/llm"
)

// TestProviderInterfaceBoundaries verifies that the Provider interface boundaries
// are working correctly and that implementations properly propagate
// errors and return values.
func TestProviderInterfaceBoundaries(t *testing.T) {
	// Create a mock provider that implements the Provider interface
	mockProvider := &mockProvider{
		createFunc: func(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error) {
			if apiKey == "" {
				return nil, ErrInvalidAPIKey
			}
			if modelID == "" {
				return nil, ErrInvalidModelID
			}
			if apiEndpoint == "invalid" {
				return nil, ErrInvalidEndpoint
			}
			if modelID == "error" {
				return nil, errors.New("generic error")
			}
			return &mockClient{
				modelName: modelID,
			}, nil
		},
	}

	// Test cases to verify the Provider interface behavior
	tests := []struct {
		name        string
		apiKey      string
		modelID     string
		apiEndpoint string
		wantErr     bool
		errCheck    func(error) bool
	}{
		{
			name:    "Empty API key",
			apiKey:  "",
			modelID: "test-model",
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, ErrInvalidAPIKey)
			},
		},
		{
			name:    "Empty model ID",
			apiKey:  "test-key",
			modelID: "",
			wantErr: true,
			errCheck: func(err error) bool {
				return errors.Is(err, ErrInvalidModelID)
			},
		},
		{
			name:        "Invalid endpoint",
			apiKey:      "test-key",
			modelID:     "test-model",
			apiEndpoint: "invalid",
			wantErr:     true,
			errCheck: func(err error) bool {
				return errors.Is(err, ErrInvalidEndpoint)
			},
		},
		{
			name:    "Error creating client",
			apiKey:  "test-key",
			modelID: "error",
			wantErr: true,
			errCheck: func(err error) bool {
				return err != nil && err.Error() == "generic error"
			},
		},
		{
			name:    "Successful client creation",
			apiKey:  "test-key",
			modelID: "test-model",
			wantErr: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := mockProvider.CreateClient(context.Background(), tt.apiKey, tt.modelID, tt.apiEndpoint)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check specific error if expected
			if tt.wantErr && tt.errCheck != nil && !tt.errCheck(err) {
				t.Errorf("CreateClient() error type mismatch: %v", err)
				return
			}

			// Check client
			if !tt.wantErr {
				if client == nil {
					t.Error("CreateClient() returned nil client but no error")
					return
				}
				// Verify client has expected model name
				if client.GetModelName() != tt.modelID {
					t.Errorf("Client model name = %s, want %s", client.GetModelName(), tt.modelID)
				}
			}
		})
	}
}

// Mock implementation of Provider interface for testing
type mockProvider struct {
	createFunc func(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error)
}

func (m *mockProvider) CreateClient(ctx context.Context, apiKey, modelID, apiEndpoint string) (llm.LLMClient, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, apiKey, modelID, apiEndpoint)
	}
	return &mockClient{}, nil
}

// Mock implementation of LLMClient interface for testing
type mockClient struct {
	modelName string
	content   string
	err       error
}

func (c *mockClient) GenerateContent(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
	if c.err != nil {
		return nil, c.err
	}
	content := c.content
	if content == "" {
		content = "Mock response for prompt: " + prompt
	}
	return &llm.ProviderResult{
		Content: content,
	}, nil
}

func (c *mockClient) GetModelName() string {
	if c.modelName == "" {
		return "mock-model"
	}
	return c.modelName
}

func (c *mockClient) Close() error {
	return nil
}

// TestErrorTypeComparison verifies that provider errors are properly defined
// and can be correctly identified with errors.Is()
func TestErrorTypeComparison(t *testing.T) {
	// Define test cases for each error type
	tests := []struct {
		name  string
		err   error
		check func(error) bool
	}{
		{
			name: "ErrProviderNotFound",
			err:  ErrProviderNotFound,
			check: func(err error) bool {
				return errors.Is(err, ErrProviderNotFound)
			},
		},
		{
			name: "ErrInvalidAPIKey",
			err:  ErrInvalidAPIKey,
			check: func(err error) bool {
				return errors.Is(err, ErrInvalidAPIKey)
			},
		},
		{
			name: "ErrInvalidModelID",
			err:  ErrInvalidModelID,
			check: func(err error) bool {
				return errors.Is(err, ErrInvalidModelID)
			},
		},
		{
			name: "ErrInvalidEndpoint",
			err:  ErrInvalidEndpoint,
			check: func(err error) bool {
				return errors.Is(err, ErrInvalidEndpoint)
			},
		},
		{
			name: "ErrClientCreation",
			err:  ErrClientCreation,
			check: func(err error) bool {
				return errors.Is(err, ErrClientCreation)
			},
		},
		{
			name: "Wrapped error",
			err:  fmt.Errorf("wrapped: %w", ErrProviderNotFound),
			check: func(err error) bool {
				return errors.Is(err, ErrProviderNotFound)
			},
		},
	}

	// Run test cases
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Verify the error exists
			if tc.err == nil {
				t.Fatalf("Error %s should not be nil", tc.name)
			}

			// Verify the error message is not empty
			if tc.err.Error() == "" {
				t.Errorf("Error %s has empty message", tc.name)
			}

			// Check that the error type is correctly identified
			if !tc.check(tc.err) {
				t.Errorf("Error type check failed for %s", tc.name)
			}
		})
	}
}
