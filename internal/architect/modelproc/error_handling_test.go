package modelproc_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
)

func TestModelProcessor_Process_ClientInitError(t *testing.T) {
	// Setup mocks
	expectedErr := errors.New("client init error")
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return nil, expectedErr
		},
	}

	mockToken := &mockTokenManager{}
	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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
	// Setup mocks
	expectedErr := errors.New("generation error")
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return nil, expectedErr
				},
			}, nil
		},
	}

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

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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
	// Setup mocks
	expectedErr := errors.New("save error")
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return &gemini.GenerationResult{
						Content:    "Generated content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
			return result.Content, nil
		},
	}

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

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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
	// Setup mocks
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				getModelInfoFunc: func(ctx context.Context) (*gemini.ModelInfo, error) {
					return &gemini.ModelInfo{
						Name:             "test-model",
						InputTokenLimit:  4000,
						OutputTokenLimit: 1000,
					}, nil
				},
				countTokensFunc: func(ctx context.Context, text string) (*gemini.TokenCount, error) {
					return &gemini.TokenCount{Total: 5000}, nil
				},
				getModelNameFunc: func() string {
					return "test-model"
				},
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					// This should not be called in the token exceeded case
					return nil, errors.New("should not be called")
				},
			}, nil
		},
	}
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

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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
	} else if err.Error() != "token limit exceeded for model test-model: prompt exceeds token limit (5000 tokens > 4000 token limit)" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestProcess_ProcessResponseError tests error handling in response processing
func TestProcess_ProcessResponseError(t *testing.T) {
	// Create expected error
	expectedError := errors.New("response processing error")

	// Create mock API service that returns an error from ProcessResponse
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
					return &gemini.GenerationResult{
						Content:    "Generated content",
						TokenCount: 50,
					}, nil
				},
			}, nil
		},
		processResponseFunc: func(result *gemini.GenerationResult) (string, error) {
			return "", expectedError
		},
	}

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

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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
	// Create expected error
	generationErr := errors.New("empty response")

	// Create mock API service that identifies empty response errors
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
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

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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
	// Create expected error
	generationErr := errors.New("safety blocked")

	// Create mock API service that identifies safety blocked errors
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return &mockClient{
				generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
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

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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
	// Setup mocks with a failing client initialization
	mockAPI := &mockAPIService{
		initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
			return nil, errors.New("client initialization failure")
		},
		getErrorDetailsFunc: func(err error) string {
			return "error details"
		},
	}

	mockToken := &mockTokenManager{}
	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	// Setup config
	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	// Create processor
	processor := modelproc.NewProcessor(
		mockAPI,
		mockToken,
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