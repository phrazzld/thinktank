package modelproc_test

import (
	"context"
	"errors"
	"testing"

	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/llm"
	"github.com/misty-step/thinktank/internal/thinktank/modelproc"
)

// TestProcess_GetModelParametersError tests handling when GetModelParameters fails
func TestProcess_GetModelParametersError(t *testing.T) {
	expectedParamError := errors.New("failed to get model parameters")

	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					// Verify empty parameters are passed when GetModelParameters fails
					if len(params) != 0 {
						t.Errorf("Expected empty parameters when GetModelParameters fails, got %d parameters", len(params))
					}
					return &llm.ProviderResult{
						Content: "Generated content",
					}, nil
				},
			}, nil
		},
		getModelParametersFunc: func(ctx context.Context, modelName string) (map[string]interface{}, error) {
			return nil, expectedParamError
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	processor := modelproc.NewProcessor(mockAPI, mockWriter, mockAudit, mockLogger, cfg)

	// Run test - should succeed despite parameter error
	output, err := processor.Process(context.Background(), "test-model", "Test prompt")

	// Verify success with fallback to empty parameters
	if err != nil {
		t.Errorf("Expected success with parameter fallback, got error: %v", err)
	}
	if output != "Generated content" {
		t.Errorf("Expected output 'Generated content', got: %s", output)
	}
}

// TestProcess_EmptyResponseError_Comprehensive tests handling of empty response errors with detailed verification
func TestProcess_EmptyResponseError_Comprehensive(t *testing.T) {
	emptyResponseErr := errors.New("empty response")

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
			return "", emptyResponseErr
		},
		isEmptyResponseErrorFunc: func(err error) bool {
			return err == emptyResponseErr
		},
		getErrorDetailsFunc: func(err error) string {
			return "empty response details"
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	processor := modelproc.NewProcessor(mockAPI, mockWriter, mockAudit, mockLogger, cfg)

	// Run test
	output, err := processor.Process(context.Background(), "test-model", "Test prompt")

	// Verify error handling
	if err == nil {
		t.Errorf("Expected error for empty response, got nil")
	} else if !errors.Is(err, modelproc.ErrEmptyModelResponse) {
		t.Errorf("Expected error to be ErrEmptyModelResponse, got '%v'", err)
	}

	if output != "" {
		t.Errorf("Expected empty output on error, got: %s", output)
	}
}

// TestProcess_SafetyBlockedError_Comprehensive tests handling of safety blocked errors with detailed verification
func TestProcess_SafetyBlockedError_Comprehensive(t *testing.T) {
	safetyBlockedErr := errors.New("content blocked by safety filters")

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
			return "", safetyBlockedErr
		},
		isEmptyResponseErrorFunc: func(err error) bool {
			return false
		},
		isSafetyBlockedErrorFunc: func(err error) bool {
			return err == safetyBlockedErr
		},
		getErrorDetailsFunc: func(err error) string {
			return "safety blocked details"
		},
	}

	mockWriter := &mockFileWriter{}
	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	processor := modelproc.NewProcessor(mockAPI, mockWriter, mockAudit, mockLogger, cfg)

	// Run test
	output, err := processor.Process(context.Background(), "test-model", "Test prompt")

	// Verify error handling
	if err == nil {
		t.Errorf("Expected error for safety blocked response, got nil")
	} else if !errors.Is(err, modelproc.ErrContentFiltered) {
		t.Errorf("Expected error to be ErrContentFiltered, got '%v'", err)
	}

	if output != "" {
		t.Errorf("Expected empty output on error, got: %s", output)
	}
}

// TestProcess_CategorizedErrors tests handling of different categorized errors
func TestProcess_CategorizedErrors(t *testing.T) {
	testCases := []struct {
		name          string
		category      llm.ErrorCategory
		expectedError error
	}{
		{
			name:          "content filtered",
			category:      llm.CategoryContentFiltered,
			expectedError: modelproc.ErrContentFiltered,
		},
		{
			name:          "rate limit",
			category:      llm.CategoryRateLimit,
			expectedError: modelproc.ErrModelRateLimited,
		},
		{
			name:          "input limit",
			category:      llm.CategoryInputLimit,
			expectedError: modelproc.ErrModelTokenLimitExceeded,
		},
		{
			name:          "server error",
			category:      llm.CategoryServer,
			expectedError: modelproc.ErrInvalidModelResponse,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			categorizedErr := &llm.LLMError{
				Message:       "test error",
				ErrorCategory: tc.category,
			}

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
					return "", categorizedErr
				},
				isEmptyResponseErrorFunc: func(err error) bool {
					return false
				},
				isSafetyBlockedErrorFunc: func(err error) bool {
					return false
				},
				getErrorDetailsFunc: func(err error) string {
					return "categorized error details"
				},
			}

			mockWriter := &mockFileWriter{}
			mockAudit := &mockAuditLogger{}
			mockLogger := newNoOpLogger()

			cfg := config.NewDefaultCliConfig()
			cfg.APIKey = "test-api-key"
			cfg.OutputDir = "/tmp/test-output"

			processor := modelproc.NewProcessor(mockAPI, mockWriter, mockAudit, mockLogger, cfg)

			// Run test
			output, err := processor.Process(context.Background(), "test-model", "Test prompt")

			// Verify error handling
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tc.name)
			} else if !errors.Is(err, tc.expectedError) {
				t.Errorf("Expected error to be %v, got '%v'", tc.expectedError, err)
			}

			if output != "" {
				t.Errorf("Expected empty output on error, got: %s", output)
			}
		})
	}
}

