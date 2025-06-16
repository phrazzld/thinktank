package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
)

func TestSetupLogging_DefaultBehavior_LogsToFile(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "thinktank_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a config with output directory but no special flags
	config := &config.CliConfig{
		OutputDir: tempDir,
		LogLevel:  logutil.InfoLevel,
		Verbose:   false,
		JsonLogs:  false,
	}

	// Setup logging
	logger := SetupLogging(config)
	if logger == nil {
		t.Fatal("SetupLogging returned nil logger")
	}

	// Log a test message
	logger.Info("Test message for file logging")

	// Check if log file was created
	logFilePath := filepath.Join(tempDir, "thinktank.log")
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logFilePath)
	}

	// Read the log file and verify it contains JSON
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Error("Log file is empty")
	}

	// Check for JSON structure (should contain "level" and "msg" fields)
	if !strings.Contains(contentStr, `"level":"INFO"`) || !strings.Contains(contentStr, `"msg":"Test message for file logging"`) {
		t.Errorf("Log file doesn't contain expected JSON content. Got: %s", contentStr)
	}
}

func TestSetupLogging_JsonLogsFlag_LogsToStderr(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "thinktank_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a config with JsonLogs enabled
	config := &config.CliConfig{
		OutputDir: tempDir,
		LogLevel:  logutil.InfoLevel,
		Verbose:   false,
		JsonLogs:  true, // This should trigger stderr logging
	}

	// Setup logging
	logger := SetupLogging(config)
	if logger == nil {
		t.Fatal("SetupLogging returned nil logger")
	}

	// Check that no log file was created
	logFilePath := filepath.Join(tempDir, "thinktank.log")
	if _, err := os.Stat(logFilePath); !os.IsNotExist(err) {
		t.Errorf("Log file should not be created when JsonLogs=true, but file exists at %s", logFilePath)
	}
}

func TestSetupLogging_VerboseFlag_LogsToStderr(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "thinktank_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a config with Verbose enabled
	config := &config.CliConfig{
		OutputDir: tempDir,
		LogLevel:  logutil.InfoLevel,
		Verbose:   true, // This should trigger stderr logging
		JsonLogs:  false,
	}

	// Setup logging
	logger := SetupLogging(config)
	if logger == nil {
		t.Fatal("SetupLogging returned nil logger")
	}

	// Check that no log file was created
	logFilePath := filepath.Join(tempDir, "thinktank.log")
	if _, err := os.Stat(logFilePath); !os.IsNotExist(err) {
		t.Errorf("Log file should not be created when Verbose=true, but file exists at %s", logFilePath)
	}
}

func TestSetupLogging_EmptyOutputDir_UsesCurrentDirectory(t *testing.T) {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Create and change to a temporary directory
	tempDir, err := os.MkdirTemp("", "thinktank_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore original directory: %v", err)
		}
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to remove temp dir: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Create a config with empty output directory
	config := &config.CliConfig{
		OutputDir: "", // Empty output dir should default to current directory
		LogLevel:  logutil.InfoLevel,
		Verbose:   false,
		JsonLogs:  false,
	}

	// Setup logging
	logger := SetupLogging(config)
	if logger == nil {
		t.Fatal("SetupLogging returned nil logger")
	}

	// Log a test message
	logger.Info("Test message")

	// Check if log file was created in current directory
	logFilePath := "thinktank.log"
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Log file was not created in current directory at %s", logFilePath)
	}
}

func TestSetupLogging_FileCreationError_FallsBackToStderr(t *testing.T) {
	// Use an invalid directory path that should cause file creation to fail
	invalidPath := "/invalid/nonexistent/directory"

	// Create a config that would try to write to an invalid directory
	config := &config.CliConfig{
		OutputDir: invalidPath,
		LogLevel:  logutil.InfoLevel,
		Verbose:   false,
		JsonLogs:  false,
	}

	// Setup logging - should not panic even with invalid directory
	logger := SetupLogging(config)
	if logger == nil {
		t.Fatal("SetupLogging returned nil logger even in error case")
	}

	// The function should fall back gracefully and return a working logger
	// We can't easily test that it's using stderr specifically, but we can
	// verify that the logger works
	logger.Info("Test message after file creation failure")

	// The test passes if we get here without panicking
}
