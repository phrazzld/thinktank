package logutil

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

// TestModernConsoleWriter focuses on testing the new modern clean CLI output functionality
// including Unicode fallback detection, color schemes, responsive layout, and environment adaptation.
// This complements the existing legacy tests while providing comprehensive coverage for T014-T015 features.

// TestModernConsoleWriter_UnicodeAdaptation verifies that output adapts correctly between
// Unicode and ASCII environments with proper symbol selection
func TestModernConsoleWriter_UnicodeAdaptation(t *testing.T) {
	tests := []struct {
		name                string
		isInteractive       bool
		forceUnicodeSupport bool
		expectedSuccess     string
		expectedError       string
		expectedWarning     string
		expectedBullet      string
	}{
		{
			name:            "Non-interactive uses ASCII symbols",
			isInteractive:   false,
			expectedSuccess: "[OK]",
			expectedError:   "[X]",
			expectedWarning: "[!]",
			expectedBullet:  "*",
		},
		{
			name:                "Interactive with Unicode support",
			isInteractive:       true,
			forceUnicodeSupport: true,
			expectedSuccess:     "✓",
			expectedError:       "✗",
			expectedWarning:     "⚠",
			expectedBullet:      "●",
		},
		{
			name:            "Interactive without Unicode falls back to ASCII",
			isInteractive:   true,
			expectedSuccess: "[OK]",
			expectedError:   "[X]",
			expectedWarning: "[!]",
			expectedBullet:  "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment for Unicode detection test
			if tt.forceUnicodeSupport {
				originalLang := os.Getenv("LANG")
				defer func() {
					if originalLang == "" {
						_ = os.Unsetenv("LANG")
					} else {
						_ = os.Setenv("LANG", originalLang)
					}
				}()
				_ = os.Setenv("LANG", "en_US.UTF-8")
			} else {
				// Clear Unicode indicators for fallback test
				originalVars := make(map[string]string)
				vars := []string{"LANG", "LC_ALL", "LC_CTYPE", "TERM", "WT_SESSION", "VSCODE_INJECTION"}
				for _, v := range vars {
					originalVars[v] = os.Getenv(v)
					_ = os.Unsetenv(v)
				}
				defer func() {
					for _, v := range vars {
						if originalVars[v] == "" {
							_ = os.Unsetenv(v)
						} else {
							_ = os.Setenv(v, originalVars[v])
						}
					}
				}()
			}

			// Create console writer with test environment
			// For Unicode support test, we need to ensure CI variables don't interfere
			getEnvFunc := os.Getenv
			if tt.forceUnicodeSupport {
				getEnvFunc = func(key string) string {
					// Filter out CI environment variables to allow interactive mode
					ciVars := []string{"CI", "GITHUB_ACTIONS", "CONTINUOUS_INTEGRATION",
						"GITLAB_CI", "TRAVIS", "CIRCLECI", "JENKINS_URL", "BUILDKITE"}
					for _, ciVar := range ciVars {
						if key == ciVar {
							return "" // Pretend CI variables don't exist
						}
					}
					return os.Getenv(key) // Return actual value for other vars like LANG
				}
			}

			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc:  func() bool { return tt.isInteractive },
				GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
				GetEnvFunc:      getEnvFunc,
			})

			// Test symbol provider selection
			symbols := writer.(*consoleWriter).symbols.GetSymbols()

			if symbols.Success != tt.expectedSuccess {
				t.Errorf("Expected success symbol %q, got %q", tt.expectedSuccess, symbols.Success)
			}
			if symbols.Error != tt.expectedError {
				t.Errorf("Expected error symbol %q, got %q", tt.expectedError, symbols.Error)
			}
			if symbols.Warning != tt.expectedWarning {
				t.Errorf("Expected warning symbol %q, got %q", tt.expectedWarning, symbols.Warning)
			}
			if symbols.Bullet != tt.expectedBullet {
				t.Errorf("Expected bullet symbol %q, got %q", tt.expectedBullet, symbols.Bullet)
			}
		})
	}
}

// TestModernConsoleWriter_ColorSchemeIntegration verifies that semantic colors
// are applied correctly in interactive mode and omitted in CI mode
func TestModernConsoleWriter_ColorSchemeIntegration(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Test interactive mode with colors
	interactiveWriter := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return true },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" }, // No CI environment
	})

	// Force Unicode support for consistent testing
	_ = os.Setenv("LANG", "en_US.UTF-8")
	defer func() { _ = os.Unsetenv("LANG") }()

	interactiveWriter.ModelCompleted(1, 3, "test-model", 850*time.Millisecond)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	interactiveOutput := buf.String()

	// Verify that interactive output contains ANSI color codes
	if !strings.Contains(interactiveOutput, "\x1b[") {
		t.Errorf("Expected interactive output to contain ANSI color codes, got: %s", interactiveOutput)
	}

	// Test CI mode without colors
	old = os.Stdout
	r, w, _ = os.Pipe()
	os.Stdout = w

	ciWriter := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "true" }, // CI environment
	})

	ciWriter.ModelCompleted(1, 3, "test-model", 850*time.Millisecond)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf = new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	ciOutput := buf.String()

	// Verify that CI output does not contain ANSI color codes
	if strings.Contains(ciOutput, "\x1b[") {
		t.Errorf("Expected CI output to NOT contain ANSI color codes, got: %s", ciOutput)
	}

	// Verify both outputs contain the essential content
	if !strings.Contains(interactiveOutput, "test-model") || !strings.Contains(ciOutput, "test-model") {
		t.Errorf("Both outputs should contain model name")
	}
}

