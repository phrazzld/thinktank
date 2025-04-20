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
	"github.com/phrazzld/architect/internal/registry"
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
		// Using empty params map to match the interface
		_, testErr := s.mockClient.GenerateContent(ctx, "test", map[string]interface{}{})
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

// GetModelDefinition retrieves the full model definition from the registry
func (s *mockIntAPIService) GetModelDefinition(modelName string) (*registry.ModelDefinition, error) {
	// Return a basic model definition for testing
	return &registry.ModelDefinition{
		Name:            modelName,
		Provider:        "gemini",
		APIModelID:      modelName,
		ContextWindow:   8192,
		MaxOutputTokens: 8192,
		Parameters: map[string]registry.ParameterDefinition{
			"temperature": {
				Type:    "float",
				Default: 0.7,
				Min:     0.0,
				Max:     1.0,
			},
			"max_tokens": {
				Type:    "integer",
				Default: 8192,
				Min:     1,
				Max:     8192,
			},
		},
	}, nil
}

// GetModelParameters retrieves parameter values from the registry for a given model
func (s *mockIntAPIService) GetModelParameters(modelName string) (map[string]interface{}, error) {
	// Return default parameters for testing
	return map[string]interface{}{
		"temperature": 0.7,
		"max_tokens":  8192,
	}, nil
}

// ValidateModelParameter validates a parameter value against its constraints
func (s *mockIntAPIService) ValidateModelParameter(modelName, paramName string, value interface{}) (bool, error) {
	// Basic validation for common parameters
	switch paramName {
	case "temperature":
		if temp, ok := value.(float64); ok {
			return temp >= 0.0 && temp <= 1.0, nil
		}
		return false, fmt.Errorf("temperature must be a float between 0 and 1")
	case "max_tokens":
		if tokens, ok := value.(int); ok {
			return tokens >= 1 && tokens <= 8192, nil
		}
		return false, fmt.Errorf("max_tokens must be an integer between 1 and 8192")
	default:
		// By default, accept unknown parameters
		return true, nil
	}
}

// GetModelTokenLimits retrieves token limits from the registry for a given model
func (s *mockIntAPIService) GetModelTokenLimits(modelName string) (contextWindow, maxOutputTokens int32, err error) {
	// Return standard limits for testing
	return 8192, 8192, nil
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
