//go:build manual_api_test
// +build manual_api_test

package e2e

import (
	"os"
	"strings"
	"testing"
)

// TestModernCLIOutput provides integration tests for the modern clean CLI output system.
// These tests verify end-to-end behavior of the console output transformation including
// environment adaptation, Unicode fallback, color schemes, and responsive layout.

// TestModernCLIOutput_BasicWorkflow tests the complete modern output workflow
func TestModernCLIOutput_BasicWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	// Create test files with varied content
	env.CreateTestFile("input1.txt", "Analyze this technical documentation.")
	env.CreateTestFile("input2.txt", "Review this architectural proposal.")

	// Set up environment for ASCII output (simulating CI)
	originalCI := os.Getenv("CI")
	defer func() {
		if originalCI == "" {
			_ = os.Unsetenv("CI")
		} else {
			_ = os.Setenv("CI", originalCI)
		}
	}()
	_ = os.Setenv("CI", "true")

	// Set up test flags for multiple models
	flags := env.DefaultFlags
	flags.Model = []string{"gemini-3-flash", "gemini-3-flash"}
	flags.Instructions = "Brief analysis:"
	flags.DryRun = true // Use dry-run to avoid actual API calls

	// Run with the test files
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{
		env.TempDir + "/input1.txt",
		env.TempDir + "/input2.txt",
	})

	if err != nil {
		t.Fatalf("Failed to run thinktank: %v", err)
	}

	// Log output for debugging
	t.Logf("Exit code: %d\nStdout: %s\nStderr: %s", exitCode, stdout, stderr)

	output := stdout + stderr

	// Verify modern clean output elements (CI mode uses ASCII symbols)
	expectedElements := []string{
		"Processing 2 models", // Clean initialization message
		"gemini-3-flash",      // Model names
		"gemini-3-flash",
		"SUMMARY",              // Uppercase section headers
		"* 2 models processed", // ASCII bullet points in CI mode
		"OUTPUT FILES",         // File listing section
		"FAILED MODELS",        // May appear if dry-run causes failures
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Logf("Expected element %q not found in output. Full output:\n%s", element, output)
			// Don't fail on optional elements like FAILED MODELS
			if !strings.Contains(element, "FAILED") {
				t.Errorf("Expected modern output to contain %q", element)
			}
		}
	}

	// Verify modern output characteristics
	t.Run("No emoji usage", func(t *testing.T) {
		emojiPatterns := []string{"🚀", "📄", "✨", "❌", "⚠️", "✅"}
		for _, emoji := range emojiPatterns {
			if strings.Contains(output, emoji) {
				t.Errorf("Found emoji %q in output, should use clean symbols. Output:\n%s", emoji, output)
			}
		}
	})

	t.Run("ASCII symbols in CI mode", func(t *testing.T) {
		// In CI mode, should use ASCII alternatives
		asciiSymbols := []string{"*", "[OK]", "[X]", "[!]"}
		foundASCII := false
		for _, symbol := range asciiSymbols {
			if strings.Contains(output, symbol) {
				foundASCII = true
				break
			}
		}
		if !foundASCII {
			t.Errorf("Expected ASCII symbols in CI mode, but none found. Output:\n%s", output)
		}
	})

	t.Run("No ANSI color codes in CI", func(t *testing.T) {
		// CI mode should not contain color codes
		if strings.Contains(output, "\x1b[") {
			t.Errorf("Found ANSI color codes in CI output. Output:\n%s", output)
		}
	})

	t.Run("Structured sections", func(t *testing.T) {
		// Verify sections are properly structured
		if strings.Contains(output, "SUMMARY") {
			summaryIndex := strings.Index(output, "SUMMARY")
			separatorIndex := strings.Index(output[summaryIndex:], "───")
			if separatorIndex == -1 {
				t.Errorf("Expected separator line after SUMMARY header")
			}
		}
	})
}

