# Todo

## Orchestrator Error Handling

- [x] **T001 · Bugfix · P0: implement aggregateErrorMessages helper**
    - **Context:** PLAN.md cr-01 Step 2 (helper to summarize modelErrors)
    - **Action:**
        1. In `internal/thinktank/orchestrator/orchestrator.go`, add a new function `aggregateErrorMessages(errs []error) string`.
        2. Concatenate each `err.Error()` with a separator (e.g., `"; "`).
    - **Done-when:**
        1. Helper properly formats multiple errors into a single message string.
    - **Depends-on:** none

- [x] **T002 · Bugfix · P0: refactor orchestrator run to return composite error on partial failure**
    - **Context:** PLAN.md cr-01 Steps 1-2 (always return composite error if any model fails)
    - **Action:**
        1. In `internal/thinktank/orchestrator/orchestrator.go` (lines 108–116), modify `Run` to collect outputs and errors from `processModels`.
        2. If `len(modelErrors) > 0`, return a formatted error using the helper:
           ```go
           fmt.Errorf(
             "processed %d/%d models successfully; %d failed: %v",
             len(modelOutputs), len(o.config.ModelNames), len(modelErrors),
             aggregateErrorMessages(modelErrors),
           )
           ```
    - **Done-when:**
        1. `Run` returns a non-nil error whenever at least one model fails.
        2. The error message includes success count, failure count, and aggregated messages.
        3. `Run` returns `nil` only if all models succeed.
    - **Depends-on:** [T001]

- [x] **T003 · Bugfix · P0: add warning logs for partial model failures**
    - **Context:** PLAN.md cr-01 Step 4 (warning-level log on partial failures)
    - **Action:**
        1. After detecting `len(modelErrors) > 0` in `Run`, emit a warning log via the structured logger.
        2. Include both success and failure counts in the log entry.
    - **Done-when:**
        1. A warning-level structured log appears whenever `Run` processes a mixture of successful and failed models.
        2. Log entries contain relevant context about which models succeeded and which failed.
    - **Depends-on:** [T002]

- [x] **T004 · Bugfix · P0: update CLI to exit non-zero on orchestrator partial failure**
    - **Context:** PLAN.md cr-01 Step 3 (ensure downstream callers propagate error)
    - **Action:**
        1. In the CLI entrypoint (likely `cmd/thinktank/main.go`), check the error returned by `orchestrator.Run`.
        2. If the error is non-nil, print the error message to stderr and ensure the CLI process exits with a non-zero status code (e.g., `os.Exit(1)`).
    - **Done-when:**
        1. Running the CLI command with a failing model results in a non-zero exit code.
        2. The composite error message from the orchestrator is printed to standard error.
        3. Manual test with mixed-failure scenario confirms exit status is non-zero.
    - **Depends-on:** [T002]

- [x] **T005 · Test · P0: add unit tests for partial-failure scenarios**
    - **Context:** PLAN.md cr-02 Steps 1-3 (test mixed success/failure paths)
    - **Action:**
        1. Create or extend `internal/thinktank/orchestrator/orchestrator_run_test.go` with table-driven tests.
        2. Mock/stub `processModels` to simulate three scenarios: all succeed, mixed succeed/fail, all fail.
        3. For each scenario, assert that:
           - `Run` returns `nil` only when all succeed; returns expected composite error otherwise.
           - Warning logs are emitted for partial failures.
           - Successful outputs are correctly passed to the next processing step.
    - **Done-when:**
        1. Unit tests cover all branches of the error handling logic (lines 108-116).
        2. Tests pass and demonstrate correct error/nil return values for all scenarios.
        3. CI shows 100% coverage for the partial-failure code paths.
    - **Depends-on:** [T002, T003]
