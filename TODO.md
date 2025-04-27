# Todo

## Integration Tests
- [x] **T001 · Feature · P2: remove internal mocks from integration tests**
    - Context: cr-01 step 1
    - Action:
        1. Remove mocks/stubs for `Orchestrator`, `Processor`, `AuditLogger` in tests under `internal/integration`.
    - Done-when:
        1. No integration test imports or uses internal collaborator mocks.
    - Depends-on: none

- [x] **T002 · Feature · P2: introduce boundary interface mocks in integration tests**
    - Context: cr-01 step 2
    - Action:
        1. Add or reuse `FileWriter`, `SynthesisService`, HTTP client interfaces in integration tests.
        2. Replace deleted internal collaborator mocks with mocks of these boundary interfaces.
    - Done-when:
        1. Integration tests mock only external boundary interfaces.
    - Depends-on: T001

- [x] **T003 · Feature · P2: invoke real orchestrator logic in integration tests**
    - Context: cr-01 step 3
    - Action:
        1. Configure integration tests to call actual `Orchestrator.Run` with mocked external dependencies.
    - Done-when:
        1. Integration tests execute real orchestrator logic without internal collaborator mocks.
    - Depends-on: T002

- [x] **T004 · Chore · P2: remove t.Skip placeholders from integration tests**
    - Context: cr-01 step 4
    - Action:
        1. Delete `t.Skip` calls in `internal/integration` tests.
    - Done-when:
        1. No `t.Skip` placeholders remain and CI integration tests run.
    - Depends-on: T003

## Logging
- [x] **T005 · Refactor · P2: remove manual correlation_id formatting from log calls**
    - Context: cr-02 steps 1–2
    - Action:
        1. Search for log format strings containing `"correlation_id=%s"` in code.
        2. Remove the `"%s"` placeholder and corresponding arguments.
    - Done-when:
        1. No log calls include literal `"correlation_id="` in messages.
    - Depends-on: none

- [x] **T006 · Refactor · P2: enforce logger.WithContext(ctx) at entry points**
    - Context: cr-02 step 3
    - Action:
        1. Ensure `main`, handlers, and CLI entrypoints assign `logger = logger.WithContext(ctx)` before any logging.
    - Done-when:
        1. Every entrypoint uses `WithContext(ctx)` before logging.
    - Depends-on: none

- [x] **T007 · Chore · P2: add linter rule to forbid literal correlation_id= in logs**
    - Context: cr-02 step 4
    - Action:
        1. Extend linter configuration to error on patterns matching `"correlation_id="` in log messages.
    - Done-when:
        1. Linter fails if code contains `"correlation_id="` literal.
    - Depends-on: none

## Orchestrator Refactoring
- [x] **T008 · Refactor · P2: extract runDryRunFlow method from Orchestrator.Run**
    - Context: cr-03 steps 1–2
    - Action:
        1. Identify dry-run flow code and move it into private method `runDryRunFlow(ctx, params)`.
        2. Update `Run` to call `runDryRunFlow`.
    - Done-when:
        1. `Orchestrator.Run` no longer contains dry-run code; `runDryRunFlow` is invoked.
    - Depends-on: none

- [x] **T009 · Refactor · P2: extract runIndividualOutputFlow method from Orchestrator.Run**
    - Context: cr-03 steps 1–2
    - Action:
        1. Move individual output flow code into `runIndividualOutputFlow(ctx, params)`.
        2. Update `Run` to call `runIndividualOutputFlow`.
    - Done-when:
        1. No individual output flow code remains in `Run`; method invocation present.
    - Depends-on: none

- [x] **T010 · Refactor · P2: extract runSynthesisFlow method from Orchestrator.Run**
    - Context: cr-03 steps 1–2
    - Action:
        1. Move synthesis flow code block into `runSynthesisFlow(ctx, params)`.
        2. Update `Run` to call `runSynthesisFlow`.
    - Done-when:
        1. Synthesis code has moved and is invoked via `runSynthesisFlow`.
    - Depends-on: none

- [x] **T011 · Refactor · P2: extract aggregateErrors method from Orchestrator.Run**
    - Context: cr-03 steps 1–2
    - Action:
        1. Extract error aggregation logic into private method `aggregateErrors(errors []error)`.
        2. Replace inline code in `Run` with `aggregateErrors`.
    - Done-when:
        1. `aggregateErrors` exists and is used by `Run`.
    - Depends-on: none

