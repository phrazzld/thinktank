// internal/integration/test_runner.go
package integration

import (
	"context"
	"fmt"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockIntAPIService implements the architect.APIService interface for testing
// This type is actually used by running tests but wasn't detected by linting
type mockIntAPIService struct {
	logger        logutil.LoggerInterface
	mockClient    gemini.Client
	mockLLMClient llm.LLMClient
}

// InitLLMClient returns the mock LLM client instead of creating a real one
func (s *mockIntAPIService) InitLLMClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
	// If mockLLMClient is not nil, return it (for tests that explicitly set it)
	if s.mockLLMClient != nil {
		return s.mockLLMClient, nil
	}

	// Get mockClient GenerateContentFunc to check if it's configured to return an error
	if s.mockClient != nil && s.mockClient.(*gemini.MockClient).GenerateContentFunc != nil {
		// Try to generate content with a test prompt to see if it would return an error
		_, testErr := s.mockClient.GenerateContent(ctx, "test")
		if testErr != nil {
			// If the mockClient is configured to return an error, propagate that setup
			return nil, fmt.Errorf("simulated client initialization error: %v", testErr)
		}
	}

	// Create adapter that wraps the mock gemini client to implement llm.LLMClient
	return NewLLMClientAdapter(s.mockClient, modelName), nil
}

// IsEmptyResponseError checks if an error is related to empty API responses
func (s *mockIntAPIService) IsEmptyResponseError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "empty content" || err.Error() == "result is nil"
}

// IsSafetyBlockedError checks if an error is related to safety filters
func (s *mockIntAPIService) IsSafetyBlockedError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "content blocked by safety filters"
}

// GetErrorDetails extracts detailed information from an error
func (s *mockIntAPIService) GetErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// ProcessLLMResponse processes the provider-agnostic API response and extracts content
func (s *mockIntAPIService) ProcessLLMResponse(result *llm.ProviderResult) (string, error) {
	// Check for nil result
	if result == nil {
		return "", fmt.Errorf("result is nil")
	}

	// Check for empty content
	if result.Content == "" {
		return "", fmt.Errorf("empty content")
	}

	// Return the content directly
	return result.Content, nil
}

// Note: The mockLLMClientForTesting struct has been replaced by the LLMClientAdapter
// which is used consistently to adapt a gemini.MockClient to the llm.LLMClient interface.

// RunTestWithConfig runs the architect application with the provided test config and environment
func RunTestWithConfig(
	ctx context.Context,
	testConfig *config.CliConfig,
	env *TestEnv,
) error {
	// Create a mock API service directly without modifying any global variables
	mockApiService := &mockIntAPIService{
		logger:     env.Logger,
		mockClient: env.MockClient,
	}

	// Run the architect application using Execute with the mock API service
	return architect.Execute(
		ctx,
		testConfig,
		env.Logger,
		env.AuditLogger,
		mockApiService,
	)
}
