# ADR-002: Orchestrator Pattern for Concurrent Execution

## Status
Accepted

## Context
thinktank's core value is comparing responses from multiple LLM models. This requires:
- Concurrent API calls (sequential would be too slow)
- Per-model rate limiting (different models have different limits)
- Graceful handling of partial failures
- Progress reporting during execution
- Optional synthesis of results

The naive approach of goroutines + channels became unwieldy as requirements grew.

## Decision
Introduce an Orchestrator component (`internal/thinktank/orchestrator/`) that:
- Coordinates all model execution
- Manages per-model rate limiters
- Collects results and errors
- Triggers synthesis when configured
- Reports progress through ConsoleWriter

The orchestrator receives all dependencies via constructor injection and exposes a simple `Execute(ctx, instructions, context)` interface.

## Consequences

**Positive:**
- Clear separation: app.go sets up, orchestrator executes
- Testable: dependencies are injectable
- Flexible: synthesis is optional, rate limits are per-model
- Observable: progress reporting centralized

**Negative:**
- More indirection (app → orchestrator → services)
- Orchestrator has many dependencies (8+ constructor params)
- Testing requires significant setup

## Alternatives Considered

1. **Keep logic in app.go**: Simpler but growing unwieldy
2. **Pipeline pattern**: Too rigid for variable model counts
3. **Actor model**: Over-engineered for this use case
