package modelproc_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/architect/modelproc"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
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
			// We're calling the unexported sanitizeFilename indirectly through Process
			// by creating a mock setup that lets us check the file path
			mockAPI := &mockAPIService{
				initClientFunc: func(ctx context.Context, apiKey, modelName, apiEndpoint string) (gemini.Client, error) {
					return &mockClient{
						generateContentFunc: func(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
							return &gemini.GenerationResult{
								Content: "Generated content",
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

			// Create processor
			processor := modelproc.NewProcessor(
				mockAPI,
				mockToken,
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