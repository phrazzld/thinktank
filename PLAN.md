# Implementation Plan: Enhance Error Handling and Logging Consistency

## Overview

This plan outlines the approach to enhance error handling and ensure logging consistency across the codebase as specified in the top-priority backlog item. This refactoring will improve reliability, debugging capabilities, and operational excellence through rich self-describing error types and standardized structured logging.

## Approach Summary

Incrementally refactor error handling to use a canonical `llm.LLMError` type with categorization and standardize all operational logging to structured JSON via `log/slog`, enforcing context propagation and strict secret sanitization.

## Architecture Blueprint

### Modules / Packages

- **`internal/llm`**: Defines canonical `LLMError` type, `ErrorCategory` enum, and error handling helpers (`Wrap`, `IsCategory`, etc.). Core error vocabulary.
- **`internal/logutil`**: Provides `LoggerInterface` abstraction, `slog`-based implementation, context propagation helpers (`WithCorrelationID`), and `SecretDetectingLogger`. Core logging infrastructure.
- **`internal/providers/*`**: Provider-specific error translation (`FormatAPIError`) mapping raw API errors to `llm.LLMError`. Uses `logutil.LoggerInterface`.
- **`internal/thinktank/registry`**: Handles model/provider lookup and client initialization errors, using `llm.LLMError` and `logutil.LoggerInterface`.
- **`internal/thinktank/interfaces`**: Defines core service interfaces (`APIService`, `ContextGatherer`, etc.) ensuring `context.Context` propagation and consistent `error` return types.
- **`internal/thinktank/modelproc`**: Handles model processing logic, wraps underlying errors (e.g., from `APIService`) using `fmt.Errorf("%w", ...)` and specific `modelproc` error types/sentinels. Uses `logutil.LoggerInterface`.
- **`internal/thinktank/orchestrator`**: Orchestrates the overall workflow, aggregates errors from `modelproc`, handles final outcome reporting/logging using `logutil.LoggerInterface` and `auditlog.AuditLogger`. Extracts user-facing details from errors.
- **`internal/auditlog`**: Provides structured audit logging (`AuditLogger` interface and implementation), aligned with `logutil` standards.
- **`cmd/thinktank`**: Entry point, sets up initial context, logging, handles top-level errors, determines exit code.

### Public Interfaces / Contracts

- **`llm.LLMError`**: Struct implementing `error` with fields: `Provider`, `Code`, `StatusCode`, `Message`, `Original error`, `ErrorCategory ErrorCategory`, `Suggestion string`, `Details string`. Includes `Unwrap() error` and `Category() ErrorCategory`.
- **`llm.ErrorCategory`**: Enum (`Unknown`, `Auth`, `RateLimit`, `InvalidRequest`, `NotFound`, `Server`, `Network`, `Cancelled`, `InputLimit`, `ContentFiltered`, `InsufficientCredits`).
- **`llm.CategorizedError`**: Interface `{ error; Category() ErrorCategory }`.
- **`llm.Wrap(error, provider, message, category) error`**: Helper to wrap errors into `LLMError`.
- **`llm.IsCategory(error, ErrorCategory) bool`**: Helper to check error category.
- **`logutil.LoggerInterface`**: Interface aligned with `slog` context-aware methods (`DebugContext`, `InfoContext`, `WarnContext`, `ErrorContext`).
- **`logutil.WithCorrelationID(ctx) context.Context`**: Injects/retrieves correlation ID.
- **`logutil.GetCorrelationID(ctx) string`**: Retrieves correlation ID.
- **`auditlog.AuditLogger`**: Interface for logging key lifecycle events structurally.

### Data Flow Diagram

