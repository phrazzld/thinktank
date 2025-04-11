// internal/integration/test_runner.go
package integration

import (
	"context"
	"fmt"

	"github.com/phrazzld/architect/internal/architect"
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
func (s *mockIntAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	// Always return the mock client, ignoring the API key and model name
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

// mockConfigManager implements a simple version of config.ManagerInterface for testing
type mockConfigManager struct {
	logger logutil.LoggerInterface
	config *config.AppConfig
}

func newMockConfigManager(logger logutil.LoggerInterface) *mockConfigManager {
	return &mockConfigManager{
		logger: logger,
		config: config.DefaultConfig(),
	}
}

// Implement the minimal required interface methods
func (m *mockConfigManager) GetConfig() *config.AppConfig {
	return m.config
}

func (m *mockConfigManager) GetUserConfigDir() string {
	return "/mock/user/config/dir"
}

func (m *mockConfigManager) GetSystemConfigDirs() []string {
	return []string{"/mock/system/config/dir"}
}

func (m *mockConfigManager) GetConfigDirs() config.ConfigDirectories {
	return config.ConfigDirectories{
		UserConfigDir:    m.GetUserConfigDir(),
		SystemConfigDirs: m.GetSystemConfigDirs(),
	}
}

func (m *mockConfigManager) LoadFromFiles() error {
	return nil
}

func (m *mockConfigManager) MergeWithFlags(cliFlags map[string]interface{}) error {
	return nil
}

func (m *mockConfigManager) EnsureConfigDirs() error {
	return nil
}

func (m *mockConfigManager) WriteDefaultConfig() error {
	return nil
}

// GetUserTemplateDir returns a mock path (this is a legacy method that will be removed)
func (m *mockConfigManager) GetUserTemplateDir() string {
	return "/mock/user/template/dir"
}

// GetSystemTemplateDirs returns mock paths (this is a legacy method that will be removed)
func (m *mockConfigManager) GetSystemTemplateDirs() []string {
	return []string{"/mock/system/template/dir"}
}

// GetTemplatePath is a legacy method that will be removed
func (m *mockConfigManager) GetTemplatePath(name string) (string, error) {
	return "", fmt.Errorf("templates not supported in new architecture")
}

// RunTestWithConfig runs the architect application with the provided test config and environment
func RunTestWithConfig(
	ctx context.Context,
	testConfig *architect.CliConfig,
	env *TestEnv,
) error {
	// Use a mock config manager for testing
	configManager := newMockConfigManager(env.Logger)

	// Create a mock API service that uses the test environment's mock client
	mockAPIService := &mockIntAPIService{
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
