package orchestrator

import (
	"context"
	"strings"
	"testing"

	"github.com/phrazzld/thinktank/internal/config"
	"github.com/phrazzld/thinktank/internal/logutil"
	"github.com/phrazzld/thinktank/internal/ratelimit"
)

// TestCorrelationIDPropagation verifies that correlation IDs are properly
// generated and propagated through the Orchestrator's workflow
func TestCorrelationIDPropagation(t *testing.T) {
	// Create a test buffer to capture logs
	logBuffer := &TestLogBuffer{}

	// Create a logger with the test buffer
	logger := logutil.NewLogger(logutil.DebugLevel, logBuffer, "")

	// Create test configuration
	cfg := &config.CliConfig{
		ModelNames: []string{"test-model-1", "test-model-2"},
		DryRun:     true, // Use dry run to avoid actual API calls
	}

	// Create test dependencies
	mockAPIService := &MockAPIService{}
	mockContextGatherer := &MockContextGatherer{}
	mockFileWriter := NewMockFileWriter()
	mockAuditLogger := &MockAuditLogger{}
	rateLimiter := ratelimit.NewRateLimiter(0, 0)

	// Create the orchestrator
	orchestrator := NewOrchestrator(
		mockAPIService,
		mockContextGatherer,
		mockFileWriter,
		mockAuditLogger,
		rateLimiter,
		cfg,
		logger,
	)

	// Run the orchestrator with an empty context
	// The function should generate and add a correlation ID
	err := orchestrator.Run(context.Background(), "test instructions")

	// An error is expected as this is just a mock test
	if err != nil {
		t.Logf("Expected error received: %v", err)
	}

	// Check that logs contain correlation_id
	logs := logBuffer.String()
	if !strings.Contains(logs, "correlation_id=") {
		t.Errorf("Logs do not contain correlation_id. Logs: %s", logs)
	}

	// Test with an existing correlation ID in the context
	logBuffer.Reset()
	testID := "test-correlation-id-123"
	ctx := logutil.WithCustomCorrelationID(context.Background(), testID)

	err = orchestrator.Run(ctx, "test instructions")
	if err != nil {
		t.Logf("Expected error received: %v", err)
	}

	// Check that logs contain our specific correlation ID
	logs = logBuffer.String()
	if !strings.Contains(logs, "correlation_id="+testID) {
		t.Errorf("Logs do not contain our specific correlation ID %q. Logs: %s", testID, logs)
	}
}

// TestLogBuffer is a simple buffer that captures log messages
type TestLogBuffer struct {
	strings.Builder
}

// Reset clears the log buffer
func (b *TestLogBuffer) Reset() {
	b.Builder.Reset()
}