// TestModernCLIOutput_InteractiveMode tests modern output in interactive terminal mode
func TestModernCLIOutput_InteractiveMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test environment
	env := NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestFile("test.txt", "Test content for interactive mode.")

	// Clear CI environment variables to simulate interactive mode
	originalVars := make(map[string]string)
	ciVars := []string{"CI", "GITHUB_ACTIONS", "CONTINUOUS_INTEGRATION"}
	for _, v := range ciVars {
		originalVars[v] = os.Getenv(v)
		_ = os.Unsetenv(v)
	}
	defer func() {
		for _, v := range ciVars {
			if originalVars[v] == "" {
				_ = os.Unsetenv(v)
			} else {
				_ = os.Setenv(v, originalVars[v])
			}
		}
	}()

	// Set UTF-8 locale for Unicode detection
	originalLang := os.Getenv("LANG")
	defer func() {
		if originalLang == "" {
			_ = os.Unsetenv("LANG")
		} else {
			_ = os.Setenv("LANG", originalLang)
		}
	}()
	_ = os.Setenv("LANG", "en_US.UTF-8")

	// Set up test flags
	flags := env.DefaultFlags
	flags.Model = []string{"gemini-3-flash"}
	flags.Instructions = "Analyze briefly:"
	flags.DryRun = true

	// Run with the test file
	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{
		env.TempDir + "/test.txt",
	})

	if err != nil {
		t.Fatalf("Failed to run thinktank: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("CLI failed with exit code %d. Stderr: %s", exitCode, stderr)
	}

	output := stdout

	t.Run("Unicode symbols or ASCII fallback", func(t *testing.T) {
		// Should have either Unicode symbols or ASCII fallback
		unicodeFound := strings.Contains(output, "✓") || strings.Contains(output, "●")
		asciiFound := strings.Contains(output, "[OK]") || strings.Contains(output, "*")

		if !unicodeFound && !asciiFound {
			t.Errorf("Expected either Unicode or ASCII symbols, found neither. Output:\n%s", output)
		}
	})

	t.Run("Color codes in interactive mode", func(t *testing.T) {
		// Interactive mode may contain color codes (depending on terminal detection)
		// This is environment-dependent, so we just log the result
		hasColors := strings.Contains(output, "\x1b[")
		t.Logf("Interactive mode color detection: %v", hasColors)
	})
}

// TestModernCLIOutput_ErrorScenarios tests modern output with various error conditions
func TestModernCLIOutput_ErrorScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	env := NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestFile("test.txt", "Test content")

	// Set CI mode for consistent output
	_ = os.Setenv("CI", "true")
	defer func() { _ = os.Unsetenv("CI") }()

	t.Run("Invalid model names", func(t *testing.T) {
		// Set up test flags for invalid model
		flags := env.DefaultFlags
		flags.Model = []string{"nonexistent-model-123"}
		flags.Instructions = "Test:"
		flags.DryRun = true

		stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{
			env.TempDir + "/test.txt",
		})
		_ = err // May be nil in mock environment

		// Should handle invalid models gracefully with clean error output
		output := stdout + stderr

		// Verify clean error presentation
		if exitCode != 0 {
			// Should show clean error format
			errorIndicators := []string{"error", "Error", "ERROR", "invalid", "not found"}
			foundError := false
			for _, indicator := range errorIndicators {
				if strings.Contains(output, indicator) {
					foundError = true
					break
				}
			}
			if !foundError {
				t.Errorf("Expected clean error messaging for invalid model")
			}
		}
	})

	t.Run("Missing files", func(t *testing.T) {
		// Set up test flags for missing file
		flags := env.DefaultFlags
		flags.Model = []string{"gemini-3-flash"}
		flags.Instructions = "Test:"
		flags.DryRun = true

		stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{
			"/nonexistent/file.txt",
		})
		_ = err // May be nil in mock environment

		// Should handle missing files with clear error messaging
		if exitCode != 0 {
			output := stdout + stderr
			if !strings.Contains(output, "file") && !strings.Contains(output, "not found") {
				t.Errorf("Expected clear file error messaging. Output: %s", output)
			}
		}
	})
}

