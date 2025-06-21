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
