# TODO

## Audit Log Structure and Interface Definition
- [x] **Define Audit Log Data Structures**
  - **Action:** Create the `internal/auditlog` directory. Define the `AuditEntry`, `TokenCountInfo`, and `ErrorInfo` structs in a new file `internal/auditlog/entry.go` as specified in the plan. Ensure correct JSON tags are used.
  - **Depends On:** None
  - **AC Ref:** PLAN.MD Detailed Step 1

- [x] **Define Audit Logger Interface**
  - **Action:** Define the `AuditLogger` interface with `Log(entry AuditEntry) error` and `Close() error` methods in a new file `internal/auditlog/logger.go`.
  - **Depends On:** Define Audit Log Data Structures
  - **AC Ref:** PLAN.MD Detailed Step 2

## Audit Logger Implementations
- [x] **Implement FileAuditLogger**
  - **Action:** Implement the `FileAuditLogger` struct and its methods (`NewFileAuditLogger`, `Log`, `Close`) in `internal/auditlog/logger.go`. Include file opening (append/create), JSON marshaling, writing JSON Lines, mutex locking for concurrent safety, and internal error logging using the provided `logutil.LoggerInterface`.
  - **Depends On:** Define Audit Logger Interface
  - **AC Ref:** PLAN.MD Detailed Step 2

- [x] **Implement NoOpAuditLogger**
  - **Action:** Implement the `NoOpAuditLogger` struct and its methods (`NewNoOpAuditLogger`, `Log`, `Close`) in `internal/auditlog/logger.go`. These methods should perform no actions and return `nil` error.
  - **Depends On:** Define Audit Logger Interface
  - **AC Ref:** PLAN.MD Detailed Step 4

- [ ] **Add Interface Satisfaction Checks**
  - **Action:** Add `var _ AuditLogger = (*FileAuditLogger)(nil)` and `var _ AuditLogger = (*NoOpAuditLogger)(nil)` lines at the end of `internal/auditlog/logger.go` to ensure implementations satisfy the interface contract at compile time.
  - **Depends On:** Implement FileAuditLogger, Implement NoOpAuditLogger
  - **AC Ref:** PLAN.MD Detailed Step 2, PLAN.MD Detailed Step 4

## Configuration and Initialization
- [ ] **Add Audit Log CLI Flag**
  - **Action:** Add a new string flag `--audit-log-file` to `cmd/architect/cli.go` using `flagSet.String`. Include appropriate usage description.
  - **Depends On:** None
  - **AC Ref:** PLAN.MD Detailed Step 3

- [ ] **Update CLI Configuration Struct (cmd)**
  - **Action:** Add the `AuditLogFile string` field to the `CliConfig` struct in `cmd/architect/cli.go`.
  - **Depends On:** Add Audit Log CLI Flag
  - **AC Ref:** PLAN.MD Detailed Step 3

- [ ] **Update Flag Parsing Logic**
  - **Action:** In `ParseFlagsWithEnv` within `cmd/architect/cli.go`, assign the value of the new `--audit-log-file` flag to the `config.AuditLogFile` field.
  - **Depends On:** Update CLI Configuration Struct (cmd)
  - **AC Ref:** PLAN.MD Detailed Step 3

- [ ] **Update Core Configuration Struct (internal)**
  - **Action:** Add the `AuditLogFile string` field to the `CliConfig` struct in `internal/architect/types.go`.
  - **Depends On:** None
  - **AC Ref:** PLAN.MD Detailed Step 4

- [ ] **Update Config Conversion Logic**
  - **Action:** Update the `convertToArchitectConfig` function in `cmd/architect/main.go` to copy the `AuditLogFile` value from the `cmd` config to the `internal/architect` config.
  - **Depends On:** Update CLI Configuration Struct (cmd), Update Core Configuration Struct (internal)
  - **AC Ref:** PLAN.MD Detailed Step 4

- [ ] **Implement Audit Logger Initialization in Main**
  - **Action:** In `cmd/architect/main.go`, import `internal/auditlog`. After initializing the console `logger`, add logic to check `cmdConfig.AuditLogFile`. If set, call `auditlog.NewFileAuditLogger`; otherwise, call `auditlog.NewNoOpAuditLogger`. Use the console `logger` for internal logging within `FileAuditLogger`.
  - **Depends On:** Implement FileAuditLogger, Implement NoOpAuditLogger, Update Config Conversion Logic
  - **AC Ref:** PLAN.MD Detailed Step 5

