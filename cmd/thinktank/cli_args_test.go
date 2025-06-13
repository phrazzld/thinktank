// Package thinktank provides the command-line interface for the thinktank tool
package main

import (
	"flag"
	"io"
	"strings"
	"testing"
)

// TestParseFlags_PartialSuccessOk tests parsing of the partial-success-ok flag
func TestParseFlags_PartialSuccessOk(t *testing.T) {
	// Create a flag set
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // suppress flag error output

	// Test cases
	testCases := []struct {
		name           string
		args           []string
		expectedValue  bool
		expectError    bool
		errorSubstring string
	}{
		{
			name:          "Flag enabled",
			args:          []string{"--partial-success-ok"},
			expectedValue: true,
			expectError:   false,
		},
		{
			name:          "Flag explicitly enabled",
			args:          []string{"--partial-success-ok=true"},
			expectedValue: true,
			expectError:   false,
		},
		{
			name:          "Flag explicitly disabled",
			args:          []string{"--partial-success-ok=false"},
			expectedValue: false,
			expectError:   false,
		},
		{
			name:          "Flag not specified",
			args:          []string{"--instructions=test.txt"},
			expectedValue: false, // Default is false
			expectError:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset flag set for each test case
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			fs.SetOutput(io.Discard)

			// Parse with the test args
			cfg, err := ParseFlagsWithEnv(fs, tc.args, func(string) string { return "mock-value" })

			// Check error expectations
			if (err != nil) != tc.expectError {
				t.Errorf("Expected error: %v, got: %v", tc.expectError, err)
			}

			// If expecting an error, check error message
			if tc.expectError && err != nil && tc.errorSubstring != "" {
				if !strings.Contains(err.Error(), tc.errorSubstring) {
					t.Errorf("Expected error containing '%s', got: %v", tc.errorSubstring, err)
				}
			}

			// Check the parsed flag value
			if !tc.expectError && cfg != nil {
				if cfg.PartialSuccessOk != tc.expectedValue {
					t.Errorf("Expected PartialSuccessOk: %v, got: %v", tc.expectedValue, cfg.PartialSuccessOk)
				}
			}
		})
	}
}
