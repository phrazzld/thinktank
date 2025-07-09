package logutil

import (
	"encoding/json"
	"io"
	"os"
	"testing"
	"time"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Bytes (less than 1024)
		{"zero bytes", 0, "0B"},
		{"single byte", 1, "1B"},
		{"small bytes", 512, "512B"},
		{"max bytes", 1023, "1023B"},

		// Kilobytes
		{"exact 1K", 1024, "1.0K"},
		{"1.5K", 1536, "1.5K"},
		{"large K", 1048575, "1024.0K"}, // Just under 1M

		// Megabytes
		{"exact 1M", 1048576, "1.0M"},      // 1024^2
		{"4.2M", 4404019, "4.2M"},          // ~4.2MB
		{"large M", 1073741823, "1024.0M"}, // Just under 1G

		// Gigabytes
		{"exact 1G", 1073741824, "1.0G"},      // 1024^3
		{"2.5G", 2684354560, "2.5G"},          // ~2.5GB
		{"large G", 1099511627775, "1024.0G"}, // Just under 1T

		// Terabytes
		{"exact 1T", 1099511627776, "1.0T"},      // 1024^4
		{"1.2T", 1319413953331, "1.2T"},          // ~1.2TB
		{"large T", 1125899906842623, "1024.0T"}, // Just under 1P

		// Petabytes
		{"exact 1P", 1125899906842624, "1.0P"},      // 1024^5
		{"2.3P", 2589569785253478, "2.3P"},          // ~2.3PB
		{"large P", 1152921504606846975, "1024.0P"}, // Just under 1E

		// Exabytes
		{"exact 1E", 1152921504606846976, "1.0E"}, // 1024^6
		{"3.7E", 4265267724775055360, "3.7E"},     // ~3.7EB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

func TestFormatFileSize_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Negative values (edge case - should handle gracefully)
		{"negative bytes", -1, "-1B"},
		{"negative large", -1024, "-1.0K"},

		// Very large values
		{"max int64", 9223372036854775807, "8.0E"}, // Close to max int64
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}

// Test that covers specific decimal precision requirements
func TestFormatFileSize_DecimalPrecision(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		// Test that we get exactly 1 decimal place
		{"1.1K", 1126, "1.1K"},           // Should be 1.1, not 1.10
		{"2.0M", 2097152, "2.0M"},        // Should show .0 for exact values
		{"3.14159G", 3373259499, "3.1G"}, // Should round to 1 decimal
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatFileSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
			}
		})
	}
}
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"zero duration", 0, "0ms"},
		{"100ms", 100 * time.Millisecond, "100ms"},
		{"500ms", 500 * time.Millisecond, "500ms"},
		{"999ms", 999 * time.Millisecond, "999ms"},
		{"1 second", 1 * time.Second, "1.0s"},
		{"1.5 seconds", 1500 * time.Millisecond, "1.5s"},
		{"2.2 seconds", 2200 * time.Millisecond, "2.2s"},
		{"10 seconds", 10 * time.Second, "10.0s"},
		{"59.9 seconds", 59900 * time.Millisecond, "59.9s"},
		{"1 minute", 60 * time.Second, "60.0s"},
		{"2 minutes", 120 * time.Second, "120.0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("FormatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
			}
		})
	}
}

func TestFormatToWidth(t *testing.T) {
	tests := []struct {
		name          string
		message       string
		width         int
		isInteractive bool
		expected      string
	}{
		// Interactive mode tests
		{"short message interactive", "hello", 10, true, "hello"},
		{"exact width interactive", "hello", 5, true, "hello"},
		{"too long interactive", "hello world", 8, true, "hello..."},
		{"very short width interactive", "hello", 3, true, "..."},
		{"width 2 interactive", "hello", 2, true, "..."},
		{"width 1 interactive", "hello", 1, true, "..."},
		{"empty message interactive", "", 10, true, ""},

		// Non-interactive mode tests (no truncation)
		{"short message non-interactive", "hello", 10, false, "hello"},
		{"too long non-interactive", "hello world", 8, false, "hello world"},
		{"very short width non-interactive", "hello", 3, false, "hello"},
		{"empty message non-interactive", "", 10, false, ""},

		// Edge cases
		{"zero width interactive", "hello", 0, true, "..."},
		{"negative width interactive", "hello", -1, true, "..."},
		{"unicode characters", "héllo wørld", 8, true, "héll..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatToWidth(tt.message, tt.width, tt.isInteractive)
			if result != tt.expected {
				t.Errorf("FormatToWidth(%q, %d, %t) = %q, want %q",
					tt.message, tt.width, tt.isInteractive, result, tt.expected)
			}
		})
	}
}

