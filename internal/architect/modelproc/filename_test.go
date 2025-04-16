package modelproc_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/llm"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/registry"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"gemini-1.0-pro", "gemini-1.0-pro"},
		{"gemini/1.0-pro", "gemini-1.0-pro"},
		{"gemini\\1.0:pro", "gemini-1.0-pro"},
		{"gemini-1.0-pro*?\"<>|", "gemini-1.0-pro------"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			// Save original factory function and restore after test
			originalNewTokenManagerWithClient := modelproc.NewTokenManagerWithClient
			defer func() {
				modelproc.NewTokenManagerWithClient = originalNewTokenManagerWithClient
			}()

			// We're calling the unexported sanitizeFilename indirectly through Process
			// by creating a mock setup that lets us check the file path
			mockAPI := &mockAPIService{
				initLLMClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (llm.LLMClient, error) {
					return &mockLLMClient{
						generateContentFunc: func(ctx context.Context, prompt string, params map[string]interface{}) (*llm.ProviderResult, error) {
							return &llm.ProviderResult{
								Content:    "Generated content",
								TokenCount: 50,
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

			// Check if the sanitized filename is used in the output path
			mockWriter := &mockFileWriter{
				saveToFileFunc: func(content, outputFile string) error {
					fileName := filepath.Base(outputFile)
					expectedFileName := test.expected + ".md"
					if fileName != expectedFileName {
						t.Errorf("Expected filename '%s', got '%s'", expectedFileName, fileName)
					}
					return nil
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

			// Run Process with the test input as the model name
			_ = processor.Process(
				context.Background(),
				test.input,
				"Test prompt",
			)
		})
	}
}
