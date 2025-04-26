package orchestrator

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

// MockFileWriter that tracks saved files and can simulate failures
type TestFileWriter struct {
	savedFiles        map[string]string
	failingFilePaths  []string
	failWithError     error
	savesToFailCount  int // Number of saves that should fail (in order)
	currentSaveCount  int
	alwaysFail        bool
	specificFailPaths map[string]error // Map of specific paths that should fail with specific errors
}

// SaveToFile implements the FileWriter interface
func (m *TestFileWriter) SaveToFile(content, outputFile string) error {
	// Check if we should always fail
	if m.alwaysFail {
		return m.failWithError
	}

	// Check if this is a failing path specifically
	if err, ok := m.specificFailPaths[outputFile]; ok {
		return err
	}

	// Check if path contains any of the failing patterns
	for _, failPath := range m.failingFilePaths {
		if filepath.Base(outputFile) == failPath {
			return m.failWithError
		}
	}

	// Check if we should fail based on count
	m.currentSaveCount++
	if m.savesToFailCount > 0 && m.currentSaveCount <= m.savesToFailCount {
		return m.failWithError
	}

	// Otherwise, save the file
	if m.savedFiles == nil {
		m.savedFiles = make(map[string]string)
	}
	m.savedFiles[outputFile] = content
	return nil
}

// newTestFileWriter creates a mock file writer for testing
func newTestFileWriter() *TestFileWriter {
	return &TestFileWriter{
		savedFiles:        make(map[string]string),
		failingFilePaths:  []string{},
		failWithError:     errors.New("simulated file save error"),
		specificFailPaths: make(map[string]error),
	}
}

// TestSaveIndividualOutputs tests the SaveIndividualOutputs method
func TestSaveIndividualOutputs(t *testing.T) {
	tests := []struct {
		name               string
		modelOutputs       map[string]string
		outputDir          string
		expectedSaveCount  int
		expectedError      bool
		fileWriterSetupFn  func(*TestFileWriter)
		expectedErrorMatch string
	}{
		{
			name: "All files save successfully",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
				"model2": "Output from model2",
			},
			outputDir:         "/tmp/test",
			expectedSaveCount: 2,
			expectedError:     false,
			fileWriterSetupFn: func(fw *TestFileWriter) {},
		},
		{
			name: "Some files fail to save",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
				"model2": "Output from model2",
				"model3": "Output from model3",
			},
			outputDir:         "/tmp/test",
			expectedSaveCount: 1,
			expectedError:     true,
			fileWriterSetupFn: func(fw *TestFileWriter) {
				fw.savesToFailCount = 2 // First two saves will fail
			},
			expectedErrorMatch: "2/3 files failed to save",
		},
		{
			name: "All files fail to save",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
				"model2": "Output from model2",
			},
			outputDir:         "/tmp/test",
			expectedSaveCount: 0,
			expectedError:     true,
			fileWriterSetupFn: func(fw *TestFileWriter) {
				fw.alwaysFail = true
			},
			expectedErrorMatch: "2/2 files failed to save",
		},
		{
			name: "Specific file fails to save",
			modelOutputs: map[string]string{
				"model1": "Output from model1",
				"model2": "Output from model2",
				"model3": "Output from model3",
			},
			outputDir:         "/tmp/test",
			expectedSaveCount: 2,
			expectedError:     true,
			fileWriterSetupFn: func(fw *TestFileWriter) {
				fw.failingFilePaths = []string{"model2.md"}
			},
			expectedErrorMatch: "1/3 files failed to save",
		},
		{
			name:               "Empty model outputs",
			modelOutputs:       map[string]string{},
			outputDir:          "/tmp/test",
			expectedSaveCount:  0,
			expectedError:      false,
			fileWriterSetupFn:  func(fw *TestFileWriter) {},
			expectedErrorMatch: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockFileWriter := newTestFileWriter()
			tt.fileWriterSetupFn(mockFileWriter)
			mockLogger := &MockLogger{}
			mockAuditLogger := &MockAuditLogger{}

			// Create output writer
			outputWriter := NewOutputWriter(mockFileWriter, mockAuditLogger, mockLogger)

			// Call SaveIndividualOutputs
			savedCount, err := outputWriter.SaveIndividualOutputs(context.Background(), tt.modelOutputs, tt.outputDir)

			// Verify results
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tt.expectedErrorMatch != "" && !strings.Contains(err.Error(), tt.expectedErrorMatch) {
					t.Errorf("Error message didn't contain expected text: got %q, want to contain %q", err.Error(), tt.expectedErrorMatch)
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check saved count
			if savedCount != tt.expectedSaveCount {
				t.Errorf("Expected saved count %d but got %d", tt.expectedSaveCount, savedCount)
			}

			// Check number of saved files
			if len(mockFileWriter.savedFiles) != tt.expectedSaveCount {
				t.Errorf("Expected %d saved files but got %d", tt.expectedSaveCount, len(mockFileWriter.savedFiles))
			}

			// Verify file paths and content for saved files
			for modelName, content := range tt.modelOutputs {
				expectedPath := filepath.Join(tt.outputDir, modelName+".md")
				if len(mockFileWriter.savedFiles) > 0 {
					// Only check files that were successfully saved
					if savedContent, ok := mockFileWriter.savedFiles[expectedPath]; ok {
						if savedContent != content {
							t.Errorf("Content mismatch for %s. Expected %q but got %q", modelName, content, savedContent)
						}
					}
				}
			}
		})
	}
}