func TestColorizeStatus(t *testing.T) {
	// Create a mock color scheme for testing
	colors := NewColorScheme(false) // Non-interactive for predictable output

	tests := []struct {
		name     string
		status   string
		expected string // We'll check if the appropriate color method was called
	}{
		{"success status", "✓ completed", "✓ completed"}, // In non-interactive, colors don't change the text
		{"error status", "✗ failed", "✗ failed"},
		{"warning status", "⚠ rate limited", "⚠ rate limited"},
		{"plain status", "processing", "processing"},
		{"success in middle", "task ✓ done", "task ✓ done"},
		{"error in middle", "task ✗ failed", "task ✗ failed"},
		{"warning in middle", "task ⚠ limited", "task ⚠ limited"},
		{"empty status", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorizeStatus(tt.status, colors)
			// In non-interactive mode, the color scheme doesn't change the text
			// but the function should still work and return the original text
			if result != tt.expected {
				t.Errorf("ColorizeStatus(%q) = %q, want %q", tt.status, result, tt.expected)
			}
		})
	}
}

func TestDetectInteractiveEnvironment(t *testing.T) {
	tests := []struct {
		name       string
		isTerminal bool
		envVars    map[string]string
		expected   bool
	}{
		{
			name:       "terminal with no CI vars",
			isTerminal: true,
			envVars:    map[string]string{},
			expected:   true,
		},
		{
			name:       "no terminal",
			isTerminal: false,
			envVars:    map[string]string{},
			expected:   false,
		},
		{
			name:       "terminal with CI=true",
			isTerminal: true,
			envVars:    map[string]string{"CI": "true"},
			expected:   false,
		},
		{
			name:       "terminal with GITHUB_ACTIONS=true",
			isTerminal: true,
			envVars:    map[string]string{"GITHUB_ACTIONS": "true"},
			expected:   false,
		},
		{
			name:       "terminal with TRAVIS=true",
			isTerminal: true,
			envVars:    map[string]string{"TRAVIS": "true"},
			expected:   false,
		},
		{
			name:       "terminal with JENKINS_URL set",
			isTerminal: true,
			envVars:    map[string]string{"JENKINS_URL": "http://jenkins.example.com"},
			expected:   false,
		},
		{
			name:       "terminal with CI=false",
			isTerminal: true,
			envVars:    map[string]string{"CI": "false"},
			expected:   true,
		},
		{
			name:       "no terminal with CI=true",
			isTerminal: false,
			envVars:    map[string]string{"CI": "true"},
			expected:   false,
		},
		{
			name:       "terminal with multiple CI vars",
			isTerminal: true,
			envVars:    map[string]string{"CI": "true", "GITHUB_ACTIONS": "true"},
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isTerminalFunc := func() bool { return tt.isTerminal }
			getEnvFunc := func(key string) string { return tt.envVars[key] }

			result := DetectInteractiveEnvironment(isTerminalFunc, getEnvFunc)
			if result != tt.expected {
				t.Errorf("DetectInteractiveEnvironment() = %t, want %t", result, tt.expected)
			}
		})
	}
}

