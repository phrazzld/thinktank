# TODO

## Task Group 1: Eliminate Duplication
- [x] **Task Title:** Remove Duplicate `cmd/architect/api.go` File
  - **Action:** Delete the file `cmd/architect/api.go`. Verify that `cmd/architect/cli.go` and any other potential callers correctly use `internal/architect.APIService`. Ensure the build passes after removal.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 1.1

- [x] **Task Title:** Create `internal/config` Package and Define Canonical `CliConfig`
  - **Action:** Create a new package `internal/config`. Move the `CliConfig` struct definition from `cmd/architect/cli.go` to `internal/config/config.go`. Define shared constants (like defaults, env vars) in this package. Use a flat structure initially with a single config object.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 1.2

- [x] **Task Title:** Update Code to Use `internal/config.CliConfig`
  - **Action:** Update `cmd/architect/cli.go` to use the `CliConfig` struct from `internal/config`. Update `internal/architect/app.go` and other internal packages to accept configuration values directly from the config package rather than depending on the CLI-specific struct.
  - **Depends On:** Create `internal/config` Package and Define Canonical `CliConfig`
  - **AC Ref:** Refactoring Goals, Task Group 1.2

- [x] **Task Title:** Remove `internal/architect/types.go`
  - **Action:** Delete the file `internal/architect/types.go` after its contents have been successfully consolidated into `internal/config/config.go`. Ensure no compilation errors arise from the removal.
  - **Depends On:** Update Code to Use `internal/config.CliConfig`
  - **AC Ref:** Refactoring Goals, Task Group 1.2

- [x] **Task Title:** Remove `RunInternal` Function from `internal/architect/app.go`
  - **Action:** Delete the `RunInternal` function. Identify any tests using `RunInternal` and refactor them to use the primary `Execute` function, potentially by injecting mocked dependencies via a new orchestrator struct's constructor or options pattern.
  - **Depends On:** Create `Orchestrator` Struct and Define `Orchestrator.Run` Method
  - **AC Ref:** Refactoring Goals, Task Group 1.3

## Task Group 2: Restructure `internal/architect/app.go`
- [x] **Task Title:** Create `internal/architect/orchestrator` Package and Define `Orchestrator` Struct
  - **Action:** Create a new package `internal/architect/orchestrator`. Define the `Orchestrator` struct within `orchestrator.go`, including fields for dependencies (APIService, ContextGatherer, TokenManager, FileWriter, AuditLogger, RateLimiter, config, logger). Define a `NewOrchestrator` constructor function.
  - **Depends On:** Create `internal/config` Package and Define Canonical `CliConfig`
  - **AC Ref:** Refactoring Goals, Task Group 2.1

- [x] **Task Title:** Create `internal/architect/prompt` Package
  - **Action:** Create a new package `internal/architect/prompt` and corresponding files.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 2.2

- [x] **Task Title:** Move Prompt Logic to `internal/architect/prompt`
  - **Action:** Move the `StitchPrompt` and `EscapeContent` functions from `internal/architect/output.go` to the new `internal/architect/prompt/prompt.go` file. Update callers (currently `internal/architect/app.go`) to use the functions from the new package.
  - **Depends On:** Create `internal/architect/prompt` Package
  - **AC Ref:** Refactoring Goals, Task Group 2.2

- [x] **Task Title:** Create `internal/architect/modelproc` Package and Define `ModelProcessor` Struct
  - **Action:** Create a new package `internal/architect/modelproc`. Define the `ModelProcessor` struct within `processor.go`, including fields for dependencies (APIService, TokenManager, FileWriter, AuditLogger, logger, and relevant config). Define a `NewProcessor` constructor function.
  - **Depends On:** Create `internal/config` Package and Define Canonical `CliConfig`
  - **AC Ref:** Refactoring Goals, Task Group 2.3

