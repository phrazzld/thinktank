package modelproc_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/llm"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/registry"
	"github.com/phrazzld/thinktank/internal/thinktank/modelproc"
)

// Store the original factory function
var originalNewTokenManagerWithClient = modelproc.NewTokenManagerWithClient

// Helper function to restore the original factory
func restoreNewTokenManagerWithClient() {
	modelproc.NewTokenManagerWithClient = originalNewTokenManagerWithClient
}

// No-op logger for tests
type noOpLogger struct{}

func (l *noOpLogger) Debug(format string, args ...interface{})  {}
func (l *noOpLogger) Info(format string, args ...interface{})   {}
func (l *noOpLogger) Warn(format string, args ...interface{})   {}
func (l *noOpLogger) Error(format string, args ...interface{})  {}
func (l *noOpLogger) Fatal(format string, args ...interface{})  {}
func (l *noOpLogger) Println(args ...interface{})               {}
func (l *noOpLogger) Printf(format string, args ...interface{}) {}

func newNoOpLogger() logutil.LoggerInterface {
	return &noOpLogger{}
}

func TestModelProcessor_Process_ClientInitError(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Setup mocks
	expectedErr := errors.New("client init error")
	mockAPI := &mockAPIService{
		// Now we use the initLLMClient function for the provider-agnostic interface
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return nil, expectedErr
		},
	}

	// Create a mock token manager for factory function mocking
	mockToken := &mockTokenManager{}

	// Mock the factory function instead of injecting the token manager
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature (no token manager)
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for client initialization, got nil")
	} else if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestModelProcessor_Process_GenerationError(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Setup mocks
	expectedErr := errors.New("generation error")
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return nil, expectedErr
				},
			}, nil
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	// Mock the factory function instead of injecting the token manager
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for generation failure, got nil")
	} else if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestModelProcessor_Process_SaveError(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Setup mocks
	expectedErr := errors.New("save error")
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return &llm.ProviderResult{
						Content: "Generated content",
					}, nil
				},
			}, nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	// Mock the factory function
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{
		saveToFileFunc: func(content, outputFile string) error {
			return expectedErr
		},
	}

	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for save failure, got nil")
	} else if !errors.Is(err, expectedErr) {
		t.Errorf("Expected error '%v', got '%v'", expectedErr, err)
	}
}

func TestModelProcessor_Process_TokenLimitExceeded(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Setup mocks
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				getModelNameFunc: func() string {
					return "test-model"
				},
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					// Return a provider error about token limits
					return nil, errors.New("token limit exceeded: provider error")
				},
			}, nil
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   5000,
				InputLimit:   4000,
				ExceedsLimit: true,
				LimitError:   "prompt exceeds token limit (5000 tokens > 4000 token limit)",
				Percentage:   125.0,
			}, nil
		},
	}

	// Mock the factory function
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify results
	if err == nil {
		t.Errorf("Expected error for token limit exceeded, got nil")
	} else if !strings.Contains(err.Error(), "token limit exceeded") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestProcess_ProcessResponseError tests error handling in response processing
func TestProcess_ProcessResponseError(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Create expected error
	expectedError := errors.New("response processing error")

	// Create mock API service that returns an error from ProcessLLMResponse
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return &llm.ProviderResult{
						Content: "Generated content",
					}, nil
				},
			}, nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return "", expectedError
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	// Mock the factory function
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify the error was returned
	if err == nil {
		t.Errorf("Expected error for response processing, got nil")
	} else if !errors.Is(err, expectedError) {
		t.Errorf("Expected error '%v', got '%v'", expectedError, err)
	}
}

// TestProcess_EmptyResponseError tests handling of empty response errors
func TestProcess_EmptyResponseError(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Create expected error
	generationErr := errors.New("empty response")

	// Create mock API service that identifies empty response errors
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return nil, generationErr
				},
			}, nil
		},
		isEmptyResponseErrorFunc: func(err error) bool {
			return true // Report that this is an empty response error
		},
		getErrorDetailsFunc: func(err error) string {
			return "Empty response error details"
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	// Mock the factory function
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify the error was returned and contains appropriate context
	if err == nil {
		t.Errorf("Expected error for empty response, got nil")
	} else if !strings.Contains(err.Error(), "empty response") {
		t.Errorf("Expected error to mention empty response, got: %v", err)
	}
}

// TestProcess_SafetyBlockedError tests handling of safety blocked errors
func TestProcess_SafetyBlockedError(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Create expected error
	generationErr := errors.New("safety blocked")

	// Create mock API service that identifies safety blocked errors
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					return nil, generationErr
				},
			}, nil
		},
		isEmptyResponseErrorFunc: func(err error) bool {
			return false
		},
		isSafetyBlockedErrorFunc: func(err error) bool {
			return true // Report that this is a safety blocked error
		},
		getErrorDetailsFunc: func(err error) string {
			return "Content blocked due to safety concerns"
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{
		getTokenInfoFunc: func(ctx context.Context, prompt string) (*modelproc.TokenResult, error) {
			return &modelproc.TokenResult{
				TokenCount:   100,
				InputLimit:   1000,
				ExceedsLimit: false,
				Percentage:   10.0,
			}, nil
		},
	}

	// Mock the factory function
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify the error was returned and contains appropriate context
	if err == nil {
		t.Errorf("Expected error for safety blocked, got nil")
	} else if !strings.Contains(err.Error(), "safety") {
		t.Errorf("Expected error to mention safety concerns, got: %v", err)
	}
}

// TestProcess_NilClientDeference tests the behavior when a nil client is returned
// This test specifically targets the bugfix for the nil pointer dereference issue
func TestProcess_NilClientDeference(t *testing.T) {
	// Save original factory function and restore after test
	defer restoreNewTokenManagerWithClient()

	// Setup mocks with a failing client initialization
	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return nil, errors.New("client initialization failure")
		},
		getErrorDetailsFunc: func(err error) string {
			return "error details"
		},
	}

	// Create mock token manager
	mockToken := &mockTokenManager{}

	// Mock the factory function
	modelproc.NewTokenManagerWithClient = func(logger logutil.LoggerInterface, auditLogger auditlog.AuditLogger, client llm.LLMClient, reg *registry.Registry) modelproc.TokenManager {
		return mockToken
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor with updated constructor signature
	processor := modelproc.NewProcessor(
		mockAPI,
		mockWriter,
		mockAudit,
		mockLogger,
		cfg,
	)

	// Run test - this should not panic due to nil pointer dereference in defer
	err := processor.Process(
		context.Background(),
		"test-model",
		"Test prompt",
	)

	// Verify an error was returned but no panic occurred
	if err == nil {
		t.Errorf("Expected error for client initialization failure, got nil")
	}
}
