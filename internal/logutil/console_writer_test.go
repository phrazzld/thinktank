package logutil

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// captureOutput captures stdout for the duration of a function call.
func captureOutput(f func()) string {
	var mu sync.Mutex
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	var buf bytes.Buffer
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		mu.Lock()
		defer mu.Unlock()
		_, _ = io.Copy(&buf, r)
		wg.Done()
	}()

	f()

	_ = w.Close()
	wg.Wait()
	os.Stdout = old
	return buf.String()
}

func TestConsoleWriter_OutputFormatting(t *testing.T) {
	testCases := []struct {
		name          string
		isInteractive bool
		quiet         bool
		noProgress    bool
		action        func(cw ConsoleWriter)
		expected      string
		notExpected   string
	}{
		// --- Interactive Mode ---
		{
			name:          "Interactive StartProcessing",
			isInteractive: true,
			action:        func(cw ConsoleWriter) { cw.StartProcessing(3) },
			expected:      "🚀 Processing 3 models...\n",
		},
		{
			name:          "Interactive ModelCompleted Success",
			isInteractive: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelCompleted("test-model", 1, 850*time.Millisecond, nil)
			},
			expected: "[1/3] test-model: ✓ completed (850ms)\n",
		},
		{
			name:          "Interactive ModelCompleted Failure",
			isInteractive: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelCompleted("test-model", 1, 1200*time.Millisecond, errors.New("API error"))
			},
			expected: "[1/3] test-model: ✗ failed (API error)\n",
		},
		{
			name:          "Interactive ModelStarted",
			isInteractive: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelStarted("test-model", 1)
			},
			expected: "[1/3] test-model: processing...\n",
		},
		{
			name:          "Interactive ModelRateLimited",
			isInteractive: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelRateLimited("test-model", 1, 2*time.Second)
			},
			expected: "[1/3] test-model: rate limited, waiting 2.0s...\n",
		},
		{
			name:          "Interactive SynthesisStarted",
			isInteractive: true,
			action:        func(cw ConsoleWriter) { cw.SynthesisStarted() },
			expected:      "📄 Synthesizing results...\n",
		},
		{
			name:          "Interactive SynthesisCompleted",
			isInteractive: true,
			action:        func(cw ConsoleWriter) { cw.SynthesisCompleted("output/path") },
			expected:      "✨ Done! Output saved to: output/path\n",
		},
		{
			name:          "Interactive StatusMessage",
			isInteractive: true,
			action:        func(cw ConsoleWriter) { cw.StatusMessage("Test status") },
			expected:      "📁 Test status\n",
		},
		// --- CI/CD (Non-Interactive) Mode ---
		{
			name:          "CI StartProcessing",
			isInteractive: false,
			action:        func(cw ConsoleWriter) { cw.StartProcessing(3) },
			expected:      "Starting processing with 3 models\n",
		},
		{
			name:          "CI ModelCompleted Success",
			isInteractive: false,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelCompleted("test-model", 1, 1200*time.Millisecond, nil)
			},
			expected: "Completed model 1/3: test-model (1.2s)\n",
		},
		{
			name:          "CI ModelCompleted Failure",
			isInteractive: false,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelCompleted("test-model", 1, 1200*time.Millisecond, errors.New("API error"))
			},
			expected: "Failed model 1/3: test-model (API error)\n",
		},
		{
			name:          "CI ModelStarted",
			isInteractive: false,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelStarted("test-model", 1)
			},
			expected: "Processing model 1/3: test-model\n",
		},
		{
			name:          "CI ModelRateLimited",
			isInteractive: false,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelRateLimited("test-model", 1, 2*time.Second)
			},
			expected: "Rate limited for model 1/3: test-model (waiting 2.0s)\n",
		},
		{
			name:          "CI SynthesisStarted",
			isInteractive: false,
			action:        func(cw ConsoleWriter) { cw.SynthesisStarted() },
			expected:      "Starting synthesis\n",
		},
		{
			name:          "CI SynthesisCompleted",
			isInteractive: false,
			action:        func(cw ConsoleWriter) { cw.SynthesisCompleted("output/path") },
			expected:      "Synthesis complete. Output: output/path\n",
		},
		{
			name:          "CI StatusMessage",
			isInteractive: false,
			action:        func(cw ConsoleWriter) { cw.StatusMessage("Test status") },
			expected:      "Test status\n",
		},
		// --- Flag-based Behavior ---
		{
			name:     "Quiet mode suppresses output",
			quiet:    true,
			action:   func(cw ConsoleWriter) { cw.StartProcessing(3) },
			expected: "",
		},
		{
			name:  "Quiet mode suppresses progress",
			quiet: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelStarted("test-model", 1)
			},
			expected: "",
		},
		{
			name:     "Quiet mode suppresses synthesis",
			quiet:    true,
			action:   func(cw ConsoleWriter) { cw.SynthesisStarted() },
			expected: "",
		},
		{
			name:          "NoProgress suppresses success ModelCompleted",
			noProgress:    true,
			isInteractive: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelCompleted("test-model", 1, 850*time.Millisecond, nil)
			},
			notExpected: "✓ completed",
		},
		{
			name:          "NoProgress suppresses ModelStarted but shows StartProcessing",
			noProgress:    true,
			isInteractive: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelStarted("test-model", 1)
			},
			expected:    "🚀 Processing 3 models...\n",
			notExpected: "processing...",
		},
		{
			name:          "NoProgress still shows failures",
			noProgress:    true,
			isInteractive: true,
			action: func(cw ConsoleWriter) {
				cw.StartProcessing(3)
				cw.ModelCompleted("test-model", 1, 1200*time.Millisecond, errors.New("API error"))
			},
			expected: "✗ failed",
		},
		{
			name:          "NoProgress still shows StartProcessing",
			noProgress:    true,
			isInteractive: true,
			action:        func(cw ConsoleWriter) { cw.StartProcessing(3) },
			expected:      "🚀 Processing 3 models...\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return tc.isInteractive },
				// Mock environment function to prevent CI detection interference
				GetEnvFunc: func(key string) string { return "" },
			})
			cw.SetQuiet(tc.quiet)
			cw.SetNoProgress(tc.noProgress)

			output := captureOutput(func() { tc.action(cw) })

			if tc.expected != "" && !strings.Contains(output, tc.expected) {
				t.Errorf("Expected output to contain %q, but got %q", tc.expected, output)
			}
			if tc.expected == "" && tc.notExpected == "" && output != "" {
				t.Errorf("Expected no output, but got %q", output)
			}
			if tc.notExpected != "" && strings.Contains(output, tc.notExpected) {
				t.Errorf("Expected output to NOT contain %q, but it did: %q", tc.notExpected, output)
			}
		})
	}
}