- [x] **T012 · Refactor · P2: simplify Orchestrator.Run to coordinator method**
    - Context: cr-03 step 3
    - Action:
        1. Refactor `Run` to <30 lines, calling `runDryRunFlow`, `runIndividualOutputFlow`, `runSynthesisFlow`, `aggregateErrors` in sequence.
    - Done-when:
        1. `Run` is under 30 lines and compiles.
    - Depends-on: [T008, T009, T010, T011]

- [x] **T013 · Test · P2: add unit tests for runDryRunFlow method**
    - Context: cr-03 step 4
    - Action:
        1. Write table-driven tests for `runDryRunFlow`, covering happy and error paths.
    - Done-when:
        1. Tests for `runDryRunFlow` pass.
    - Depends-on: T008

- [ ] **T014 · Test · P2: add unit tests for runIndividualOutputFlow method**
    - Context: cr-03 step 4
    - Action:
        1. Write tests for typical and edge-case behavior of `runIndividualOutputFlow`.
    - Done-when:
        1. Tests for `runIndividualOutputFlow` pass.
    - Depends-on: T009

- [ ] **T015 · Test · P2: add unit tests for runSynthesisFlow method**
    - Context: cr-03 step 4
    - Action:
        1. Cover concurrency and error scenarios in `runSynthesisFlow` tests.
    - Done-when:
        1. `runSynthesisFlow` tests pass.
    - Depends-on: T010

- [ ] **T016 · Test · P2: add unit tests for aggregateErrors method**
    - Context: cr-03 step 4
    - Action:
        1. Validate that `aggregateErrors` combines multiple errors correctly.
    - Done-when:
        1. `aggregateErrors` tests pass.
    - Depends-on: T011

## Audit Logging
- [ ] **T017 · Feature · P2: implement logAuditEvent helper in Orchestrator**
    - Context: cr-04 step 1
    - Action:
        1. Add method `logAuditEvent(ctx, op, status string, inputs, outputs map[string]interface{}, err error)` on `*Orchestrator`.
    - Done-when:
        1. Method compiles and constructs `auditlog.AuditEntry` and calls `o.auditLogger.LogOp`.
    - Depends-on: none

- [ ] **T018 · Refactor · P2: replace inline AuditEntry and LogOp calls with logAuditEvent**
    - Context: cr-04 step 3
    - Action:
        1. Locate all direct `auditlog.AuditEntry` constructions and `LogOp` calls.
        2. Replace them with calls to `o.logAuditEvent`.
    - Done-when:
        1. No direct `AuditEntry` or `LogOp` outside `logAuditEvent`.
    - Depends-on: T017

- [ ] **T019 · Test · P2: add unit tests for logAuditEvent helper**
    - Context: cr-04 step 2
    - Action:
        1. Write tests verifying duration calculation, `ErrorInfo` wrapping, and `LogOp` invocation.
    - Done-when:
        1. Tests for `logAuditEvent` pass.
    - Depends-on: T017

## Test Utilities
- [ ] **T020 · Chore · P2: create MockLogger in internal/testutil**
    - Context: cr-05 step 1
    - Action:
        1. Add `internal/testutil/mocklogger.go` with type `MockLogger`.
    - Done-when:
        1. `MockLogger` type exists and compiles.
    - Depends-on: none

- [ ] **T021 · Feature · P2: implement InfoContext, ErrorContext, WithContext, and LogOp on MockLogger**
    - Context: cr-05 step 2
    - Action:
        1. Implement methods capturing parameters for test assertions.
    - Done-when:
        1. `MockLogger` satisfies `logutil.LoggerInterface` and `auditlog.AuditLogger`.
    - Depends-on: T020

- [ ] **T022 · Refactor · P2: update tests to use MockLogger**
    - Context: cr-05 step 3
    - Action:
        1. Replace per-test logger mocks with `testutil.NewMockLogger()` in all tests.
    - Done-when:
        1. No custom logger mocks in tests; `MockLogger` is used everywhere.
    - Depends-on: T021

- [ ] **T023 · Test · P2: verify MockLogger captures log and audit calls**
    - Context: cr-05 step 2
    - Action:
        1. Add tests asserting `MockLogger` recorded `InfoContext`, `ErrorContext`, and `LogOp` invocations.
    - Done-when:
        1. Added tests pass and confirm captured calls.
    - Depends-on: T021