```mermaid
graph TD
    A[cmd/thinktank/main.go] --> B(Execute)
    B -- Creates Context + CorrelationID --> C{Validate Inputs}
    C -- Error --> X[Log Error (slog) & Exit]
    C -- Success --> D(Setup Logging - slog)
    D --> E(Setup Audit Logger)
    E --> F(Setup Core Services - Registry, APIService, etc.)
    F -- Inject Logger --> G(Orchestrator.Run)
    G -- Pass Context --> H(Gather Project Context)
    H -- Error --> G
    H -- Success --> I{Dry Run?}
    I -- Yes --> J(Display Dry Run Info)
    J --> K[Exit 0]
    I -- No --> L(Build Prompt)
    L --> M(Process Models - modelproc)
    M -- Pass Context --> N(APIService Call - e.g., InitClient, GenerateContent)
    N -- Raw Provider Error --> O(Provider Error Formatting - FormatAPIError)
    O --> P[llm.LLMError]
    P -- Propagate as error --> M
    M -- Wrapped Error --> G
    G -- Aggregates Errors --> Q(Handle Output Flow / Synthesis)
    Q -- Error --> G
    Q -- Success --> R(Handle Outcome)
    R -- Log Audit Event --> E
    R -- Log Final Status (slog) --> D
    R -- Error --> X
    R -- Success --> K

    subgraph Logging & Error Handling Flow
        direction LR
        P --> S{Error Handling Logic}
        S -- errors.Is/As --> T[Specific Action]
        S -- Log Details --> U(logutil.LoggerInterface - ErrorContext)
        U -- Add CorrelationID --> V(slog JSON Output)
        P -- User Msg --> W[CLI Output Formatting]
        W --> X
    end
```

### Error & Edge-Case Strategy

- **Canonical Type:** Use `llm.LLMError` for detailed errors, especially from providers. Use standard Go sentinel errors (`var ErrX = errors.New("...")`) for simpler cases.
- **Categorization:** Use `llm.ErrorCategory` for classifying common failure modes (Auth, RateLimit, etc.). Check using `llm.IsCategory` or `errors.As`.
- **Wrapping:** Consistently use `fmt.Errorf("context: %w", originalErr)` to add context while preserving the error chain for inspection via `errors.Is`/`errors.As`.
- **Propagation:** Return errors up the stack. Avoid panics for expected/recoverable errors.
- **Handling Boundaries:** Top-level (`cmd/thinktank`) catches final errors. Logs detailed, *sanitized* error info using `slog`. Provides clear, actionable user messages. Determines exit code based on error type/category.
- **Logging:** Log errors at `ERROR` level via `logutil.LoggerInterface.ErrorContext`. Include `correlation_id`, `error_category`, sanitized `error_message`, `error_type`, and potentially `stack_trace` (configurable).
- **Sanitization:** Implement strict sanitization before logging error details (especially `err.Error()` or `LLMError.Details`) to prevent leaking API keys, PII, or internal paths. Custom errors might provide a `SafeLogFields()` method.
- **Audit Logging:** Use `auditlog.AuditLogger` for key events (start/end, major operations, failures), including correlation ID and error summaries.
- **Edge Cases:** Explicitly handle context cancellation/deadline errors, empty/whitespace responses, safety blocks (classify appropriately), partial vs. total model failures, file I/O errors, configuration errors.

## Implementation Plan

### Phase 1: Core Error and Logging Framework

1. **Define Canonical Errors in `internal/llm/errors.go`**
   - Implement `LLMError` struct, `ErrorCategory` enum, `CategorizedError` interface
   - Create helper functions (`Wrap`, `IsCategory`, etc.)
   - Add comprehensive unit tests

2. **Setup `slog` Logging in `internal/logutil`**
   - Adapt `LoggerInterface` for `slog` context methods
   - Implement `slog`-based logger producing JSON
   - Configure default `slog` JSON handler in `cmd/thinktank`

3. **Implement Correlation ID**
   - Add `logutil.WithCorrelationID` to create initial context in `cmd/thinktank`
   - Ensure `context.Context` is passed through all relevant function calls
   - Update `logutil` to extract `correlation_id` from context for all logs

### Phase 2: Provider Error Handling

4. **Refactor Provider Error Handling**
   - In each `internal/providers/*/errors.go`, implement `FormatAPIError`
   - Update provider client implementations to use `FormatAPIError`
   - Add/update unit tests for provider error translation

### Phase 3: Core Component Refactoring

5. **Refactor Core Components (Registry, ModelProc, Orchestrator)**
   - Update method signatures to accept and propagate `context.Context`
   - Replace old logging with context-aware `logutil.LoggerInterface` methods
   - Update error handling with proper wrapping using `fmt.Errorf("%w", ...)`
   - Update tests to verify context propagation and structured logging

