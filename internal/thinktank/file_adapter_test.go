// Package thinktank contains the core application logic for the thinktank tool.
// This file tests the FileWriterAdapter implementation.
package thinktank

import (
	"errors"
	"strings"
	"testing"
)

// TestFileWriterAdapter_SaveToFile tests the SaveToFile method of the FileWriterAdapter
func TestFileWriterAdapter_SaveToFile(t *testing.T) {
	// Test constants
	const (
		testContent    = "This is test content to be saved to a file."
		testOutputFile = "/path/to/output.md"
	)

	// Test cases
	tests := []struct {
		name           string
		mockSetup      func(mock *MockFileWriterForAdapter)
		content        string
		outputFile     string
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "success case - correctly delegates to the underlying writer",
			mockSetup: func(mock *MockFileWriterForAdapter) {
				// Track what content and outputFile were passed to verify delegation
				var capturedContent, capturedOutputFile string

				mock.SaveToFileFunc = func(content, outputFile string) error {
					// Capture the parameters for later verification
					capturedContent = content
					capturedOutputFile = outputFile

					// Return no error
					return nil
				}

				// Verify after the function call that parameters were passed through
				t.Cleanup(func() {
					if capturedContent != testContent {
						t.Errorf("Expected content to be %s, got %s", testContent, capturedContent)
					}
					if capturedOutputFile != testOutputFile {
						t.Errorf("Expected outputFile to be %s, got %s", testOutputFile, capturedOutputFile)
					}
				})
			},
			content:       testContent,
			outputFile:    testOutputFile,
			expectedError: false,
		},
		{
			name: "error case - handles error from underlying service",
			mockSetup: func(mock *MockFileWriterForAdapter) {
				mock.SaveToFileFunc = func(content, outputFile string) error {
					return errors.New("failed to save file")
				}
			},
			content:        testContent,
			outputFile:     testOutputFile,
			expectedError:  true,
			expectedErrMsg: "failed to save file",
		},
		{
			name: "nil file writer - panics",
			mockSetup: func(mock *MockFileWriterForAdapter) {
				// No setup needed for nil test
			},
			content:        testContent,
			outputFile:     testOutputFile,
			expectedError:  true,
			expectedErrMsg: "nil FileWriter",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *FileWriterAdapter

			// For the nil FileWriter test
			if tc.name == "nil file writer - panics" {
				// Create adapter with nil FileWriter - should panic
				adapter = &FileWriterAdapter{
					FileWriter: nil,
				}

				// Call should panic, recover and mark as error
				defer func() {
					if r := recover(); r != nil {
						// Expected panic, test passed
					} else {
						t.Error("Expected a panic but none occurred")
					}
				}()

				// This should panic
				_ = adapter.SaveToFile(tc.content, tc.outputFile)
				return
			}

			// Create mock for non-nil test cases
			mockFileWriter := &MockFileWriterForAdapter{}

			// Setup the mock
			tc.mockSetup(mockFileWriter)

			// Create adapter with mock
			adapter = &FileWriterAdapter{
				FileWriter: mockFileWriter,
			}

			// Call the method being tested
			err := adapter.SaveToFile(tc.content, tc.outputFile)

			// Check error expectation
			if tc.expectedError && err == nil {
				t.Error("Expected an error but got nil")
			} else if !tc.expectedError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Check error message if applicable
			if tc.expectedError && err != nil && tc.expectedErrMsg != "" {
				if !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("Expected error message to contain '%s', got: '%s'", tc.expectedErrMsg, err.Error())
				}
			}
		})
	}
}

// MockFileWriterForAdapter is a testing mock for the FileWriter interface, specifically for adapter tests
type MockFileWriterForAdapter struct {
	SaveToFileFunc func(content, outputFile string) error
}

func (m *MockFileWriterForAdapter) SaveToFile(content, outputFile string) error {
	if m.SaveToFileFunc != nil {
		return m.SaveToFileFunc(content, outputFile)
	}
	return errors.New("SaveToFile not implemented")
}
