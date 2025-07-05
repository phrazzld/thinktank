// Package integration provides comprehensive coverage tests for audit logging and test utilities
// Following TDD principles to target remaining 0% coverage functions
package integration

import (
	"testing"

	"github.com/phrazzld/thinktank/internal/auditlog"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/stretchr/testify/assert"
)

// TestBoundaryAuditLoggerMethods tests uncovered audit logger methods
// Targets: LogLegacy, LogOpLegacy, Close (all 0% coverage)
func TestBoundaryAuditLoggerMethods(t *testing.T) {
	filesystem := NewMockFilesystemIO()
	logger := logutil.NewLogger(logutil.InfoLevel, nil, "test")
	auditLogger := NewBoundaryAuditLogger(filesystem, logger)

	t.Run("LogLegacy with valid entry", func(t *testing.T) {
		entry := auditlog.AuditEntry{
			Operation: "test_operation",
			Status:    "success",
			Message:   "test message",
		}
		err := auditLogger.LogLegacy(entry)
		assert.NoError(t, err, "LogLegacy should handle legacy format successfully")
	})

	t.Run("LogOpLegacy with valid parameters", func(t *testing.T) {
		inputs := map[string]interface{}{
			"param1": "value1",
			"param2": 42,
		}
		outputs := map[string]interface{}{
			"result": "success",
		}

		err := auditLogger.LogOpLegacy("test_operation", "success", inputs, outputs, nil)
		assert.NoError(t, err, "LogOpLegacy should handle legacy operation format successfully")
	})

	t.Run("LogOpLegacy with nil inputs and outputs", func(t *testing.T) {
		err := auditLogger.LogOpLegacy("test_operation", "success", nil, nil, nil)
		assert.NoError(t, err, "LogOpLegacy should handle nil inputs and outputs")
	})

	t.Run("LogOpLegacy with error", func(t *testing.T) {
		testError := assert.AnError
		err := auditLogger.LogOpLegacy("test_operation", "failure", nil, nil, testError)
		assert.NoError(t, err, "LogOpLegacy should handle operations with errors")
	})

	t.Run("Close releases resources", func(t *testing.T) {
		err := auditLogger.Close()
		assert.NoError(t, err, "Close should successfully release resources")
	})

	t.Run("Close is idempotent", func(t *testing.T) {
		filesystem2 := NewMockFilesystemIO()
		logger2 := logutil.NewLogger(logutil.InfoLevel, nil, "test")
		auditLogger2 := NewBoundaryAuditLogger(filesystem2, logger2)

		err := auditLogger2.Close()
		assert.NoError(t, err, "First close should succeed")

		err = auditLogger2.Close()
		assert.NoError(t, err, "Second close should also succeed (idempotent)")
	})
}
