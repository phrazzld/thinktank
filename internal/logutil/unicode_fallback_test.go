package logutil

import (
	"os"
	"testing"
)

// TestSymbolProvider verifies that the symbol provider correctly detects Unicode support
// and provides appropriate symbol sets for different environments
func TestSymbolProvider(t *testing.T) {
	tests := []struct {
		name          string
		isInteractive bool
		envVars       map[string]string
		expectUnicode bool
	}{
		{
			name:          "Non-interactive always uses ASCII",
			isInteractive: false,
			envVars:       map[string]string{"LANG": "en_US.UTF-8"},
			expectUnicode: false,
		},
		{
			name:          "Interactive with UTF-8 locale uses Unicode",
			isInteractive: true,
			envVars:       map[string]string{"LANG": "en_US.UTF-8"},
			expectUnicode: true,
		},
		{
			name:          "Interactive with minimal environment uses ASCII",
			isInteractive: true,
			envVars:       map[string]string{}, // Clear all, no Unicode indicators
			expectUnicode: false,
		},
		{
			name:          "Interactive with modern terminal uses Unicode",
			isInteractive: true,
			envVars:       map[string]string{"TERM": "xterm-256color"},
			expectUnicode: true,
		},
		{
			name:          "Interactive with VS Code uses Unicode",
			isInteractive: true,
			envVars:       map[string]string{"VSCODE_INJECTION": "1"},
			expectUnicode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment for all Unicode-related variables
			allKeys := []string{"LANG", "LC_ALL", "LC_CTYPE", "TERM", "WT_SESSION", "VSCODE_INJECTION"}
			originalEnv := make(map[string]string)
			for _, key := range allKeys {
				originalEnv[key] = os.Getenv(key)
			}

			// For minimal environment test, clear all variables first
			if len(tt.envVars) == 0 {
				for _, key := range allKeys {
					_ = os.Unsetenv(key)
				}
			}

			// Set test environment
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Create symbol provider
			provider := NewSymbolProvider(tt.isInteractive)
			symbols := provider.GetSymbols()

			// Restore original environment
			for _, key := range allKeys {
				originalValue := originalEnv[key]
				if originalValue == "" {
					_ = os.Unsetenv(key)
				} else {
					_ = os.Setenv(key, originalValue)
				}
			}

			// Verify symbol selection
			if tt.expectUnicode {
				if symbols.Success != "✓" {
					t.Errorf("Expected Unicode symbols, got ASCII. Success symbol: %q", symbols.Success)
				}
			} else {
				if symbols.Success != "[OK]" {
					t.Errorf("Expected ASCII symbols, got Unicode. Success symbol: %q", symbols.Success)
				}
			}
		})
	}
}

// TestUnicodeDetection verifies the Unicode detection logic
func TestUnicodeDetection(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "UTF-8 locale detected",
			envVars:  map[string]string{"LANG": "en_US.UTF-8"},
			expected: true,
		},
		{
			name:     "Modern terminal detected",
			envVars:  map[string]string{"TERM": "xterm-256color"},
			expected: true,
		},
		{
			name:     "No Unicode indicators",
			envVars:  map[string]string{"LANG": "C", "TERM": "vt100"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalEnv := make(map[string]string)
			allKeys := []string{"LANG", "LC_ALL", "LC_CTYPE", "TERM", "WT_SESSION", "VSCODE_INJECTION"}
			for _, key := range allKeys {
				originalEnv[key] = os.Getenv(key)
				_ = os.Unsetenv(key) // Clear all first
			}

			// Set test environment
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
			}

			// Test Unicode detection
			result := supportsUnicode()

			// Restore original environment
			for key, originalValue := range originalEnv {
				if originalValue == "" {
					_ = os.Unsetenv(key)
				} else {
					_ = os.Setenv(key, originalValue)
				}
			}

			if result != tt.expected {
				t.Errorf("Expected Unicode detection to return %v, got %v", tt.expected, result)
			}
		})
	}
}

// TestSymbolSets verifies the symbol sets contain expected values
func TestSymbolSets(t *testing.T) {
	// Test Unicode symbols
	if UnicodeSymbols.Success != "✓" {
		t.Errorf("Expected Unicode success symbol ✓, got %q", UnicodeSymbols.Success)
	}
	if UnicodeSymbols.Error != "✗" {
		t.Errorf("Expected Unicode error symbol ✗, got %q", UnicodeSymbols.Error)
	}

	// Test ASCII symbols
	if ASCIISymbols.Success != "[OK]" {
		t.Errorf("Expected ASCII success symbol [OK], got %q", ASCIISymbols.Success)
	}
	if ASCIISymbols.Error != "[X]" {
		t.Errorf("Expected ASCII error symbol [X], got %q", ASCIISymbols.Error)
	}
}
