package logutil

import (
	"os"

	"golang.org/x/term"
)

// ColorScheme defines the color palette for modern clean CLI output.
// It provides semantic color mapping for different types of output elements,
// automatically adapting between interactive (colored) and CI (uncolored) environments.
type ColorScheme struct {
	ModelName     string // Subtle blue for model names
	Success       string // Green for success indicators
	Warning       string // Yellow for warning indicators
	Error         string // Red for error indicators
	Duration      string // Gray for timing information
	FileSize      string // Gray for file size information
	FilePath      string // Default/white for file paths
	SectionHeader string // Bold white for section headers
	Separator     string // Gray for section separators
	Symbol        string // Default/white for Unicode symbols
	Reset         string // Reset code to clear formatting
}

// ANSI color codes for interactive terminals
const (
	// Foreground colors
	ansiReset     = "\033[0m"
	ansiBlue      = "\033[34m"   // ModelName - subtle blue
	ansiGreen     = "\033[32m"   // Success - green
	ansiYellow    = "\033[33m"   // Warning - yellow
	ansiRed       = "\033[31m"   // Error - red
	ansiGray      = "\033[90m"   // Duration, FileSize, Separator - gray
	ansiDefault   = "\033[39m"   // FilePath, Symbol - default/white
	ansiBoldWhite = "\033[1;37m" // SectionHeader - bold white
)

// NewColorScheme creates a new ColorScheme based on environment type.
// If interactive is true, returns a scheme with ANSI color codes.
// If interactive is false, returns a scheme with empty strings (no colors).
func NewColorScheme(interactive bool) *ColorScheme {
	if !interactive {
		// Non-interactive mode: no colors, only semantic meaning preserved
		return &ColorScheme{
			ModelName:     "",
			Success:       "",
			Warning:       "",
			Error:         "",
			Duration:      "",
			FileSize:      "",
			FilePath:      "",
			SectionHeader: "",
			Separator:     "",
			Symbol:        "",
			Reset:         "",
		}
	}

	// Interactive mode: full ANSI color support
	return &ColorScheme{
		ModelName:     ansiBlue,
		Success:       ansiGreen,
		Warning:       ansiYellow,
		Error:         ansiRed,
		Duration:      ansiGray,
		FileSize:      ansiGray,
		FilePath:      ansiDefault,
		SectionHeader: ansiBoldWhite,
		Separator:     ansiGray,
		Symbol:        ansiDefault,
		Reset:         ansiReset,
	}
}

// ApplyColor applies the specified color to text if the scheme supports colors.
// In non-interactive mode, returns the text unchanged.
// In interactive mode, wraps text with color codes and reset.
func (cs *ColorScheme) ApplyColor(color, text string) string {
	// If no color code or reset code, we're in non-interactive mode
	if color == "" || cs.Reset == "" {
		return text
	}

	return color + text + cs.Reset
}

// Convenience methods for applying semantic colors

// ModelName applies the model name color to text
func (cs *ColorScheme) ColorModelName(text string) string {
	return cs.ApplyColor(cs.ModelName, text)
}

// Success applies the success color to text
func (cs *ColorScheme) ColorSuccess(text string) string {
	return cs.ApplyColor(cs.Success, text)
}

// Warning applies the warning color to text
func (cs *ColorScheme) ColorWarning(text string) string {
	return cs.ApplyColor(cs.Warning, text)
}

// Error applies the error color to text
func (cs *ColorScheme) ColorError(text string) string {
	return cs.ApplyColor(cs.Error, text)
}

// Duration applies the duration color to text
func (cs *ColorScheme) ColorDuration(text string) string {
	return cs.ApplyColor(cs.Duration, text)
}

// FileSize applies the file size color to text
func (cs *ColorScheme) ColorFileSize(text string) string {
	return cs.ApplyColor(cs.FileSize, text)
}

// FilePath applies the file path color to text
func (cs *ColorScheme) ColorFilePath(text string) string {
	return cs.ApplyColor(cs.FilePath, text)
}

// SectionHeader applies the section header color to text
func (cs *ColorScheme) ColorSectionHeader(text string) string {
	return cs.ApplyColor(cs.SectionHeader, text)
}

// Separator applies the separator color to text
func (cs *ColorScheme) ColorSeparator(text string) string {
	return cs.ApplyColor(cs.Separator, text)
}

// Symbol applies the symbol color to text
func (cs *ColorScheme) ColorSymbol(text string) string {
	return cs.ApplyColor(cs.Symbol, text)
}

// NewColorSchemeFromEnvironment creates a ColorScheme by detecting the current environment.
// Uses the same detection logic as the console writer for consistency.
func NewColorSchemeFromEnvironment() *ColorScheme {
	isInteractive := detectInteractiveEnvironmentForColors(defaultIsTerminalForColors)
	return NewColorScheme(isInteractive)
}

// detectInteractiveEnvironmentForColors determines if we're running in an interactive
// environment based on TTY detection and CI environment variables.
// This is a copy of the logic from console_writer.go to avoid circular dependencies.
func detectInteractiveEnvironmentForColors(isTerminalFunc func() bool) bool {
	return detectInteractiveEnvironmentWithEnvForColors(isTerminalFunc, getEnvForColors)
}

// detectInteractiveEnvironmentWithEnvForColors determines if we're running in an interactive
// environment with injectable environment function for testing.
func detectInteractiveEnvironmentWithEnvForColors(isTerminalFunc func() bool, getEnvFunc func(string) string) bool {
	// Check common CI environment variables
	ciVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"CONTINUOUS_INTEGRATION",
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"BUILDKITE",
	}

	for _, envVar := range ciVars {
		value := getEnvFunc(envVar)
		if value != "" && (value == "true" || envVar == "JENKINS_URL") {
			return false
		}
	}

	// If not in CI and stdout is a terminal, we're interactive
	return isTerminalFunc()
}

// defaultIsTerminalForColors uses golang.org/x/term to detect if stdout is a terminal
func defaultIsTerminalForColors() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// getEnvForColors gets environment variables using os.Getenv
func getEnvForColors(key string) string {
	return os.Getenv(key)
}