- [ ] **Implement Error Handling for File Logger Init**
  - **Action:** In `cmd/architect/main.go`, add error handling for the `auditlog.NewFileAuditLogger` call. If an error occurs, log it using the console logger and fall back to using `auditlog.NewNoOpAuditLogger`.
  - **Depends On:** Implement Audit Logger Initialization in Main
  - **AC Ref:** PLAN.MD Detailed Step 5

- [ ] **Ensure Audit Logger Closure**
  - **Action:** Add `defer auditLogger.Close()` in `cmd/architect/main.go` immediately after successful initialization of `auditLogger` to ensure the log file is closed properly on application exit.
  - **Depends On:** Implement Audit Logger Initialization in Main
  - **AC Ref:** PLAN.MD Detailed Step 5

## Dependency Injection
- [ ] **Update Execute Function Signature**
  - **Action:** Modify the `Execute` function signature in `internal/architect/app.go` to accept `auditLogger auditlog.AuditLogger` as a parameter.
  - **Depends On:** Define Audit Logger Interface
  - **AC Ref:** PLAN.MD Detailed Step 6

- [ ] **Pass Audit Logger to Execute**
  - **Action:** Update the call to `architect.Execute` in `cmd/architect/main.go` to pass the initialized `auditLogger` instance.
  - **Depends On:** Implement Audit Logger Initialization in Main, Update Execute Function Signature
  - **AC Ref:** PLAN.MD Detailed Step 5, PLAN.MD Detailed Step 6

## Instrumentation (Core Logic)
- [ ] **Instrument Execute Start**
  - **Action:** In `internal/architect/app.go`'s `Execute` function, add a call to `auditLogger.Log` at the beginning to record `Operation: "ExecuteStart"`, `Status: "InProgress"`, and relevant `Inputs` (CLI flags from `cliConfig`).
  - **Depends On:** Update Execute Function Signature
  - **AC Ref:** PLAN.MD Detailed Step 6, PLAN.MD Step 8

- [ ] **Instrument Instruction Reading**
  - **Action:** In `internal/architect/app.go`'s `Execute` function, add `auditLogger.Log` calls after reading the instructions file to record `Operation: "ReadInstructions"`, `Status: "Success/Failure"`, `Inputs: {path}`, and `Error` if applicable.
  - **Depends On:** Update Execute Function Signature
  - **AC Ref:** PLAN.MD Detailed Step 6, PLAN.MD Step 8

- [ ] **Instrument Context Gathering**
  - **Action:** In `internal/architect/app.go`'s `Execute` function, wrap the call to `contextGatherer.GatherContext`. Log `Operation: "GatherContextStart"`, `Status: "InProgress"`, `Inputs` before the call. After the call, log `Operation: "GatherContextEnd"`, `Status: "Success/Failure"`, calculated `DurationMs`, `Outputs` (file count/stats), and `Error` if applicable.
  - **Depends On:** Update Execute Function Signature
  - **AC Ref:** PLAN.MD Detailed Step 6, PLAN.MD Step 8

- [ ] **Instrument Token Check**
  - **Action:** In `internal/architect/app.go`'s `Execute` function, add `auditLogger.Log` calls after `tokenManager.GetTokenInfo` to record `Operation: "CheckTokens"`, `Status: "Success/Failure"`, `Inputs` (prompt length approximation if feasible, or just indicate check occurred), `Outputs` (the `TokenCountInfo` struct), and `Error` if applicable. Also log the final decision if the limit was exceeded.
  - **Depends On:** Update Execute Function Signature, Define Audit Log Data Structures
  - **AC Ref:** PLAN.MD Detailed Step 6, PLAN.MD Step 8

- [ ] **Instrument Content Generation**
  - **Action:** In `internal/architect/app.go`'s `Execute` function, wrap the call to `geminiClient.GenerateContent`. Log `Operation: "GenerateContentStart"`, `Status: "InProgress"`, `Inputs` (model name) before the call. After the call, log `Operation: "GenerateContentEnd"`, `Status: "Success/Failure"`, calculated `DurationMs`, `Outputs` (finish reason, safety ratings, token counts), and `Error` if applicable.
  - **Depends On:** Update Execute Function Signature, Define Audit Log Data Structures
  - **AC Ref:** PLAN.MD Detailed Step 6, PLAN.MD Step 8

