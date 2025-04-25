# Remediation Plan – Sprint 1

## Executive Summary
This sprint fixes critical correctness and reliability bugs in the orchestrator’s model-processing and I/O paths, then rapidly improves observability and audit logging. We start by repairing partial-failure counting and blank-output propagation (cr-01, cr-02), add fast CLI validation and proper error surfacing for file writes (cr-03, cr-04), then enhance logging with correlation IDs and centralized audit events (cr-07, cr-08). Foundational refactors to modularize the orchestrator (cr-05), harden tests (cr-06), and sweep minor nits (cr-09) follow to restore simplicity, testability, and code-style compliance.

## Strike List

| Seq | CR-ID | Title                                            | Effort | Owner   |
|-----|-------|--------------------------------------------------|--------|---------|
| 1   | cr-01 | Fix partial-failure success counting             | s      | backend |
| 2   | cr-02 | Exclude failed models from synthesis prompt      | s      | backend |
| 3   | cr-03 | Fail fast on invalid `--synthesis-model`         | s      | backend |
| 4   | cr-04 | Surface file-save errors to exit code            | s      | backend |
| 5   | cr-07 | Add `correlation_id` to all structured logs      | s      | backend |
| 6   | cr-08 | Centralize and trim audit logging                | s      | backend |
| 7   | cr-05 | Extract SynthesisService & OutputWriter          | m      | backend |
| 8   | cr-06 | Refactor tests to mock only external boundaries  | m      | backend |
| 9   | cr-09 | Minor naming, lint & documentation fixes         | xs     | backend |

## Detailed Remedies

### cr-01 Fix partial-failure success counting
- **Problem:** `Run` and its logs use `len(modelOutputs)` (including empty entries) to report success, so partial failures always appear as 100% success.
- **Impact:** Silent logic bombs: failed models are never detected or surfaced, undermining reliability.
- **Chosen Fix:**
  • In `processModels`, only add entries for models where `err == nil`.
  • In `Run`, compute `successCount` from the slice/map of truly successful models.
  • Update warning/error logs and return value to use the new counts.
- **Steps:**
  1. Change `processModels` result-channel handling to `if err==nil { emit result }`.
  2. Build `successfulModelsSlice` and use `len(successfulModelsSlice)` in `Run`.
  3. Adjust all logs and error messages to reflect accurate success/fail counts.
  4. Add unit tests covering 0%, partial, and 100% success scenarios.
- **Done-When:**
  • Partial-failure tests pass, logs show correct counts; no empty entries in `modelOutputs`.

### cr-02 Exclude failed models from synthesis prompt
- **Problem:** `processModels` inserts empty strings for failed models into `modelOutputs`, and synthesis includes blank `<model_result>` sections.
- **Impact:** Misleading or confusing prompts cause poor summaries and potential misinterpretation by the LLM.
- **Chosen Fix:** Only include successful model results in the map passed to synthesis.
- **Steps:**
  1. Filter out `result.content` when `err != nil`; do not add blank entries.
  2. In prompt builder, iterate only over `modelOutputs` keys for which content exists.
  3. Write a test verifying that failed-model keys are absent from the final prompt.
- **Done-When:**
  • Synthesized prompt contains no placeholders or blank sections for failures; related test passes.

### cr-03 Fail fast on invalid `--synthesis-model`
- **Problem:** CLI validation falls back silently when registry is nil, accepting any synthesis-model string.
- **Impact:** Long runs and wasted API calls before encountering an unsupported model, degrading UX.
- **Chosen Fix:** Always validate the flag against a static or configured whitelist; error out immediately if invalid.
- **Steps:**
  1. Introduce a constant list of allowed synthesis models or load from config.
  2. In `ValidateInputsWithEnv`, if `regManager==nil` or lookup fails, return a clear error:
     `invalid synthesis model: %q`.
  3. Update CLI tests to assert that invalid values exit with non-zero code and message.
- **Done-When:**
  • CLI rejects bad `--synthesis-model` immediately; tests confirm failure path.

### cr-04 Surface file-save errors to exit code
- **Problem:** Write failures (disk full, perms) are only logged, not propagated to the exit status—users assume success despite data loss.
- **Impact:** Silent data loss and mismatched user expectations.
- **Chosen Fix:** Collect save errors in loops, then return a summarized error and non-zero exit.
- **Steps:**
  1. In both individual and synthesis save loops, track `failCount`.
  2. After saving all files, if `failCount > 0`, return `fmt.Errorf("%d/%d files failed to save", failCount, total)`.
  3. Ensure `main` or CLI runner calls `os.Exit(1)` on non-nil error.
  4. Add tests simulating permission/disk errors verifying non-zero exit and correct message.
- **Done-When:**
  • Simulated save failures surface an error and exit code ≠ 0; tests pass.

