// internal/integration/test_runner.go
package integration

import (
	"context"
	"fmt"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockIntAPIService implements the architect.APIService interface for testing
// This type is actually used by running tests but wasn't detected by linting
type mockIntAPIService struct {
	logger     logutil.LoggerInterface
	mockClient gemini.Client
}

// InitClient returns the mock client instead of creating a real one
func (s *mockIntAPIService) InitClient(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
	// Always return the mock client, ignoring the API key, model name, and API endpoint
	return s.mockClient, nil
}

// ProcessResponse processes the API response and extracts content
func (s *mockIntAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	// Check for nil result
	if result == nil {
		return "", fmt.Errorf("result is nil")
	}

	// Check for empty content
	if result.Content == "" {
		return "", fmt.Errorf("empty content")
	}

	// Get the original content - with the new architecture, we just return the content
	// directly without any template processing
	return result.Content, nil
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

// RunInternal is a replacement for the removed architect.RunInternal function
// It's used for backwards compatibility with the existing integration tests
func RunInternal(
	ctx context.Context,
	testConfig *config.CliConfig,
	logger logutil.LoggerInterface,
	apiService architect.APIService,
	auditLogger auditlog.AuditLogger,
) error {
	// Save the original NewAPIService function
	originalNewAPIService := architect.NewAPIService

	// Override the API Service creation function for this test to use the provided service
	architect.NewAPIService = func(logger logutil.LoggerInterface) architect.APIService {
		return apiService
	}

	// Restore the original NewAPIService when done
	defer func() {
		architect.NewAPIService = originalNewAPIService
	}()

	// Run the Execute function directly
	return architect.Execute(
		ctx,
		testConfig,
		logger,
		auditLogger,
	)
}

// RunTestWithConfig runs the architect application with the provided test config and environment
func RunTestWithConfig(
	ctx context.Context,
	testConfig *config.CliConfig,
	env *TestEnv,
) error {
	// Set up the original API Service for architect.Execute
	originalNewAPIService := architect.NewAPIService

	// Override the API Service creation function for this test
	architect.NewAPIService = func(logger logutil.LoggerInterface) architect.APIService {
		return &mockIntAPIService{
			logger:     env.Logger,
			mockClient: env.MockClient,
		}
	}

	// Restore the original API Service creation function when done
	defer func() {
		architect.NewAPIService = originalNewAPIService
	}()

	// Run the architect application using Execute directly
	return architect.Execute(
		ctx,
		testConfig,
		env.Logger,
		env.AuditLogger,
	)
}