func TestDetectInteractiveEnvironment(t *testing.T) {
	ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "TRAVIS", "CIRCLECI", "JENKINS_URL", "CONTINUOUS_INTEGRATION", "BUILDKITE"}

	for _, envVar := range ciVars {
		t.Run("Terminal with "+envVar, func(t *testing.T) {
			originalValue := os.Getenv(envVar)
			defer func() {
				if originalValue != "" {
					_ = os.Setenv(envVar, originalValue)
				} else {
					_ = os.Unsetenv(envVar)
				}
			}()

			_ = os.Setenv(envVar, "true")
			result := detectInteractiveEnvironment(func() bool { return true })
			if result {
				t.Errorf("Expected interactive=false when %s is set, got true", envVar)
			}
		})
	}

	t.Run("Terminal without CI vars", func(t *testing.T) {
		ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "TRAVIS", "CIRCLECI", "JENKINS_URL", "CONTINUOUS_INTEGRATION", "BUILDKITE"}
		originalValues := make(map[string]string)
		for _, envVar := range ciVars {
			originalValues[envVar] = os.Getenv(envVar)
			_ = os.Unsetenv(envVar)
		}
		defer func() {
			for envVar, value := range originalValues {
				if value != "" {
					_ = os.Setenv(envVar, value)
				}
			}
		}()

		result := detectInteractiveEnvironment(func() bool { return true })
		if !result {
			t.Errorf("Expected interactive=true in a terminal with no CI vars, got false")
		}
	})

	t.Run("Non-terminal without CI vars", func(t *testing.T) {
		ciVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "TRAVIS", "CIRCLECI", "JENKINS_URL", "CONTINUOUS_INTEGRATION", "BUILDKITE"}
		originalValues := make(map[string]string)
		for _, envVar := range ciVars {
			originalValues[envVar] = os.Getenv(envVar)
			_ = os.Unsetenv(envVar)
		}
		defer func() {
			for envVar, value := range originalValues {
				if value != "" {
					_ = os.Setenv(envVar, value)
				}
			}
		}()

		result := detectInteractiveEnvironment(func() bool { return false })
		if result {
			t.Errorf("Expected interactive=false for non-terminal, got true")
		}
	})
}

