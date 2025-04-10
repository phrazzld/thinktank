package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/logutil"
)

func TestInitAuditLogger(t *testing.T) {
	logger := logutil.NewLogger(logutil.DebugLevel, nil, "[test] ", false)

	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "architect-audit-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test audit logging disabled
	t.Run("AuditLoggingDisabled", func(t *testing.T) {
		cfg := &config.AppConfig{
			AuditLogEnabled: false,
		}

		auditLogger := initAuditLogger(cfg, logger)

		// Verify it's a NoopLogger
		_, isNoopLogger := auditLogger.(*auditlog.NoopLogger)
		if !isNoopLogger {
			t.Errorf("Expected NoopLogger when AuditLogEnabled is false, got %T", auditLogger)
		}

		// Close logger (should have no effect for NoopLogger)
		err := auditLogger.Close()
		if err != nil {
			t.Errorf("Expected no error when closing NoopLogger, got: %v", err)
		}
	})

	// Test audit logging enabled with custom path
	t.Run("AuditLoggingEnabledCustomPath", func(t *testing.T) {
		logFilePath := filepath.Join(tmpDir, "audit.log")

		cfg := &config.AppConfig{
			AuditLogEnabled: true,
			AuditLogFile:    logFilePath,
		}

		auditLogger := initAuditLogger(cfg, logger)
		defer auditLogger.Close()

		// Verify it's a FileLogger
		_, isFileLogger := auditLogger.(*auditlog.FileLogger)
		if !isFileLogger {
			t.Errorf("Expected FileLogger when AuditLogEnabled is true, got %T", auditLogger)
		}

		// Verify the log file was created
		if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
			t.Errorf("Log file was not created at %s", logFilePath)
		}

		// Log a test event
		event := auditlog.NewAuditEvent("INFO", "TestOperation", "Test message")
		auditLogger.Log(event)

		// Close logger
		err := auditLogger.Close()
		if err != nil {
			t.Errorf("Error closing FileLogger: %v", err)
		}

		// Verify content was written to the log file
		content, err := os.ReadFile(logFilePath)
		if err != nil {
			t.Fatalf("Failed to read log file: %v", err)
		}

		if len(content) == 0 {
			t.Error("Log file is empty, expected content to be written")
		}
	})

	// Test with invalid path (should fall back to NoopLogger)
	t.Run("InvalidPath", func(t *testing.T) {
		// Use an invalid path (a directory that we don't have permission to write to)
		invalidDir := filepath.Join(tmpDir, "invalid-dir")
		os.Mkdir(invalidDir, 0000)       // Directory with no permissions
		defer os.Chmod(invalidDir, 0755) // Restore permissions for cleanup

		invalidPath := filepath.Join(invalidDir, "audit.log")

		cfg := &config.AppConfig{
			AuditLogEnabled: true,
			AuditLogFile:    invalidPath,
		}

		auditLogger := initAuditLogger(cfg, logger)
		defer auditLogger.Close()

		// Should fall back to NoopLogger on error
		_, isNoopLogger := auditLogger.(*auditlog.NoopLogger)
		if !isNoopLogger {
			t.Errorf("Expected fallback to NoopLogger with invalid path, got %T", auditLogger)
		}
	})

	// Test with default path (empty string)
	t.Run("DefaultPath", func(t *testing.T) {
		// Temporarily override the getCacheDir function
		origGetCacheDir := getCacheDir
		defer func() { getCacheDir = origGetCacheDir }()

		getCacheDir = func() string {
			return tmpDir
		}

		cfg := &config.AppConfig{
			AuditLogEnabled: true,
			AuditLogFile:    "", // Empty string should use default path
		}

		auditLogger := initAuditLogger(cfg, logger)
		defer auditLogger.Close()

		// Verify it's a FileLogger
		_, isFileLogger := auditLogger.(*auditlog.FileLogger)
		if !isFileLogger {
			t.Errorf("Expected FileLogger with default path, got %T", auditLogger)
		}

		// Verify the log file was created in the expected location
		expectedPath := filepath.Join(tmpDir, "audit.log")
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("Log file was not created at default path %s", expectedPath)
		}
	})
}
