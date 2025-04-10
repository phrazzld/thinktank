// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
	"github.com/phrazzld/architect/internal/prompt"
)

// mockTokenManager implements the TokenManager interface for testing
type mockTokenManager struct{}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
	return &TokenResult{
		TokenCount:   100,
		InputLimit:   1000,
		ExceedsLimit: false,
		Percentage:   10.0,
	}, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, confirmTokens int) bool {
	return true
}

// mockPromptManager implements the prompt.ManagerInterface for testing
type mockPromptManager struct{}

func (m *mockPromptManager) BuildPrompt(templateName string, data *prompt.TemplateData) (string, error) {
	return "mock prompt", nil
}

func (m *mockPromptManager) ListExampleTemplates() ([]string, error) {
	return []string{"example.tmpl"}, nil
}

func (m *mockPromptManager) GetExampleTemplate(name string) (string, error) {
	return "example template content", nil
}

// mockGeminiClient implements a simplified gemini.Client for testing
type mockGeminiClient struct{}

func (m *mockGeminiClient) GenerateContent(ctx context.Context, prompt string) (*gemini.GenerationResult, error) {
	return &gemini.GenerationResult{
		Content:    "generated content",
		TokenCount: 50,
	}, nil
}

func (m *mockGeminiClient) GetModelInfo(ctx context.Context) (*gemini.ModelInfo, error) {
	return &gemini.ModelInfo{
		InputTokenLimit: 1000,
	}, nil
}

func (m *mockGeminiClient) CountTokens(ctx context.Context, prompt string) (*gemini.TokenCount, error) {
	return &gemini.TokenCount{Total: 100}, nil
}

func (m *mockGeminiClient) Close() {}

// mockConfigManager implements a simplified config.ManagerInterface for testing
type mockConfigManager struct{}

func (m *mockConfigManager) LoadFromFiles() error {
	return nil
}

func (m *mockConfigManager) EnsureConfigDirs() error {
	return nil
}

func (m *mockConfigManager) GetConfig() *config.AppConfig {
	return &config.AppConfig{}
}

func (m *mockConfigManager) MergeWithFlags(flags map[string]interface{}) error {
	return nil
}

func (m *mockConfigManager) GetConfigDirs() config.ConfigDirectories {
	return config.ConfigDirectories{}
}

// mockAPIService implements a simplified APIService for testing
type mockAPIService struct{}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	if result == nil {
		return "", errors.New("empty response")
	}
	return result.Content, nil
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return "mock error details"
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	return false
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	return false
}

// TestSaveToFile tests the SaveToFile method
func TestSaveToFile(t *testing.T) {
	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ", false)

	// Create a token manager for testing
	tokenManager := &mockTokenManager{}

	// Create an output writer
	outputWriter := NewOutputWriter(logger, tokenManager)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "output_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define test cases
	tests := []struct {
		name       string
		content    string
		outputFile string
		wantErr    bool
	}{
		{
			name:       "Valid file path",
			content:    "Test content",
			outputFile: filepath.Join(tempDir, "test_output.md"),
			wantErr:    false,
		},
		{
			name:       "Relative file path",
			content:    "Test content with relative path",
			outputFile: "test_output_relative.md",
			wantErr:    false,
		},
	}

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Save to file
			err := outputWriter.SaveToFile(tc.content, tc.outputFile)

			// Check error
			if (err != nil) != tc.wantErr {
				t.Errorf("SaveToFile() error = %v, wantErr %v", err, tc.wantErr)
				return
			}

			// Determine output path for validation
			outputPath := tc.outputFile
			if !filepath.IsAbs(outputPath) {
				cwd, _ := os.Getwd()
				outputPath = filepath.Join(cwd, outputPath)
				// Clean up relative path file after test
				defer os.Remove(outputPath)
			}

			// Verify file was created and content matches
			if !tc.wantErr {
				content, err := os.ReadFile(outputPath)
				if err != nil {
					t.Errorf("Failed to read output file: %v", err)
					return
				}

				if string(content) != tc.content {
					t.Errorf("File content = %v, want %v", string(content), tc.content)
				}
			}
		})
	}
}

// TestGenerateAndSavePlan tests the GenerateAndSavePlan method
func TestGenerateAndSavePlan(t *testing.T) {
	// Skip this test until implementation is complete
	t.Skip("Skipping until GenerateAndSavePlan is implemented")

	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ", false)

	// Create a token manager for testing
	tokenManager := &mockTokenManager{}

	// Create an output writer
	outputWriter := NewOutputWriter(logger, tokenManager)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "output_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock gemini client
	geminiClient := &mockGeminiClient{}

	// Create a mock prompt manager
	promptManager := &mockPromptManager{}

	// Define test parameters
	ctx := context.Background()
	taskDescription := "Test task"
	projectContext := "Test project context"
	outputFile := filepath.Join(tempDir, "test_plan.md")

	// Call the method being tested
	err = outputWriter.GenerateAndSavePlan(ctx, geminiClient, taskDescription, projectContext, outputFile, promptManager)

	// Verify the output
	if err != nil {
		t.Errorf("GenerateAndSavePlan() error = %v", err)
		return
	}

	// Verify the file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("output file not created: %v", err)
		return
	}

	// Read the file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	// Verify content
	if string(content) != "generated content" {
		t.Errorf("File content = %v, want %v", string(content), "generated content")
	}
}

// TestGenerateAndSavePlanWithConfig tests the GenerateAndSavePlanWithConfig method
func TestGenerateAndSavePlanWithConfig(t *testing.T) {
	// Skip this test until implementation is complete
	t.Skip("Skipping until GenerateAndSavePlanWithConfig is implemented")

	// Create a logger for testing
	logger := logutil.NewLogger(logutil.InfoLevel, os.Stderr, "[test] ", false)

	// Create a token manager for testing
	tokenManager := &mockTokenManager{}

	// Create an output writer
	outputWriter := NewOutputWriter(logger, tokenManager)

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "output_test")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock gemini client
	geminiClient := &mockGeminiClient{}

	// Create a mock config manager
	configManager := &mockConfigManager{}

	// Define test parameters
	ctx := context.Background()
	taskDescription := "Test task"
	projectContext := "Test project context"
	outputFile := filepath.Join(tempDir, "test_plan.md")

	// Call the method being tested
	err = outputWriter.GenerateAndSavePlanWithConfig(ctx, geminiClient, taskDescription, projectContext, outputFile, configManager)

	// Verify the output
	if err != nil {
		t.Errorf("GenerateAndSavePlanWithConfig() error = %v", err)
		return
	}

	// Verify the file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Errorf("output file not created: %v", err)
		return
	}

	// Read the file content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Errorf("Failed to read output file: %v", err)
		return
	}

	// Verify content
	if string(content) != "generated content" {
		t.Errorf("File content = %v, want %v", string(content), "generated content")
	}
}
