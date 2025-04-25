# Todo

## processModels success counting
- [x] **T001 · Bugfix · P0: fix partial-failure success counting**
    - **Context:** Detailed Remedies → cr-01 → Steps 1-4
    - **Action:**
        1. In `processModels`, emit results only when `err == nil`.
        2. In `Run`, build a `successfulModelsSlice` and compute `successCount` from its length.
        3. Update all logs and error messages in `Run` to use the new counts.
        4. Write unit tests for 0%, partial, and 100% success scenarios.
    - **Done-when:**
        1. No empty entries in `modelOutputs`.
        2. Partial-failure tests pass and logs reflect accurate success/fail counts.
    - **Depends-on:** none

## synthesis prompt filtering
- [x] **T002 · Bugfix · P1: exclude failed models from synthesis prompt**
    - **Context:** Detailed Remedies → cr-02 → Steps 1-3
    - **Action:**
        1. Filter out `result.content` when `err != nil` in `processModels`.
        2. In the prompt builder, iterate only over keys with non-empty content.
        3. Add a test verifying no blank or placeholder sections in the final prompt.
    - **Done-when:**
        1. Synthesized prompt contains no entries for failed models.
        2. Related test passes.
    - **Depends-on:** none

## cli input validation
- [x] **T003 · Bugfix · P1: fail fast on invalid `--synthesis-model`**
    - **Context:** Detailed Remedies → cr-03 → Steps 1-3
    - **Action:**
        1. Define a constant whitelist of allowed synthesis models or load from config.
        2. In `ValidateInputsWithEnv`, return `invalid synthesis model: %q` error if lookup fails or `regManager==nil`.
        3. Update CLI tests to assert non-zero exit code and appropriate error message on invalid input.
    - **Done-when:**
        1. CLI rejects invalid `--synthesis-model` immediately.
        2. Tests confirm failure path with correct exit code and message.
    - **Depends-on:** none

## output writer error propagation
- [ ] **T004 · Bugfix · P1: surface file-save errors to exit code**
    - **Context:** Detailed Remedies → cr-04 → Steps 1-4
    - **Action:**
        1. In individual and synthesis save loops, track a `failCount`.
        2. After all saves, if `failCount > 0`, return `fmt.Errorf("%d/%d files failed to save", failCount, total)`.
        3. Ensure `main` or CLI runner calls `os.Exit(1)` on non-nil error.
        4. Write tests simulating permission/disk errors to verify non-zero exit and error message.
    - **Done-when:**
        1. Simulated save failures cause non-zero exit and correct error message.
        2. Tests pass.
    - **Depends-on:** none

## structured logging
- [ ] **T005 · Feature · P2: add correlation_id to all structured logs**
    - **Context:** Detailed Remedies → cr-07 → Steps 1-3
    - **Action:**
        1. In `Orchestrator.Run`, generate or accept a `correlation_id` and store in `context.Context`.
        2. Pass `ctx` through downstream calls; update logger calls to include `ctx.Value("correlation_id")` or use `WithContext`.
        3. Update unit tests for log stubs to assert the presence of `correlation_id`.
    - **Done-when:**
        1. Every structured log entry (Info/Warn/Error) includes `correlation_id`.
        2. Trace tests confirm consistent ID propagation.
    - **Depends-on:** none

## audit logging
- [ ] **T006 · Feature · P2: centralize and trim audit logging**
    - **Context:** Detailed Remedies → cr-08 → Steps 1-4
    - **Action:**
        1. Remove granular audit calls, retaining only Start/API call/End events.
        2. Implement an `AuditLogger.Log(operation, status)` helper.
        3. Replace in-line audit calls with helper invocations.
        4. Add tests to verify helper output for each key event.
    - **Done-when:**
        1. Audit logs show only Start, API call, and End events.
        2. Helper tests pass.
    - **Depends-on:** none

## orchestrator refactoring
- [ ] **T007 · Refactor · P2: extract SynthesisService & OutputWriter**
    - **Context:** Detailed Remedies → cr-05 → Steps 1-4
    - **Action:**
        1. Define `SynthesisService` interface; move `synthesizeResults` into `synthesis_service.go`.
        2. Define `OutputWriter` interface; move file save loops into `output_writer.go`.
        3. Inject these services into `Orchestrator.Run`, reducing its size to <100 lines.
        4. Write unit tests for both new services.
    - **Done-when:**
        1. `orchestrator.go` is <100 lines and orchestrates via interfaces.
        2. `SynthesisService` and `OutputWriter` are independently tested.
        3. All existing tests pass.
    - **Depends-on:** none

## tests mocking boundaries
- [ ] **T008 · Refactor · P2: refactor tests to mock only external boundaries**
    - **Context:** Detailed Remedies → cr-06 → Steps 1-4
    - **Action:**
        1. Remove mocks of internal collaborators in integration tests.
        2. Define boundary interfaces (HTTP client, filesystem) and adjust code injection.
        3. Update tests to use real internal components and lightweight fakes at external boundaries.
        4. Run full-flow integration tests to confirm no internal mocks remain.
    - **Done-when:**
        1. No tests mock internal packages; only external interfaces are faked.
        2. Integration tests pass using real orchestration path.
    - **Depends-on:** [T007]

## lint & docs cleanup
- [ ] **T009 · Chore · P3: fix naming, lint & documentation issues**
    - **Context:** Detailed Remedies → cr-09 → Steps 1-3
    - **Action:**
        1. Align function names (`SanitizeFilename` vs `sanitizeFilename`), remove dead imports.
        2. Refactor redundant tests into table-driven subtests.
        3. Replace `fmt.Errorf("static")` with `errors.New` and remove `TODO.md`.
        4. Run `golangci-lint`, `go vet`, and `staticcheck` to confirm zero warnings.
    - **Done-when:**
        1. Codebase is lint-clean with no warnings.
        2. Documentation is updated and `TODO.md` has been removed.
    - **Depends-on:** none