// TestModernConsoleWriter_ResponsiveLayout verifies that layout adapts to different terminal widths
func TestModernConsoleWriter_ResponsiveLayout(t *testing.T) {
	widthTests := []struct {
		name     string
		width    int
		expected string // Expected behavior description
	}{
		{"Narrow terminal", 40, "narrow"},
		{"Standard terminal", 80, "standard"},
		{"Wide terminal", 120, "wide"},
		{"Very narrow terminal", 20, "minimal"},
	}

	for _, tt := range widthTests {
		t.Run(tt.name, func(t *testing.T) {
			// Test layout calculation directly (avoiding cache issues)
			layout := CalculateLayout(tt.width)
			if layout.TerminalWidth != tt.width {
				t.Errorf("Expected layout terminal width %d, got %d", tt.width, layout.TerminalWidth)
			}

			// Verify layout makes sense for the width
			totalUsed := layout.ModelNameWidth + layout.StatusWidth + layout.MinPadding
			if totalUsed > tt.width+5 { // Allow some flexibility for edge cases
				t.Errorf("Layout uses %d characters but terminal is only %d wide", totalUsed, tt.width)
			}

			// Since the console writer caches width, we test the layout system directly
			// which is the core responsive functionality
			testLayout := CalculateLayout(tt.width)
			if testLayout.TerminalWidth != tt.width {
				t.Errorf("Layout calculation failed for width %d", tt.width)
			}

			// Verify specific layout behavior for different widths
			switch {
			case tt.width < 50:
				if testLayout.MinPadding > 1 {
					t.Errorf("Narrow terminals should use minimal padding, got %d", testLayout.MinPadding)
				}
			case tt.width >= 120:
				if testLayout.ModelNameWidth < 30 {
					t.Errorf("Wide terminals should have generous model name width, got %d", testLayout.ModelNameWidth)
				}
			}
		})
	}
}

// TestModernConsoleWriter_SummaryFormatting verifies the new structured summary output
func TestModernConsoleWriter_SummaryFormatting(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false }, // Non-interactive for cleaner testing
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	summaryData := SummaryData{
		ModelsProcessed:  4,
		SuccessfulModels: 3,
		FailedModels:     1,
		SynthesisStatus:  "completed",
		OutputDirectory:  "/tmp/thinktank-output",
	}

	writer.ShowSummarySection(summaryData)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify structured summary elements (using ASCII symbols for non-interactive mode)
	expectedElements := []string{
		"SUMMARY",              // Uppercase header
		"───────",              // Separator line
		"* 4 models processed", // ASCII bullet point with stats
		"* 3 successful, 1 failed",
		"* Synthesis: [OK] completed", // ASCII success symbol
		"* Output directory: /tmp/thinktank-output",
		"[!] Partial success", // ASCII warning symbol for partial success
		"* Success rate: 75%", // Success rate calculation
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected summary to contain %q, but it was missing.\nActual output:\n%s", element, output)
		}
	}
}

// TestModernConsoleWriter_OutputFilesFormatting verifies file list formatting with human-readable sizes
func TestModernConsoleWriter_OutputFilesFormatting(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	files := []OutputFile{
		{Name: "model-a.md", Size: 1024},      // 1.0K
		{Name: "model-b.md", Size: 1536000},   // 1.5M
		{Name: "synthesis.md", Size: 2097152}, // 2.0M
	}

	writer.ShowOutputFiles(files)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify file list formatting
	expectedElements := []string{
		"OUTPUT FILES", // Uppercase header
		"────────────", // Separator line
		"model-a.md",   // File names
		"model-b.md",
		"synthesis.md",
		"1.0K", // Human-readable sizes
		"1.5M",
		"2.0M",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected output files to contain %q, but it was missing.\nActual output:\n%s", element, output)
		}
	}

	// Verify proper alignment (files should be indented)
	lines := strings.Split(output, "\n")
	fileLines := 0
	for _, line := range lines {
		if strings.Contains(line, ".md") {
			if !strings.HasPrefix(line, "  ") {
				t.Errorf("Expected file line to be indented: %q", line)
			}
			fileLines++
		}
	}

	if fileLines != 3 {
		t.Errorf("Expected 3 file lines, found %d", fileLines)
	}
}

// TestModernConsoleWriter_FailedModelsFormatting verifies failed models section formatting
func TestModernConsoleWriter_FailedModelsFormatting(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	failedModels := []FailedModel{
		{Name: "gpt-4o", Reason: "rate limited"},
		{Name: "claude-3-5-sonnet", Reason: "invalid API key"},
	}

	writer.ShowFailedModels(failedModels)

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify failed models formatting
	expectedElements := []string{
		"FAILED MODELS", // Uppercase header
		"─────────────", // Separator line
		"gpt-4o", // Model names
		"claude-3-5-sonnet",
		"rate limited", // Failure reasons
		"invalid API key",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected failed models to contain %q, but it was missing.\nActual output:\n%s", element, output)
		}
	}
}

