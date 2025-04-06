# TODO

## Refactor runThinktank Workflow

- [x] **Analyze Current Flow Structure**: Map the existing control flow and error handling approaches in runThinktank.
  - **Action:** Identify each distinct operation phase, data flow between stages, and existing error handling patterns. Document key function dependencies.
  - **Depends On:** None.
  - **AC Ref:** T6.1

- [x] **Design Helper Function Interfaces**: Define the inputs, outputs, and contracts for each helper function.
  - **Action:** For each identified phase (`_setupWorkflow`, `_processInput`, `_selectModels`, `_executeQueries`, `_processOutput`, `_logCompletionSummary`, `_handleWorkflowError`), create TypeScript interfaces defining required parameters and return types. Define proper error handling contracts.
  - **Depends On:** Analyze Current Flow Structure
  - **AC Ref:** T6.2

- [x] **Implement Setup Workflow Helper**: Create the `_setupWorkflow` helper function.
  - **Action:** Implement function to handle configuration loading, run name generation, and output directory creation with proper error handling and spinner updates.
  - **Depends On:** Design Helper Function Interfaces
  - **AC Ref:** T6.2, T6.4, T6.5

- [x] **Implement Input Processing Helper**: Create the `_processInput` helper function.
  - **Action:** Implement function that handles input processing with appropriate spinner text updates and error wrapping using `FileSystemError`.
  - **Depends On:** Design Helper Function Interfaces
  - **AC Ref:** T6.2, T6.4, T6.5

- [x] **Implement Model Selection Helper**: Create the `_selectModels` helper function.
  - **Action:** Implement function that handles model selection with warnings display, error handling, and appropriate spinner updates. Use `ConfigError` for wrapping errors.
  - **Depends On:** Design Helper Function Interfaces
  - **AC Ref:** T6.2, T6.4, T6.5

- [x] **Implement Query Execution Helper**: Create the `_executeQueries` helper function.
  - **Action:** Implement function that handles query execution with spinner updates and proper error propagation using `ApiError` when appropriate.
  - **Depends On:** Design Helper Function Interfaces
  - **AC Ref:** T6.2, T6.4, T6.5

- [x] **Implement Output Processing Helper**: Create the `_processOutput` helper function.
  - **Action:** Implement function that handles file writing and console output formatting with spinner updates. Catch and wrap errors using `FileSystemError` when needed.
  - **Depends On:** Design Helper Function Interfaces
  - **AC Ref:** T6.2, T6.4, T6.5

- [x] **Implement Completion Summary Helper**: Create the `_logCompletionSummary` helper function.
  - **Action:** Implement function that formats and logs the completion summary, handling both success and partial failure scenarios.
  - **Depends On:** Design Helper Function Interfaces
  - **AC Ref:** T6.2, T6.7

- [x] **Implement Error Handling Helper**: Create the `_handleWorkflowError` helper function.
  - **Action:** Implement function that categorizes unknown errors, ensures proper ThinktankError types, logs contextual information, and rethrows for upstream handling.
  - **Depends On:** Design Helper Function Interfaces
  - **AC Ref:** T6.2, T6.4, T6.7

- [x] **Refactor Main runThinktank Function**: Restructure the main function to use the new helpers.
  - **Action:** Replace the current implementation with a simpler orchestration function that calls the helper functions in sequence and handles top-level error cases.
  - **Depends On:** All helper implementation tasks
  - **AC Ref:** T6.3

- [x] **Audit Spinner Lifecycle**: Ensure spinner is properly managed throughout workflow.
  - **Action:** Verify that spinner is started, updated, and properly terminated (succeed/fail/warn) at appropriate points in all workflow phases, including error paths.
  - **Depends On:** Refactor Main runThinktank Function
  - **AC Ref:** T6.5

- [ ] **Audit Resource Cleanup**: Review potential hanging issues.
  - **Action:** Check for any unhandled promises, missing async/await patterns, or potential connection leaks in provider SDKs that could cause the program to hang.
  - **Depends On:** Refactor Main runThinktank Function
  - **AC Ref:** T6.6

- [ ] **Add Unit Tests for Helper Functions**: Create tests for each helper function.
  - **Action:** Create new test files or update existing ones to test each helper function in isolation with mocked dependencies.
  - **Depends On:** All helper implementation tasks
  - **AC Ref:** T6.8

- [ ] **Update Integration Tests**: Refactor existing tests to work with new structure.
  - **Action:** Update `runThinktank.test.ts` to mock helper functions and verify orchestration logic. Test error propagation through the entire workflow.
  - **Depends On:** Add Unit Tests for Helper Functions
  - **AC Ref:** T6.8

- [ ] **Update Error Handling Tests**: Ensure tests cover all error scenarios.
  - **Action:** Update `runThinktank-error-handling.test.ts` to verify each helper properly catches, wraps, and propagates errors. Test the main error handling catch block.
  - **Depends On:** Update Integration Tests
  - **AC Ref:** T6.8

- [ ] **Run End-to-End Tests**: Verify behavior of refactored workflow.
  - **Action:** Run E2E tests to confirm the refactored runThinktank behaves exactly as before from external perspective.
  - **Depends On:** Update Error Handling Tests
  - **AC Ref:** T6.8
