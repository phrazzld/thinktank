# Remediation Plan – Sprint 1

## Executive Summary
This sprint fixes a critical flaw in the orchestrator where partial model failures were silently treated as success, breaking CLI, scripts, and CI pipelines. We will first change the `Run` method to always return a composite error when any model fails, ensuring no errors are masked. Immediately after, we will add targeted unit tests to cover all partial-failure paths and prevent regressions.

## Strike List
| Seq | CR-ID | Title                                          | Effort | Owner   |
|-----|-------|------------------------------------------------|--------|---------|
| 1   | cr-01 | Do not swallow errors on partial model failures | s      | backend |
| 2   | cr-02 | Add unit tests for partial-failure error path   | s      | backend |

## Detailed Remedies

### cr-01 Do not swallow errors on partial model failures
- **Problem:** The orchestrator's `Run` returns `nil` whenever at least one model succeeds, even if others fail.
- **Impact:** Partial failures go undetected, causing CLI/CI tooling to report success and masking real issues.
- **Chosen Fix:** After processing all models, check for any errors and, if present, return a composite error summarizing success vs. failure counts along with underlying messages.
- **Steps:**
  1. In `internal/thinktank/orchestrator/orchestrator.go` (lines 108–116), collect outputs and errors from `processModels`.
  2. If `len(modelErrors) > 0`, call a helper (e.g., `aggregateErrorMessages(modelErrors)`) and return:
     ```go
     fmt.Errorf(
       "processed %d/%d models successfully; %d failed: %v",
       len(modelOutputs), len(o.config.ModelNames), len(modelErrors), aggregateErrorMessages(modelErrors),
     )
     ```
  3. Ensure downstream callers (CLI, scripts) propagate non-zero exit codes on this error.
  4. Add a warning-level log when partial failures occur.
- **Done-When:**
  - Any model failure yields a non-nil error from `Run`.
  - CLI and CI detect and report partial-failure runs as failures.
  - Warning logs appear with failure summary.

### cr-02 Add unit tests for partial-failure error path
- **Problem:** The new partial-failure logic in `Run` is untested.
- **Impact:** Future changes may reintroduce silent-error swallowing without detection.
- **Chosen Fix:** Write table-driven unit tests simulating mixed success and failure in `processModels`, asserting correct error return, logs, and output handling.
- **Steps:**
  1. Create or extend `internal/thinktank/orchestrator/orchestrator_run_test.go`.
  2. Mock or stub `processModels` (or the underlying model calls) to return:
     - All succeed
     - Mixed succeed/fail
     - All fail
  3. For each scenario, assert:
     - `Run` returns the expected composite error when failures exist.
     - `Run` returns `nil` only when all succeed.
     - Warning logs are emitted for partial failures.
     - Successful outputs are still passed to the next step.
- **Done-When:**
  - Tests covering lines 108–116 pass and fail appropriately on regressions.
  - CI shows 100% coverage for partial-failure branches.

## Standards Alignment
- Simplicity: Adds minimal code around existing `Run` logic, keeping error-handling straightforward.
- Design for Observability: Exposes partial failures via return values and warning logs.
- Design for Testability: Introduces focused tests for every partial-failure branch.
- Coding Standards: Follows idiomatic Go error patterns; no panics or hidden control flows.
- Security: No sensitive data is leaked in error messages or logs.

## Validation Checklist
- [ ] Automated tests (unit and integration) pass, including new tests.
- [ ] `golangci-lint` and `go vet` show no new warnings.
- [ ] Manual test: simulate model failures and confirm CLI/CI exit code is non-zero with correct error summary.
- [ ] No new lint, audit, or security warnings introduced.
