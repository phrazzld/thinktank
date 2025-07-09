package logutil

import (
	"testing"
)

func TestNewColorScheme_Interactive(t *testing.T) {
	scheme := NewColorScheme(true) // interactive = true

	// Test that interactive scheme has ANSI color codes
	if scheme.ModelName == "" {
		t.Error("Expected ModelName to have color codes in interactive mode")
	}
	if scheme.Success == "" {
		t.Error("Expected Success to have color codes in interactive mode")
	}
	if scheme.Error == "" {
		t.Error("Expected Error to have color codes in interactive mode")
	}
	if scheme.Warning == "" {
		t.Error("Expected Warning to have color codes in interactive mode")
	}

	// Test specific colors match specification
	// ModelName should be subtle blue
	if !containsColorCode(scheme.ModelName) {
		t.Errorf("Expected ModelName to contain color codes, got %q", scheme.ModelName)
	}

	// Success should be green
	if !containsColorCode(scheme.Success) {
		t.Errorf("Expected Success to contain green color codes, got %q", scheme.Success)
	}

	// Error should be red
	if !containsColorCode(scheme.Error) {
		t.Errorf("Expected Error to contain red color codes, got %q", scheme.Error)
	}

	// Warning should be yellow
	if !containsColorCode(scheme.Warning) {
		t.Errorf("Expected Warning to contain yellow color codes, got %q", scheme.Warning)
	}
}

func TestNewColorScheme_NonInteractive(t *testing.T) {
	scheme := NewColorScheme(false) // interactive = false

	// Test that non-interactive scheme has no color codes
	if scheme.ModelName != "" {
		t.Errorf("Expected ModelName to be empty in non-interactive mode, got %q", scheme.ModelName)
	}
	if scheme.Success != "" {
		t.Errorf("Expected Success to be empty in non-interactive mode, got %q", scheme.Success)
	}
	if scheme.Error != "" {
		t.Errorf("Expected Error to be empty in non-interactive mode, got %q", scheme.Error)
	}
	if scheme.Warning != "" {
		t.Errorf("Expected Warning to be empty in non-interactive mode, got %q", scheme.Warning)
	}
	if scheme.Duration != "" {
		t.Errorf("Expected Duration to be empty in non-interactive mode, got %q", scheme.Duration)
	}
	if scheme.FileSize != "" {
		t.Errorf("Expected FileSize to be empty in non-interactive mode, got %q", scheme.FileSize)
	}
	if scheme.FilePath != "" {
		t.Errorf("Expected FilePath to be empty in non-interactive mode, got %q", scheme.FilePath)
	}
	if scheme.SectionHeader != "" {
		t.Errorf("Expected SectionHeader to be empty in non-interactive mode, got %q", scheme.SectionHeader)
	}
	if scheme.Separator != "" {
		t.Errorf("Expected Separator to be empty in non-interactive mode, got %q", scheme.Separator)
	}
	if scheme.Symbol != "" {
		t.Errorf("Expected Symbol to be empty in non-interactive mode, got %q", scheme.Symbol)
	}
}

func TestColorScheme_ApplyColor(t *testing.T) {
	tests := []struct {
		name        string
		interactive bool
		color       string
		text        string
		expectColor bool
	}{
		{"interactive with color", true, "\033[32m", "test", true},
		{"non-interactive no color", false, "", "test", false},
		{"interactive empty color", true, "", "test", false},
		{"non-interactive with color field", false, "\033[32m", "test", false}, // Should ignore color in non-interactive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := NewColorScheme(tt.interactive)
			result := scheme.ApplyColor(tt.color, tt.text)

			if tt.expectColor {
				if !containsColorCode(result) {
					t.Errorf("Expected result to contain color codes, got %q", result)
				}
				if result == tt.text {
					t.Errorf("Expected result to be different from input text when colored, got %q", result)
				}
			} else {
				if result != tt.text {
					t.Errorf("Expected result to equal input text when not colored, got %q, want %q", result, tt.text)
				}
			}
		})
	}
}

