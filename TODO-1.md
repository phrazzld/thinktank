# Core Error & Logging Framework Refactor Tasks

This document contains the detailed task breakdown for implementing the first part of the "Enhance Error Handling and Logging Consistency" epic, focusing on establishing the foundational error handling and logging infrastructure.

## Core LLM Error Handling (`internal/llm`)

- [x] **T001 · Feature · P0: Define canonical LLMError struct**
    - **Context:** Phase 1, Step 1 from PLAN.md (Define Canonical Errors)
    - **Action:**
        1. Create `internal/llm/errors.go`.
        2. Define `LLMError` struct with fields: `Provider`, `Code`, `StatusCode`, `Message`, `OriginalError error`, `ErrCategory ErrorCategory`, `Suggestion string`, `Details string`.
        3. Implement the standard `error` interface (`Error() string`) and `Unwrap() error` method for `LLMError`.
    - **Done‑when:**
        1. `LLMError` struct is defined and compiles.
        2. `LLMError` implements `error` and `Unwrap()`.
    - **Verification:**
        1. Code review confirms struct fields and method signatures.
    - **Depends‑on:** none

- [x] **T002 · Feature · P0: Define ErrorCategory enum**
    - **Context:** Phase 1, Step 1 from PLAN.md (Define Canonical Errors)
    - **Action:**
        1. In `internal/llm/errors.go`, define `ErrorCategory` enum with values: `Unknown`, `Auth`, `RateLimit`, `InvalidRequest`, `NotFound`, `Server`, `Network`, `Cancelled`, `InputLimit`, `ContentFiltered`, `InsufficientCredits`.
    - **Done‑when:**
        1. `ErrorCategory` enum is defined and compiles.
    - **Verification:**
        1. Code review confirms enum values.
    - **Depends‑on:** none

- [x] **T003 · Feature · P0: Define CategorizedError interface and implement on LLMError**
    - **Context:** Phase 1, Step 1 from PLAN.md (Define Canonical Errors)
    - **Action:**
        1. In `internal/llm/errors.go`, define `CategorizedError` interface: `{ error; Category() ErrorCategory }`.
        2. Implement the `Category() ErrorCategory` method on the `LLMError` struct, returning its `ErrCategory` field.
    - **Done‑when:**
        1. `CategorizedError` interface is defined.
        2. `LLMError` struct implements `CategorizedError`.
    - **Verification:**
        1. Ensure an `*LLMError` can be assigned to a `CategorizedError` variable.
    - **Depends‑on:** [T001, T002]

- [x] **T004 · Feature · P1: Implement llm.Wrap helper function**
    - **Context:** Phase 1, Step 1 from PLAN.md (Define Canonical Errors)
    - **Action:**
        1. In `internal/llm/errors.go`, implement `Wrap(originalErr error, provider string, message string, category ErrorCategory, details ...string) error`.
        2. The function should create, populate, and return an `*LLMError`, correctly setting `OriginalError`.
    - **Done‑when:**
        1. `Wrap` function is implemented and compiles.
        2. Function correctly creates and populates an `LLMError`.
    - **Verification:**
        1. Unit tests verify field population and error wrapping.
    - **Depends‑on:** [T001, T002]

- [x] **T005 · Feature · P1: Implement llm.IsCategory helper function**
    - **Context:** Phase 1, Step 1 from PLAN.md (Define Canonical Errors)
    - **Action:**
        1. In `internal/llm/errors.go`, implement `IsCategory(err error, category ErrorCategory) bool`.
        2. Use `errors.As` to check if `err` or its wrapped errors implement `CategorizedError` and match the given category.
    - **Done‑when:**
        1. `IsCategory` function is implemented and compiles.
        2. Function correctly identifies categories in an error chain.
    - **Verification:**
        1. Unit tests verify behavior with direct and wrapped `LLMError` instances.
    - **Depends‑on:** [T003]

- [x] **T006 · Test · P1: Add unit tests for llm error types and helpers**
    - **Context:** Phase 1, Step 1 from PLAN.md; Testing Strategy
    - **Action:**
        1. Create `internal/llm/errors_test.go`.
        2. Write tests for `LLMError` methods (`Error`, `Unwrap`, `Category`).
        3. Write tests for `Wrap` (field population, error chain).
        4. Write tests for `IsCategory` (various scenarios: match, no match, nil error).
    - **Done‑when:**
        1. Unit tests for `internal/llm/errors.go` achieve >95% coverage.
        2. All tests pass.
    - **Verification:**
        1. Review test coverage report.
    - **Depends‑on:** [T001, T002, T003, T004, T005]

## Core Logging Framework (`internal/logutil`)

- [x] **T007 · Feature · P0: Define LoggerInterface with context-aware methods**
    - **Context:** Phase 1, Step 2 from PLAN.md (Setup slog Logging)
    - **Action:**
        1. Create `internal/logutil/logger.go`.
        2. Define `LoggerInterface` with methods like `DebugContext(ctx context.Context, msg string, args ...any)`, `InfoContext`, `WarnContext`, `ErrorContext`.
    - **Done‑when:**
        1. `LoggerInterface` is defined and compiles.
    - **Verification:**
        1. Code review confirms interface method signatures.
    - **Depends‑on:** none