- [x] **Task Title:** Extract Model Processing Logic into `ModelProcessor.Process`
  - **Action:** Create a `Process(ctx context.Context, modelName string, stitchedPrompt string) error` method on the `ModelProcessor` struct. Move the logic currently inside the `processModel` / `processModelConcurrently` functions (client init per model, token check, generation, response processing, saving) into this method. Ensure it uses its injected dependencies. Remove the old functions from `app.go`.
  - **Depends On:** Create `internal/architect/modelproc` Package and Define `ModelProcessor` Struct, Move Prompt Logic to `internal/architect/prompt`
  - **AC Ref:** Refactoring Goals, Task Group 2.3

- [x] **Task Title:** Extract Configuration Handling Logic from `app.go`
  - **Action:** Move the logic related to setting up the output directory (checking if empty, generating run name, creating directory) from the beginning of `Execute` into a dedicated setup phase or helper function within `app.go`. This setup should occur before the `Orchestrator` is initialized.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 2.5

- [x] **Task Title:** Simplify `internal/architect/app.go#Execute` Function
  - **Action:** Refactor the `Execute` function in `app.go` to primarily handle: 1) Initial setup (output dir, logging, audit logging), 2) Reading instructions, 3) Initializing dependencies (APIService, TokenManager, ContextGatherer, FileWriter, RateLimiter, AuditLogger, ModelProcessor, Orchestrator), 4) Calling `Orchestrator.Run`, and 5) Handling the final error and top-level audit logging (ExecuteEnd).
  - **Depends On:** Create `internal/architect/orchestrator` Package and Define `Orchestrator` Struct, Create `internal/architect/modelproc` Package and Define `ModelProcessor` Struct, Extract Configuration Handling Logic from `app.go`
  - **AC Ref:** Refactoring Goals, Task Group 2.1

- [x] **Task Title:** Implement `Orchestrator.Run` Method
  - **Action:** Create a `Run(ctx context.Context) error` method on the `Orchestrator` struct. Move the main execution flow logic (context gathering, dry run handling, prompt stitching, concurrent model processing loop using rate limiter and calling `ModelProcessor.Process`, error aggregation, final audit log) from the old `Execute` function into this method.
  - **Depends On:** Create `internal/architect/orchestrator` Package and Define `Orchestrator` Struct, Extract Model Processing Logic into `ModelProcessor.Process`, Simplify `internal/architect/app.go#Execute` Function
  - **AC Ref:** Refactoring Goals, Task Group 2.1, Task Group 2.4

## Task Group 3: Improve Separation of Concerns & Dependencies
- [ ] **Task Title:** Decouple Audit Logging from Orchestration Flow
  - **Action:** Remove detailed audit logging calls (e.g., `GatherContextStart/End`, `CheckTokensStart/End`, `GenerateContentStart/End`, `SaveOutputStart/End`) from the `Orchestrator.Run` method. Pass the `AuditLogger` instance to the relevant components (`ContextGatherer`, `ModelProcessor`, `FileWriter`, `TokenManager`) via their constructors. Modify these components to perform their own specific audit logging for their primary operations.
  - **Depends On:** Implement `Orchestrator.Run` Method, Extract Model Processing Logic into `ModelProcessor.Process`
  - **AC Ref:** Refactoring Goals, Task Group 3.1

- [x] **Task Title:** Remove `CalculateStatisticsWithTokenCounting` from `fileutil`
  - **Action:** Delete the `CalculateStatisticsWithTokenCounting` function from `internal/fileutil/fileutil.go`. Update `internal/fileutil/fileutil_test.go` accordingly. Ensure `fileutil` no longer imports `internal/gemini`.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 3.2

- [ ] **Task Title:** Implement Token Counting within `ContextGatherer`
  - **Action:** Modify `ContextGatherer.GatherContext` to calculate the token count for the gathered context *after* collecting all file content. It should use the `gemini.Client` (passed via dependency injection) to perform the count. Update the `ContextStats` struct and return values accordingly.
  - **Depends On:** Remove `CalculateStatisticsWithTokenCounting` from `fileutil`, Ensure `gemini.Client` is Injected into `ContextGatherer`
  - **AC Ref:** Refactoring Goals, Task Group 3.2, Task Group 3.3