## Context Deadlines
- [ ] **T024 · Feature · P2: add timeout CLI flag to main.go**
    - Context: cr-06 step 1
    - Action:
        1. Introduce `--timeout` flag with default `60s` in CLI setup.
    - Done-when:
        1. Flag recognized and parsed in `main`.
    - Depends-on: none

- [ ] **T025 · Feature · P2: wrap root context with WithTimeout and propagate**
    - Context: cr-06 steps 2–3
    - Action:
        1. Replace `context.Background()` with `context.WithTimeout` using the flag value.
        2. Pass timeout context to `Orchestrator.Run` and downstream services.
    - Done-when:
        1. External API calls receive a context with deadline.
    - Depends-on: T024

- [ ] **T026 · Test · P2: add integration test for context deadline enforcement**
    - Context: cr-06 step 4
    - Action:
        1. Simulate a hanging `SynthesisService` or `client.GenerateContent`.
        2. Assert operation fails with “context deadline exceeded” without hanging.
    - Done-when:
        1. Integration test triggers and validates deadline error.
    - Depends-on: T025

## File Permissions
- [ ] **T027 · Refactor · P2: change default directory permissions to 0750**
    - Context: cr-07 step 1
    - Action:
        1. Update `os.MkdirAll` calls in `output_writer.go` and tests to use `0750`.
    - Done-when:
        1. Directories created have mode `0750`; affected tests updated.
    - Depends-on: none

- [ ] **T028 · Refactor · P2: change default file write permissions to 0640**
    - Context: cr-07 step 2
    - Action:
        1. Update `os.WriteFile` or `FileWriter` default mode to `0640` and update tests.
    - Done-when:
        1. Files written have mode `0640`; tests updated.
    - Depends-on: none

- [ ] **T029 · Feature · P2: add config option for file and directory permissions**
    - Context: cr-07 step 3
    - Action:
        1. Expose permissions as configurable parameters in the configuration struct.
        2. Wire config into `output_writer`.
    - Done-when:
        1. Config option exists and controls permission flags.
    - Depends-on: [T027, T028]

## Error Sentinels
- [ ] **T030 · Chore · P2: define sentinel errors for synthesis and processing failures**
    - Context: cr-08 step 1
    - Action:
        1. Add exported variables `ErrInvalidSynthesisModel`, etc., using `errors.New` in relevant packages.
    - Done-when:
        1. Sentinel error variables compile and are documented.
    - Depends-on: none

- [ ] **T031 · Refactor · P2: wrap error cases with %w for sentinel propagation**
    - Context: cr-08 step 2
    - Action:
        1. Update return statements to use `fmt.Errorf("%w")` wrapping sentinel errors.
    - Done-when:
        1. Code wraps errors and `errors.Is` can match sentinel values.
    - Depends-on: T030

- [ ] **T032 · Refactor · P2: replace magic-string error assertions with errors.Is in tests**
    - Context: cr-08 step 3
    - Action:
        1. Update tests to use `errors.Is`/`errors.As` instead of `strings.Contains(err.Error(), …)`.
    - Done-when:
        1. No `strings.Contains(err.Error())` in tests; `errors.Is` used.
    - Depends-on: [T030, T031]

## Filesystem Abstraction
- [ ] **T033 · Chore · P2: create in-memory FilesystemIO stub**
    - Context: cr-09 step 2
    - Action:
        1. Implement `testutil.NewMemFS()` returning a stub implementing `FilesystemIO`.
    - Done-when:
        1. `NewMemFS` compiles and provides read/write methods.
    - Depends-on: none

- [ ] **T034 · Refactor · P2: replace direct os/filepath calls with FilesystemIO in tests**
    - Context: cr-09 step 1
    - Action:
        1. Update tests to use `FilesystemIO` methods instead of `os.MkdirAll`, `os.WriteFile`, `filepath.Join`.
    - Done-when:
        1. No direct `os` or `filepath` calls in tests; `FilesystemIO` used.
    - Depends-on: T033

- [ ] **T035 · Refactor · P2: update test assertions to use FilesystemIO reading API**
    - Context: cr-09 step 3
    - Action:
        1. Change test reads and existence checks to use `FilesystemIO` methods.
    - Done-when:
        1. Tests read and verify files via `FilesystemIO` and pass on all platforms.
    - Depends-on: T033
