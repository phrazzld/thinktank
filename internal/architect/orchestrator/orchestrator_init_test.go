package orchestrator

import (
	"testing"

	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/ratelimit"
)

// TestNewOrchestrator verifies that the constructor creates an Orchestrator instance
// with the provided dependencies.
func TestNewOrchestrator(t *testing.T) {
	// Create mock dependencies
	apiSvc := &mockAPIService{}
	gatherer := &mockContextGatherer{}
	tokenMgr := &mockTokenManager{}
	writer := &mockFileWriter{}
	auditor := &mockAuditLogger{}
	limiter := ratelimit.NewRateLimiter(1, 1)
	cfg := &config.CliConfig{}
	logger := &mockLogger{}

	// Create the orchestrator with the mock dependencies
	orch := NewOrchestrator(
		apiSvc,
		gatherer,
		tokenMgr,
		writer,
		auditor,
		limiter,
		cfg,
		logger,
	)

	// Verify the orchestrator is not nil
	if orch == nil {
		t.Fatal("Expected non-nil Orchestrator")
	}

	// Verify that the dependencies were properly set
	if orch.apiService != apiSvc {
		t.Errorf("Expected apiService to be %v, got %v", apiSvc, orch.apiService)
	}
	if orch.contextGatherer != gatherer {
		t.Errorf("Expected contextGatherer to be %v, got %v", gatherer, orch.contextGatherer)
	}
	if orch.tokenManager != tokenMgr {
		t.Errorf("Expected tokenManager to be %v, got %v", tokenMgr, orch.tokenManager)
	}
	if orch.fileWriter != writer {
		t.Errorf("Expected fileWriter to be %v, got %v", writer, orch.fileWriter)
	}
	if orch.auditLogger != auditor {
		t.Errorf("Expected auditLogger to be %v, got %v", auditor, orch.auditLogger)
	}
	if orch.rateLimiter != limiter {
		t.Errorf("Expected rateLimiter to be %v, got %v", limiter, orch.rateLimiter)
	}
	if orch.config != cfg {
		t.Errorf("Expected config to be %v, got %v", cfg, orch.config)
	}
	if orch.logger != logger {
		t.Errorf("Expected logger to be %v, got %v", logger, orch.logger)
	}
}
