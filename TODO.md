# Error Handling and Logging Consistency Implementation Tasks

This document contains the detailed task breakdown for implementing the "Enhance Error Handling and Logging Consistency" epic, organized by component and priority.

## Thinktank Output Usability Improvement

These tasks address the usability issue where thinktank appears to error out despite successfully generating output.

- [x] **T025 · Investigation · P1: Analyze current logging implementation in thinktank/thinktank-wrapper**
    - **Context:** Solution 2 - Log Stream Separation
    - **Action:**
        1. Examine source code to identify how logging is currently implemented
        2. Determine where logs are directed (STDOUT vs STDERR)
        3. Identify the log handler/appender configuration
        4. Document the current logging flow and configuration points
    - **Done‑when:**
        1. Current logging implementation is fully understood and documented
        2. Key points for modification are identified
    - **Verification:**
        1. Validate understanding by tracing some sample log messages through the code
    - **Depends‑on:** none

- [x] **T026 · Feature · P1: Implement proper log stream separation**
    - **Context:** Solution 2 - Log Stream Separation
    - **Action:**
        1. Modify logging configuration to route INFO/DEBUG logs to STDOUT
        2. Route only ERROR/WARN logs to STDERR
        3. Ensure context and correlation IDs are preserved in both streams
        4. Update any custom logging handlers to respect this separation
    - **Done‑when:**
        1. INFO/DEBUG logs appear on STDOUT
        2. Only ERROR/WARN logs appear on STDERR
        3. Existing logging functionality is otherwise preserved
    - **Verification:**
        1. Run thinktank with various scenarios and verify stream routing
        2. Test with Claude's Bash tool to confirm error reporting is accurate
    - **Depends‑on:** [T025]

- [x] **T027 · Feature · P1: Add tolerant mode flag to thinktank-wrapper**
    - **Context:** Solution 4 - Tolerant Mode
    - **Action:**
        1. Add `--partial-success-ok` flag to CLI argument parsing
        2. Update help documentation to explain the flag
        3. Add configuration field to store flag value
        4. Pass this configuration to relevant components
    - **Done‑when:**
        1. Flag is properly parsed from command line
        2. Help documentation includes clear explanation
        3. Configuration is made available to exit code determination logic
    - **Verification:**
        1. Run with `--help` to confirm flag is documented
        2. Test parsing with and without flag
    - **Depends‑on:** none

- [x] **T028 · Feature · P1: Modify exit code logic based on tolerant mode**
    - **Context:** Solution 4 - Tolerant Mode
    - **Action:**
        1. Identify where exit code is determined in thinktank-wrapper
        2. Add logic to consider partial success as success when flag is enabled
        3. Return exit code 0 if synthesis file was generated, even if some models failed
        4. Preserve existing strict behavior when flag is not used
    - **Done‑when:**
        1. With flag enabled: exit code 0 if synthesis file exists
        2. Without flag: original exit code behavior is preserved
    - **Verification:**
        1. Test partial success scenarios with and without flag
        2. Verify exit codes match expected behavior
    - **Depends‑on:** [T027]

- [x] **T029 · Feature · P2: Implement improved results summary output**
    - **Context:** Solution 4 - Improved Summary
    - **Action:**
        1. Create code to track individual model successes/failures
        2. Design a concise summary format showing success/failure counts
        3. Include path to synthesis file when it exists
        4. Add optional terminal color coding (green/red) for success/failure
        5. Ensure summary appears as final output regardless of exit code
    - **Done‑when:**
        1. Clear summary is displayed at end of execution
        2. Summary shows success/failure counts and synthesis path
        3. Summary is visible in both success and failure cases
    - **Verification:**
        1. Test with varying numbers of successful/failed models
        2. Confirm summary clarity and visibility
    - **Depends‑on:** [T026, T028]

- [x] **T030 · Test · P2: Add comprehensive tests for improved output handling**
    - **Context:** Testing for Solutions 2 & 4
    - **Action:**
        1. Create tests for proper log stream routing
        2. Add tests for tolerant mode flag behavior
        3. Implement tests for exit code determination logic
        4. Create tests for summary generation
        5. Test integration of all components
        6. Fix package import issues with orchestrator dependencies
    - **Done‑when:**
        1. Test coverage for new features is >90%
        2. All tests pass without import errors or build failures
        3. Test infrastructure properly handles module dependencies
    - **Verification:**
        1. Review test coverage report
        2. Manual verification of key scenarios
        3. Verify build runs cleanly without vendor directory issues
    - **Depends‑on:** [T026, T028, T029]