func TestConsoleWriter_ConcurrencySafety(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
		GetEnvFunc:     func(key string) string { return "" },
	})

	cw.StartProcessing(10)

	var wg sync.WaitGroup
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			modelName := "concurrent-model"
			cw.ModelStarted(modelName, index+1)

			time.Sleep(1 * time.Millisecond)

			if index%3 == 0 {
				cw.ModelCompleted(modelName, index+1, 100*time.Millisecond, errors.New("simulated error"))
			} else {
				cw.ModelCompleted(modelName, index+1, 100*time.Millisecond, nil)
			}
		}(i)
	}

	wg.Wait()
}

func TestConsoleWriter_SettersAndGetters(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
		GetEnvFunc:     func(key string) string { return "" },
	})

	if !cw.IsInteractive() {
		t.Error("Expected IsInteractive to be true")
	}

	cw.SetQuiet(true)
	output := captureOutput(func() { cw.StartProcessing(1) })
	if output != "" {
		t.Errorf("Expected no output when quiet=true, got: %q", output)
	}

	cw.SetQuiet(false)
	output = captureOutput(func() { cw.StartProcessing(1) })
	if output == "" {
		t.Error("Expected output when quiet=false")
	}

	cw.SetNoProgress(true)
	output = captureOutput(func() {
		cw.StartProcessing(1)
		cw.ModelStarted("test", 1)
	})
	if strings.Contains(output, "processing...") {
		t.Errorf("Expected no progress output when noProgress=true, got: %q", output)
	}

	cw.SetNoProgress(false)
	output = captureOutput(func() {
		cw.StartProcessing(1)
		cw.ModelStarted("test", 1)
	})
	if !strings.Contains(output, "processing...") {
		t.Errorf("Expected progress output when noProgress=false, got: %q", output)
	}
}

func TestConsoleWriter_NonInteractiveEnvironment(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false },
		GetEnvFunc:     func(key string) string { return "" },
	})

	if cw.IsInteractive() {
		t.Error("Expected IsInteractive to be false for non-terminal")
	}

	output := captureOutput(func() {
		cw.StartProcessing(1)
		cw.ModelCompleted("test", 1, 500*time.Millisecond, nil)
		cw.SynthesisStarted()
		cw.SynthesisCompleted("output/path")
	})

	if strings.Contains(output, "🚀") || strings.Contains(output, "✓") || strings.Contains(output, "📄") || strings.Contains(output, "✨") {
		t.Errorf("Non-interactive mode should not contain emojis, got: %q", output)
	}

	expectedStrings := []string{
		"Starting processing with 1 models",
		"Completed model 1/1: test",
		"Starting synthesis",
		"Synthesis complete. Output: output/path",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected CI output to contain %q, got: %q", expected, output)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{1 * time.Second, "1.0s"},
		{1500 * time.Millisecond, "1.5s"},
		{2200 * time.Millisecond, "2.2s"},
		{100 * time.Millisecond, "100ms"},
		{0, "0ms"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatDuration(tc.duration)
			if result != tc.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tc.duration, result, tc.expected)
			}
		})
	}
}

func TestNewConsoleWriter_DefaultBehavior(t *testing.T) {
	cw := NewConsoleWriter()
	if cw == nil {
		t.Fatal("NewConsoleWriter returned nil")
	}

	output := captureOutput(func() {
		cw.StartProcessing(1)
		cw.StatusMessage("Test message")
	})

	if output == "" {
		t.Error("Expected some output from default ConsoleWriter")
	}
}

