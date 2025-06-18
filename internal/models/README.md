# Models Package

The `models` package provides a simple, hardcoded system for managing LLM model metadata. It replaces the previous complex registry system with direct map-based lookups, eliminating the need for YAML configuration files and complex initialization logic.

## Overview

This package contains metadata for all supported LLM models, including their providers, API model IDs, token limits, and default parameters. The design prioritizes simplicity and explicitness over configurability.

**Supported Models (7 total):**
- **OpenAI**: gpt-4.1, o4-mini
- **Gemini**: gemini-2.5-pro, gemini-2.5-flash
- **OpenRouter**: openrouter/deepseek/deepseek-chat-v3-0324, openrouter/deepseek/deepseek-r1, openrouter/x-ai/grok-3-beta

## Package Structure

### Core Components

- **`ModelInfo` struct**: Defines the structure for model metadata
- **`ModelDefinitions` map**: Contains all model definitions, keyed by model name
- **API functions**: Provide safe, validated access to model information

### Files

- `models.go` - Core implementation with ModelInfo struct, ModelDefinitions map, and API functions
- `models_test.go` - Comprehensive test suite with 100% coverage
- `doc.go` - Package-level documentation and usage examples
- `README.md` - This documentation file

## API Reference

### Data Structures

```go
type ModelInfo struct {
    Provider        string                 // Provider name (openai, gemini, openrouter)
    APIModelID      string                 // Model ID used in API calls
    ContextWindow   int                    // Maximum input + output tokens
    MaxOutputTokens int                    // Maximum output tokens
    DefaultParams   map[string]interface{} // Provider-specific parameters
}
```

### Functions

#### `GetModelInfo(name string) (ModelInfo, error)`
Returns complete model metadata for the given model name.

```go
info, err := models.GetModelInfo("gpt-4.1")
if err != nil {
    // Handle unknown model
}
fmt.Printf("Provider: %s, Context: %d tokens\n", info.Provider, info.ContextWindow)
```

#### `GetProviderForModel(name string) (string, error)`
Returns the provider name for a given model.

```go
provider, err := models.GetProviderForModel("gemini-2.5-pro")
// provider == "gemini"
```

#### `ListAllModels() []string`
Returns a sorted slice of all supported model names.

```go
allModels := models.ListAllModels()
// ["gemini-2.5-flash", "gemini-2.5-pro", "gpt-4.1", "o4-mini", ...]
```

#### `ListModelsForProvider(provider string) []string`
Returns models for a specific provider.

```go
openaiModels := models.ListModelsForProvider("openai")
// ["gpt-4.1", "o4-mini"]
```

#### `GetAPIKeyEnvVar(provider string) string`
Returns the environment variable name for a provider's API key.

```go
envVar := models.GetAPIKeyEnvVar("openai")
// "OPENAI_API_KEY"
```

#### `IsModelSupported(name string) bool`
Checks if a model is supported.

```go
if models.IsModelSupported("gpt-4.1") {
    // Model is supported
}
```

## Adding New Models

To add a new model:

1. **Edit `models.go`**: Add a new entry to the `ModelDefinitions` map
2. **Run tests**: Ensure all tests pass with `go test ./internal/models`
3. **Update tests**: Add test cases for the new model if needed
4. **Submit PR**: Follow the standard contribution process

### Example: Adding a New OpenAI Model

```go
"gpt-5": {
    Provider:        "openai",
    APIModelID:      "gpt-5",
    ContextWindow:   300000,
    MaxOutputTokens: 100000,
    DefaultParams: map[string]interface{}{
        "temperature":       0.7,
        "top_p":             1.0,
        "frequency_penalty": 0.0,
        "presence_penalty":  0.0,
    },
},
```

### Adding a New Provider

To add a completely new provider:

1. Add models to `ModelDefinitions` with the new provider name
2. Update `GetAPIKeyEnvVar()` function to include the new provider
3. Ensure the provider is supported in the client creation logic (outside this package)
4. Add comprehensive tests for the new provider

## Testing

The package maintains 100% test coverage with comprehensive table-driven tests:

- **Model validation**: Tests all 7 supported models
- **Error handling**: Tests invalid model names and edge cases
- **Provider filtering**: Tests provider-specific model listings
- **API key mapping**: Tests environment variable mappings
- **Sorting**: Ensures consistent model ordering

Run tests:
```bash
go test ./internal/models
go test -cover ./internal/models  # With coverage report
```

## Architecture Decisions

### Why Hardcoded Models?

The previous registry system used YAML files and complex initialization logic. This approach was replaced with hardcoded definitions for several reasons:

- **Simplicity**: Direct map lookups vs complex registry initialization
- **Reliability**: No file I/O or parsing errors at runtime
- **Performance**: O(1) lookups with no startup overhead
- **Maintainability**: Changes require code review and testing
- **Explicitness**: All supported models are visible in the codebase

### Design Principles

- **No External Dependencies**: Only uses standard library (fmt, sort)
- **Immutable Data**: Model definitions are read-only after initialization
- **Fail Fast**: Unknown models return errors immediately
- **Consistent Interface**: All functions follow Go idioms and error handling
- **Zero Configuration**: No setup required, works out of the box

### Trade-offs

**Benefits:**
- Simple, fast, reliable
- No configuration management
- Type-safe at compile time
- Easy to understand and debug

**Limitations:**
- Adding models requires code changes
- No runtime configuration
- Fixed set of supported providers

This trade-off prioritizes simplicity and reliability over configurability, which aligns with the project's focus on a curated set of well-tested models rather than supporting arbitrary model configurations.
