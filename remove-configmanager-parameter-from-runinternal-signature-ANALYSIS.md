# Analysis: Remove `configManager` parameter from `RunInternal` signature

## Task Description
Modify the function signature of `RunInternal` in `internal/architect/app.go` to remove the `configManager config.ManagerInterface` parameter.

## Current Implementation
The current `RunInternal` function signature included an unused `configManager` parameter:

```go
func RunInternal(
    ctx context.Context,
    cliConfig *CliConfig,
    logger logutil.LoggerInterface,
    configManager config.ManagerInterface,
    apiService APIService,
) error {
    // Implementation doesn't use configManager
    // ...
}
```

## Analysis
After reviewing the `RunInternal` function in `internal/architect/app.go`, I found:

1. The `configManager` parameter is included in the function signature
2. The parameter isn't used anywhere within the function body
3. The function already works directly with the `cliConfig` parameter for all configuration needs
4. Removing the parameter won't affect the function's logic

This aligns with the refactoring goal of removing the file-based configuration system and simplifying the API.

## Changes Made
I modified the function signature to remove the unused parameter:

```go
func RunInternal(
    ctx context.Context,
    cliConfig *CliConfig,
    logger logutil.LoggerInterface,
    apiService APIService,
) error {
    // Implementation unchanged
    // ...
}
```

## Impact
This change:
- Simplifies the function signature
- Reduces coupling between the core logic and the configuration system
- Is a step toward removing all `config.ManagerInterface` references
- Will require updating call sites in integration tests

## Next Steps
- Update integration tests that call `RunInternal` to remove the `configManager` argument
- Continue refactoring to remove remaining `configManager` usages