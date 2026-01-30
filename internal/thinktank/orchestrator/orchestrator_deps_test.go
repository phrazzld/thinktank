package orchestrator

import (
	"testing"

	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/ratelimit"
	"github.com/misty-step/thinktank/internal/testutil"
)

// TestNewOrchestratorWithNilDependencies verifies that NewOrchestrator panics
// when essential dependencies are nil. This is idiomatic Go behavior to signal
// programming errors early at construction time.
func TestNewOrchestratorWithNilDependencies(t *testing.T) {
	// Create valid dependencies for baseline
	validDeps := func() OrchestratorDeps {
		return OrchestratorDeps{
			APIService:           &MockAPIService{},
			ContextGatherer:      &MockContextGatherer{},
			FileWriter:           &MockFileWriter{},
			AuditLogger:          NewMockAuditLogger(),
			RateLimiter:          ratelimit.NewRateLimiter(10, 60),
			Config:               &config.CliConfig{ModelNames: []string{"test-model"}},
			Logger:               testutil.NewMockLogger(),
			ConsoleWriter:        &MockConsoleWriter{},
			TokenCountingService: &MockTokenCountingService{},
		}
	}

	tests := []struct {
		name        string
		modifyDeps  func(*OrchestratorDeps)
		expectedMsg string
	}{
		{
			name: "nil APIService panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.APIService = nil
			},
			expectedMsg: "NewOrchestrator: APIService cannot be nil",
		},
		{
			name: "nil ContextGatherer panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.ContextGatherer = nil
			},
			expectedMsg: "NewOrchestrator: ContextGatherer cannot be nil",
		},
		{
			name: "nil FileWriter panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.FileWriter = nil
			},
			expectedMsg: "NewOrchestrator: FileWriter cannot be nil",
		},
		{
			name: "nil AuditLogger panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.AuditLogger = nil
			},
			expectedMsg: "NewOrchestrator: AuditLogger cannot be nil",
		},
		{
			name: "nil RateLimiter panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.RateLimiter = nil
			},
			expectedMsg: "NewOrchestrator: RateLimiter cannot be nil",
		},
		{
			name: "nil Config panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.Config = nil
			},
			expectedMsg: "NewOrchestrator: Config cannot be nil",
		},
		{
			name: "nil Logger panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.Logger = nil
			},
			expectedMsg: "NewOrchestrator: Logger cannot be nil",
		},
		{
			name: "nil ConsoleWriter panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.ConsoleWriter = nil
			},
			expectedMsg: "NewOrchestrator: ConsoleWriter cannot be nil",
		},
		{
			name: "nil TokenCountingService panics",
			modifyDeps: func(d *OrchestratorDeps) {
				d.TokenCountingService = nil
			},
			expectedMsg: "NewOrchestrator: TokenCountingService cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := validDeps()
			tt.modifyDeps(&deps)

			defer func() {
				r := recover()
				if r == nil {
					t.Fatalf("Expected panic but none occurred")
				}
				panicMsg, ok := r.(string)
				if !ok {
					t.Fatalf("Expected panic message to be string, got %T: %v", r, r)
				}
				if panicMsg != tt.expectedMsg {
					t.Errorf("Panic message mismatch\nExpected: %q\nActual:   %q", tt.expectedMsg, panicMsg)
				}
			}()

			NewOrchestrator(deps)
		})
	}
}

// TestNewOrchestratorWithValidDependencies verifies that NewOrchestrator
// successfully creates an Orchestrator when all dependencies are provided.
func TestNewOrchestratorWithValidDependencies(t *testing.T) {
	deps := OrchestratorDeps{
		APIService:           &MockAPIService{},
		ContextGatherer:      &MockContextGatherer{},
		FileWriter:           &MockFileWriter{},
		AuditLogger:          NewMockAuditLogger(),
		RateLimiter:          ratelimit.NewRateLimiter(10, 60),
		Config:               &config.CliConfig{ModelNames: []string{"test-model"}},
		Logger:               testutil.NewMockLogger(),
		ConsoleWriter:        &MockConsoleWriter{},
		TokenCountingService: &MockTokenCountingService{},
	}

	// Should not panic
	orchestrator := NewOrchestrator(deps)

	if orchestrator == nil {
		t.Fatal("Expected non-nil orchestrator")
	}
	if orchestrator.apiService == nil {
		t.Error("Expected apiService to be set")
	}
	if orchestrator.contextGatherer == nil {
		t.Error("Expected contextGatherer to be set")
	}
	if orchestrator.fileWriter == nil {
		t.Error("Expected fileWriter to be set")
	}
	if orchestrator.auditLogger == nil {
		t.Error("Expected auditLogger to be set")
	}
	if orchestrator.rateLimiter == nil {
		t.Error("Expected rateLimiter to be set")
	}
	if orchestrator.config == nil {
		t.Error("Expected config to be set")
	}
	if orchestrator.logger == nil {
		t.Error("Expected logger to be set")
	}
	if orchestrator.consoleWriter == nil {
		t.Error("Expected consoleWriter to be set")
	}
	if orchestrator.tokenCountingService == nil {
		t.Error("Expected tokenCountingService to be set")
	}
}
