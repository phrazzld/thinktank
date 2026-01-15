# internal/thinktank

Core application logic for thinktank.

## Overview

This package contains the main execution flow: setup, file gathering, LLM interaction, and output writing. It's the "application layer" sitting between CLI parsing and infrastructure.

## Key Components

| File | Purpose |
|------|---------|
| `app.go` | Main Execute() function, orchestrates workflow phases |
| `orchestrator.go` | Factory for creating orchestrator with dependencies |
| `context.go` | Context gathering coordination |
| `errors.go` | Application-level error definitions |

## Subpackages

- `orchestrator/` - Concurrent model execution, synthesis, progress
- `interfaces/` - Contracts for dependency injection
- `prompt/` - Prompt building and formatting
- `workflow/` - Workflow state management
- `tokenizers/` - Token counting implementations

## Entry Point

```go
// internal/cli calls this
result := thinktank.Execute(ctx, cliConfig, logger, auditLogger)
```

## Testing

Most tests are in `*_test.go` files. Key test categories:
- `app_test.go` - Integration tests for Execute()
- `app_error_test.go` - Error handling scenarios
- `app_partial_success_test.go` - Partial failure handling