// TestModernCLIOutput_MultipleModels tests the modern output with multiple models to verify
// the new processing display and summary sections work correctly
func TestModernCLIOutput_MultipleModels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	env := NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestFile("document.txt", "This is a test document for analysis.")

	// Use CI mode for consistent ASCII output
	_ = os.Setenv("CI", "true")
	defer func() { _ = os.Unsetenv("CI") }()

	// Set up test flags for multiple models
	flags := env.DefaultFlags
	flags.Model = []string{"gemini-3-flash", "gemini-3-flash"}
	flags.Instructions = "Provide analysis:"
	flags.DryRun = true

	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{
		env.TempDir + "/document.txt",
	})

	if err != nil {
		t.Fatalf("Failed to run thinktank: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("CLI failed with exit code %d. Stderr: %s", exitCode, stderr)
	}

	output := stdout

	t.Run("Processing multiple models", func(t *testing.T) {
		// Should show processing for multiple models
		if !strings.Contains(output, "Processing 2 models") {
			t.Errorf("Expected 'Processing 2 models' message. Output:\n%s", output)
		}

		// Should mention both models
		models := []string{"gemini-3-flash", "gemini-3-flash"}
		for _, model := range models {
			if !strings.Contains(output, model) {
				t.Errorf("Expected model %s to be mentioned in output", model)
			}
		}
	})

	t.Run("Summary statistics", func(t *testing.T) {
		// Should show summary with statistics
		if strings.Contains(output, "SUMMARY") {
			// Look for statistics pattern
			summaryPart := output[strings.Index(output, "SUMMARY"):]
			if !strings.Contains(summaryPart, "2 models processed") {
				t.Errorf("Expected '2 models processed' in summary")
			}
		}
	})

	t.Run("Clean section formatting", func(t *testing.T) {
		// Verify sections are properly formatted with headers and separators
		sections := []string{"SUMMARY", "OUTPUT FILES"}
		for _, section := range sections {
			if strings.Contains(output, section) {
				// Find the section and verify it has a separator line after it
				sectionIndex := strings.Index(output, section)
				afterSection := output[sectionIndex+len(section):]
				if !strings.Contains(afterSection[:50], "───") { // Look in next 50 chars
					t.Errorf("Expected separator line after %s section", section)
				}
			}
		}
	})
}

// TestModernCLIOutput_TerminalWidthAdaptation tests that the output adapts to different terminal widths
func TestModernCLIOutput_TerminalWidthAdaptation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test verifies the responsive layout system works in integration
	// While we can't easily control terminal width in E2E tests, we can verify
	// that the output remains properly formatted and readable

	env := NewTestEnv(t)
	defer env.Cleanup()

	env.CreateTestFile("content.txt", "Content for testing terminal width adaptation.")

	// Use CI mode for consistent testing
	_ = os.Setenv("CI", "true")
	defer func() { _ = os.Unsetenv("CI") }()

	// Set up test flags
	flags := env.DefaultFlags
	flags.Model = []string{"gemini-3-flash"}
	flags.Instructions = "Brief analysis:"
	flags.DryRun = true

	stdout, stderr, exitCode, err := env.RunWithFlags(flags, []string{
		env.TempDir + "/content.txt",
	})

	if err != nil {
		t.Fatalf("Failed to run thinktank: %v", err)
	}

	if exitCode != 0 {
		t.Fatalf("CLI failed with exit code %d. Stderr: %s", exitCode, stderr)
	}

	output := stdout

	t.Run("Readable output format", func(t *testing.T) {
		// Verify output is readable and well-formatted
		lines := strings.Split(output, "\n")

		// Check that lines aren't excessively long (basic readability test)
		for i, line := range lines {
			if len(line) > 200 { // Reasonable maximum line length
				t.Errorf("Line %d is excessively long (%d chars): %s", i, len(line), line)
			}
		}
	})

	t.Run("Proper indentation", func(t *testing.T) {
		// Verify that sections use proper indentation
		if strings.Contains(output, "OUTPUT FILES") || strings.Contains(output, "FAILED MODELS") {
			lines := strings.Split(output, "\n")
			inFileSection := false
			for _, line := range lines {
				if strings.Contains(line, "OUTPUT FILES") || strings.Contains(line, "FAILED MODELS") {
					inFileSection = true
					continue
				}
				if inFileSection && strings.Contains(line, ".") && len(line) > 0 {
					// File lines should be indented
					if !strings.HasPrefix(line, "  ") {
						t.Errorf("Expected file line to be indented: %q", line)
					}
					break
				}
				if inFileSection && (strings.HasPrefix(line, "=") || strings.Contains(line, "SUMMARY")) {
					break // End of file section
				}
			}
		}
	})
}