// TestProcess_FileWriteError tests error handling in saveOutputToFile
func TestProcess_FileWriteError(t *testing.T) {
	writeErr := errors.New("file write error")

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

	mockWriter := &mockFileWriter{
		saveToFileFunc: func(ctx context.Context, content, outputFile string) error {
			return writeErr
		},
	}

	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	processor := modelproc.NewProcessor(mockAPI, mockWriter, mockAudit, mockLogger, cfg)

	// Run test
	output, err := processor.Process(context.Background(), "test-model", "Test prompt")

	// Verify error handling
	if err == nil {
		t.Errorf("Expected error for file write failure, got nil")
	} else if !errors.Is(err, modelproc.ErrOutputWriteFailed) {
		t.Errorf("Expected error to be ErrOutputWriteFailed, got '%v'", err)
	}

	if output != "" {
		t.Errorf("Expected empty output on error, got: %s", output)
	}
}

// TestProcess_SuccessfulExecution tests the complete successful execution path
func TestProcess_SuccessfulExecution(t *testing.T) {
	expectedContent := "Successfully generated content"
	expectedParams := map[string]interface{}{
		"temperature": 0.7,
		"max_tokens":  1000,
	}

	mockAPI := &mockAPIService{
		initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
			return &mockLLMClient{
				generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
					// Verify parameters are passed correctly
					if len(params) != len(expectedParams) {
						t.Errorf("Expected %d parameters, got %d", len(expectedParams), len(params))
					}
					return &llm.ProviderResult{
						Content:      expectedContent,
						FinishReason: "stop",
					}, nil
				},
			}, nil
		},
		getModelParametersFunc: func(ctx context.Context, modelName string) (map[string]interface{}, error) {
			return expectedParams, nil
		},
		processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
			return result.Content, nil
		},
	}

	var savedContent, savedPath string
	mockWriter := &mockFileWriter{
		saveToFileFunc: func(ctx context.Context, content, outputFile string) error {
			savedContent = content
			savedPath = outputFile
			return nil
		},
	}

	mockAudit := &mockAuditLogger{}
	mockLogger := newNoOpLogger()

	cfg := config.NewDefaultCliConfig()
	cfg.APIKey = "test-api-key"
	cfg.OutputDir = "/tmp/test-output"

	processor := modelproc.NewProcessor(mockAPI, mockWriter, mockAudit, mockLogger, cfg)

	// Run test
	output, err := processor.Process(context.Background(), "test-model", "Test prompt")

	// Verify successful execution
	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}
	if output != expectedContent {
		t.Errorf("Expected output '%s', got: %s", expectedContent, output)
	}
	if savedContent != expectedContent {
		t.Errorf("Expected saved content '%s', got: %s", expectedContent, savedContent)
	}
	if savedPath != "/tmp/test-output/test-model.md" {
		t.Errorf("Expected saved path '/tmp/test-output/test-model.md', got: %s", savedPath)
	}
}

// TestProcess_ModelNameSanitization tests filename sanitization for various model names
func TestProcess_ModelNameSanitization(t *testing.T) {
	testCases := []struct {
		modelName    string
		expectedFile string
	}{
		{
			modelName:    "gpt-4",
			expectedFile: "/tmp/test-output/gpt-4.md",
		},
		{
			modelName:    "gpt/3.5/turbo",
			expectedFile: "/tmp/test-output/gpt-3.5-turbo.md",
		},
		{
			modelName:    "claude:v1",
			expectedFile: "/tmp/test-output/claude-v1.md",
		},
		{
			modelName:    "gemini pro",
			expectedFile: "/tmp/test-output/gemini_pro.md",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.modelName, func(t *testing.T) {
			mockAPI := &mockAPIService{
				initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
					return &mockLLMClient{
						generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
							return &llm.ProviderResult{
								Content: "Test content",
							}, nil
						},
					}, nil
				},
				processLLMResponseFunc: func(result *llm.ProviderResult) (string, error) {
					return result.Content, nil
				},
			}

			var savedPath string
			mockWriter := &mockFileWriter{
				saveToFileFunc: func(ctx context.Context, content, outputFile string) error {
					savedPath = outputFile
					return nil
				},
			}

			mockAudit := &mockAuditLogger{}
			mockLogger := newNoOpLogger()

			cfg := config.NewDefaultCliConfig()
			cfg.APIKey = "test-api-key"
			cfg.OutputDir = "/tmp/test-output"

			processor := modelproc.NewProcessor(mockAPI, mockWriter, mockAudit, mockLogger, cfg)

			// Run test
			_, err := processor.Process(context.Background(), tc.modelName, "Test prompt")

			// Verify sanitization
			if err != nil {
				t.Errorf("Expected success, got error: %v", err)
			}
			if savedPath != tc.expectedFile {
				t.Errorf("Expected sanitized path '%s', got: %s", tc.expectedFile, savedPath)
			}
		})
	}
}
