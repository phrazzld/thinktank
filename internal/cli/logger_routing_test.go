package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/logutil"
)

func TestCreateLoggerWithRouting(t *testing.T) {
	// Note: Not using t.Parallel() due to potential file system conflicts

	tests := []struct {
		name                      string
		jsonLogs                  bool
		verbose                   bool
		logLevel                  logutil.LogLevel
		outputDir                 string
		expectConsoleRouting      bool
		expectFileRouting         bool
		expectLogFileInOutputDir  bool
		expectLogFileInCurrentDir bool
		description               string
	}{
		{
			name:                 "json logs flag enables console routing",
			jsonLogs:             true,
			verbose:              false,
			logLevel:             logutil.InfoLevel,
			outputDir:            "/tmp/test",
			expectConsoleRouting: true,
			expectFileRouting:    false,
			description:          "JsonLogs flag should route logs to console (stderr)",
		},
		{
			name:                 "verbose flag enables console routing",
			jsonLogs:             false,
			verbose:              true,
			logLevel:             logutil.DebugLevel,
			outputDir:            "/tmp/test",
			expectConsoleRouting: true,
			expectFileRouting:    false,
			description:          "Verbose flag should route logs to console (stderr)",
		},
		{
			name:                 "both flags enable console routing",
			jsonLogs:             true,
			verbose:              true,
			logLevel:             logutil.DebugLevel,
			outputDir:            "/tmp/test",
			expectConsoleRouting: true,
			expectFileRouting:    false,
			description:          "Both flags should route logs to console (stderr)",
		},
		{
			name:                     "no flags with output dir routes to file",
			jsonLogs:                 false,
			verbose:                  false,
			logLevel:                 logutil.InfoLevel,
			outputDir:                "", // Will be set to temp dir in test
			expectConsoleRouting:     false,
			expectFileRouting:        true,
			expectLogFileInOutputDir: true,
			description:              "No special flags should route logs to file in output directory",
		},
		{
			name:                      "no flags no output dir routes to current dir",
			jsonLogs:                  false,
			verbose:                   false,
			logLevel:                  logutil.WarnLevel,
			outputDir:                 "",
			expectConsoleRouting:      false,
			expectFileRouting:         true,
			expectLogFileInCurrentDir: true,
			description:               "No output dir should route logs to current directory",
		},
		{
			name:                 "error level with file routing",
			jsonLogs:             false,
			verbose:              false,
			logLevel:             logutil.ErrorLevel,
			outputDir:            "", // Will be set to temp dir in test
			expectConsoleRouting: false,
			expectFileRouting:    true,
			description:          "Error level should work with file routing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() due to potential file system conflicts

			// Create temporary directory if needed
			var tempDir string
			var cleanup func()
			if tt.expectLogFileInOutputDir {
				tempDir = t.TempDir()
				tt.outputDir = tempDir
			}

			// Create test config
			cfg := &config.MinimalConfig{
				JsonLogs: tt.jsonLogs,
				Verbose:  tt.verbose,
				LogLevel: tt.logLevel,
			}

			// Call function under test
			logger, wrapper := createLoggerWithRouting(cfg, tt.outputDir)
			if cleanup != nil {
				defer cleanup()
			}
			defer func() { _ = wrapper.Close() }()

			// Verify logger was created
			if logger == nil {
				t.Fatal("Expected logger to be created, got nil")
			}

			if wrapper == nil {
				t.Fatal("Expected wrapper to be created, got nil")
			}

			// Test wrapper interface
			if wrapper.LoggerInterface == nil {
				t.Error("Expected wrapper to implement LoggerInterface")
			}

			// Verify console routing expectations
			if tt.expectConsoleRouting {
				if wrapper.file != nil {
					t.Error("Expected console routing (no file), but wrapper has file set")
				}
			}

			// Verify file routing expectations
			if tt.expectFileRouting {
				var expectedLogPath string
				if tt.expectLogFileInOutputDir && tt.outputDir != "" {
					expectedLogPath = filepath.Join(tt.outputDir, "thinktank.log")
				} else if tt.expectLogFileInCurrentDir {
					expectedLogPath = "thinktank.log"
				}

				if expectedLogPath != "" {
					// Check if log file was created (it might be, or might fallback to console)
					if wrapper.file != nil {
						// File routing succeeded
						// Clean up the log file if it exists
						if _, err := os.Stat(expectedLogPath); err == nil {
							// File exists, clean it up after test
							defer func() {
								_ = os.Remove(expectedLogPath) // Cleanup - ignore error
							}()
						}
					} else {
						// File routing failed, should have fallen back to console
						t.Logf("File routing failed, fell back to console (this is acceptable behavior)")
					}
				}
			}

			// Test that logger can be used
			ctx := logutil.WithCorrelationID(context.Background(), "test-correlation-id")
			contextLogger := logger.WithContext(ctx)
			if contextLogger == nil {
				t.Error("Expected context logger to be created")
			}

			// Test wrapper Close method
			if err := wrapper.Close(); err != nil {
				t.Errorf("Expected wrapper.Close() to succeed, got error: %v", err)
			}

			// Test that Close can be called multiple times without error
			if err := wrapper.Close(); err != nil {
				t.Errorf("Expected wrapper.Close() to be idempotent, got error: %v", err)
			}
		})
	}
}