### cr-07 Add `correlation_id` to all structured logs
- **Problem:** Logs omit a `correlation_id`, thwarting end-to-end request tracing.
- **Impact:** Poor observability—hard to follow a single CLI run across concurrent operations.
- **Chosen Fix:** Generate or extract a `correlation_id` at `Run` entry, store it in `context.Context`, and have every log call include it.
- **Steps:**
  1. In `Orchestrator.Run`, generate a UUID or accept a passed-in ID and insert into `ctx`.
  2. Pass `ctx` through all downstream calls; update logger calls to `WithContext(ctx)` or include `ctx.Value("correlation_id")`.
  3. Update unit tests for logger stubs to assert presence of the field.
- **Done-When:**
  • Every `Info/Warn/Error` call logs `correlation_id`; trace tests confirm consistency.

### cr-08 Centralize and trim audit logging
- **Problem:** Excessive, low-value audit events clutter logs and slow execution.
- **Impact:** Harder debugging, inflated log volumes, violation of simplicity.
- **Chosen Fix:** Audit only high-value steps (Start, API call, End) via a single helper.
- **Steps:**
  1. Remove granular audit calls; leave only key events.
  2. Create `AuditLogger.Log(operation, status)` helper.
  3. Replace in-line audit calls with helper invocations.
  4. Add tests for helper outputs.
- **Done-When:**
  • Audit logs show only Start/API/End events; helper covers all cases.

### cr-05 Extract SynthesisService & OutputWriter
- **Problem:** `orchestrator.go` exceeds SRP, mixing context gathering, concurrency, prompt building, synthesis, file I/O, and logging.
- **Impact:** Poor modularity, testability, and maintainability.
- **Chosen Fix:**
  • Introduce `SynthesisService` for synthesis logic.
  • Introduce `OutputWriter` for file writes.
  • Refactor `Orchestrator.Run` to orchestrate these interfaces and remain <100 lines.
- **Steps:**
  1. Define interfaces and move `synthesizeResults` into `synthesis_service.go`.
  2. Move save loops into `output_writer.go`.
  3. Inject instances into `Orchestrator`.
  4. Update dependency wiring and tests.
- **Done-When:**
  • `orchestrator.go` shrinks significantly; new services are independently unit-tested; all tests pass.

### cr-06 Refactor tests to mock only external boundaries
- **Problem:** Integration tests mock internal services (ContextGatherer, ModelProcessor), hiding real interactions.
- **Impact:** Missed coupling issues; false confidence; high maintenance.
- **Chosen Fix:**
  • Define clear boundary interfaces (HTTP client, FS).
  • Mock only those; use real internal implementations in tests.
- **Steps:**
  1. Identify and remove mocks of internal collaborators.
  2. Introduce boundary interfaces and adjust production code.
  3. Update tests to use real components + lightweight fakes at boundaries.
  4. Verify full-flow integration tests pass without internal mocks.
- **Done-When:**
  • No test mocks of internal packages; integration tests cover real orchestration.

### cr-09 Minor naming, lint & documentation fixes
- **Problem:** Inconsistent export naming, unused imports, redundant tests, `fmt.Errorf` used without verbs, overlapping docs.
- **Impact:** Linter warnings, confusion, minor maintenance friction.
- **Chosen Fix:**
  • Align function names (`SanitizeFilename` vs `sanitizeFilename`).
  • Remove dead imports; merge redundant tests into tables.
  • Replace `fmt.Errorf("static")` with `errors.New`.
  • Consolidate `TODO.md` into `PLAN.md`.
- **Steps:**
  1. Apply lint fixes and run `golangci-lint`.
  2. Refactor test suites for table-driven subtests.
  3. Update docs and remove `TODO.md`.
- **Done-When:**
  • Code is lint-clean; no warnings; documentation is unambiguous.

## Standards Alignment
- Simplicity: Early bug fixes remove hidden logic; audit logging trimmed.
- Modularity: Synthesis and output concerns extracted; test boundaries tightened.
- Testability: New and updated tests cover real flows, partial failures, and error paths.
- Explicitness: Fail-fast CLI validation; explicit error returns for I/O.
- Logging Strategy: All logs carry `correlation_id`; audit via centralized helper.
- Security & Reliability: No silent failures; clear user feedback on error conditions.

## Validation Checklist
- [ ] Unit & integration tests pass with new coverage for partial failures, CLI validation, and save errors.
- [ ] `golangci-lint`, `go vet`, and `staticcheck` show zero new issues.
- [ ] Manual CLI validation:
  - Partial failures report correct counts.
  - Synthesis prompt omits blanks.
  - Invalid `--synthesis-model` fails immediately.
  - File‐save errors produce non-zero exit.
  - Logs include consistent `correlation_id`.
- [ ] Documentation updated; `PLAN.md` reflects all tasks; `TODO.md` removed.