- [x] **T031 · Docs · P2: Update documentation for output handling improvements**
    - **Context:** User Documentation for Solutions 2 & 4
    - **Action:**
        1. Update CLI help text for new flag
        2. Add explanation of exit code behavior to documentation
        3. Document the meaning of the summary output
        4. Update any relevant README files
    - **Done‑when:**
        1. Documentation accurately reflects new features
        2. Help text is clear and helpful
    - **Verification:**
        1. Review documentation for clarity and completeness
    - **Depends‑on:** [T027, T029]

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

- [ ] **T010 · Refactor · P1: Update slog logger to include correlation ID**
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

- [ ] **T011 · Test · P1: Add unit tests for logutil logger and correlation ID**
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

- [ ] **T012 · Refactor · P1: Implement FormatAPIError in provider packages**
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

- [ ] **T013 · Test · P1: Add unit tests for provider error translation**
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

- [ ] **T014 · Refactor · P1: Update thinktank/registry for context, logging, and LLMError**
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

- [ ] **T015 · Refactor · P1: Update thinktank/modelproc for context, logging, and LLMError**
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

- [ ] **T016 · Refactor · P1: Update thinktank/orchestrator for context, logging, and error aggregation**
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

- [ ] **T017 · Refactor · P1: Update auditlog for LoggerInterface and correlation ID**
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

## Input/Output & Top-Level Application (`internal/io`, `cmd/thinktank`)

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

- [ ] **T020 · Refactor · P1: Setup initial context and logger in cmd/thinktank**
    - **Context:** Phase 1, Step 3 & Phase 4, Step 8 from PLAN.md
    - **Action:**
        1. In `cmd/thinktank/main.go`, create the root `context.Context`.
        2. Initialize it with a correlation ID using `logutil.WithCorrelationID`.
        3. Instantiate the `LoggerInterface` (e.g., `SlogLogger`) and `AuditLogger`.
        4. Pass the root context and loggers to top-level application components.
    - **Done‑when:**
        1. Application entry point correctly initializes context and loggers.
        2. Correlation ID is generated/set at startup.
    - **Verification:**
        1. Run the application and observe the first logs to ensure correlation ID is present.
    - **Depends‑on:** [T010, T017]

- [ ] **T021 · Feature · P1: Implement top-level error handling in cmd/thinktank**
    - **Context:** Phase 4, Step 8 from PLAN.md (Refactor Top-Level Error Handling)
    - **Action:**
        1. In `cmd/thinktank/main.go`, implement a central error handling mechanism for errors bubbling up to the main function.
        2. Log the full error using `LoggerInterface` (which should include sanitization if T022 is done).
        3. Based on `llm.IsCategory` or error type, determine a user-friendly message and an appropriate exit code.
        4. Print the user-friendly message to stderr.
    - **Done‑when:**
        1. Application exits with appropriate codes and user messages for different error types.
        2. Detailed errors are logged.
    - **Verification:**
        1. Manually trigger different error scenarios (e.g., auth failure, file not found) and check CLI output, exit code, and logs.
    - **Depends‑on:** [T005, T010, T022]

## Security & Cleanup

- [x] **T022 · Feature · P1: Implement error detail sanitization in logutil**
    - **Context:** Phase 5, Step 9 from PLAN.md (Implement Sanitization)
    - **Action:**
        1. In `internal/logutil`, create logic to sanitize sensitive information (e.g., API keys, secrets) from error messages or details before they are logged.
        2. This could be a `SanitizingLogger` wrapper around `LoggerInterface` or a handler option for `slog`.
        3. Define patterns for secrets to be detected and masked.
    - **Done‑when:**
        1. Sanitization logic is implemented and integrated into the logging pipeline.
        2. Secrets are masked in log outputs.
    - **Verification:**
        1. Write unit tests that attempt to log errors/messages containing fake secrets and verify they are masked in the output.
    - **Depends‑on:** [T008]

- [ ] **T023 · Chore · P2: Audit codebase and remove legacy logging/error handling**
    - **Context:** Phase 5, Step 10 from PLAN.md (Perform Codebase Audit)
    - **Action:**
        1. Search the entire codebase for old logging patterns (e.g., `fmt.Println`, `log.Printf`, direct `log` package usage for application logging).
        2. Replace them with the new `LoggerInterface`.
        3. Ensure error handling consistently uses `LLMError` or standard wrapping.
    - **Done‑when:**
        1. Legacy logging and inconsistent error handling are eliminated.
        2. Application functionality remains intact.
    - **Verification:**
        1. Code review and static analysis confirm removal of old patterns.
        2. Full test suite passes.
    - **Depends‑on:** [T006, T011, T013, T014, T015, T016, T017, T019, T020, T021, T022]

