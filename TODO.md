# TODO

## [Issue 1: Redundant Validation Logic]
- [x] **Task Title:** Remove Redundant `validateInputs` Function from `internal/architect/app.go`
  - **Action:** Delete the `validateInputs` function located at `internal/architect/app.go:950-969`. Verify that all necessary input validation is comprehensively handled by `cmd/architect/cli.go:ValidateInputs` before `architect.Execute` is called.
  - **Depends On:** None
  - **AC Ref:** Issue 1, `CORE_PRINCIPLES.md` (Simplicity), `ARCHITECTURE_GUIDELINES.md` (Separation of Concerns)

## [Issue 2: Output Directory Standardization]
- [x] **Task Title:** Remove Backward Compatibility for --output Flag
  - **Action:** Decision made to eliminate backward compatibility for the --output flag. Will use --output-dir exclusively for simplicity and clarity.
  - **Depends On:** None
  - **AC Ref:** Issue 2, `CORE_PRINCIPLES.md` (Simplicity)
- [x] **Task Title:** Update Integration Tests for Model-Specific Outputs
  - **Action:** Modify `internal/integration/integration_test.go` and `internal/integration/xml_integration_test.go` to use the new --output-dir flag and assert against the model-specific output files (e.g., `filepath.Join(outputDir, "test-model.md")`).
  - **Depends On:** None
  - **AC Ref:** Issue 2
- [x] **Task Title:** Remove --output Flag and Logic
  - **Action:** Remove the --output flag and all related logic from the codebase, ensuring that --output-dir is the only option for specifying output location.
  - **Depends On:** Update Integration Tests for Model-Specific Outputs
  - **AC Ref:** Issue 2, `CORE_PRINCIPLES.md` (Simplicity)
- [x] **Task Title:** Update `savePlanToFile` Function
  - **Action:** Simplify the `savePlanToFile` function in `internal/architect/app.go` to only write to model-specific files in the output directory. Remove any legacy output path handling.
  - **Depends On:** Remove --output Flag and Logic
  - **AC Ref:** Issue 2, `CORE_PRINCIPLES.md` (Simplicity)
- [x] **Task Title:** Update Documentation
  - **Action:** Update README.md and any other documentation to remove all references to the --output flag, ensuring only --output-dir is mentioned.
  - **Depends On:** Remove --output Flag and Logic
  - **AC Ref:** Issue 2, `DOCUMENTATION_APPROACH.md` (Clarity and Consistency)

## [Issue 3: Performance Optimization for Multi-Model Requests]
- [x] **Task Title:** Implement Concurrent Model Processing
  - **Action:** Modify the application to process requests for multiple models concurrently rather than sequentially. Use Go's concurrency primitives (goroutines and channels) to implement this feature while maintaining proper error handling and logging.
  - **Depends On:** None
  - **AC Ref:** Issue 3, `CORE_PRINCIPLES.md` (Modularity), `ARCHITECTURE_GUIDELINES.md` (Embrace the Unix Philosophy)
- [ ] **Task Title:** Add Concurrency Control for API Rate Limits
  - **Action:** Implement a mechanism to control concurrency levels based on API rate limits. This should include configurable settings to prevent overwhelming the Gemini API.
  - **Depends On:** Implement Concurrent Model Processing
  - **AC Ref:** Issue 3
- [ ] **Task Title:** Update Integration Tests for Concurrent Processing
  - **Action:** Extend existing integration tests to verify that multiple model requests are processed concurrently and results are correctly saved to their respective output files.
  - **Depends On:** Implement Concurrent Model Processing
  - **AC Ref:** Issue 3, `TESTING_STRATEGY.md` (Integration Testing Approach)
- [ ] **Task Title:** Update Documentation for Concurrent Processing
  - **Action:** Update README.md to explain the concurrent processing of multiple models, including any new configuration options for controlling concurrency.
  - **Depends On:** Implement Concurrent Model Processing
  - **AC Ref:** Issue 3, `DOCUMENTATION_APPROACH.md` (README.md: The Essential Entry Point)

## [Issue 4: Improve Logging for Broader Use Cases]
- [ ] **Task Title:** Update Logging Terminology
  - **Action:** Refactor logging messages throughout the codebase to replace specific "plan" terminology with more general terms reflective of the tool's broader use cases (e.g., "output", "analysis", "result").
  - **Depends On:** None
  - **AC Ref:** Issue 4, `CODING_STANDARDS.md` (Meaningful Naming: Communicate Purpose)
- [ ] **Task Title:** Enhance Logging Verbosity and Clarity
  - **Action:** Improve log messages to be more informative about the current operation, including more context about files being processed, models being used, and operation progress. Add additional log points at appropriate places in the execution flow.
  - **Depends On:** None
  - **AC Ref:** Issue 4, `DOCUMENTATION_APPROACH.md` (Explicit is Better than Implicit)
- [ ] **Task Title:** Standardize Log Level Usage
  - **Action:** Review and standardize the use of different log levels (debug, info, warn, error) throughout the application to ensure consistency and appropriate verbosity at each level.
  - **Depends On:** None
  - **AC Ref:** Issue 4

## [Issue 5: Minor Documentation Inconsistencies]
- [ ] **Task Title:** Standardize "architect" Casing in README.md
  - **Action:** Edit `README.md`. Change the main title from `# Code Review: ...` (or similar) to `# architect`. Ensure all other references to the tool name within the body text, examples, and configuration sections consistently use the lowercase "architect".
  - **Depends On:** None
  - **AC Ref:** Issue 5, `DOCUMENTATION_APPROACH.md` (Clarity and Consistency)
- [ ] **Task Title:** Add Missing Newline Before License Link in README.md
  - **Action:** Edit `README.md`. Locate the license link at the very end (e.g., `[MIT License](LICENSE)`) and ensure there is a blank line immediately preceding it.
  - **Depends On:** None
  - **AC Ref:** Issue 5, `DOCUMENTATION_APPROACH.md` (Clarity and Consistency)