func TestConsoleWriter_ModelQueued(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
		GetEnvFunc:     func(key string) string { return "" },
	})

	output := captureOutput(func() {
		cw.ModelQueued("test-model", 1)
	})

	if output != "" {
		t.Logf("ModelQueued produced output: %q", output)
	}

	cw.SetNoProgress(true)
	output = captureOutput(func() {
		cw.ModelQueued("test-model", 1)
	})

	if output != "" {
		t.Errorf("ModelQueued should not produce output with noProgress=true, got: %q", output)
	}
}

func TestConsoleWriter_TerminalWidth(t *testing.T) {
	tests := []struct {
		name           string
		terminalWidth  int
		terminalHeight int
		getTermSizeErr error
		isTerminal     bool
		expectedWidth  int
	}{
		{
			name:           "Normal terminal width",
			terminalWidth:  100,
			terminalHeight: 30,
			getTermSizeErr: nil,
			isTerminal:     true,
			expectedWidth:  100,
		},
		{
			name:           "Very narrow terminal",
			terminalWidth:  5,
			terminalHeight: 30,
			getTermSizeErr: nil,
			isTerminal:     true,
			expectedWidth:  5, // Should be allowed for testing edge cases
		},
		{
			name:           "Very wide terminal",
			terminalWidth:  200,
			terminalHeight: 30,
			getTermSizeErr: nil,
			isTerminal:     true,
			expectedWidth:  MaxTerminalWidth, // Should be clamped to maximum
		},
		{
			name:           "Non-terminal environment",
			terminalWidth:  0,
			terminalHeight: 0,
			getTermSizeErr: nil,
			isTerminal:     false,
			expectedWidth:  DefaultTerminalWidth,
		},
		{
			name:           "Terminal size detection error",
			terminalWidth:  0,
			terminalHeight: 0,
			getTermSizeErr: fmt.Errorf("cannot detect size"),
			isTerminal:     true,
			expectedWidth:  DefaultTerminalWidth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return tt.isTerminal },
				GetTermSizeFunc: func() (int, int, error) {
					return tt.terminalWidth, tt.terminalHeight, tt.getTermSizeErr
				},
				GetEnvFunc: func(key string) string { return "" },
			})

			width := cw.GetTerminalWidth()
			if width != tt.expectedWidth {
				t.Errorf("GetTerminalWidth() = %d, want %d", width, tt.expectedWidth)
			}

			// Test caching behavior - second call should return same value
			width2 := cw.GetTerminalWidth()
			if width2 != width {
				t.Errorf("GetTerminalWidth() caching failed: first=%d, second=%d", width, width2)
			}
		})
	}
}

func TestConsoleWriter_FormatMessage(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		terminalWidth int
		isInteractive bool
		expectedMsg   string
	}{
		{
			name:          "Short message fits",
			message:       "Hello world",
			terminalWidth: 80,
			isInteractive: true,
			expectedMsg:   "Hello world",
		},
		{
			name:          "Long message truncated in interactive mode",
			message:       "This is a very long message that should be truncated",
			terminalWidth: 20,
			isInteractive: true,
			expectedMsg:   "This is a very lo...", // 20 chars total
		},
		{
			name:          "Long message not truncated in non-interactive mode",
			message:       "This is a very long message that should not be truncated",
			terminalWidth: 20,
			isInteractive: false,
			expectedMsg:   "This is a very long message that should not be truncated",
		},
		{
			name:          "Very narrow terminal",
			message:       "Hello",
			terminalWidth: 3,
			isInteractive: true,
			expectedMsg:   "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return tt.isInteractive },
				GetTermSizeFunc: func() (int, int, error) {
					return tt.terminalWidth, 24, nil
				},
				GetEnvFunc: func(key string) string { return "" },
			})

			formatted := cw.FormatMessage(tt.message)
			if formatted != tt.expectedMsg {
				t.Errorf("FormatMessage() = %q, want %q", formatted, tt.expectedMsg)
			}
		})
	}
}

