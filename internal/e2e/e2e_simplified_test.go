//go:build manual_api_test
// +build manual_api_test

// Package e2e contains end-to-end tests for the simplified CLI
// These tests verify that the simplified interface works correctly in various scenarios.
package e2e

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/cli"
)

// TestSimplifiedInterfaceE2E tests the simplified CLI interface end-to-end
// using Kent Beck's TDD approach. Focuses on validating the core functionality
// works as expected in a real environment.
func TestSimplifiedInterfaceE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e test in short mode")
	}

	// Create test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Test 1: Argument Parsing Validation
	// Ensure the simplified parser correctly handles various flag combinations
	t.Run("ArgumentParsingValidation", func(t *testing.T) {
		// Create test files
		instructionsFile := env.CreateTestFile("instructions.md", "Test instructions for parsing")
		srcDir := env.CreateTestDirectory("src")
		env.CreateTestFile("src/main.go", "package main\n\nfunc main() {}")

		// Test simplified parser with various flags
		testCases := []struct {
			name      string
			args      []string
			checkFunc func(*testing.T, *cli.SimplifiedConfig)
		}{
			{
				name: "dry-run flag",
				args: []string{"thinktank", instructionsFile, srcDir, "--dry-run"},
				checkFunc: func(t *testing.T, config *cli.SimplifiedConfig) {
					if !config.HasFlag(cli.FlagDryRun) {
						t.Error("DryRun flag not set")
					}
				},
			},
			{
				name: "verbose flag",
				args: []string{"thinktank", instructionsFile, srcDir, "--verbose"},
				checkFunc: func(t *testing.T, config *cli.SimplifiedConfig) {
					if !config.HasFlag(cli.FlagVerbose) {
						t.Error("Verbose flag not set")
					}
				},
			},
			{
				name: "synthesis flag",
				args: []string{"thinktank", instructionsFile, srcDir, "--synthesis"},
				checkFunc: func(t *testing.T, config *cli.SimplifiedConfig) {
					if !config.HasFlag(cli.FlagSynthesis) {
						t.Error("Synthesis flag not set")
					}
				},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				config, err := cli.ParseSimpleArgsWithArgs(tc.args)
				if err != nil {
					t.Fatalf("Parser failed: %v", err)
				}

				tc.checkFunc(t, config)

				// Verify basic fields are always set correctly
				if config.InstructionsFile != instructionsFile {
					t.Errorf("Instructions file mismatch: got %s, want %s",
						config.InstructionsFile, instructionsFile)
				}
				if config.TargetPath != srcDir {
					t.Errorf("Target path mismatch: got %s, want %s", config.TargetPath, srcDir)
				}
			})
		}
	})

	// Test 2: Dry Run Functionality
	// Verify the simplified interface handles dry-run mode correctly
	t.Run("DryRunFunctionality", func(t *testing.T) {
		// Create test files
		instructionsFile := env.CreateTestFile("test-instructions.md", "Analyze the code structure")
		srcDir := env.CreateTestDirectory("test-src")
		env.CreateTestFile("test-src/example.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

		// Run with simplified interface in dry-run mode
		args := []string{instructionsFile, srcDir, "--dry-run", "--verbose"}
		stdout, stderr, exitCode, err := env.RunThinktank(args, nil)
		if err != nil {
			t.Logf("Dry run error (acceptable in mock): %v", err)
		}

		// Verify dry-run mode produces expected output
		expectedMarkers := []string{"DRY RUN MODE", "Files that would be processed", "Total characters"}
		for _, marker := range expectedMarkers {
			if !containsIgnoreCase(stdout, marker) && !containsIgnoreCase(stderr, marker) {
				t.Errorf("Dry run output missing marker '%s'", marker)
			}
		}

		// Verify dry-run exits successfully (or with acceptable error codes in mock)
		if exitCode != 0 && exitCode != 1 {
			t.Errorf("Unexpected exit code in dry-run: %d", exitCode)
		}
	})

	// Test 3: Performance Validation
	// Ensure the simplified interface performs reasonably
	t.Run("PerformanceValidation", func(t *testing.T) {
		// Create test files
		instructionsFile := env.CreateTestFile("perf-instructions.md", "Performance test instructions")
		srcDir := env.CreateTestDirectory("perf-src")
		for i := 0; i < 5; i++ {
			env.CreateTestFile(filepath.Join("perf-src", "file"+string(rune('0'+i))+".go"),
				"package main\n\nfunc main() {}")
		}

		// Measure performance in dry-run mode
		start := time.Now()
		args := []string{instructionsFile, srcDir, "--dry-run"}
		_, _, _, err := env.RunThinktank(args, nil)
		duration := time.Since(start)
		if err != nil {
			t.Logf("Performance test error (acceptable in mock): %v", err)
		}

		// Verify reasonable performance (should complete quickly in dry-run)
		maxDuration := 30 * time.Second // Very generous for a dry-run
		if duration > maxDuration {
			t.Errorf("Performance issue: dry-run took too long: %.2fs (max: %.2fs)",
				duration.Seconds(), maxDuration.Seconds())
		} else {
			t.Logf("Performance acceptable: dry-run completed in %.2fms",
				float64(duration.Nanoseconds())/1e6)
		}
	})
}