func TestFormatConsoleMessage(t *testing.T) {
	colors := NewColorScheme(false) // Non-interactive for predictable testing
	symbols := NewSymbolProvider(false)

	tests := []struct {
		name     string
		message  string
		opts     ConsoleMessageOptions
		expected string
	}{
		{
			name:    "info message interactive",
			message: "hello world",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeInfo,
				IsInteractive: true,
				Width:         20,
				Colors:        colors,
				Symbols:       symbols,
			},
			expected: "hello world",
		},
		{
			name:    "success message interactive",
			message: "task completed",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeSuccess,
				IsInteractive: true,
				Width:         20,
				Colors:        colors,
				Symbols:       symbols,
			},
			expected: "[OK] task completed", // Non-interactive symbols use ASCII
		},
		{
			name:    "error message interactive",
			message: "task failed",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeError,
				IsInteractive: true,
				Width:         20,
				Colors:        colors,
				Symbols:       symbols,
			},
			expected: "[X] task failed", // Non-interactive symbols use ASCII
		},
		{
			name:    "warning message interactive",
			message: "rate limited",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeWarning,
				IsInteractive: true,
				Width:         20,
				Colors:        colors,
				Symbols:       symbols,
			},
			expected: "[!] rate limited", // Non-interactive symbols use ASCII
		},
		{
			name:    "status message interactive",
			message: "processing",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeStatus,
				IsInteractive: true,
				Width:         20,
				Colors:        colors,
				Symbols:       symbols,
			},
			expected: "* processing", // Non-interactive symbols use ASCII
		},
		{
			name:    "success message non-interactive",
			message: "task completed",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeSuccess,
				IsInteractive: false,
				Width:         20,
				Colors:        nil,
				Symbols:       nil,
			},
			expected: "SUCCESS: task completed",
		},
		{
			name:    "error message non-interactive",
			message: "task failed",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeError,
				IsInteractive: false,
				Width:         20,
				Colors:        nil,
				Symbols:       nil,
			},
			expected: "ERROR: task failed",
		},
		{
			name:    "warning message non-interactive",
			message: "rate limited",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeWarning,
				IsInteractive: false,
				Width:         20,
				Colors:        nil,
				Symbols:       nil,
			},
			expected: "WARNING: rate limited",
		},
		{
			name:    "long message truncated",
			message: "this is a very long message that should be truncated",
			opts: ConsoleMessageOptions{
				Type:          MessageTypeInfo,
				IsInteractive: true,
				Width:         20,
				Colors:        colors,
				Symbols:       symbols,
			},
			expected: "this is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatConsoleMessage(tt.message, tt.opts)
			if result != tt.expected {
				t.Errorf("FormatConsoleMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestFormatJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		pretty   bool
		expected string
		wantErr  bool
	}{
		{
			name: "simple object compact",
			data: map[string]interface{}{
				"name": "test",
				"age":  25,
			},
			pretty:   false,
			expected: `{"age":25,"name":"test"}`,
			wantErr:  false,
		},
		{
			name: "simple object pretty",
			data: map[string]interface{}{
				"name": "test",
				"age":  25,
			},
			pretty: true,
			expected: `{
  "age": 25,
  "name": "test"
}`,
			wantErr: false,
		},
		{
			name:     "array compact",
			data:     []string{"a", "b", "c"},
			pretty:   false,
			expected: `["a","b","c"]`,
			wantErr:  false,
		},
		{
			name:   "array pretty",
			data:   []string{"a", "b", "c"},
			pretty: true,
			expected: `[
  "a",
  "b",
  "c"
]`,
			wantErr: false,
		},
		{
			name:     "string",
			data:     "hello world",
			pretty:   false,
			expected: `"hello world"`,
			wantErr:  false,
		},
		{
			name:     "number",
			data:     42,
			pretty:   false,
			expected: `42`,
			wantErr:  false,
		},
		{
			name:     "boolean",
			data:     true,
			pretty:   false,
			expected: `true`,
			wantErr:  false,
		},
		{
			name:     "null",
			data:     nil,
			pretty:   false,
			expected: `null`,
			wantErr:  false,
		},
		{
			name:    "invalid data",
			data:    make(chan int), // channels can't be marshaled
			pretty:  false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatJSON(tt.data, tt.pretty)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FormatJSON() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("FormatJSON() unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("FormatJSON() = %q, want %q", result, tt.expected)
			}

			// Verify the result is valid JSON by unmarshaling it
			if !tt.wantErr {
				var parsed interface{}
				if err := json.Unmarshal([]byte(result), &parsed); err != nil {
					t.Errorf("FormatJSON() produced invalid JSON: %v", err)
				}
			}
		})
	}
}

// TestIOOperations tests the extracted I/O functions
func TestIOOperations(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func()
		expectedOut string
		expectedErr string
	}{
		{
			name: "WriteToConsole",
			testFunc: func() {
				WriteToConsole("test message")
			},
			expectedOut: "test message\n",
		},
		{
			name: "WriteToConsoleF",
			testFunc: func() {
				WriteToConsoleF("test %s %d", "format", 42)
			},
			expectedOut: "test format 42",
		},
		{
			name: "WriteLineToConsole",
			testFunc: func() {
				WriteLineToConsole("line message")
			},
			expectedOut: "line message\n",
		},
		{
			name: "WriteEmptyLineToConsole",
			testFunc: func() {
				WriteEmptyLineToConsole()
			},
			expectedOut: "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Run the test function
			tt.testFunc()

			// Restore stdout and read captured output
			_ = w.Close()
			os.Stdout = oldStdout
			captured, _ := io.ReadAll(r)

			if string(captured) != tt.expectedOut {
				t.Errorf("expected output %q, got %q", tt.expectedOut, string(captured))
			}
		})
	}
}

// TestStderrOperations tests stderr I/O functions
func TestStderrOperations(t *testing.T) {
	tests := []struct {
		name        string
		testFunc    func()
		expectedErr string
	}{
		{
			name: "WriteToStderr",
			testFunc: func() {
				WriteToStderr("error message")
			},
			expectedErr: "error message\n",
		},
		{
			name: "WriteToStderrF",
			testFunc: func() {
				WriteToStderrF("error %s %d", "format", 123)
			},
			expectedErr: "error format 123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			// Run the test function
			tt.testFunc()

			// Restore stderr and read captured output
			_ = w.Close()
			os.Stderr = oldStderr
			captured, _ := io.ReadAll(r)

			if string(captured) != tt.expectedErr {
				t.Errorf("expected stderr output %q, got %q", tt.expectedErr, string(captured))
			}
		})
	}
}