func TestConsoleWriter_ErrorWarningSuccessMessages(t *testing.T) {
	tests := []struct {
		name          string
		method        string
		message       string
		isInteractive bool
		expectedRegex string
	}{
		{
			name:          "Error message interactive",
			method:        "ErrorMessage",
			message:       "Something went wrong",
			isInteractive: true,
			expectedRegex: `❌ Something went wrong`,
		},
		{
			name:          "Error message non-interactive",
			method:        "ErrorMessage",
			message:       "Something went wrong",
			isInteractive: false,
			expectedRegex: `ERROR: Something went wrong`,
		},
		{
			name:          "Warning message interactive",
			method:        "WarningMessage",
			message:       "This is a warning",
			isInteractive: true,
			expectedRegex: `⚠️  This is a warning`,
		},
		{
			name:          "Warning message non-interactive",
			method:        "WarningMessage",
			message:       "This is a warning",
			isInteractive: false,
			expectedRegex: `WARNING: This is a warning`,
		},
		{
			name:          "Success message interactive",
			method:        "SuccessMessage",
			message:       "Operation completed",
			isInteractive: true,
			expectedRegex: `✅ Operation completed`,
		},
		{
			name:          "Success message non-interactive",
			method:        "SuccessMessage",
			message:       "Operation completed",
			isInteractive: false,
			expectedRegex: `SUCCESS: Operation completed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return tt.isInteractive },
				GetEnvFunc:     func(key string) string { return "" },
			})

			var output string
			switch tt.method {
			case "ErrorMessage":
				output = captureOutput(func() {
					cw.ErrorMessage(tt.message)
				})
			case "WarningMessage":
				output = captureOutput(func() {
					cw.WarningMessage(tt.message)
				})
			case "SuccessMessage":
				output = captureOutput(func() {
					cw.SuccessMessage(tt.message)
				})
			}

			if !strings.Contains(output, tt.expectedRegex) {
				t.Errorf("%s output %q does not contain expected pattern %q", tt.method, output, tt.expectedRegex)
			}
		})
	}
}

func TestConsoleWriter_QuietModeWithNewMethods(t *testing.T) {
	cw := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
		GetEnvFunc:     func(key string) string { return "" },
	})
	cw.SetQuiet(true)

	// Test that errors and warnings are STILL SHOWN in quiet mode (as documented)
	t.Run("ErrorMessage shows in quiet mode", func(t *testing.T) {
		output := captureOutput(func() { cw.ErrorMessage("test error") })
		if output == "" {
			t.Error("ErrorMessage should produce output in quiet mode (errors are essential)")
		}
		if !strings.Contains(output, "test error") {
			t.Errorf("ErrorMessage output should contain error message, got: %q", output)
		}
	})

	t.Run("WarningMessage shows in quiet mode", func(t *testing.T) {
		output := captureOutput(func() { cw.WarningMessage("test warning") })
		if output == "" {
			t.Error("WarningMessage should produce output in quiet mode (warnings are essential)")
		}
		if !strings.Contains(output, "test warning") {
			t.Errorf("WarningMessage output should contain warning message, got: %q", output)
		}
	})

	// Test that success messages are suppressed in quiet mode
	t.Run("SuccessMessage suppressed in quiet mode", func(t *testing.T) {
		output := captureOutput(func() { cw.SuccessMessage("test success") })
		if output != "" {
			t.Errorf("SuccessMessage should not produce output in quiet mode, got: %q", output)
		}
	})

	// Test that ModelCompleted errors are shown in quiet mode
	t.Run("ModelCompleted error shows in quiet mode", func(t *testing.T) {
		cw.StartProcessing(1)
		testErr := fmt.Errorf("model failed")
		output := captureOutput(func() {
			cw.ModelCompleted("test-model", 1, time.Millisecond*100, testErr)
		})
		if output == "" {
			t.Error("ModelCompleted error should produce output in quiet mode (errors are essential)")
		}
		if !strings.Contains(output, "model failed") {
			t.Errorf("ModelCompleted error output should contain error message, got: %q", output)
		}
	})

	// Test that ModelCompleted success is suppressed in quiet mode
	t.Run("ModelCompleted success suppressed in quiet mode", func(t *testing.T) {
		cw.StartProcessing(1)
		output := captureOutput(func() {
			cw.ModelCompleted("test-model", 1, time.Millisecond*100, nil)
		})
		if output != "" {
			t.Errorf("ModelCompleted success should not produce output in quiet mode, got: %q", output)
		}
	})
}