- [x] **T008 · Feature · P1: Implement slog-based JSON logger**
    - **Context:** Phase 1, Step 2 from PLAN.md (Setup slog Logging)
    - **Action:**
        1. In `internal/logutil`, implement a struct (e.g., `SlogLogger`) that satisfies `LoggerInterface`.
        2. Use `log/slog` with `slog.NewJSONHandler` for output.
        3. Provide a constructor (e.g., `NewSlogLogger(writer io.Writer, level slog.Leveler) LoggerInterface`).
    - **Done‑when:**
        1. `SlogLogger` implementation is complete and compiles.
        2. Logger produces structured JSON logs.
    - **Verification:**
        1. Manually inspect sample log output for correct JSON structure and fields (timestamp, level, msg).
    - **Depends‑on:** [T007]

- [x] **T009 · Feature · P1: Implement correlation ID context helpers**
    - **Context:** Phase 1, Step 3 from PLAN.md (Implement Correlation ID)
    - **Action:**
        1. In `internal/logutil/context.go` (or similar), add `WithCorrelationID(ctx context.Context, id string) context.Context`. If `id` is empty, generate a new UUID.
        2. Add `GetCorrelationID(ctx context.Context) string` to retrieve the ID.
    - **Done‑when:**
        1. `WithCorrelationID` and `GetCorrelationID` functions are implemented.
        2. Correlation ID can be set and retrieved from context.
    - **Verification:**
        1. Unit tests verify setting and getting correlation IDs.
    - **Depends‑on:** none

- [x] **T010 · Refactor · P1: Update slog logger to include correlation ID**
    - **Context:** Phase 1, Step 3 from PLAN.md (Implement Correlation ID)
    - **Action:**
        1. Modify the `SlogLogger` implementation.
        2. In each logging method, use `GetCorrelationID` to retrieve the ID from the context.
        3. If a correlation ID exists, add it as a field (e.g., `correlation_id`) to the slog record.
    - **Done‑when:**
        1. Logs include `correlation_id` field when present in context.
    - **Verification:**
        1. Test logging with a context containing a correlation ID and verify its presence in the output.
    - **Depends‑on:** [T008, T009]

- [x] **T011 · Test · P1: Add unit tests for logutil logger and correlation ID**
    - **Context:** Phase 1, Steps 2 & 3 from PLAN.md; Testing Strategy
    - **Action:**
        1. Create `internal/logutil/logger_test.go` and `context_test.go`.
        2. Test `SlogLogger` output structure for different levels and fields.
        3. Test `WithCorrelationID` and `GetCorrelationID` for ID generation, setting, and retrieval.
        4. Test that `SlogLogger` correctly includes the correlation ID from context.
    - **Done‑when:**
        1. Unit tests for `internal/logutil` achieve >90% coverage.
        2. All tests pass.
    - **Verification:**
        1. Review test coverage report.
        2. Manually inspect test log outputs for correctness.
    - **Depends‑on:** [T008, T009, T010]

## Provider Error Handling (`internal/providers/*`)

- [x] **T012 · Refactor · P1: Implement FormatAPIError in provider packages**
    - **Context:** Phase 2, Step 4 from PLAN.md (Refactor Provider Error Handling)
    - **Action:**
        1. For each provider package in `internal/providers/*`, create/update `errors.go`.
        2. Implement `FormatAPIError(rawError error, providerName string) error` function that uses `llm.Wrap` to convert provider-specific errors into `LLMError`.
        3. Update provider client code to call `FormatAPIError` for all API errors.
    - **Done‑when:**
        1. All relevant provider packages have and use `FormatAPIError`.
        2. Provider errors are consistently translated to `LLMError`.
    - **Verification:**
        1. Code review confirms consistent usage across providers.
    - **Depends‑on:** [T004]

- [x] **T013 · Test · P1: Add unit tests for provider error translation**
    - **Context:** Phase 2, Step 4 from PLAN.md; Testing Strategy
    - **Action:**
        1. For each provider, add tests in `errors_test.go` or `client_test.go`.
        2. Test `FormatAPIError` with various mock provider errors, verifying the resulting `LLMError`'s category, message, and wrapped original error.
    - **Done‑when:**
        1. Unit tests cover common error scenarios for each provider.
        2. All tests pass.
    - **Verification:**
        1. Review tests to ensure different error types (auth, rate limit, server error) are mapped correctly.
    - **Depends‑on:** [T012]

## Core Component Refactoring (`internal/thinktank/*`, `internal/auditlog`)

