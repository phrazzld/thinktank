# internal/llm

LLM client interface and error handling.

## Overview

Defines the contract for LLM clients and provides categorized error handling. Provider-specific implementations live in `internal/providers/`.

## Key Components

| File | Purpose |
|------|---------|
| `client.go` | LLMClient interface definition |
| `errors.go` | Error categories and wrapping |

## Interface

```go
type LLMClient interface {
    GenerateContent(ctx, prompt string, params map[string]any) (string, error)
    GetModelID() string
    CountTokens(ctx, text string) (int, error)
}
```

## Error Categories

All LLM errors are wrapped with a category for appropriate handling:

| Category | Meaning | Typical Response |
|----------|---------|------------------|
| `CategoryAuth` | Invalid API key | Fail fast |
| `CategoryRateLimit` | Rate limited | Retry with backoff |
| `CategoryNotFound` | Model not found | User error |
| `CategoryContext` | Context too long | Reduce input |
| `CategoryServer` | Provider error | Retry |

```go
if catErr, ok := llm.IsCategorizedError(err); ok {
    switch catErr.Category() {
    case llm.CategoryRateLimit:
        // wait and retry
    }
}
```
