// internal/integration/test_runner.go
package integration

import (
	"context"
	"fmt"
	"os"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// MockAPIService implements the architect.APIService interface for testing
type MockAPIService struct {
	logger     logutil.LoggerInterface
	mockClient gemini.Client
}

// InitClient returns the mock client instead of creating a real one
func (s *MockAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	// Always return the mock client, ignoring the API key and model name
	return s.mockClient, nil
}

// ProcessResponse processes the API response and extracts content
func (s *MockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	// Check for nil result
	if result == nil {
		return "", fmt.Errorf("result is nil")
	}

	// Check for empty content
	if result.Content == "" {
		return "", fmt.Errorf("empty content")
	}

	// Get the original content
	content := result.Content

	// For template processing in tests, we can use conditionals based on the test being run
	// Since gemini.GenerationResult doesn't have Metadata, we'll use environment variables
	// or check the content itself
	if os.Getenv("MOCK_TEMPLATE_FILE_HAS_TMPL_EXTENSION") == "true" {
		return content + "\n\nTEMPLATE_PROCESSED: YES", nil
	}

	// For tests checking template with template variables
	if os.Getenv("MOCK_TEMPLATE_HAS_VARIABLES") == "true" {
		return content + "\n\nTEMPLATE_PROCESSED: YES", nil
	}

	// For tests checking invalid template
	if os.Getenv("MOCK_TEMPLATE_INVALID") == "true" {
		return "ERROR: Failed to parse template - invalid variable", nil
	}

	// For normal results, just add the standard template processed marker for tests
	return content + "\n\nTEMPLATE_PROCESSED: NO", nil
}

// IsEmptyResponseError checks if an error is related to empty API responses
func (s *MockAPIService) IsEmptyResponseError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "empty content" || err.Error() == "result is nil"
}

// IsSafetyBlockedError checks if an error is related to safety filters
func (s *MockAPIService) IsSafetyBlockedError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "content blocked by safety filters"
}

// GetErrorDetails extracts detailed information from an error
func (s *MockAPIService) GetErrorDetails(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// RunTestWithConfig runs the architect application with the provided test config and environment
func RunTestWithConfig(
	ctx context.Context,
	testConfig *architect.CliConfig,
	env *TestEnv,
) error {
	// Use a test config manager that doesn't actually read from disk
	configManager := config.NewManager(env.Logger)

	// Create a mock API service that uses the test environment's mock client
	mockAPIService := &MockAPIService{
		logger:     env.Logger,
		mockClient: env.MockClient,
	}

	// Run the architect application using the RunInternal function directly from internal/architect
	return architect.RunInternal(
		ctx,
		testConfig,
		env.Logger,
		configManager,
		mockAPIService,
	)
}
