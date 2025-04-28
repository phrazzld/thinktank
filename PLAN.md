# Remediation Plan – Sprint 1

## Executive Summary
This plan tackles two **blockers** (cr-01, cr-02), four **high-severity** issues (cr-03–cr-06), and three **medium** fixes that are low‐effort but unlock further refactoring (cr-07–cr-09). Blockers clear the way for stable tests and clean structured logging; high-severity work modularizes core logic, audit trails, and context handling; medium fixes secure outputs and harden tests.

## Strike List
| Seq | CR-ID | Title                                          | Effort | Owner    |
|-----|-------|------------------------------------------------|--------|----------|
| 1   | cr-01 | Illegal Internal Mocking in Integration Tests  | l      | backend  |
| 2   | cr-02 | Manual Correlation ID in Log Messages          | xs     | backend  |
| 3   | cr-03 | Refactor Monolithic `Orchestrator.Run`         | m      | backend  |
| 4   | cr-04 | Consolidate Audit-Logging Boilerplate          | s      | backend  |
| 5   | cr-05 | Complete Logger Interface in Tests             | s      | backend  |
| 6   | cr-06 | Add Context Deadlines on API Calls             | s      | backend  |
| 7   | cr-07 | Lock Down Default File Permissions             | xs     | backend  |
| 8   | cr-08 | Replace Magic-String Error Checks with Sentinels | s    | backend  |
| 9   | cr-09 | Use FilesystemIO Abstraction in Tests          | s      | backend  |

## Detailed Remedies

### cr-01 Illegal Internal Mocking in Integration Tests
- **Problem:** Integration tests in `internal/integration/**` mock internal orchestrator, processor, and audit-log types instead of only external boundaries.
- **Impact:** Tests are brittle, tightly coupled to implementation details, break on refactor, violating our Mocking Policy.
- **Chosen Fix:** Refactor all integration tests to remove mocks of internal collaborators and instead mock only true external interfaces (HTTP clients, filesystem I/O, environment providers).
- **Steps:**
  1. Audit each test in `internal/integration` and delete mocks/stubs for `Orchestrator`, `Processor`, `AuditLogger`, etc.
  2. Introduce or reuse boundary interfaces (`FileWriter`, `SynthesisService`, HTTP clients) and mock only those.
  3. Invoke real orchestrator logic in tests, injecting mocked external dependencies.
  4. Remove any `t.Skip` placeholders once tests pass.
- **Done-When:** No integration test uses an internal collaborator mock; CI green; coverage unchanged or improved.

### cr-02 Manual Correlation ID in Log Messages
- **Problem:** `InfoContext`/`ErrorContext` calls manually include `correlation_id=%s` in their message format.
- **Impact:** Correlation IDs are duplicated in JSON logs—confusing, violates structured-logging policy.
- **Chosen Fix:** Strip manual `correlation_id` formatting and rely on `WithContext(ctx)` to auto-inject the field.
- **Steps:**
  1. Search for format strings containing `correlation_id=` in `orchestrator.go` (and other packages).
  2. Remove the `%s` specifier and argument from these calls.
  3. Ensure every entry point uses `logger = logger.WithContext(ctx)`.
  4. Add a linter rule to forbid literal `correlation_id=` in log messages.
- **Done-When:** No manual CID in code; structured logs contain a single `correlation_id` field; linter catch regressions.

### cr-03 Refactor Monolithic `Orchestrator.Run`
- **Problem:** `Orchestrator.Run` (~130 lines) handles context gathering, dry-run, prompt construction, concurrent processing, error aggregation, two output flows.
- **Impact:** Violates Single-Responsibility and Simplicity principles; difficult to read, test, and maintain.
- **Chosen Fix:** Extract well-named private methods for each logical step.
- **Steps:**
  1. Identify core flows: `runDryRunFlow`, `runIndividualOutputFlow`, `runSynthesisFlow`, `aggregateErrors`.
  2. Move corresponding code blocks into private methods on `*Orchestrator`.
  3. Refactor `Run` to call these methods in sequence, passing data and `ctx`.
  4. Write unit tests for each extracted method.
- **Done-When:** `Run` is a <30-line coordinator; each private method has focused tests; CI passes.

### cr-04 Consolidate Audit-Logging Boilerplate
- **Problem:** Multiple manual constructions of `auditlog.AuditEntry` and direct `auditLogger.LogOp` calls across orchestrator and synthesis code.
- **Impact:** DRY violation; inconsistent audit schema; change-heavy maintenance.
- **Chosen Fix:** Create a single helper (`logAuditEvent`) that builds and logs entries.
- **Steps:**
  1. In `internal/thinktank/orchestrator`, add `func (o *Orchestrator) logAuditEvent(ctx, op, status string, inputs, outputs map[string]interface{}, err error)`.
  2. Let helper compute duration, wrap `err` into `auditlog.ErrorInfo`, and call `o.auditLogger.LogOp`.
  3. Replace all inline `AuditEntry` builds and `LogOp` calls with this helper.
- **Done-When:** No direct `AuditEntry` or `LogOp` outside the helper; audit behavior unchanged; tests pass.