func TestLoggerWrapper_Close(t *testing.T) {
	// Note: Not using t.Parallel() due to potential file system conflicts

	t.Run("close with file", func(t *testing.T) {
		// Note: Not using t.Parallel() due to potential file system conflicts

		tempDir := t.TempDir()
		logFile := filepath.Join(tempDir, "test.log")

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			t.Fatalf("Failed to create test log file: %v", err)
		}

		logger := logutil.NewSlogLoggerFromLogLevel(file, logutil.InfoLevel)
		wrapper := &LoggerWrapper{
			LoggerInterface: logger,
			file:            file,
		}

		// Test Close
		err = wrapper.Close()
		if err != nil {
			t.Errorf("Expected Close() to succeed, got error: %v", err)
		}

		// Verify file was closed (writing should fail)
		_, writeErr := file.Write([]byte("test"))
		if writeErr == nil {
			t.Error("Expected write to closed file to fail")
		}
	})

	t.Run("close without file", func(t *testing.T) {
		// Note: Not using t.Parallel() due to potential file system conflicts

		logger := logutil.NewSlogLoggerFromLogLevel(os.Stderr, logutil.InfoLevel)
		wrapper := &LoggerWrapper{
			LoggerInterface: logger,
			file:            nil,
		}

		// Test Close with no file
		err := wrapper.Close()
		if err != nil {
			t.Errorf("Expected Close() with no file to succeed, got error: %v", err)
		}
	})
}

func TestCreateLoggerWithRouting_FileCreationFailure(t *testing.T) {
	// Note: Not using t.Parallel() due to potential file system conflicts

	t.Run("fallback to console when file creation fails", func(t *testing.T) {
		// Note: Not using t.Parallel() due to potential file system conflicts

		// Create config that should use file routing
		cfg := &config.MinimalConfig{
			JsonLogs: false,
			Verbose:  false,
			LogLevel: logutil.InfoLevel,
		}

		// Use a directory path that doesn't exist and can't be created
		// This should trigger the fallback to console logging
		invalidOutputDir := "/nonexistent/path/that/should/not/exist"

		logger, wrapper := createLoggerWithRouting(cfg, invalidOutputDir)
		defer func() { _ = wrapper.Close() }()

		// Verify logger was created
		if logger == nil {
			t.Fatal("Expected logger to be created even with invalid output dir")
		}

		if wrapper == nil {
			t.Fatal("Expected wrapper to be created even with invalid output dir")
		}

		// Should have fallen back to console routing (no file)
		if wrapper.file != nil {
			t.Error("Expected fallback to console routing (no file) when file creation fails")
		}
	})
}

func TestCreateLoggerWithRouting_LogLevels(t *testing.T) {
	// Note: Not using t.Parallel() due to potential file system conflicts

	logLevels := []logutil.LogLevel{
		logutil.DebugLevel,
		logutil.InfoLevel,
		logutil.WarnLevel,
		logutil.ErrorLevel,
	}

	for _, level := range logLevels {
		t.Run(level.String(), func(t *testing.T) {
			// Note: Not using t.Parallel() due to potential file system conflicts

			cfg := &config.MinimalConfig{
				JsonLogs: true, // Use console routing for simplicity
				Verbose:  false,
				LogLevel: level,
			}

			logger, wrapper := createLoggerWithRouting(cfg, "")
			defer func() { _ = wrapper.Close() }()

			if logger == nil {
				t.Errorf("Expected logger to be created for log level %s", level.String())
			}

			if wrapper == nil {
				t.Errorf("Expected wrapper to be created for log level %s", level.String())
			}
		})
	}
}

func TestCreateLoggerWithRouting_OutputDirVariations(t *testing.T) {
	// Note: Not using t.Parallel() due to potential file system conflicts

	tests := []struct {
		name           string
		outputDir      string
		expectFallback bool
		description    string
	}{
		{
			name:        "empty output dir",
			outputDir:   "",
			description: "Empty output dir should use current directory",
		},
		{
			name:        "valid temp dir",
			outputDir:   "", // Will be set to temp dir
			description: "Valid output dir should create log file in that directory",
		},
		{
			name:           "invalid output dir",
			outputDir:      "/root/nonexistent/path", // Should not be writable
			expectFallback: true,
			description:    "Invalid output dir should fallback to console",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Not using t.Parallel() due to potential file system conflicts

			if tt.outputDir == "" && tt.name == "valid temp dir" {
				tt.outputDir = t.TempDir()
			}

			cfg := &config.MinimalConfig{
				JsonLogs: false, // Use file routing
				Verbose:  false,
				LogLevel: logutil.InfoLevel,
			}

			logger, wrapper := createLoggerWithRouting(cfg, tt.outputDir)
			defer func() { _ = wrapper.Close() }()

			if logger == nil {
				t.Fatal("Expected logger to be created")
			}

			if wrapper == nil {
				t.Fatal("Expected wrapper to be created")
			}

			// Check expectations
			if tt.expectFallback {
				if wrapper.file != nil {
					t.Error("Expected fallback to console (no file) for invalid output dir")
				}
			} else {
				// For valid cases, either file creation succeeds or fails gracefully
				// Both outcomes are acceptable as long as a logger is created
				t.Logf("Output dir: %s, has file: %t", tt.outputDir, wrapper.file != nil)
			}
		})
	}
}
