// Package architect provides the command-line interface for the architect tool
package architect

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/gemini"
	"github.com/phrazzld/architect/internal/logutil"
)

// mockTokenManager implements the TokenManager interface for testing
type mockTokenManager struct{}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*TokenResult, error) {
	return &TokenResult{}, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, confirmTokens int) bool {
	return true
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