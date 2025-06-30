package cli

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test 1: RED - The simplest failing test for --help flag
func TestParseSimpleArgs_HelpFlag(t *testing.T) {
	t.Parallel()

	args := []string{"thinktank", "--help"}
	config, err := ParseSimpleArgsWithArgs(args)

	// We expect this to succeed but with help flag set
	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.True(t, config.HelpRequested())
}

// Test 2: Short form support
func TestParseSimpleArgs_HelpFlag_ShortForm(t *testing.T) {
	t.Parallel()

	args := []string{"thinktank", "-h"}
	config, err := ParseSimpleArgsWithArgs(args)

	assert.Nil(t, err)
	assert.NotNil(t, config)
	assert.True(t, config.HelpRequested())
}

// Test 3: Help flag position independence
func TestParseSimpleArgs_HelpFlag_PositionIndependence(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "help_at_end",
			args: []string{"thinktank", "instructions.txt", "src/", "--help"},
		},
		{
			name: "help_in_middle",
			args: []string{"thinktank", "instructions.txt", "--help", "src/"},
		},
		{
			name: "help_with_other_flags",
			args: []string{"thinktank", "--dry-run", "--help", "--verbose"},
		},
		{
			name: "help_after_model_flag",
			args: []string{"thinktank", "--model", "gpt-4", "-h"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config, err := ParseSimpleArgsWithArgs(tc.args)
			assert.Nil(t, err)
			assert.NotNil(t, config)
			assert.True(t, config.HelpRequested())
		})
	}
}

// Test 4: Help text content
func TestPrintHelp_ContainsEssentialInformation(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	PrintHelp(&buf)
	helpText := buf.String()

	// Check for essential components
	essentialContent := []string{
		// Usage pattern
		"thinktank instructions.txt target_path",
		// Key flags
		"--help",
		"--dry-run",
		"--verbose",
		"--model",
		"--synthesis",
		// Examples
		"EXAMPLES:",
		"thinktank instructions.md ./src",
		// Environment variables
		"GEMINI_API_KEY",
		"OPENAI_API_KEY",
		// Troubleshooting
		"TROUBLESHOOTING:",
	}

	for _, content := range essentialContent {
		assert.Contains(t, helpText, content, "Help text should contain: %s", content)
	}
}

// Test 5: Help text is properly formatted
func TestPrintHelp_ProperFormatting(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	PrintHelp(&buf)
	helpText := buf.String()

	// Should have proper sections
	assert.Contains(t, helpText, "USAGE:")
	assert.Contains(t, helpText, "DESCRIPTION:")
	assert.Contains(t, helpText, "ARGUMENTS:")
	assert.Contains(t, helpText, "FLAGS:")
	assert.Contains(t, helpText, "EXAMPLES:")
	assert.Contains(t, helpText, "ENVIRONMENT VARIABLES:")
	assert.Contains(t, helpText, "TROUBLESHOOTING:")

	// Should not have any template variables
	assert.NotContains(t, helpText, "{{")
	assert.NotContains(t, helpText, "}}")

	// Should end with a newline
	assert.True(t, strings.HasSuffix(helpText, "\n"))
}

// Test 6: Integration - Main handles help correctly
func TestMain_HelpFlagExitsWithZero(t *testing.T) {
	// Save original values
	oldArgs := os.Args
	oldExit := osExit
	oldStdout := os.Stdout

	// Create a pipe to capture stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Track exit code
	var exitCode int
	osExit = func(code int) {
		exitCode = code
		panic("exit") // Use panic to stop execution
	}

	// Restore after test
	defer func() {
		os.Args = oldArgs
		osExit = oldExit
		os.Stdout = oldStdout
		_ = w.Close()
	}()

	// Test help flag
	os.Args = []string{"thinktank", "--help"}

	// Run Main and catch the exit
	assert.Panics(t, func() {
		Main()
	})

	// Close writer and read output
	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	// Verify exit code and output
	assert.Equal(t, ExitCodeSuccess, exitCode)
	assert.Contains(t, buf.String(), "USAGE:")
	assert.Contains(t, buf.String(), "thinktank instructions.txt target_path")
}