- [ ] **T024 · Chore · P2: Update documentation for error and logging standards**
    - **Context:** Phase 5, Step 11 from PLAN.md (Update Documentation)
    - **Action:**
        1. Update `DEVELOPMENT_PHILOSOPHY.md` (or similar) with new error handling patterns (using `LLMError`, `Wrap`, `IsCategory`).
        2. Document the structured logging format, mandatory fields (like `correlation_id`), and `ErrorCategory` enum in `README.md` or a dedicated logging document.
    - **Done‑when:**
        1. Documentation accurately reflects the new error handling and logging standards.
    - **Verification:**
        1. Review documentation for clarity, accuracy, and completeness.
    - **Depends‑on:** [T006, T011, T022, T023]

## Cleanup Tasks Related to T028 Build Issues

- [x] **T032 · Cleanup · P0: Remove refactored duplicate files**
    - **Context:** Build and test errors from unintentional file duplication
    - **Action:**
        1. Remove all *_refactored* files to prevent duplicate declarations
        2. Ensure proper merging of any needed changes from refactored files
        3. Fix issues with circular imports between packages
    - **Done‑when:**
        1. Build succeeds without duplicate declaration errors
        2. All tests pass without import errors
    - **Verification:**
        1. Run complete test suite
        2. Confirm clean build
    - **Depends‑on:** none

- [x] **T033 · Refactor · P1: Fix import paths and package structure**
    - **Context:** Package organization and vendor issues
    - **Action:**
        1. Resolve issues with import paths for internal packages
        2. Ensure vendor directory is properly configured
        3. Fix circular dependencies if present
        4. Review dependency management approach (vendor vs. go modules)
    - **Done‑when:**
        1. All imports resolve correctly
        2. No vendor-related errors in build process
    - **Verification:**
        1. Run clean build with -mod=vendor
    - **Depends‑on:** [T032]

- [x] **T036 · Fix · P1: Fix synthesis model response truncation**
    - **Context:** Synthesis models frequently truncate responses when combining outputs from multiple models
    - **Action:**
        1. Identify where token limits are enforced in the codebase
        2. Update the GetModelTokenLimits method to use actual values from model definition
        3. Add the ContextWindow and MaxOutputTokens fields to ModelDefinition struct
        4. Set very high default values for models without explicit limits
    - **Done‑when:**
        1. The synthesis prompt can handle much larger inputs and outputs
        2. Truncation issues are resolved for large synthesis operations
    - **Verification:**
        1. Run synthesis tests with multiple model outputs
        2. Verify models use their configured token limits from models.yaml
    - **Depends‑on:** none

- [ ] **T034 · Refactor · P1: Update tests for error handling improvements**
    - **Context:** Missing or incompatible tests for new error handling features
    - **Action:**
        1. Create comprehensive tests for all error handling components
        2. Ensure tests work properly with both string-based and type-based error detection
        3. Verify exit code handling across different error scenarios
    - **Done‑when:**
        1. Complete test coverage for error handling features
        2. All tests pass consistently
    - **Verification:**
        1. Verify tests for different error handling scenarios
    - **Depends‑on:** [T032, T033]

- [x] **T035 · Fix · P2: Fix flaky TestGenerateTimestampedRunNameUniqueness test**
    - **Context:** The test occasionally fails due to non-deterministic random number generation
    - **Action:**
        1. Review the implementation of `generateTimestampedRunName` in `app.go`
        2. Fix the randomness mechanism to ensure uniqueness
        3. Update the test to be more robust against timing issues
    - **Done‑when:**
        1. The test passes consistently across multiple runs
    - **Verification:**
        1. Run the test multiple times to verify consistency
    - **Depends‑on:** none

## Clarifications & Assumptions

- [ ] **Issue:** Confirm specific list of provider packages in `internal/providers/*` requiring `FormatAPIError` implementation.
    - **Context:** PLAN.md - Phase 2, Step 4
    - **Blocking?:** no (can proceed with known ones, but full scope needed for T012 completion)
- [ ] **Issue:** Define precise patterns for secret detection and masking for sanitization logic (T022).
    - **Context:** PLAN.md - Phase 5, Step 9
    - **Blocking?:** no (can start with common patterns, but refinement needed)
- [ ] **Issue:** Standardize mandatory fields and their exact names for all structured logs (beyond `timestamp`, `level`, `msg`, `correlation_id`).
    - **Context:** PLAN.md - Logging & Observability
    - **Blocking?:** no (can use defaults, but consistency is key for log processing)
- [ ] **Issue:** Determine if stack traces should be included in logs for certain error severities/categories, and if so, how (e.g., specific field, configurable).
    - **Context:** PLAN.md - Logging & Observability
    - **Blocking?:** no
- [ ] **Issue:** Clarify policy on `LLMError` usage for non-LLM specific errors within `thinktank` components (e.g., should internal validation errors become `LLMError` or use standard Go errors?).
    - **Context:** PLAN.md - Architecture Blueprint / Error Handling
    - **Blocking?:** no (default to standard Go errors unless directly tied to an LLM interaction flow)