6. **Refactor Audit Logging**
   - Update `internal/auditlog` implementation to use `logutil.LoggerInterface`
   - Ensure consistent fields and correlation ID in audit logs
   - Update tests

### Phase 4: Input/Output Components

7. **Refactor File I/O Components**
   - Update `context.Context` propagation in I/O components
   - Replace logging calls with context-aware methods
   - Update unit tests

8. **Refactor Top-Level (`cmd/thinktank`)**
   - Add initial context creation with correlation ID
   - Correctly pass context, logger, audit logger into components
   - Implement final error handling with proper sanitization
   - Format user-friendly messages and determine appropriate exit codes

### Phase 5: Security and Cleanup

9. **Implement Sanitization & Secret Detection**
   - Add sanitization logic for error details logging
   - Integrate `logutil.SecretDetectingLogger` in all tests

10. **Codebase Audit & Cleanup**
    - Remove any remaining `fmt.Println`, `log.Printf`, direct `log` package usage
    - Ensure consistent error/logging patterns throughout
    - Run final tests and verify coverage

11. **Documentation Update**
    - Update `DEVELOPMENT_PHILOSOPHY.md`, `README.md`, and code comments
    - Document error handling patterns, log field definitions, and error categories

## Testing Strategy

### Test Layers

- **Unit Tests:** Verify individual functions, error creation/wrapping, error checking logic, logging helpers
- **Integration Tests:** Verify interactions between components, error propagation, context flow, and structured log output
- **E2E Tests:** Verify the compiled binary by simulating various failure scenarios

### What to Mock

- **Externals Only:** HTTP clients for LLM APIs, Filesystem I/O, Environment variables
- **Test Loggers:** Use `internal/testutil.MockLogger` to assert log messages, levels, and structured fields
- **No Internal Mocking:** Test integration by controlling external dependencies

### Coverage and Edge Cases

- Aim for >80% overall coverage, >95% for critical error handling and logging code
- Test all defined `ErrorCategory` paths
- Test nested error wrapping and unwrapping
- Test context cancellation propagation and logging
- Test log output structure (valid JSON, mandatory fields)
- Test secret sanitization to ensure no leaks

## Logging & Observability

### Log Format and Fields

- **Format:** Structured JSON via `log/slog`
- **Mandatory Fields:** `timestamp`, `level`, `msg`, `correlation_id`, `service`
- **Contextual Fields:** `func`, `model`, `provider`, `phase`
- **Error Fields:** `err_type`, `err_msg` (sanitized), `error_category`, `suggestion`, `stack_trace`

### Key Log Events

- Log start/end of execution
- Log start/end of major phases (context gathering, model processing, output saving)
- Log configuration loading, client initialization, significant decisions
- Log all errors with appropriate context

## Security Considerations

- **Input Validation:** Validate all external API responses, configuration values, and user inputs
- **Secrets Handling:** Ensure API keys and other secrets are never logged
- **Permissions:** Restrict file permissions for log files and output directories

## Risk Assessment

| Risk                                      | Severity | Mitigation                                                                                     |
|-------------------------------------------|----------|------------------------------------------------------------------------------------------------|
| Accidental Logging of Secrets/API Keys    | Critical | Strict sanitization, `SecretDetectingLogger` in tests, focused code reviews                     |
| Inconsistent Error Handling/Wrapping      | High     | Standardize on `LLMError` & `%w`, use linters, comprehensive tests                              |
| Loss of Error Context/Information         | Medium   | Enforce `%w` wrapping, ensure sufficient detail capture, test error chains                      |
| Incomplete Context/Correlation ID Flow    | Medium   | Consistent context passing, integration tests verifying correlation ID                          |
| Log Structure Inconsistency               | Medium   | Centralize logger setup, define mandatory fields, test log output structure                     |
| Performance Overhead                      | Low      | Use efficient `slog` library, make stack traces configurable                                    |
| Legacy Code Bypasses New System           | Medium   | Thorough refactoring, remove old code, code reviews, test coverage                             |