// TestSimplifiedParserEdgeCases tests edge cases specific to the simplified parser
func TestSimplifiedParserEdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping simplified parser edge case tests in short mode")
	}

	// Create test environment with actual files
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files once
	instructionsFile := env.CreateTestFile("test.txt", "Test instructions")
	srcDir := env.CreateTestDirectory("src")
	env.CreateTestFile("src/main.go", "package main\n\nfunc main() {}")

	testCases := []struct {
		name           string
		args           []string
		expectError    bool
		expectedError  string
		createTestFile bool
	}{
		{
			name:        "MinimalValidArgs",
			args:        []string{"thinktank", instructionsFile, srcDir},
			expectError: false,
		},
		{
			name:          "MissingTargetPath",
			args:          []string{"thinktank", instructionsFile},
			expectError:   true,
			expectedError: "usage:",
		},
		{
			name:          "EmptyArgs",
			args:          []string{},
			expectError:   true,
			expectedError: "usage:",
		},
		{
			name:        "WithDryRunFlag",
			args:        []string{"thinktank", instructionsFile, srcDir, "--dry-run"},
			expectError: false,
		},
		{
			name:        "WithMultipleFlags",
			args:        []string{"thinktank", instructionsFile, srcDir, "--dry-run", "--verbose", "--synthesis"},
			expectError: false,
		},
		{
			name:          "UnknownFlag",
			args:          []string{"thinktank", instructionsFile, srcDir, "--unknown"},
			expectError:   true,
			expectedError: "unknown flag",
		},
		{
			name:        "ModelFlagWithValue",
			args:        []string{"thinktank", instructionsFile, srcDir, "--model", "gpt-5.2"},
			expectError: false,
		},
		{
			name:          "ModelFlagWithoutValue",
			args:          []string{"thinktank", instructionsFile, srcDir, "--model"},
			expectError:   true,
			expectedError: "--model flag requires a value",
		},
		{
			name:        "ModelFlagEqualsFormat",
			args:        []string{"thinktank", instructionsFile, srcDir, "--model=gpt-5.2"},
			expectError: false,
		},
		{
			name:          "ModelFlagEmptyEqualsFormat",
			args:          []string{"thinktank", instructionsFile, srcDir, "--model="},
			expectError:   true,
			expectedError: "--model flag requires a non-empty value",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := cli.ParseSimpleArgsWithArgs(tc.args)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tc.expectedError)
				} else if !strings.Contains(err.Error(), tc.expectedError) {
					t.Errorf("Expected error containing '%s', got: %v", tc.expectedError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if config == nil {
					t.Error("Expected non-nil config")
				}
			}
		})
	}
}

// Helper functions

// containsIgnoreCase checks if text contains substring (case-insensitive)
func containsIgnoreCase(text, substring string) bool {
	return strings.Contains(strings.ToLower(text), strings.ToLower(substring))
}

// abs returns absolute value of integer difference
func abs(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}
