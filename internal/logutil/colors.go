package logutil

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// Color semantic purposes:
// - Blue (#3B82F6): Model names, interactive elements
// - Cyan (#06B6D4): Processing state, info
// - Green (#22C55E): Success, completion
// - Yellow (#FBBF24): Warnings, rate limits
// - Red (#EF4444): Errors, failures
// - Gray (#9CA3AF): Muted: timing, paths, separators
// - White (#FFFFFF): Section headers (bold)

// ColorScheme defines the color palette for modern clean CLI output.
// It provides semantic color mapping for different types of output elements,
// automatically adapting between interactive (colored) and CI (uncolored) environments.
type ColorScheme struct {
	ModelName     lipgloss.Style     // Subtle blue for model names
	Success       lipgloss.Style     // Green for success indicators
	Warning       lipgloss.Style     // Yellow for warning indicators
	Error         lipgloss.Style     // Red for error indicators
	Info          lipgloss.Style     // Cyan for processing indicators
	Duration      lipgloss.Style     // Gray for timing information
	FileSize      lipgloss.Style     // Gray for file size information
	FilePath      lipgloss.Style     // Default/white for file paths
	SectionHeader lipgloss.Style     // Bold white for section headers
	Separator     lipgloss.Style     // Gray for section separators
	Symbol        lipgloss.Style     // Default/white for Unicode symbols
	Reset         lipgloss.Style     // Reset style to clear formatting
	enabled       bool               // Whether colors are active
	renderer      *lipgloss.Renderer // Custom renderer for color output
}

// createStylesForRenderer creates lipgloss styles bound to a specific renderer.
// This ensures colors work even when running in non-TTY environments (like tests).
func createStylesForRenderer(r *lipgloss.Renderer) (modelName, success, warning, errorStyle, info, muted, sectionHeader, noStyle lipgloss.Style) {
	modelName = r.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#2563EB", Dark: "#3B82F6"})
	success = r.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#22C55E"})
	warning = r.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"})
	errorStyle = r.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"})
	info = r.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#0891B2", Dark: "#06B6D4"})
	muted = r.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#6B7280", Dark: "#9CA3AF"})
	sectionHeader = r.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"})
	noStyle = r.NewStyle()
	return
}

// NewColorScheme creates a new ColorScheme based on environment type.
// If interactive is true, returns a scheme with adaptive colors.
// If interactive is false, returns a scheme with no colors.
func NewColorScheme(interactive bool) *ColorScheme {
	if !interactive {
		// Non-interactive mode: use ASCII profile (no colors)
		r := lipgloss.NewRenderer(os.Stdout)
		r.SetColorProfile(termenv.Ascii)
		_, _, _, _, _, _, _, noStyle := createStylesForRenderer(r)
		return &ColorScheme{
			ModelName:     noStyle,
			Success:       noStyle,
			Warning:       noStyle,
			Error:         noStyle,
			Info:          noStyle,
			Duration:      noStyle,
			FileSize:      noStyle,
			FilePath:      noStyle,
			SectionHeader: noStyle,
			Separator:     noStyle,
			Symbol:        noStyle,
			Reset:         noStyle,
			enabled:       false,
			renderer:      r,
		}
	}

	// Interactive mode: detect terminal capabilities
	r := lipgloss.NewRenderer(os.Stdout)
	// Let termenv auto-detect the color profile instead of forcing TrueColor
	// Only override if detection fails or returns Ascii in interactive mode
	detectedProfile := termenv.NewOutput(os.Stdout).ColorProfile()
	if detectedProfile == termenv.Ascii {
		// We determined we're interactive via our own detection, so at minimum support ANSI
		r.SetColorProfile(termenv.ANSI)
	} else {
		r.SetColorProfile(detectedProfile)
	}
	// Detect dark background if possible, default to true for most terminals
	if output := termenv.NewOutput(os.Stdout); output.HasDarkBackground() {
		r.SetHasDarkBackground(true)
	} else {
		r.SetHasDarkBackground(false)
	}
	modelName, success, warning, errorStyle, info, muted, sectionHeader, noStyle := createStylesForRenderer(r)

	return &ColorScheme{
		ModelName:     modelName,
		Success:       success,
		Warning:       warning,
		Error:         errorStyle,
		Info:          info,
		Duration:      muted,
		FileSize:      muted,
		FilePath:      muted, // Changed to gray for muted appearance
		SectionHeader: sectionHeader,
		Separator:     muted,
		Symbol:        modelName, // Changed to blue for bullet points
		Reset:         noStyle,
		enabled:       true,
		renderer:      r,
	}
}

// ApplyColor applies the specified color to text if the scheme supports colors.
// In non-interactive mode, returns the text unchanged.
// In interactive mode, renders text with the requested style.
func (cs *ColorScheme) ApplyColor(color, text string) string {
	// If no color value or colors disabled, return text unchanged
	if color == "" || !cs.enabled {
		return text
	}

	style := cs.renderer.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: color, Dark: color})
	return style.Render(text)
}

