// Package architect contains the core application logic for the architect tool.
// This file tests the ContextGathererAdapter implementation.
package architect

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/phrazzld/architect/internal/architect/interfaces"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/logutil"
)

// TestContextGathererAdapter_GatherContext tests the GatherContext method of the ContextGathererAdapter
func TestContextGathererAdapter_GatherContext(t *testing.T) {
	// Create test context
	ctx := context.Background()

	// Sample test files
	testFiles := []fileutil.FileMeta{
		{
			Path:    "/path/to/file1.go",
			Content: "package main\n\nfunc main() {\n\tfmt.Println(\"Hello, World!\")\n}\n",
		},
		{
			Path:    "/path/to/file2.go",
			Content: "package test\n\nfunc Test() bool {\n\treturn true\n}\n",
		},
	}

	// Sample test stats
	testStats := &ContextStats{
		ProcessedFilesCount: 2,
		CharCount:           100,
		LineCount:           10,
		TokenCount:          25,
		ProcessedFiles:      []string{"/path/to/file1.go", "/path/to/file2.go"},
	}

	// Sample test config
	testConfig := interfaces.GatherConfig{
		Paths:        []string{"/path/to"},
		Include:      "*.go",
		Exclude:      "vendor/",
		ExcludeNames: "generated.go",
		Format:       "markdown",
		Verbose:      true,
		LogLevel:     logutil.DebugLevel,
	}

	// Test cases
	tests := []struct {
		name           string
		mockSetup      func(mock *MockContextGathererForAdapter)
		config         interfaces.GatherConfig
		expectedFiles  []fileutil.FileMeta
		expectedStats  *interfaces.ContextStats
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "success case - correctly adapts config and returns results",
			mockSetup: func(mock *MockContextGathererForAdapter) {
				// Track what internal GatherConfig was passed to verify conversion
				var capturedConfig GatherConfig

				mock.GatherContextFunc = func(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error) {
					// Capture the config for later verification
					capturedConfig = config

					// Return test data
					return testFiles, testStats, nil
				}

				// Verify after the function call that config was correctly adapted
				t.Cleanup(func() {
					// Verify all fields were properly converted
					if len(capturedConfig.Paths) != len(testConfig.Paths) || capturedConfig.Paths[0] != testConfig.Paths[0] {
						t.Errorf("Expected Paths to be %v, got %v", testConfig.Paths, capturedConfig.Paths)
					}
					if capturedConfig.Include != testConfig.Include {
						t.Errorf("Expected Include to be %s, got %s", testConfig.Include, capturedConfig.Include)
					}
					if capturedConfig.Exclude != testConfig.Exclude {
						t.Errorf("Expected Exclude to be %s, got %s", testConfig.Exclude, capturedConfig.Exclude)
					}
					if capturedConfig.ExcludeNames != testConfig.ExcludeNames {
						t.Errorf("Expected ExcludeNames to be %s, got %s", testConfig.ExcludeNames, capturedConfig.ExcludeNames)
					}
					if capturedConfig.Format != testConfig.Format {
						t.Errorf("Expected Format to be %s, got %s", testConfig.Format, capturedConfig.Format)
					}
					if capturedConfig.Verbose != testConfig.Verbose {
						t.Errorf("Expected Verbose to be %v, got %v", testConfig.Verbose, capturedConfig.Verbose)
					}
					if capturedConfig.LogLevel != testConfig.LogLevel {
						t.Errorf("Expected LogLevel to be %d, got %d", testConfig.LogLevel, capturedConfig.LogLevel)
					}
				})
			},
			config:        testConfig,
			expectedFiles: testFiles,
			expectedStats: &interfaces.ContextStats{
				ProcessedFilesCount: testStats.ProcessedFilesCount,
				CharCount:           testStats.CharCount,
				LineCount:           testStats.LineCount,
				TokenCount:          testStats.TokenCount,
				ProcessedFiles:      testStats.ProcessedFiles,
			},
			expectedError: false,
		},
		{
			name: "error case - handles error from underlying service",
			mockSetup: func(mock *MockContextGathererForAdapter) {
				mock.GatherContextFunc = func(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error) {
					return nil, nil, errors.New("failed to gather context")
				}
			},
			config:         testConfig,
			expectedFiles:  nil,
			expectedStats:  nil,
			expectedError:  true,
			expectedErrMsg: "failed to gather context",
		},
		{
			name: "nil context gatherer - panics",
			mockSetup: func(mock *MockContextGathererForAdapter) {
				// No setup needed for nil test
			},
			config:         testConfig,
			expectedFiles:  nil,
			expectedStats:  nil,
			expectedError:  true,
			expectedErrMsg: "nil ContextGatherer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *ContextGathererAdapter

			// For the nil ContextGatherer test
			if tc.name == "nil context gatherer - panics" {
				// Create adapter with nil ContextGatherer - should panic
				adapter = &ContextGathererAdapter{
					ContextGatherer: nil,
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
				_, _, _ = adapter.GatherContext(ctx, tc.config)
				return
			}

			// Create mock for non-nil test cases
			mockContextGatherer := &MockContextGathererForAdapter{}

			// Setup the mock
			tc.mockSetup(mockContextGatherer)

			// Create adapter with mock
			adapter = &ContextGathererAdapter{
				ContextGatherer: mockContextGatherer,
			}

			// Call the method being tested
			files, stats, err := adapter.GatherContext(ctx, tc.config)

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

			// For success case, verify returned files and stats
			if !tc.expectedError {
				// Check returned files
				if len(files) != len(tc.expectedFiles) {
					t.Errorf("Expected %d files, got %d", len(tc.expectedFiles), len(files))
				} else {
					for i, file := range files {
						if file.Path != tc.expectedFiles[i].Path {
							t.Errorf("Expected file path %s, got %s", tc.expectedFiles[i].Path, file.Path)
						}
						if file.Content != tc.expectedFiles[i].Content {
							t.Errorf("Expected file content %s, got %s", tc.expectedFiles[i].Content, file.Content)
						}
					}
				}

				// Check returned stats
				if stats == nil {
					t.Error("Expected non-nil stats but got nil")
				} else {
					if stats.ProcessedFilesCount != tc.expectedStats.ProcessedFilesCount {
						t.Errorf("Expected ProcessedFilesCount %d, got %d", tc.expectedStats.ProcessedFilesCount, stats.ProcessedFilesCount)
					}
					if stats.CharCount != tc.expectedStats.CharCount {
						t.Errorf("Expected CharCount %d, got %d", tc.expectedStats.CharCount, stats.CharCount)
					}
					if stats.LineCount != tc.expectedStats.LineCount {
						t.Errorf("Expected LineCount %d, got %d", tc.expectedStats.LineCount, stats.LineCount)
					}
					if stats.TokenCount != tc.expectedStats.TokenCount {
						t.Errorf("Expected TokenCount %d, got %d", tc.expectedStats.TokenCount, stats.TokenCount)
					}
					if len(stats.ProcessedFiles) != len(tc.expectedStats.ProcessedFiles) {
						t.Errorf("Expected %d processed files, got %d", len(tc.expectedStats.ProcessedFiles), len(stats.ProcessedFiles))
					} else {
						for i, file := range stats.ProcessedFiles {
							if file != tc.expectedStats.ProcessedFiles[i] {
								t.Errorf("Expected processed file %s, got %s", tc.expectedStats.ProcessedFiles[i], file)
							}
						}
					}
				}
			}
		})
	}
}

