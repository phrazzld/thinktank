// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import (
	"github.com/misty-step/thinktank/internal/auditlog"
	"github.com/misty-step/thinktank/internal/config"
	"github.com/misty-step/thinktank/internal/logutil"
	"github.com/misty-step/thinktank/internal/ratelimit"
	"github.com/misty-step/thinktank/internal/thinktank/interfaces"
	"github.com/misty-step/thinktank/internal/thinktank/orchestrator"
)

// NewOrchestrator creates an orchestrator instance with all required dependencies
// This function allows tests to create orchestrators without relying on
// the orchestratorConstructor variable
func NewOrchestrator(
	apiService interfaces.APIService,
	contextGatherer interfaces.ContextGatherer,
	fileWriter interfaces.FileWriter,
	auditLogger auditlog.AuditLogger,
	rateLimiter *ratelimit.RateLimiter,
	config *config.CliConfig,
	logger logutil.LoggerInterface,
	consoleWriter logutil.ConsoleWriter,
	tokenCountingService interfaces.TokenCountingService,
) Orchestrator {
	return orchestrator.NewOrchestrator(
		apiService,
		contextGatherer,
		fileWriter,
		auditLogger,
		rateLimiter,
		config,
		logger,
		consoleWriter,
		tokenCountingService,
	)
}