// Convenience methods for applying semantic colors
func (cs *ColorScheme) applyStyle(style lipgloss.Style, text string) string {
	if !cs.enabled {
		return text
	}
	return style.Render(text)
}

// ModelName applies the model name color to text
func (cs *ColorScheme) ColorModelName(text string) string {
	return cs.applyStyle(cs.ModelName, text)
}

// Success applies the success color to text
func (cs *ColorScheme) ColorSuccess(text string) string {
	return cs.applyStyle(cs.Success, text)
}

// Warning applies the warning color to text
func (cs *ColorScheme) ColorWarning(text string) string {
	return cs.applyStyle(cs.Warning, text)
}

// Error applies the error color to text
func (cs *ColorScheme) ColorError(text string) string {
	return cs.applyStyle(cs.Error, text)
}

// Info applies the info color to text
func (cs *ColorScheme) ColorInfo(text string) string {
	return cs.applyStyle(cs.Info, text)
}

// Duration applies the duration color to text
func (cs *ColorScheme) ColorDuration(text string) string {
	return cs.applyStyle(cs.Duration, text)
}

// FileSize applies the file size color to text
func (cs *ColorScheme) ColorFileSize(text string) string {
	return cs.applyStyle(cs.FileSize, text)
}

// FilePath applies the file path color to text
func (cs *ColorScheme) ColorFilePath(text string) string {
	return cs.applyStyle(cs.FilePath, text)
}

// SectionHeader applies the section header color to text
func (cs *ColorScheme) ColorSectionHeader(text string) string {
	return cs.applyStyle(cs.SectionHeader, text)
}

// Separator applies the separator color to text
func (cs *ColorScheme) ColorSeparator(text string) string {
	return cs.applyStyle(cs.Separator, text)
}

// Symbol applies the symbol color to text
func (cs *ColorScheme) ColorSymbol(text string) string {
	return cs.applyStyle(cs.Symbol, text)
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

// SymbolSet defines the symbols used for different states and UI elements.
// Provides both Unicode and ASCII alternatives for maximum compatibility.
type SymbolSet struct {
	Success    string // ✓ or [OK]
	Error      string // ✗ or [X]
	Warning    string // ⚠ or [!]
	Bullet     string // ● or *
	Sparkles   string // ✨ or **
	Processing string // ... or ...
}

// UnicodeSymbols provides modern Unicode symbols for interactive terminals
var UnicodeSymbols = SymbolSet{
	Success:    "✓",
	Error:      "✗",
	Warning:    "⚠",
	Bullet:     "●",
	Sparkles:   "✨",
	Processing: "...",
}

// ASCIISymbols provides ASCII alternatives for limited terminals
var ASCIISymbols = SymbolSet{
	Success:    "[OK]",
	Error:      "[X]",
	Warning:    "[!]",
	Bullet:     "*",
	Sparkles:   "**",
	Processing: "...",
}

// SymbolProvider handles Unicode fallback detection and symbol selection
type SymbolProvider struct {
	symbols SymbolSet
}

// NewSymbolProvider creates a symbol provider with Unicode detection
func NewSymbolProvider(isInteractive bool) *SymbolProvider {
	// In non-interactive environments, always use ASCII for better compatibility
	if !isInteractive {
		return &SymbolProvider{symbols: ASCIISymbols}
	}

	// For interactive environments, detect Unicode support
	if supportsUnicode() {
		return &SymbolProvider{symbols: UnicodeSymbols}
	}

	return &SymbolProvider{symbols: ASCIISymbols}
}

// GetSymbols returns the current symbol set
func (sp *SymbolProvider) GetSymbols() SymbolSet {
	return sp.symbols
}

// supportsUnicode detects if the current terminal supports Unicode properly.
// This is a heuristic approach that checks common indicators.
func supportsUnicode() bool {
	// Check locale environment variables for UTF-8 support
	locale := os.Getenv("LC_ALL")
	if locale == "" {
		locale = os.Getenv("LC_CTYPE")
	}
	if locale == "" {
		locale = os.Getenv("LANG")
	}

	// If locale contains UTF-8, Unicode is likely supported
	if strings.Contains(strings.ToUpper(locale), "UTF-8") || strings.Contains(strings.ToUpper(locale), "UTF8") {
		return true
	}

	// Check terminal type indicators
	term := os.Getenv("TERM")

	// Modern terminals typically support Unicode
	modernTerms := []string{"xterm-256color", "screen-256color", "tmux-256color", "alacritty", "kitty"}
	for _, modernTerm := range modernTerms {
		if strings.Contains(term, modernTerm) {
			return true
		}
	}

	// Check for Windows Terminal, VS Code terminal, etc.
	if os.Getenv("WT_SESSION") != "" || os.Getenv("VSCODE_INJECTION") != "" {
		return true
	}

	// Conservative fallback: if we can't detect Unicode support, use ASCII
	return false
}
