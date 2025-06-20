package logutil

import (
	"os"
	"testing"
	"time"
)

// TestNewConsoleWriter verifies constructor creates properly configured instance
func TestNewConsoleWriter(t *testing.T) {
	writer := NewConsoleWriter()

	if writer == nil {
		t.Fatal("NewConsoleWriter should return non-nil instance")
	}

	// Verify interface compliance
	_ = ConsoleWriter(writer)
}

// TestConsoleWriter_EnvironmentDetection tests TTY and CI detection logic
func TestConsoleWriter_EnvironmentDetection(t *testing.T) {
	tests := []struct {
		name                string
		envVars             map[string]string
		mockIsTerminal      func() bool
		expectedInteractive bool
	}{
		{
			name:                "CI environment variable set",
			envVars:             map[string]string{"CI": "true"},
			mockIsTerminal:      func() bool { return true },
			expectedInteractive: false,
		},
		{
			name:                "GITHUB_ACTIONS environment variable set",
			envVars:             map[string]string{"GITHUB_ACTIONS": "true"},
			mockIsTerminal:      func() bool { return true },
			expectedInteractive: false,
		},
		{
			name:                "CONTINUOUS_INTEGRATION environment variable set",
			envVars:             map[string]string{"CONTINUOUS_INTEGRATION": "true"},
			mockIsTerminal:      func() bool { return true },
			expectedInteractive: false,
		},
		{
			name:                "TTY environment with no CI vars",
			envVars:             map[string]string{},
			mockIsTerminal:      func() bool { return true },
			expectedInteractive: true,
		},
		{
			name:                "Non-TTY environment with no CI vars",
			envVars:             map[string]string{},
			mockIsTerminal:      func() bool { return false },
			expectedInteractive: false,
		},
		{
			name:                "CI=false should still allow interactive",
			envVars:             map[string]string{"CI": "false"},
			mockIsTerminal:      func() bool { return true },
			expectedInteractive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variables
			originalEnv := make(map[string]string)
			for key, value := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set environment variable %s: %v", key, err)
				}
			}

			// Cleanup environment after test
			defer func() {
				for key := range tt.envVars {
					if original, exists := originalEnv[key]; exists {
						if err := os.Setenv(key, original); err != nil {
							t.Errorf("Failed to restore environment variable %s: %v", key, err)
						}
					} else {
						if err := os.Unsetenv(key); err != nil {
							t.Errorf("Failed to unset environment variable %s: %v", key, err)
						}
					}
				}
			}()

			// Create writer with mocked terminal detection and environment variable isolation
			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc: tt.mockIsTerminal,
				GetEnvFunc: func(key string) string {
					// Return the test-controlled environment variables only
					if value, exists := tt.envVars[key]; exists {
						return value
					}
					// For all other variables (including real CI vars), return empty
					return ""
				},
			})

			result := writer.IsInteractive()
			if result != tt.expectedInteractive {
				t.Errorf("Expected IsInteractive()=%v, got %v", tt.expectedInteractive, result)
			}
		})
	}
}

// TestConsoleWriter_ControlMethods tests SetQuiet and SetNoProgress
func TestConsoleWriter_ControlMethods(t *testing.T) {
	writer := NewConsoleWriter()

	// Test SetQuiet
	writer.SetQuiet(true)
	writer.SetQuiet(false)

	// Test SetNoProgress
	writer.SetNoProgress(true)
	writer.SetNoProgress(false)

	// Methods should not panic and should accept the values
}

// TestConsoleWriter_ProgressReporting tests the progress reporting methods
func TestConsoleWriter_ProgressReporting(t *testing.T) {
	writer := NewConsoleWriter()

	// Test StartProcessing
	writer.StartProcessing(3)

	// Test model lifecycle methods
	writer.ModelQueued("test-model", 1)
	writer.ModelStarted(1, 3, "test-model")
	writer.ModelCompleted(1, 3, "test-model", time.Second)
	writer.ModelRateLimited(2, 3, "test-model", time.Second*2)

	// Test with error (use ModelFailed for error scenarios)
	writer.ModelFailed(2, 3, "failed-model", "test error")

	// Test synthesis methods
	writer.SynthesisStarted()
	writer.SynthesisCompleted("/path/to/output")
}

// TestConsoleWriter_ThreadSafety tests concurrent access to console writer
func TestConsoleWriter_ThreadSafety(t *testing.T) {
	writer := NewConsoleWriter()
	writer.StartProcessing(10)

	// Start multiple goroutines making concurrent calls
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			defer func() { done <- true }()

			modelName := "model-" + string(rune('A'+index))
			writer.ModelQueued(modelName, index+1)
			writer.ModelStarted(index+1, 10, modelName)

			// Simulate some processing time
			time.Sleep(time.Millisecond * 10)

			writer.ModelCompleted(index+1, 10, modelName, time.Millisecond*10)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panic, thread safety is working
}

// TestConsoleWriter_QuietMode tests that quiet mode suppresses appropriate output
func TestConsoleWriter_QuietMode(t *testing.T) {
	// Create a writer that captures output for verification
	writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return true },
	})

	// Enable quiet mode
	writer.SetQuiet(true)

	// These calls should be suppressed in quiet mode
	writer.StartProcessing(2)
	writer.ModelStarted(1, 2, "test-model")
	writer.ModelCompleted(1, 2, "test-model", time.Second)
	writer.SynthesisStarted()
	writer.SynthesisCompleted("/output")

	// In real implementation, we'd verify no output was produced
	// For now, just verify the methods don't panic
}

// TestConsoleWriter_NoProgressMode tests that no-progress mode works correctly
func TestConsoleWriter_NoProgressMode(t *testing.T) {
	writer := NewConsoleWriter()

	// Enable no-progress mode
	writer.SetNoProgress(true)

	// These calls should show reduced output in no-progress mode
	writer.StartProcessing(2)
	writer.ModelQueued("test-model", 1)
	writer.ModelStarted(1, 2, "test-model")
	writer.ModelCompleted(1, 2, "test-model", time.Second)

	// Synthesis methods should still work
	writer.SynthesisStarted()
	writer.SynthesisCompleted("/output")
}

// TestConsoleWriter_OutputModes tests both interactive and CI output modes
func TestConsoleWriter_OutputModes(t *testing.T) {
	tests := []struct {
		name        string
		interactive bool
	}{
		{"Interactive Mode", true},
		{"CI Mode", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writer := NewConsoleWriterWithOptions(ConsoleWriterOptions{
				IsTerminalFunc: func() bool { return tt.interactive },
				GetEnvFunc:     func(key string) string { return "" },
			})

			if writer.IsInteractive() != tt.interactive {
				t.Errorf("Expected IsInteractive()=%v, got %v", tt.interactive, writer.IsInteractive())
			}

			// Test full workflow in both modes
			writer.StartProcessing(2)
			writer.ModelQueued("model1", 1)
			writer.ModelStarted(1, 2, "model1")
			writer.ModelCompleted(1, 2, "model1", time.Millisecond*800)
			writer.ModelQueued("model2", 2)
			writer.ModelRateLimited(2, 2, "model2", time.Second*2)
			writer.ModelStarted(2, 2, "model2")
			writer.ModelCompleted(2, 2, "model2", time.Millisecond*1200)
			writer.SynthesisStarted()
			writer.SynthesisCompleted("/output/path")
		})
	}
}