### cr-05 Complete Logger Interface in Tests
- **Problem:** Test loggers only implement `Log`; omit `LogOp`, causing silent failures or panics when audit paths exercise `LogOp`.
- **Impact:** Audit logging untested; test suite instability.
- **Chosen Fix:** Provide a shared `MockLogger` in `internal/testutil` that implements both `logutil.LoggerInterface` and `auditlog.AuditLogger`.
- **Steps:**
  1. Create `internal/testutil/mocklogger.go` with `type MockLogger struct { … }`.
  2. Implement `InfoContext`, `ErrorContext`, `WithContext`, and `LogOp`, capturing inputs for assertions.
  3. Update all tests to import and use `testutil.NewMockLogger()`.
- **Done-When:** No per-test logger mocks remain; `MockLogger` compiles and tests verify captured calls; CI passes.

### cr-06 Add Context Deadlines on API Calls
- **Problem:** `client.GenerateContent(ctx, …)` and `processor.Process(ctx, …)` use a `ctx` without deadline.
- **Impact:** Hangs on unresponsive LLM API; goroutine/resource leaks; violates Architecture Guidelines.
- **Chosen Fix:** Introduce `context.WithTimeout` at entry point before calling `Orchestrator.Run`.
- **Steps:**
  1. Add a CLI flag `--timeout` (default e.g. 60s).
  2. In `main.go`, wrap `context.Background()` with `context.WithTimeout` using that flag.
  3. Pass this timed context throughout orchestrator and services.
  4. Add integration tests simulating a hanging service to assert timeout.
- **Done-When:** All external calls respect deadlines; tests trigger context-deadline errors; no hanging goroutines.

### cr-07 Lock Down Default File Permissions
- **Problem:** `os.MkdirAll(..., 0755)` and file writes `0644` create world-readable/executable outputs.
- **Impact:** Exposes sensitive data; violates least-privilege security.
- **Chosen Fix:** Change defaults to `0750` for dirs and `0640` for files.
- **Steps:**
  1. In `output_writer.go` and tests, update `MkdirAll` flags to `0750`.
  2. Update `os.WriteFile` or `FileWriter` to use `0640`.
  3. Add a config option for permissions if customization is needed.
- **Done-When:** Outputs have restrictive modes; tests assert file modes; CI passes.

### cr-08 Replace Magic-String Error Checks with Sentinels
- **Problem:** Tests use `strings.Contains(err.Error(), "...")` on literal messages.
- **Impact:** Fragile tests break on wording changes; violates Testing Strategy.
- **Chosen Fix:** Export sentinel errors (`ErrInvalidSynthesisModel`, etc.) and use `errors.Is` in tests.
- **Steps:**
  1. Define `var ErrInvalidSynthesisModel = errors.New("invalid synthesis model")` and similar in code.
  2. Wrap or return these errors with `%w`.
  3. Refactor tests to use `errors.Is(err, ErrInvalidSynthesisModel)`.
- **Done-When:** No `strings.Contains` in error tests; all use `errors.Is`/`errors.As`; tests pass.

### cr-09 Use FilesystemIO Abstraction in Tests
- **Problem:** Tests call `os.*` and `filepath.Join` directly instead of the `FilesystemIO` interface.
- **Impact:** OS-dependent tests; path bugs; leaks implementation details into tests.
- **Chosen Fix:** Migrate tests to use `FilesystemIO` for setup and assertions.
- **Steps:**
  1. Replace direct `os.MkdirAll`/`os.WriteFile`/`filepath.Join` calls in tests with methods on a `FilesystemIO` stub.
  2. Provide a test helper `NewMemFS()` or similar to simulate file system actions.
  3. Adjust assertions to read via the `FilesystemIO` API.
- **Done-When:** No direct FS calls in tests; all path ops use the abstraction; cross-platform CI green.

## Standards Alignment
- **Mocking Policy:** cr-01 enforces “mock only true external boundaries.”
- **Structured Logging:** cr-02, cr-06 rely on `WithContext` for auto-fields.
- **Simplicity & SRP:** cr-03 reduces cyclomatic complexity via private methods.
- **DRY Principle:** cr-04 centralizes audit logic; cr-05 removes logger duplication.
- **Error Handling:** cr-08 and the planned `%w` audit reinforce `errors.Is` patterns.
- **Architecture Guidelines:** cr-06 context deadlines; cr-09 FS abstraction.
- **Security Considerations:** cr-07 least-privilege defaults for file modes.

## Validation Checklist
- [ ] All unit and integration tests pass; CI green.
- [ ] No internal mocks in `internal/integration/**`; 100% adherence to mocking policy.
- [ ] Logs contain a single `correlation_id` field; linter flags manual inserts.
- [ ] `Orchestrator.Run` <30 lines; each private method covered by unit tests.
- [ ] All audit events flow through `logAuditEvent`; helper fully exercised.
- [ ] `MockLogger` used everywhere; `LogOp` calls captured and asserted.
- [ ] External API calls fail with “context deadline exceeded” when timeout triggers.
- [ ] Files and directories in outputs have `0750/0640` by default; tests verify modes.
- [ ] Tests use `errors.Is`/`errors.As` for sentinel errors; no magic-string assertions.
- [ ] Tests import and use `FilesystemIO` abstraction; no direct `os` or `filepath` usage.
- [ ] No new lint, static-analysis, or security warnings introduced.