// TestSaveSynthesisOutput tests the SaveSynthesisOutput method
func TestSaveSynthesisOutput(t *testing.T) {
	tests := []struct {
		name               string
		content            string
		modelName          string
		outputDir          string
		expectedError      bool
		fileWriterSetupFn  func(*TestFileWriter)
		expectedErrorMatch string
	}{
		{
			name:          "Save synthesis output successfully",
			content:       "Synthesized output content",
			modelName:     "synth-model",
			outputDir:     "/tmp/test",
			expectedError: false,
			fileWriterSetupFn: func(fw *TestFileWriter) {
				// No failures
			},
		},
		{
			name:          "Failed to save synthesis output",
			content:       "Synthesized output content",
			modelName:     "synth-model",
			outputDir:     "/tmp/test",
			expectedError: true,
			fileWriterSetupFn: func(fw *TestFileWriter) {
				fw.alwaysFail = true
			},
			// Just look for the start of the expected error message
			expectedErrorMatch: "failed to save synthesis output",
		},
		{
			name:          "Model name with special characters",
			content:       "Synthesized output content",
			modelName:     "synth/model:v1*?",
			outputDir:     "/tmp/test",
			expectedError: false,
			fileWriterSetupFn: func(fw *TestFileWriter) {
				// No failures
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockFileWriter := newTestFileWriter()
			tt.fileWriterSetupFn(mockFileWriter)
			mockLogger := &MockLogger{}
			mockAuditLogger := &MockAuditLogger{}

			// Create output writer
			outputWriter := NewOutputWriter(mockFileWriter, mockAuditLogger, mockLogger)

			// Call SaveSynthesisOutput
			err := outputWriter.SaveSynthesisOutput(context.Background(), tt.content, tt.modelName, tt.outputDir)

			// Verify results
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected an error but got nil")
				} else if tt.expectedErrorMatch != "" && !strings.Contains(err.Error(), tt.expectedErrorMatch) {
					t.Errorf("Error message didn't contain expected text: got %q, want to contain %q", err.Error(), tt.expectedErrorMatch)
				}
			} else if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Verify the correct file path and content if no error is expected
			if !tt.expectedError {
				// Sanitize the model name as expected
				sanitizedModelName := sanitizeFilename(tt.modelName)
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

// TestSanitizeFilename tests the sanitizeFilename function
func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"with spaces", "with_spaces"},
		{"with/slashes", "with-slashes"},
		{"with\\backslashes", "with-backslashes"},
		{"with:colon", "with-colon"},
		{"with*asterisk", "with-asterisk"},
		{"with?question", "with-question"},
		{"with\"quote", "with-quote"},
		{"with<angle>brackets", "with-angle-brackets"},
		{"with|pipe", "with-pipe"},
		{"combo/of:all*the?\"'<special>chars|and spaces", "combo-of-all-the----special-chars-and_spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
