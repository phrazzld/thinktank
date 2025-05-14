package orchestrator

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/testutil"
	"github.com/phrazzld/thinktank/internal/thinktank/modelproc"
)

// mockFileWriter implements the interfaces.FileWriter interface for testing
type mockFileWriter struct {
	savedFiles map[string]string
	failPath   string
	failErr    error
}

func newMockFileWriter() *mockFileWriter {
	return &mockFileWriter{
		savedFiles: make(map[string]string),
	}
}

// SetupFailure configures the mock to fail when saving to a specific path
func (m *mockFileWriter) SetupFailure(path string, err error) {
	m.failPath = path
	m.failErr = err
}

// SaveToFile implements the interfaces.FileWriter interface
func (m *mockFileWriter) SaveToFile(content, path string) error {
	if path == m.failPath && m.failErr != nil {
		return m.failErr
	}
	m.savedFiles[path] = content
	return nil
}

// TestDefaultOutputWriter_SaveIndividualOutputs tests the SaveIndividualOutputs method
func TestDefaultOutputWriter_SaveIndividualOutputs(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name          string
		modelOutputs  map[string]string
		outputDir     string
		failPath      string
		failErr       error
		expectedCount int
		expectedError bool
	}{
		{
			name: "Success - All files saved successfully",
			modelOutputs: map[string]string{
				"model1": "content1",
				"model2": "content2",
				"model3": "content3",
			},
			outputDir:     "/test/output",
			expectedCount: 3,
			expectedError: false,
		},
		{
			name: "Partial failure - Some files fail to save",
			modelOutputs: map[string]string{
				"model1": "content1",
				"model2": "content2",
				"model3": "content3",
			},
			outputDir:     "/test/output",
			failPath:      "/test/output/model2.md",
			failErr:       errors.New("permission denied"),
			expectedCount: 2,
			expectedError: true,
		},
		{
			name:          "Empty model outputs - No files to save",
			modelOutputs:  map[string]string{},
			outputDir:     "/test/output",
			expectedCount: 0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock dependencies
			mockFileWriter := newMockFileWriter()
			if tt.failPath != "" {
				mockFileWriter.SetupFailure(tt.failPath, tt.failErr)
			}

			mockAuditLogger := auditlog.NewNoOpAuditLogger()
			mockLogger := testutil.NewMockLogger()

			// Create the output writer with mock dependencies
			writer := NewOutputWriter(mockFileWriter, mockAuditLogger, mockLogger)

			// Call the method under test
			ctx := context.Background()
			count, _, err := writer.SaveIndividualOutputs(ctx, tt.modelOutputs, tt.outputDir)

			// Verify the results
			if tt.expectedError && err == nil {
				t.Errorf("Expected an error but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if count != tt.expectedCount {
				t.Errorf("Expected %d files saved but got %d", tt.expectedCount, count)
			}

			// Verify each file was saved correctly (except the failing one)
			for modelName, content := range tt.modelOutputs {
				// Sanitize the model name as expected
				sanitizedModelName := modelproc.SanitizeFilename(modelName)
				expectedPath := filepath.Join(tt.outputDir, sanitizedModelName+".md")

				// Skip checking the file that was set up to fail
				if expectedPath == tt.failPath {
					continue
				}

				// Check if the file was saved with the correct content
				if savedContent, ok := mockFileWriter.savedFiles[expectedPath]; ok {
					if savedContent != content {
						t.Errorf("Content mismatch for model %s. Expected %q but got %q", modelName, content, savedContent)
					}
				} else {
					t.Errorf("Expected file at %s but it wasn't saved", expectedPath)
				}
			}
		})
	}
}

// TestDefaultOutputWriter_SaveSynthesisOutput tests the SaveSynthesisOutput method
func TestDefaultOutputWriter_SaveSynthesisOutput(t *testing.T) {
	// Setup test cases
	tests := []struct {
		name          string
		content       string
		modelName     string
		outputDir     string
		failPath      string
		failErr       error
		expectedError bool
	}{
		{
			name:          "Success - Synthesis output saved successfully",
			content:       "Synthesis content",
			modelName:     "synthesis-model",
			outputDir:     "/test/output",
			expectedError: false,
		},
		{
			name:          "Failure - Error saving synthesis output",
			content:       "Synthesis content",
			modelName:     "synthesis-model",
			outputDir:     "/test/output",
			failPath:      "/test/output/synthesis-model-synthesis.md",
			failErr:       errors.New("disk full"),
			expectedError: true,
		},
		{
			name:          "Edge case - Empty content",
			content:       "",
			modelName:     "synthesis-model",
			outputDir:     "/test/output",
			expectedError: false,
		},
		{
			name:          "Edge case - Model name with special characters",
			content:       "Special character model",
			modelName:     "model/with:special*chars?",
			outputDir:     "/test/output",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock dependencies
			mockFileWriter := newMockFileWriter()
			if tt.failPath != "" {
				mockFileWriter.SetupFailure(tt.failPath, tt.failErr)
			}

			mockAuditLogger := auditlog.NewNoOpAuditLogger()
			mockLogger := testutil.NewMockLogger()

			// Create the output writer with mock dependencies
			writer := NewOutputWriter(mockFileWriter, mockAuditLogger, mockLogger)

			// Call the method under test
			ctx := context.Background()
			_, err := writer.SaveSynthesisOutput(ctx, tt.content, tt.modelName, tt.outputDir)

			// Verify the results
			if tt.expectedError && err == nil {
				t.Errorf("Expected an error but got nil")
			}
			if !tt.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify the correct file path and content if no error is expected
			if !tt.expectedError {
				// Sanitize the model name as expected
				sanitizedModelName := modelproc.SanitizeFilename(tt.modelName)
				expectedPath := filepath.Join(tt.outputDir, sanitizedModelName+"-synthesis.md")

				// Check if the file was saved with the correct content
				if savedContent, ok := mockFileWriter.savedFiles[expectedPath]; ok {
					if savedContent != tt.content {
						t.Errorf("Content mismatch for synthesis output. Expected %q but got %q", tt.content, savedContent)
					}
				} else {
					t.Errorf("Expected synthesis file at %s but it wasn't saved", expectedPath)
				}
			}
		})
	}
}

// TestSanitizeFilename tests the SanitizeFilename function
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "normal-model-name",
			expected: "normal-model-name",
		},
		{
			input:    "model/with/slashes",
			expected: "model-with-slashes",
		},
		{
			input:    "model:with:colons",
			expected: "model-with-colons",
		},
		{
			input:    "model*with*stars",
			expected: "model-with-stars",
		},
		{
			input:    "model?with?questions",
			expected: "model-with-questions",
		},
		{
			input:    "model\"with\"quotes",
			expected: "model-with-quotes",
		},
		{
			input:    "model<with>brackets",
			expected: "model-with-brackets",
		},
		{
			input:    "model|with|pipes",
			expected: "model-with-pipes",
		},
		{
			input:    "model with spaces",
			expected: "model_with_spaces",
		},
		{
			input:    "model'with'apostrophes",
			expected: "model-with-apostrophes",
		},
		{
			input:    "model.with.dots",
			expected: "model.with.dots", // dots are allowed in filenames
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := modelproc.SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
