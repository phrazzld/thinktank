package orchestrator

import (
	"context"
	"testing"

	"github.com/phrazzld/architect/internal/architect"
	"github.com/phrazzld/architect/internal/auditlog"
	"github.com/phrazzld/architect/internal/config"
	"github.com/phrazzld/architect/internal/fileutil"
	"github.com/phrazzld/architect/internal/gemini"
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

	// Verify the dependencies were properly assigned
	// Note: We can't directly compare the interfaces, so we just check for nil
	if orch.apiService == nil {
		t.Errorf("Expected apiService to be non-nil")
	}
	if orch.contextGatherer == nil {
		t.Errorf("Expected contextGatherer to be non-nil")
	}
	if orch.tokenManager == nil {
		t.Errorf("Expected tokenManager to be non-nil")
	}
	if orch.fileWriter == nil {
		t.Errorf("Expected fileWriter to be non-nil")
	}
	if orch.auditLogger == nil {
		t.Errorf("Expected auditLogger to be non-nil")
	}
	if orch.rateLimiter != limiter {
		t.Errorf("Expected rateLimiter to be set correctly")
	}
	if orch.config != cfg {
		t.Errorf("Expected config to be set correctly")
	}
	if orch.logger == nil {
		t.Errorf("Expected logger to be non-nil")
	}
}

// Mock implementations of the required interfaces

// mockAPIService implements architect.APIService
type mockAPIService struct{}

func (m *mockAPIService) InitClient(ctx context.Context, apiKey, modelName string) (gemini.Client, error) {
	return nil, nil
}

func (m *mockAPIService) ProcessResponse(result *gemini.GenerationResult) (string, error) {
	return "", nil
}

func (m *mockAPIService) IsEmptyResponseError(err error) bool {
	return false
}

func (m *mockAPIService) IsSafetyBlockedError(err error) bool {
	return false
}

func (m *mockAPIService) GetErrorDetails(err error) string {
	return ""
}

// mockContextGatherer implements architect.ContextGatherer
type mockContextGatherer struct{}

func (m *mockContextGatherer) GatherContext(ctx context.Context, client gemini.Client, config architect.GatherConfig) ([]fileutil.FileMeta, *architect.ContextStats, error) {
	return nil, nil, nil
}

func (m *mockContextGatherer) DisplayDryRunInfo(ctx context.Context, client gemini.Client, stats *architect.ContextStats) error {
	return nil
}

// mockTokenManager implements architect.TokenManager
type mockTokenManager struct{}

func (m *mockTokenManager) GetTokenInfo(ctx context.Context, client gemini.Client, prompt string) (*architect.TokenResult, error) {
	return nil, nil
}

func (m *mockTokenManager) CheckTokenLimit(ctx context.Context, client gemini.Client, prompt string) error {
	return nil
}

func (m *mockTokenManager) PromptForConfirmation(tokenCount int32, threshold int) bool {
	return true
}

// mockFileWriter implements architect.FileWriter
type mockFileWriter struct{}

func (m *mockFileWriter) SaveToFile(content, outputFile string) error {
	return nil
}

// mockAuditLogger implements auditlog.AuditLogger
type mockAuditLogger struct{}

func (m *mockAuditLogger) Log(entry auditlog.AuditEntry) error {
	return nil
}

func (m *mockAuditLogger) Close() error {
	return nil
}

// mockLogger implements logutil.LoggerInterface
type mockLogger struct{}

func (m *mockLogger) Debug(format string, args ...interface{})  {}
func (m *mockLogger) Info(format string, args ...interface{})   {}
func (m *mockLogger) Warn(format string, args ...interface{})   {}
func (m *mockLogger) Error(format string, args ...interface{})  {}
func (m *mockLogger) Fatal(format string, args ...interface{})  {}
func (m *mockLogger) Println(args ...interface{})               {}
func (m *mockLogger) Printf(format string, args ...interface{}) {}