- [x] **Task Title:** Refactor ContextGatherer Implementation to Eliminate Duplication
  - **Action:** Remove the duplicated `ContextGatherer` implementation in `cmd/architect/context.go` and update any necessary imports to use the implementation from `internal/architect/context.go`. Ensure all required functionality is maintained.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 3.3, Core Principles (Simplicity, Modularity)

- [x] **Task Title:** Clean Up ContextGatherer Interface by Removing Redundant Parameters
  - **Action:** Remove the redundant `gemini.Client` parameter from the `GatherContext` and `DisplayDryRunInfo` method signatures in the `interfaces.ContextGatherer` interface and its implementations. Update all callers to use the clean interface.
  - **Depends On:** Refactor ContextGatherer Implementation to Eliminate Duplication
  - **AC Ref:** Refactoring Goals, Task Group 3.3, Core Principles (Simplicity, Explicit over Implicit)

- [x] **Task Title:** Update ContextGatherer Tests for New Interface
  - **Action:** Update the test files to work with the new ContextGatherer interface, providing mock clients via constructor injection. Ensure the tests verify the behavior correctly.
  - **Depends On:** Clean Up ContextGatherer Interface by Removing Redundant Parameters
  - **AC Ref:** Refactoring Goals, Task Group 3.3, Testing Strategy

- [x] **Task Title:** Ensure `gemini.Client` is Injected into `ContextGatherer` (Original)
  - **Action:** Modify `NewContextGatherer` to accept a `gemini.Client` interface as a dependency. Update the initialization point (in `app.go#Execute`) to pass the client. Ensure `GatherContext` uses this injected client for token counting.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 3.3

- [ ] **Task Title:** Ensure `gemini.Client` is Correctly Handled in `TokenManager`
  - **Action:** Review `token.go` to confirm the best approach for `gemini.Client` dependency injection. Modify `NewTokenManager` or its methods if needed to ensure proper dependency management for operations like `GetModelInfo` and token counting.
  - **Depends On:** None
  - **AC Ref:** Refactoring Goals, Task Group 3.3

- [ ] **Task Title:** Refine and Rename `internal/architect/output.go`
  - **Action:** After moving prompt logic, rename `internal/architect/output.go` to `internal/architect/filewriter.go`. Ensure it only contains the `FileWriter` interface and its implementation (`fileWriter`, `NewFileWriter`, `SaveToFile`). Update import paths if necessary.
  - **Depends On:** Move Prompt Logic to `internal/architect/prompt`
  - **AC Ref:** Refactoring Goals, Task Group 3.4

## Task Group 4: Improve Readability and Maintainability
- [ ] **Task Title:** Review and Improve Naming Conventions
  - **Action:** Review variable, function, struct, and package names in all newly created and modified files (`orchestrator`, `modelproc`, `prompt`, `config`, `filewriter`, refactored `app`, `context`, `token`, `api`). Ensure names are clear, consistent, and adhere to Go conventions (PascalCase for exported, camelCase for unexported, short package names).
  - **Depends On:** Refine and Rename `internal/architect/output.go`, Implement `Orchestrator.Run` Method, Decouple Audit Logging from Orchestration Flow
  - **AC Ref:** Refactoring Goals, Task Group 4.1

- [ ] **Task Title:** Simplify Control Flow in `Orchestrator.Run`
  - **Action:** Review the implemented `Orchestrator.Run` method. Ensure it presents a clear, high-level view of the application workflow (Setup, Gather Context, Build Prompt, Process Models, Aggregate Results), effectively delegating implementation details to the injected components. Refactor for clarity if needed.
  - **Depends On:** Implement `Orchestrator.Run` Method
  - **AC Ref:** Refactoring Goals, Task Group 4.2

- [ ] **Task Title:** Add Package and Function Documentation
  - **Action:** Add package comments (`// package ...`) explaining the purpose of the new packages (`orchestrator`, `modelproc`, `prompt`, `config`, `filewriter`). Update function/method comments (Go doc comments `// ...`) for clarity, focusing on the "why" and the contracts, especially for the new public interfaces and methods.
  - **Depends On:** Review and Improve Naming Conventions, Simplify Control Flow in `Orchestrator.Run`
  - **AC Ref:** Refactoring Goals, Task Group 4.3

