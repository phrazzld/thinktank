# Top-Level Handling, Output Usability & Finalization Tasks

This document contains the detailed task breakdown for implementing the second part of the "Enhance Error Handling and Logging Consistency" epic, focusing on top-level handling, output usability improvements, and finalization of the transition.

## Thinktank Output Usability Improvement

These tasks address the usability issue where thinktank appears to error out despite successfully generating output.

- [ ] **T025 · Investigation · P1: Analyze current logging implementation in thinktank/thinktank-wrapper**
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

- [ ] **T026 · Feature · P1: Implement proper log stream separation**
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

- [ ] **T027 · Feature · P1: Add tolerant mode flag to thinktank-wrapper**
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

- [ ] **T028 · Feature · P1: Modify exit code logic based on tolerant mode**
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

- [ ] **T029 · Feature · P2: Implement improved results summary output**
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

- [ ] **T030 · Test · P2: Add comprehensive tests for improved output handling**
    - **Context:** Testing for Solutions 2 & 4
    - **Action:**
        1. Create tests for proper log stream routing
        2. Add tests for tolerant mode flag behavior
        3. Implement tests for exit code determination logic
        4. Create tests for summary generation
        5. Test integration of all components
    - **Done‑when:**
        1. Test coverage for new features is >90%
        2. Tests pass for all new functionality
    - **Verification:**
        1. Review test coverage report
        2. Manual verification of key scenarios
    - **Depends‑on:** [T026, T028, T029]

- [ ] **T031 · Docs · P2: Update documentation for output handling improvements**
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

## Top-Level Application (`cmd/thinktank`)

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
    - **Depends‑on:** [T010, T017] (From TODO-1.md)

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
    - **Depends‑on:** [T005, T010, T022] (T005 and T010 from TODO-1.md)

## Security & Cleanup

- [ ] **T022 · Feature · P1: Implement error detail sanitization in logutil**
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
    - **Depends‑on:** [T008] (From TODO-1.md)

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
    - **Depends‑on:** All tasks from TODO-1.md and [T020, T021, T022]

- [ ] **T024 · Chore · P2: Update documentation for error and logging standards**
    - **Context:** Phase 5, Step 11 from PLAN.md (Update Documentation)
    - **Action:**
        1. Update `DEVELOPMENT_PHILOSOPHY.md` (or similar) with new error handling patterns (using `LLMError`, `Wrap`, `IsCategory`).
        2. Document the structured logging format, mandatory fields (like `correlation_id`), and `ErrorCategory` enum in `README.md` or a dedicated logging document.
    - **Done‑when:**
        1. Documentation accurately reflects the new error handling and logging standards.
    - **Verification:**
        1. Review documentation for clarity, accuracy, and completeness.
    - **Depends‑on:** [T006, T011, T022, T023] (T006 and T011 from TODO-1.md)

## Clarifications & Assumptions

- [ ] **Issue:** Define precise patterns for secret detection and masking for sanitization logic (T022).
    - **Context:** PLAN.md - Phase 5, Step 9
    - **Blocking?:** no (can start with common patterns, but refinement needed)
- [ ] **Issue:** Standardize mandatory fields and their exact names for all structured logs (beyond `timestamp`, `level`, `msg`, `correlation_id`).
    - **Context:** PLAN.md - Logging & Observability
    - **Blocking?:** no (can use defaults, but consistency is key for log processing)
- [ ] **Issue:** Determine if stack traces should be included in logs for certain error severities/categories, and if so, how (e.g., specific field, configurable).
    - **Context:** PLAN.md - Logging & Observability
    - **Blocking?:** no
