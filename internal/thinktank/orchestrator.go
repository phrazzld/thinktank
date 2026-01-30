// Package thinktank contains the core application logic for the thinktank tool
package thinktank

import "github.com/misty-step/thinktank/internal/thinktank/orchestrator"

// OrchestratorDeps aliases the orchestrator package dependencies for callers.
type OrchestratorDeps = orchestrator.OrchestratorDeps

// NewOrchestrator creates an orchestrator instance with all required dependencies
// This function allows tests to create orchestrators without relying on
// the orchestratorConstructor variable
func NewOrchestrator(deps OrchestratorDeps) Orchestrator {
	return orchestrator.NewOrchestrator(deps)
}
