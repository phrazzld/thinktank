# Migration Guide: Provider-Agnostic API

This guide helps you migrate from deprecated Gemini-specific methods to the new provider-agnostic API in Architect.

## Background

Architect is transitioning from Gemini-specific methods to provider-agnostic APIs that work with multiple LLM providers. The deprecated methods will be removed in v0.8.0 (Q4 2024).

## Deprecated APIs

The following APIs are deprecated and scheduled for removal:

| Deprecated API | Replacement API | Location |
|----------------|----------------|----------|
| `InitClient` | `InitLLMClient` | `internal/architect/api.go` |
| `ProcessResponse` | `ProcessLLMResponse` | `internal/architect/api.go` |
| `llmToGeminiClientAdapter` | Use `llm.LLMClient` directly | `internal/architect/compat/compat.go` |
| Entire `compat` package | N/A | `internal/architect/compat/` |

## Step-by-Step Migration

### 1. Replace Client Initialization

```go
// Old code (deprecated)
client, err := apiService.InitClient(ctx, apiKey, modelName, apiEndpoint)
if err != nil {
    return err
}
defer client.Close()

// New code (provider-agnostic)
client, err := apiService.InitLLMClient(ctx, apiKey, modelName, apiEndpoint)
if err != nil {
    return err
}
defer client.Close()
```

### 2. Replace Response Processing

```go
// Old code (deprecated)
result, err := client.GenerateContent(ctx, prompt)
if err != nil {
    return err
}
content, err := apiService.ProcessResponse(result)

// New code (provider-agnostic)
result, err := client.GenerateContent(ctx, prompt)
if err != nil {
    return err
}
content, err := apiService.ProcessLLMResponse(result)
```

### 3. Update Interface Usage

If you're using `gemini.Client` directly:

```go
// Old code (Gemini-specific)
func DoSomething(client gemini.Client) error {
    // ...
}

// New code (provider-agnostic)
func DoSomething(client llm.LLMClient) error {
    // ...
}
```

### 4. Model Name Handling

The provider-agnostic API automatically detects the appropriate provider based on the model name:

- `"gemini-1.5-pro"` → Gemini provider
- `"gpt-4"` → OpenAI provider

No changes are needed if you're already specifying the model name correctly.

### 5. Error Handling

Error types and messages are consistent between the old and new APIs. However, if you're using string matching or type assertions with Gemini-specific error types, update to use the provider-agnostic error helpers:

```go
// Old code (might be brittle)
if strings.Contains(err.Error(), "safety") {
    // Handle safety block
}

// New code (recommended)
if apiService.IsSafetyBlockedError(err) {
    // Handle safety block
}
```

## Testing Your Migration

1. Replace all instances of deprecated methods with their provider-agnostic equivalents
2. Run your test suite to verify functionality
3. Test with different model names to ensure provider detection works correctly
4. Check error handling to make sure it's still functioning properly

## Need Help?

If you encounter issues during migration:

1. Check the [GitHub repository](https://github.com/phrazzld/architect) for examples
2. Open an issue describing your specific migration challenge
3. Refer to the full API documentation for detailed interface descriptions

Thank you for migrating to our provider-agnostic API!
