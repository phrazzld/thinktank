package orchestrator

import (
	"context"
	"testing"

	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/fileutil"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/metrics"
	"github.com/misty-step/thinktank/internal/ratelimit"
	"github.com/misty-step/thinktank/internal/thinktank/interfaces"
)

// TestOrchestratorMetricsIntegration verifies that metrics are recorded during orchestrator execution
func TestOrchestratorMetricsIntegration(t *testing.T) {
	ctx := context.Background()

	// Create a real metrics collector (no exporter needed for test)
	metricsCollector := metrics.NewCollector(nil)

	// Create mock dependencies
	mockLogger := &MockLogger{}
	mockAuditLogger := NewMockAuditLogger()
	mockAPIService := &MockAPIService{}
	mockFileWriter := &MockFileWriter{}
	mockContextGatherer := &MockContextGatherer{}
	mockTokenCountingService := &MockTokenCountingService{
		CountTokensResult: interfaces.TokenCountingResult{
			TotalTokens: 1000,
		},
	}

	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false },
	})
	rateLimiter := ratelimit.NewRateLimiter(5, 60)

	cfg := &config.CliConfig{
		ModelNames: []string{"mock-model"},
		OutputDir:  t.TempDir(),
		Paths:      []string{"."},
		DryRun:     true, // Use dry run to avoid actual API calls
	}

	deps := OrchestratorDeps{
		APIService:           mockAPIService,
		ContextGatherer:      mockContextGatherer,
		FileWriter:           mockFileWriter,
		AuditLogger:          mockAuditLogger,
		RateLimiter:          rateLimiter,
		Config:               cfg,
		Logger:               mockLogger,
		ConsoleWriter:        consoleWriter,
		TokenCountingService: mockTokenCountingService,
		MetricsCollector:     metricsCollector,
	}

	orch := NewOrchestrator(deps)

	// Run the orchestrator in dry-run mode
	err := orch.Run(ctx, "Test instructions")
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	// Verify metrics were recorded
	allMetrics := metricsCollector.Metrics()
	if len(allMetrics) == 0 {
		t.Fatal("Expected metrics to be recorded, got none")
	}

	// Check for required metrics
	metricNames := make(map[string]bool)
	for _, m := range allMetrics {
		metricNames[m.Name] = true
	}

	// total_duration_ms should always be recorded
	if !metricNames["total_duration_ms"] {
		t.Error("Expected total_duration_ms metric to be recorded")
	}

	// context_gather_duration_ms should be recorded
	if !metricNames["context_gather_duration_ms"] {
		t.Error("Expected context_gather_duration_ms metric to be recorded")
	}

	// Verify metric types
	for _, m := range allMetrics {
		if m.Name == "total_duration_ms" || m.Name == "context_gather_duration_ms" {
			if m.Type != metrics.TypeDuration {
				t.Errorf("Expected %s to be type duration, got %s", m.Name, m.Type)
			}
			if m.Value < 0 {
				t.Errorf("Expected %s value >= 0, got %f", m.Name, m.Value)
			}
		}
	}
}

// TestOrchestratorMetricsWithNilCollector verifies that nil collector is replaced with noop
func TestOrchestratorMetricsWithNilCollector(t *testing.T) {
	ctx := context.Background()

	// Create mock dependencies WITHOUT a metrics collector
	mockLogger := &MockLogger{}
	mockAuditLogger := NewMockAuditLogger()
	mockAPIService := &MockAPIService{}
	mockFileWriter := &MockFileWriter{}
	mockContextGatherer := &MockContextGatherer{}
	mockTokenCountingService := &MockTokenCountingService{
		CountTokensResult: interfaces.TokenCountingResult{
			TotalTokens: 1000,
		},
	}

	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false },
	})
	rateLimiter := ratelimit.NewRateLimiter(5, 60)

	cfg := &config.CliConfig{
		ModelNames: []string{"mock-model"},
		OutputDir:  t.TempDir(),
		Paths:      []string{"."},
		DryRun:     true,
	}

	deps := OrchestratorDeps{
		APIService:           mockAPIService,
		ContextGatherer:      mockContextGatherer,
		FileWriter:           mockFileWriter,
		AuditLogger:          mockAuditLogger,
		RateLimiter:          rateLimiter,
		Config:               cfg,
		Logger:               mockLogger,
		ConsoleWriter:        consoleWriter,
		TokenCountingService: mockTokenCountingService,
		MetricsCollector:     nil, // Explicitly nil
	}

	orch := NewOrchestrator(deps)

	// Run should not panic with nil metrics collector
	err := orch.Run(ctx, "Test instructions")
	if err != nil {
		t.Fatalf("Run() with nil metrics collector unexpected error: %v", err)
	}
}

// TestOrchestratorMetricsErrorPhases verifies error counters are incremented on failures
func TestOrchestratorMetricsErrorPhases(t *testing.T) {
	// This test verifies that when context gathering fails,
	// the error counter is incremented with the correct phase label
	ctx := context.Background()

	metricsCollector := metrics.NewCollector(nil)

	mockLogger := &MockLogger{}
	mockAuditLogger := NewMockAuditLogger()
	mockAPIService := &MockAPIService{}
	mockFileWriter := &MockFileWriter{}
	mockTokenCountingService := &MockTokenCountingService{}

	// Create a context gatherer that returns an error
	errorContextGatherer := &erroringContextGatherer{}

	consoleWriter := logutil.NewConsoleWriterWithOptions(logutil.ConsoleWriterOptions{
		IsTerminalFunc: func() bool { return false },
	})
	rateLimiter := ratelimit.NewRateLimiter(5, 60)

	cfg := &config.CliConfig{
		ModelNames: []string{"mock-model"},
		OutputDir:  t.TempDir(),
		Paths:      []string{"."},
	}

	deps := OrchestratorDeps{
		APIService:           mockAPIService,
		ContextGatherer:      errorContextGatherer,
		FileWriter:           mockFileWriter,
		AuditLogger:          mockAuditLogger,
		RateLimiter:          rateLimiter,
		Config:               cfg,
		Logger:               mockLogger,
		ConsoleWriter:        consoleWriter,
		TokenCountingService: mockTokenCountingService,
		MetricsCollector:     metricsCollector,
	}

	orch := NewOrchestrator(deps)

	// Run should fail due to context gathering error
	err := orch.Run(ctx, "Test instructions")
	if err == nil {
		t.Fatal("Expected Run() to fail due to context gathering error")
	}

	// Verify error counter was incremented
	allMetrics := metricsCollector.Metrics()
	foundErrorCounter := false
	for _, m := range allMetrics {
		if m.Name == "execution_errors_total" && m.Type == metrics.TypeCounter {
			foundErrorCounter = true
			if m.Labels["phase"] != "context_gather" {
				t.Errorf("Expected error phase 'context_gather', got '%s'", m.Labels["phase"])
			}
		}
	}

	if !foundErrorCounter {
		t.Error("Expected execution_errors_total counter to be recorded")
	}
}

// erroringContextGatherer is a mock that always returns an error
type erroringContextGatherer struct{}

func (e *erroringContextGatherer) GatherContext(ctx context.Context, config interfaces.GatherConfig) ([]fileutil.FileMeta, *interfaces.ContextStats, error) {
	return nil, nil, context.DeadlineExceeded
}

func (e *erroringContextGatherer) DisplayDryRunInfo(ctx context.Context, stats *interfaces.ContextStats) error {
	return nil
}