- [x] **T014 · Refactor · P1: Update thinktank/registry for context, logging, and LLMError**
    - **Context:** Phase 3, Step 5 from PLAN.md (Refactor Core Components)
    - **Action:**
        1. Modify method signatures in `internal/thinktank/registry` to accept `context.Context` as the first argument.
        2. Replace existing logging calls with `LoggerInterface` methods (e.g., `logger.InfoContext(ctx, ...)`).
        3. Refactor error handling to return/wrap errors as `LLMError` where appropriate, using `llm.Wrap` or standard wrapping.
    - **Done‑when:**
        1. `internal/thinktank/registry` uses context, `LoggerInterface`, and `LLMError`.
        2. Existing unit/integration tests pass after refactoring.
    - **Verification:**
        1. Review logs generated by registry operations for correlation ID and structured format.
    - **Depends‑on:** [T004, T007, T009]

- [x] **T015 · Refactor · P1: Update thinktank/modelproc for context, logging, and LLMError**
    - **Context:** Phase 3, Step 5 from PLAN.md (Refactor Core Components)
    - **Action:**
        1. Modify method signatures in `internal/thinktank/modelproc` to accept `context.Context`.
        2. Replace existing logging with `LoggerInterface` methods.
        3. Ensure errors (especially from providers) are wrapped using `llm.Wrap` or `fmt.Errorf("%w", ...)` and categorized correctly.
    - **Done‑when:**
        1. `internal/thinktank/modelproc` uses context, `LoggerInterface`, and consistent error wrapping.
        2. Existing unit/integration tests pass.
    - **Verification:**
        1. Test error propagation from modelproc, ensuring errors are `LLMError` or wrap one.
    - **Depends‑on:** [T004, T007, T009, T012]

- [x] **T016 · Refactor · P1: Update thinktank/orchestrator for context, logging, and error aggregation**
    - **Context:** Phase 3, Step 5 from PLAN.md (Refactor Core Components)
    - **Action:**
        1. Modify method signatures in `internal/thinktank/orchestrator` to accept `context.Context`.
        2. Use `LoggerInterface` for all logging.
        3. Aggregate errors from called components (e.g., modelproc) and handle them, potentially using `llm.IsCategory` for decision making.
    - **Done‑when:**
        1. `internal/thinktank/orchestrator` uses context, `LoggerInterface`, and handles aggregated errors.
        2. Existing unit/integration tests pass.
    - **Verification:**
        1. Test orchestrator workflows with simulated component failures, check logs and error handling.
    - **Depends‑on:** [T005, T007, T009, T015]

- [x] **T017 · Refactor · P1: Update auditlog for LoggerInterface and correlation ID**
    - **Context:** Phase 3, Step 6 from PLAN.md (Refactor Audit Logging)
    - **Action:**
        1. Refactor `internal/auditlog.AuditLogger` (or equivalent) to use `LoggerInterface` for its underlying logging.
        2. Ensure all audit log entries include the correlation ID from the context.
        3. Ensure audit logs are structured (JSON).
    - **Done‑when:**
        1. Audit logs are structured JSON and include correlation ID.
        2. Existing audit logging functionality is preserved.
    - **Verification:**
        1. Trigger audit events and verify the log output format and content.
    - **Depends‑on:** [T007, T009]

- [ ] **T018 · Test · P2: Add/update integration tests for core thinktank components**
    - **Context:** Phase 3, Step 5 from PLAN.md; Testing Strategy
    - **Action:**
        1. Review and update existing integration tests for `registry`, `modelproc`, and `orchestrator`.
        2. Add new tests focusing on context propagation, correlation ID in logs, and error handling patterns (wrapping, categorization).
    - **Done‑when:**
        1. Integration tests cover key workflows and verify new logging/error handling.
        2. All tests pass.
    - **Verification:**
        1. Manually inspect logs from test runs for consistency.
    - **Depends‑on:** [T014, T015, T016]

## Input/Output Components (`internal/io`)

- [ ] **T019 · Refactor · P2: Update file I/O components for context and logging**
    - **Context:** Phase 4, Step 7 from PLAN.md (Refactor File I/O)
    - **Action:**
        1. Identify file I/O components/utility functions.
        2. Update their signatures to accept `context.Context`.
        3. Replace any direct logging with `LoggerInterface` methods.
        4. Ensure I/O errors are wrapped appropriately (though may not always be `LLMError` unless directly related to LLM operations).
    - **Done‑when:**
        1. File I/O components use context and `LoggerInterface`.
        2. Tests for I/O operations pass.
    - **Verification:**
        1. Simulate I/O errors and check log output and error wrapping.
    - **Depends‑on:** [T007, T009]

## Clarifications & Assumptions

- [ ] **Issue:** Confirm specific list of provider packages in `internal/providers/*` requiring `FormatAPIError` implementation.
    - **Context:** PLAN.md - Phase 2, Step 4
    - **Blocking?:** no (can proceed with known ones, but full scope needed for T012 completion)
- [ ] **Issue:** Clarify policy on `LLMError` usage for non-LLM specific errors within `thinktank` components (e.g., should internal validation errors become `LLMError` or use standard Go errors?).
    - **Context:** PLAN.md - Architecture Blueprint / Error Handling
    - **Blocking?:** no (default to standard Go errors unless directly tied to an LLM interaction flow)