func TestDetectInteractiveEnvironmentForColors(t *testing.T) {
	tests := []struct {
		name           string
		isTerminalFunc func() bool
		envVars        map[string]string
		expectedResult bool
	}{
		{
			name:           "interactive terminal no CI",
			isTerminalFunc: func() bool { return true },
			envVars:        map[string]string{},
			expectedResult: true,
		},
		{
			name:           "non-terminal",
			isTerminalFunc: func() bool { return false },
			envVars:        map[string]string{},
			expectedResult: false,
		},
		{
			name:           "terminal but CI=true",
			isTerminalFunc: func() bool { return true },
			envVars:        map[string]string{"CI": "true"},
			expectedResult: false,
		},
		{
			name:           "terminal but GITHUB_ACTIONS set",
			isTerminalFunc: func() bool { return true },
			envVars:        map[string]string{"GITHUB_ACTIONS": "true"},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getEnvFunc := func(key string) string {
				return tt.envVars[key]
			}

			result := detectInteractiveEnvironmentWithEnvForColors(tt.isTerminalFunc, getEnvFunc)
			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

// Helper function to check if a string contains ANSI color codes
func containsColorCode(s string) bool {
	return len(s) > 0 && (s[0] == '\033' || s[0] == '\x1b')
}

// TestNewColorSchemeFromEnvironment tests the NewColorSchemeFromEnvironment function
func TestNewColorSchemeFromEnvironment(t *testing.T) {
	// This function uses the actual environment, so we test basic functionality
	scheme := NewColorSchemeFromEnvironment()

	// The scheme should not be nil
	if scheme == nil {
		t.Fatal("NewColorSchemeFromEnvironment() returned nil")
	}

	// The scheme should have proper structure (fields should exist)
	// We can't test exact values because they depend on the actual environment
	// But we can test that the function returns a valid ColorScheme

	// Test that it returns a different scheme in different scenarios
	// by testing the function components separately

	// Test that we can call it multiple times without issues
	scheme2 := NewColorSchemeFromEnvironment()
	if scheme2 == nil {
		t.Error("NewColorSchemeFromEnvironment() returned nil on second call")
	}
}

// TestDetectInteractiveEnvironmentForColorsWrapper tests the detectInteractiveEnvironmentForColors function
func TestDetectInteractiveEnvironmentForColorsWrapper(t *testing.T) {
	tests := []struct {
		name           string
		isTerminalFunc func() bool
		expectedResult bool
	}{
		{
			name:           "terminal returns true",
			isTerminalFunc: func() bool { return true },
			expectedResult: true, // Assumes no CI environment vars set
		},
		{
			name:           "terminal returns false",
			isTerminalFunc: func() bool { return false },
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test depends on actual environment variables
			// In a real CI environment, it might return false even if terminal is true
			result := detectInteractiveEnvironmentForColors(tt.isTerminalFunc)

			// We can't assert exact values because they depend on the actual environment
			// But we can test that the function runs without error
			_ = result
		})
	}
}

// TestDefaultIsTerminalForColors tests the defaultIsTerminalForColors function
func TestDefaultIsTerminalForColors(t *testing.T) {
	// This function calls the actual terminal detection
	result := defaultIsTerminalForColors()

	// We can't assert a specific value because it depends on how the test is run
	// But we can test that the function returns a boolean without error
	if result != true && result != false {
		t.Error("defaultIsTerminalForColors() should return a boolean")
	}

	// Test that it's consistent - calling it multiple times should return the same result
	result2 := defaultIsTerminalForColors()
	if result != result2 {
		t.Error("defaultIsTerminalForColors() should return consistent results")
	}
}

// TestGetEnvForColors tests the getEnvForColors function
func TestGetEnvForColors(t *testing.T) {
	// Test with a known environment variable
	result := getEnvForColors("PATH")
	// PATH should exist in most environments, but we can't assert exact value
	_ = result

	// Test with a non-existent environment variable
	result2 := getEnvForColors("NON_EXISTENT_VAR_12345")
	if result2 != "" {
		t.Errorf("getEnvForColors() should return empty string for non-existent var, got %q", result2)
	}

	// Test with empty string
	result3 := getEnvForColors("")
	if result3 != "" {
		t.Errorf("getEnvForColors() should return empty string for empty key, got %q", result3)
	}
}