// TestContextGathererAdapter_DisplayDryRunInfo tests the DisplayDryRunInfo method of the ContextGathererAdapter
func TestContextGathererAdapter_DisplayDryRunInfo(t *testing.T) {
	// Create test context
	ctx := context.Background()

	// Sample test stats
	testStats := &interfaces.ContextStats{
		ProcessedFilesCount: 2,
		CharCount:           100,
		LineCount:           10,
		TokenCount:          25,
		ProcessedFiles:      []string{"/path/to/file1.go", "/path/to/file2.go"},
	}

	// Test cases
	tests := []struct {
		name           string
		mockSetup      func(mock *MockContextGathererForAdapter)
		stats          *interfaces.ContextStats
		expectedError  bool
		expectedErrMsg string
	}{
		{
			name: "success case - correctly adapts stats and calls underlying method",
			mockSetup: func(mock *MockContextGathererForAdapter) {
				// Track what internal ContextStats was passed to verify conversion
				var capturedStats *ContextStats

				mock.DisplayDryRunInfoFunc = func(ctx context.Context, stats *ContextStats) error {
					// Capture the stats for later verification
					capturedStats = stats

					// Return no error
					return nil
				}

				// Verify after the function call that stats was correctly adapted
				t.Cleanup(func() {
					// Verify all fields were properly converted
					if capturedStats.ProcessedFilesCount != testStats.ProcessedFilesCount {
						t.Errorf("Expected ProcessedFilesCount to be %d, got %d", testStats.ProcessedFilesCount, capturedStats.ProcessedFilesCount)
					}
					if capturedStats.CharCount != testStats.CharCount {
						t.Errorf("Expected CharCount to be %d, got %d", testStats.CharCount, capturedStats.CharCount)
					}
					if capturedStats.LineCount != testStats.LineCount {
						t.Errorf("Expected LineCount to be %d, got %d", testStats.LineCount, capturedStats.LineCount)
					}
					if capturedStats.TokenCount != testStats.TokenCount {
						t.Errorf("Expected TokenCount to be %d, got %d", testStats.TokenCount, capturedStats.TokenCount)
					}
					if len(capturedStats.ProcessedFiles) != len(testStats.ProcessedFiles) {
						t.Errorf("Expected ProcessedFiles to have %d items, got %d", len(testStats.ProcessedFiles), len(capturedStats.ProcessedFiles))
					} else {
						for i, file := range capturedStats.ProcessedFiles {
							if file != testStats.ProcessedFiles[i] {
								t.Errorf("Expected ProcessedFiles[%d] to be %s, got %s", i, testStats.ProcessedFiles[i], file)
							}
						}
					}
				})
			},
			stats:         testStats,
			expectedError: false,
		},
		{
			name: "error case - handles error from underlying service",
			mockSetup: func(mock *MockContextGathererForAdapter) {
				mock.DisplayDryRunInfoFunc = func(ctx context.Context, stats *ContextStats) error {
					return errors.New("failed to display dry run info")
				}
			},
			stats:          testStats,
			expectedError:  true,
			expectedErrMsg: "failed to display dry run info",
		},
		{
			name: "nil context gatherer - panics",
			mockSetup: func(mock *MockContextGathererForAdapter) {
				// No setup needed for nil test
			},
			stats:          testStats,
			expectedError:  true,
			expectedErrMsg: "nil ContextGatherer",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var adapter *ContextGathererAdapter

			// For the nil ContextGatherer test
			if tc.name == "nil context gatherer - panics" {
				// Create adapter with nil ContextGatherer - should panic
				adapter = &ContextGathererAdapter{
					ContextGatherer: nil,
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
				_ = adapter.DisplayDryRunInfo(ctx, tc.stats)
				return
			}

			// Create mock for non-nil test cases
			mockContextGatherer := &MockContextGathererForAdapter{}

			// Setup the mock
			tc.mockSetup(mockContextGatherer)

			// Create adapter with mock
			adapter = &ContextGathererAdapter{
				ContextGatherer: mockContextGatherer,
			}

			// Call the method being tested
			err := adapter.DisplayDryRunInfo(ctx, tc.stats)

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

// MockContextGathererForAdapter is a testing mock for the ContextGatherer interface, specifically for adapter tests
type MockContextGathererForAdapter struct {
	GatherContextFunc     func(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error)
	DisplayDryRunInfoFunc func(ctx context.Context, stats *ContextStats) error
}

func (m *MockContextGathererForAdapter) GatherContext(ctx context.Context, config GatherConfig) ([]fileutil.FileMeta, *ContextStats, error) {
	if m.GatherContextFunc != nil {
		return m.GatherContextFunc(ctx, config)
	}
	return nil, nil, errors.New("GatherContext not implemented")
}

func (m *MockContextGathererForAdapter) DisplayDryRunInfo(ctx context.Context, stats *ContextStats) error {
	if m.DisplayDryRunInfoFunc != nil {
		return m.DisplayDryRunInfoFunc(ctx, stats)
	}
	return errors.New("DisplayDryRunInfo not implemented")
}