## Testing Strategy Implementation
- [ ] **Task Title:** Establish Testing Baseline
  - **Action:** Ensure all existing automated tests are passing before starting refactoring. If minimal tests exist, write basic integration tests for the current `Execute` function covering key scenarios (happy path, dry run, simple error case) to establish a baseline for comparison.
  - **Depends On:** None
  - **AC Ref:** Testing Strategy Section 5.1

- [ ] **Task Title:** Implement Unit Tests for New/Refactored Logic
  - **Action:** Write unit tests for newly extracted pure functions (e.g., `prompt.StitchPrompt`, `prompt.EscapeContent`) and any complex, isolatable logic within components. Aim to significantly increase overall test coverage from current levels.
  - **Depends On:** Move Prompt Logic to `internal/architect/prompt` (and other relevant refactoring tasks)
  - **AC Ref:** Testing Strategy Section 5.2

- [ ] **Task Title:** Implement Integration Tests for Orchestrator
  - **Action:** Write integration tests for `Orchestrator.Run`. Mock dependencies (`ContextGatherer`, `ModelProcessor`, `FileWriter`, etc.) to verify the orchestration logic calls collaborators correctly based on inputs and configuration (e.g., dry run behavior, multiple model processing).
  - **Depends On:** Implement `Orchestrator.Run` Method
  - **AC Ref:** Testing Strategy Section 5.3

- [ ] **Task Title:** Implement Integration Tests for Component Interactions
  - **Action:** Write integration tests verifying interactions between key components: 1) `ContextGatherer` with `TokenManager`/`gemini.Client` mock, 2) `ModelProcessor` with `APIService`/`TokenManager`/`FileWriter` mocks. Verify API calls, token checks, and file writes happen as expected.
  - **Depends On:** Implement Token Counting within `ContextGatherer`, Extract Model Processing Logic into `ModelProcessor.Process`
  - **AC Ref:** Testing Strategy Section 5.3

- [ ] **Task Title:** Implement End-to-End (CLI) Tests
  - **Action:** Write tests that execute the compiled binary with various command-line arguments and fixtures. Verify exit codes, output files (may require API mocking), audit logs, and behavior with key flags (`--dry-run`, filters, multiple models, error conditions).
  - **Depends On:** All major refactoring tasks completed
  - **AC Ref:** Testing Strategy Section 5.3

- [ ] **Task Title:** Perform Manual Verification and Comparison
  - **Action:** Manually run the refactored CLI with representative inputs used before refactoring. Compare generated output files, dry-run behavior, console output, and audit logs against the pre-refactoring version to ensure functional equivalence. Test edge cases.
  - **Depends On:** All major refactoring tasks completed
  - **AC Ref:** Testing Strategy Section 5.4, Refactoring Goals (Preserve Functionality)

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Issue/Assumption:** The audit log must maintain exact structure, content, and critical timing of entries, even as logging is decoupled from the main flow. We will ensure audit logs remain strict and exhaustive while maintaining all existing details.
  - **Context:** PLAN.md Section 4 (Risk 5: Audit Log Changes), Section 6 (Open Question 2: Audit Log Strictness).

- [ ] **Issue/Assumption:** The `internal/config` package will use a flat structure initially with a single config object containing the `CliConfig` struct and related constants/defaults. Further organization can be done later if needed.
  - **Context:** PLAN.md Section 6 (Open Question 3: Config Package Structure).

- [ ] **Issue/Assumption:** We are significantly increasing test coverage as part of this refactoring to ensure functionality is preserved and regressions are caught early.
  - **Context:** PLAN.md Section 5 (Testing Strategy), Task Group "Testing Strategy Implementation".

- [ ] **Issue/Assumption:** When injecting the `AuditLogger` into components, each component will maintain identical logging semantics as the original implementation (same event names, parameters, timing) to ensure audit trail equivalence.
  - **Context:** PLAN.md Section 3.1 (Decouple Audit Logging), Section 4 (Risk 5: Audit Log Changes).