// TestModernConsoleWriter_ProcessingLineAlignment verifies the new aligned processing display
func TestModernConsoleWriter_ProcessingLineAlignment(t *testing.T) {
	// Capture stdout for testing
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc:  func() bool { return false },
		GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
		GetEnvFunc:      func(key string) string { return "" },
	})

	// Test ShowProcessingLine
	writer.ShowProcessingLine("gemini-2.5-pro")

	// Test UpdateProcessingLine with success
	writer.UpdateProcessingLine("gemini-2.5-pro", "[OK] 2.3s")

	// Test UpdateProcessingLine with failure
	writer.UpdateProcessingLine("gpt-4o", "[X] rate limited")

	// Restore stdout and read captured output
	_ = w.Close()
	os.Stdout = old

	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Verify processing line content
	expectedElements := []string{
		"gemini-2.5-pro",   // Model name
		"processing...",    // Initial processing status
		"[OK] 2.3s",        // Success status
		"gpt-4o",           // Second model name
		"[X] rate limited", // Failure status
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected processing lines to contain %q, but it was missing.\nActual output:\n%s", element, output)
		}
	}
}

// TestModernConsoleWriter_EnvironmentDetection verifies proper environment adaptation
func TestModernConsoleWriter_EnvironmentDetection(t *testing.T) {
	envTests := []struct {
		name         string
		isTerminal   bool
		envVars      map[string]string
		expectColors bool
		expectEmoji  bool
	}{
		{
			name:         "Interactive terminal",
			isTerminal:   true,
			envVars:      map[string]string{},
			expectColors: true,
			expectEmoji:  false, // Should use Unicode symbols, not emojis
		},
		{
			name:         "CI environment",
			isTerminal:   false,
			envVars:      map[string]string{"CI": "true"},
			expectColors: false,
			expectEmoji:  false,
		},
		{
			name:         "GitHub Actions",
			isTerminal:   true,
			envVars:      map[string]string{"GITHUB_ACTIONS": "true"},
			expectColors: false,
			expectEmoji:  false,
		},
	}

	for _, tt := range envTests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc:  func() bool { return tt.isTerminal },
				GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
				GetEnvFunc: func(key string) string {
					return tt.envVars[key]
				},
			})

			// Verify interactive detection
			isInteractive := writer.IsInteractive()
			expectedInteractive := tt.expectColors // Colors imply interactive mode
			if isInteractive != expectedInteractive {
				t.Errorf("Expected IsInteractive() = %v, got %v", expectedInteractive, isInteractive)
			}

			// Test output contains/doesn't contain color codes
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			writer.SuccessMessage("test message")

			_ = w.Close()
			os.Stdout = old

			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			hasColors := strings.Contains(output, "\x1b[")
			if hasColors != tt.expectColors {
				t.Errorf("Expected colors=%v, got colors=%v in output: %q", tt.expectColors, hasColors, output)
			}
		})
	}
}

// TestShowFileOperations tests the ShowFileOperations method for proper behavior
func TestShowFileOperations(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		quiet         bool
		expectOutput  bool
		expectNewline bool
	}{
		{
			name:         "Normal file operation message",
			message:      "Processing file.txt",
			quiet:        false,
			expectOutput: true,
		},
		{
			name:          "Saving operation adds newline",
			message:       "Saving output to result.json",
			quiet:         false,
			expectOutput:  true,
			expectNewline: true,
		},
		{
			name:         "Quiet mode suppresses output",
			message:      "Processing file.txt",
			quiet:        true,
			expectOutput: false,
		},
		{
			name:         "Quiet mode suppresses saving message too",
			message:      "Saving output to result.json",
			quiet:        true,
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Create writer with quiet mode setting
			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc:  func() bool { return true },
				GetTermSizeFunc: func() (int, int, error) { return 80, 24, nil },
				GetEnvFunc:      func(key string) string { return "" },
			})

			// Set quiet mode if needed
			if tt.quiet {
				writer.(*consoleWriter).quiet = true
			}

			// Call the method
			writer.ShowFileOperations(tt.message)

			// Restore stdout
			_ = w.Close()
			os.Stdout = old

			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(r)
			output := buf.String()

			if tt.expectOutput {
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected output to contain %q, got %q", tt.message, output)
				}

				// Check for extra newline before "Saving" messages
				if tt.expectNewline {
					// Should have at least 2 newlines (one from WriteEmptyLineToConsole, one from WriteToConsole)
					newlineCount := strings.Count(output, "\n")
					if newlineCount < 2 {
						t.Errorf("Expected at least 2 newlines for 'Saving' message, got %d in output: %q", newlineCount, output)
					}
				}
			} else {
				if output != "" {
					t.Errorf("Expected no output in quiet mode, got %q", output)
				}
			}
		})
	}
}