- [ ] **Instrument Save Output**
  - **Action:** In `internal/architect/app.go`'s `Execute` function, wrap the call to `fileWriter.SaveToFile`. Log `Operation: "SaveOutputStart"`, `Status: "InProgress"`, `Inputs` (output path) before the call. After the call, log `Operation: "SaveOutputEnd"`, `Status: "Success/Failure"`, calculated `DurationMs`, and `Error` if applicable.
  - **Depends On:** Update Execute Function Signature
  - **AC Ref:** PLAN.MD Detailed Step 6, PLAN.MD Step 8

- [ ] **Instrument Execute End**
  - **Action:** In `internal/architect/app.go`'s `Execute` function, add a call to `auditLogger.Log` at the very end (e.g., in a defer block or just before returning) to record `Operation: "ExecuteEnd"`, `Status: "Success/Failure"`, and the final `Error` returned by the function.
  - **Depends On:** Update Execute Function Signature
  - **AC Ref:** PLAN.MD Detailed Step 6, PLAN.MD Step 8

## Testing
- [ ] **Create Audit Logger Test File**
  - **Action:** Create a new test file `internal/auditlog/logger_test.go`.
  - **Depends On:** Define Audit Logger Interface
  - **AC Ref:** PLAN.MD Detailed Step 9

- [ ] **Add Unit Tests for FileAuditLogger**
  - **Action:** In `internal/auditlog/logger_test.go`, add unit tests for `FileAuditLogger`. Use `t.TempDir()` to create temporary log files. Test file creation, writing valid JSON Lines format, appending to existing files, correct timestamping, handling concurrent writes (if possible in unit test context), `Close()` method, and error conditions (e.g., invalid path, write errors - potentially using filesystem mocks if needed).
  - **Depends On:** Implement FileAuditLogger, Create Audit Logger Test File
  - **AC Ref:** PLAN.MD Detailed Step 9

- [ ] **Add Unit Tests for NoOpAuditLogger**
  - **Action:** In `internal/auditlog/logger_test.go`, add unit tests for `NoOpAuditLogger` to verify that its `Log` and `Close` methods execute without error and have no side effects.
  - **Depends On:** Implement NoOpAuditLogger, Create Audit Logger Test File
  - **AC Ref:** PLAN.MD Detailed Step 9

- [ ] **Add/Update Integration Tests for Audit Logging**
  - **Action:** Modify existing integration tests (or create new ones) that run the main application flow (`architect.Execute` or via `cmd/architect/main.go`). Run these tests with the `--audit-log-file` flag pointing to a temporary file. After the test run, read the temporary log file and assert that it exists, is not empty, and contains valid JSON Lines entries corresponding to the expected operations performed during the test run.
  - **Depends On:** Implement Audit Logger Initialization in Main, Pass Audit Logger to Execute, All Instrumentation Tasks
  - **AC Ref:** PLAN.MD Detailed Step 9

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Issue/Assumption:** Assumed detailed steps in PLAN.MD serve as implicit Acceptance Criteria (ACs).
  - **Context:** No explicit AC IDs were provided in PLAN.MD. Task AC Refs point to PLAN.MD steps or overall description.

- [ ] **Issue/Assumption:** Assumed the `logutil.LoggerInterface` passed to `NewFileAuditLogger` is the existing console logger.
  - **Context:** PLAN.MD Step 5 implies using the console logger for `FileAuditLogger`'s internal errors.

- [ ] **Issue/Assumption:** Assumed the primary instrumentation points are within `internal/architect.Execute` as listed in PLAN.MD Step 8. Finer-grained logging within `ContextGatherer`, `APIService`, `FileWriter` is deferred.
  - **Context:** PLAN.MD Step 7 mentions potentially passing `AuditLogger` deeper, but Step 8 focuses on `Execute`. Initial implementation focuses on `Execute`.

- [ ] **Issue/Assumption:** Assumed the desired behavior on `NewFileAuditLogger` failure is to log the error and continue with `NoOpAuditLogger`, as shown in the PLAN.MD Step 5 code snippet.
  - **Context:** PLAN.MD Step 5 suggests logging and potentially exiting OR continuing. The code follows the continuation path.

- [ ] **Issue/Assumption:** `ErrorInfo` struct implementation will initially include `Message` and `Type`. The specific values for `Type` (e.g., "ValidationError", "APIError") need further definition/refinement during implementation. `StackTrace` is omitted for now.
  - **Context:** PLAN.MD Step 1 defines `ErrorInfo` with optional `Type` and potential `StackTrace`. Scope needs confirmation.