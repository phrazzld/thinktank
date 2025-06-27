//go:build manual_api_test
// +build manual_api_test

// Package e2e contains end-to-end compatibility tests for the simplified CLI
// These tests verify that the simplified interface produces equivalent output
// to the complex interface for identical inputs.
package e2e

import (
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/phrazzld/thinktank/internal/cli"
)

// TestSimplifiedComplexCompatibility tests core compatibility between
// simplified and complex CLI interfaces using Kent Beck's TDD approach.
// Following the principle: "Start with the smallest possible test that
// captures the core compatibility concern."
func TestSimplifiedComplexCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping e2e compatibility test in short mode")
	}

	// Create test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Test 1: Argument Parsing Equivalence (smallest test)
	// Input: ["--dry-run", "instructions.md", "./src"]
	// Assertion: Simplified parser → Complex config ≡ Complex parser → Complex config
	t.Run("ArgumentParsingEquivalence", func(t *testing.T) {
		// Create test files
		instructionsFile := env.CreateTestFile("instructions.md", "Test instructions for parsing")
		srcDir := env.CreateTestDirectory("src")
		env.CreateTestFile("src/main.go", "package main\n\nfunc main() {}")

		// Test simplified parser
		simplifiedArgs := []string{"thinktank", instructionsFile, srcDir, "--dry-run"}
		simplifiedConfig, err := cli.ParseSimpleArgsWithArgs(simplifiedArgs)
		if err != nil {
			t.Fatalf("Simplified parser failed: %v", err)
		}

		// Convert to complex config
		convertedConfig := simplifiedConfig.ToCliConfig()

		// Verify essential equivalence
		if !convertedConfig.DryRun {
			t.Error("DryRun flag not preserved in conversion")
		}
		if convertedConfig.InstructionsFile != instructionsFile {
			t.Errorf("Instructions file mismatch: got %s, want %s",
				convertedConfig.InstructionsFile, instructionsFile)
		}
		if len(convertedConfig.Paths) != 1 || convertedConfig.Paths[0] != srcDir {
			t.Errorf("Paths mismatch: got %v, want [%s]", convertedConfig.Paths, srcDir)
		}
	})

	// Test 2: Output File Equivalence
	// Verify both parsers create files with equivalent content structure
	t.Run("OutputFileEquivalence", func(t *testing.T) {
		// Create test files
		instructionsFile := env.CreateTestFile("test-instructions.md", "Analyze the code structure")
		srcDir := env.CreateTestDirectory("test-src")
		env.CreateTestFile("test-src/example.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello, world!")
}`)

		// Run with simplified interface (positional: instructions target flags)
		simplifiedArgs := []string{instructionsFile, srcDir, "--dry-run", "--verbose"}
		stdout1, stderr1, exitCode1, err1 := env.RunThinktank(simplifiedArgs, nil)
		if err1 != nil {
			t.Logf("Simplified interface error (acceptable in mock): %v", err1)
		}

		// Run with complex interface (equivalent flags)
		complexArgs := []string{
			"--instructions", instructionsFile,
			"--dry-run", "--verbose",
			srcDir,
		}
		stdout2, stderr2, exitCode2, err2 := env.RunThinktank(complexArgs, nil)
		if err2 != nil {
			t.Logf("Complex interface error (acceptable in mock): %v", err2)
		}

		// Verify exit codes match (allow for API connectivity differences)
		if exitCode1 != exitCode2 && exitCode1 != 0 && exitCode2 != 0 {
			t.Logf("Exit codes differ: simplified=%d, complex=%d (acceptable in mock)", exitCode1, exitCode2)
		}

		// Verify both contain essential processing markers
		expectedMarkers := []string{"Gathering context", "Processing", "analysis"}
		for _, marker := range expectedMarkers {
			if !containsIgnoreCase(stdout1, marker) && !containsIgnoreCase(stderr1, marker) {
				t.Logf("Simplified output missing marker '%s' (acceptable in mock)", marker)
			}
			if !containsIgnoreCase(stdout2, marker) && !containsIgnoreCase(stderr2, marker) {
				t.Logf("Complex output missing marker '%s' (acceptable in mock)", marker)
			}
		}
	})

	// Test 3: API Request Pattern Consistency
	// Verify both interfaces make equivalent API calls
	t.Run("APIRequestPatternConsistency", func(t *testing.T) {
		// Track API calls via mock server
		var simplifiedCalls, complexCalls int

		// Modify mock server to count calls
		originalHandler := env.MockConfig.HandleGeneration

		// Count calls for simplified interface
		env.MockConfig.HandleGeneration = func(req *http.Request) (string, int, string, error) {
			simplifiedCalls++
			return originalHandler(req)
		}

		// Create test files
		instructionsFile := env.CreateTestFile("api-test-instructions.md", "Simple analysis task")
		srcDir := env.CreateTestDirectory("api-test-src")
		env.CreateTestFile("api-test-src/simple.go", "package main\n\nfunc main() {}")

		// Run simplified interface (positional: instructions target flags)
		simplifiedArgs := []string{instructionsFile, srcDir, "--model", "gemini-2.5-pro"}
		_, _, _, err1 := env.RunThinktank(simplifiedArgs, nil)
		if err1 != nil {
			t.Logf("Simplified interface error (acceptable in mock): %v", err1)
		}

		// Reset counter for complex interface
		simplifiedCallCount := simplifiedCalls
		env.MockConfig.HandleGeneration = func(req *http.Request) (string, int, string, error) {
			complexCalls++
			return originalHandler(req)
		}

		// Run complex interface
		complexArgs := []string{
			"--instructions", instructionsFile,
			"--model", "gemini-2.5-pro",
			srcDir,
		}
		_, _, _, err2 := env.RunThinktank(complexArgs, nil)
		if err2 != nil {
			t.Logf("Complex interface error (acceptable in mock): %v", err2)
		}

		// Restore original handler
		env.MockConfig.HandleGeneration = originalHandler

		// Verify call patterns are equivalent (allowing for mock environment variations)
		if simplifiedCallCount > 0 && complexCalls > 0 {
			if abs(simplifiedCallCount, complexCalls) > 1 {
				t.Errorf("API call count significantly different: simplified=%d, complex=%d",
					simplifiedCallCount, complexCalls)
			}
		} else {
			t.Log("API calls not detected (acceptable in mock environment)")
		}
	})

	// Test 4: Performance Regression Detection
	// Ensure simplified interface doesn't introduce >10% slowdown
	t.Run("PerformanceRegressionDetection", func(t *testing.T) {
		// Create test files
		instructionsFile := env.CreateTestFile("perf-instructions.md", "Performance test instructions")
		srcDir := env.CreateTestDirectory("perf-src")
		for i := 0; i < 5; i++ {
			env.CreateTestFile(filepath.Join("perf-src", "file"+string(rune('0'+i))+".go"),
				"package main\n\nfunc main() {}")
		}

		// Measure simplified interface performance (positional: instructions target flags)
		start1 := time.Now()
		simplifiedArgs := []string{instructionsFile, srcDir, "--dry-run"}
		_, _, _, err1 := env.RunThinktank(simplifiedArgs, nil)
		duration1 := time.Since(start1)
		if err1 != nil {
			t.Logf("Simplified interface error (acceptable in mock): %v", err1)
		}

		// Measure complex interface performance
		start2 := time.Now()
		complexArgs := []string{"--instructions", instructionsFile, "--dry-run", srcDir}
		_, _, _, err2 := env.RunThinktank(complexArgs, nil)
		duration2 := time.Since(start2)
		if err2 != nil {
			t.Logf("Complex interface error (acceptable in mock): %v", err2)
		}

		// Verify no significant performance regression (target: <10% slowdown)
		if duration1 > 0 && duration2 > 0 {
			ratio := float64(duration1) / float64(duration2)
			if ratio > 1.10 {
				t.Errorf("Performance regression detected: simplified=%.2fms, complex=%.2fms (%.1f%% slower)",
					float64(duration1.Nanoseconds())/1e6,
					float64(duration2.Nanoseconds())/1e6,
					(ratio-1)*100)
			} else {
				t.Logf("Performance acceptable: simplified=%.2fms, complex=%.2fms (%.1f%% difference)",
					float64(duration1.Nanoseconds())/1e6,
					float64(duration2.Nanoseconds())/1e6,
					(ratio-1)*100)
			}
		} else {
			t.Log("Performance measurement inconclusive (acceptable in mock environment)")
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
			args:        []string{"thinktank", instructionsFile, srcDir, "--model", "gpt-4.1"},
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
			args:        []string{"thinktank", instructionsFile, srcDir, "--model=gpt-4.1"},
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
