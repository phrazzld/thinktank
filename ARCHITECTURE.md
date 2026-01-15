# Architecture

thinktank is a Go CLI that sends codebase context + instructions to LLM models via OpenRouter, optionally synthesizing responses from multiple models.

## System Map

```
┌─────────────────────────────────────────────────────────────────────┐
│                              CLI Layer                               │
│  cmd/thinktank/main.go → internal/cli/                              │
│  Parses args, validates config, handles errors, returns exit codes  │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          Application Core                            │
│  internal/thinktank/app.go                                          │
│  Coordinates: setup → gather files → call LLMs → write outputs      │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                           Orchestrator                               │
│  internal/thinktank/orchestrator/                                   │
│  Manages concurrent LLM calls, rate limiting, synthesis             │
└───────────┬─────────────────────────────────┬───────────────────────┘
            │                                 │
            ▼                                 ▼
┌───────────────────────┐       ┌─────────────────────────────────────┐
│   Context Gathering   │       │          LLM Integration            │
│   internal/fileutil/  │       │   internal/llm/ + providers/        │
│   File discovery,     │       │   OpenRouter client, error          │
│   filtering, reading  │       │   handling, retries                 │
└───────────────────────┘       └─────────────────────────────────────┘
```

## Module Responsibilities

### Deep Modules (Rich Behavior, Simple Interface)

| Module | Owns | Does NOT Own |
|--------|------|--------------|
| `internal/thinktank/orchestrator/` | Concurrent model execution, synthesis coordination, progress reporting | Model definitions, API protocols |
| `internal/fileutil/` | File discovery, .gitignore handling, content reading, filtering | What to do with files |
| `internal/llm/` | LLM client interface, error categorization, retry logic | Provider-specific protocols |
| `internal/providers/openrouter/` | OpenRouter API specifics, auth, request formatting | Other providers |

### Supporting Modules

| Module | Purpose |
|--------|---------|
| `internal/cli/` | Argument parsing, config validation, error presentation |
| `internal/config/` | Configuration structures and defaults |
| `internal/models/` | Model definitions (context windows, parameters) |
| `internal/ratelimit/` | Token bucket rate limiting |
| `internal/logutil/` | Structured logging, console output |
| `internal/auditlog/` | Operation audit trail |

## Data Flow

```
User Input                  Processing                      Output
──────────────────────────────────────────────────────────────────────
instructions.txt  ──┐
                    ├──► Prompt Builder ──► OpenRouter ──► model-1.md
target paths     ───┤                   ──► OpenRouter ──► model-2.md
                    │                                          │
                    │         ┌────────────────────────────────┘
                    │         ▼
                    └──► Synthesis Model ──────────────────► synthesis.md
```

1. **Input**: Instructions file + target paths
2. **Context Gathering**: Walk directories, filter by patterns, read files
3. **Prompt Assembly**: Instructions + file contents + system prompt
4. **Concurrent Execution**: Send to N models via OpenRouter (rate limited)
5. **Synthesis** (optional): Combine model outputs into unified response
6. **Output**: Write individual + synthesis files to output directory

## Key Abstractions

### Interfaces (`internal/thinktank/interfaces/`)

```go
// Core orchestration contracts
type APIService interface { InitLLMClient(...); GetModelInfo(...) }
type ContextGatherer interface { GatherContext(...) }
type FileWriter interface { WriteOutput(...) }
type TokenCountingService interface { CountTokens(...) }
```

### Error Categorization (`internal/llm/errors.go`)

All LLM errors are categorized for appropriate handling:
- `CategoryAuth` - Invalid API key
- `CategoryRateLimit` - Rate limited, retry later
- `CategoryNotFound` - Model not found
- `CategoryContext` - Context too long
- `CategoryServer` - Provider issues

## Where to Start Reading

1. **Entry point**: `cmd/thinktank/main.go` → `internal/cli/main.go`
2. **Core logic**: `internal/thinktank/app.go` (Execute function)
3. **Orchestration**: `internal/thinktank/orchestrator/orchestrator.go`
4. **Model config**: `internal/models/models.go` (add new models here)

## Architectural Decisions

See `docs/adr/` for recorded decisions including:
- ADR-001: OpenRouter as unified provider
- ADR-002: Orchestrator pattern for concurrent execution
