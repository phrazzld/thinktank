# TODO

## CLI Command Modifications

- [x] **Update `commander` definition in run command**
  - **Action:** Modify `src/cli/commands/run.ts` to add `[contextPaths...]` as a variadic argument after `<promptFile>`, including appropriate help text.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1 - CLI must support variadic path arguments after the prompt file.

- [x] **Update action handler to receive contextPaths**
  - **Action:** Update the action handler function in `src/cli/commands/run.ts` to receive and process the contextPaths parameter correctly.
  - **Depends On:** Update `commander` definition in run command.
  - **AC Ref:** AC 1.2 - CLI must pass context paths to the workflow function.

- [x] **Update RunOptions interface**
  - **Action:** Update `RunOptions` interface in `src/workflow/runThinktank.ts` to include the `contextPaths?: string[]` property.
  - **Depends On:** Update action handler to receive contextPaths.
  - **AC Ref:** AC 1.2 - CLI must pass context paths to the workflow function.

## Context Reading Utility

- [x] **Add dependency for gitignore parsing**
  - **Action:** Research and add a dependency for parsing .gitignore files (e.g., 'ignore', 'globby', or similar library).
  - **Depends On:** None.
  - **AC Ref:** AC 2.3 - System must ignore files and directories according to .gitignore rules.

- [x] **Create context file reader utility function**
  - **Action:** Create a new function to read content from a single file, handling errors appropriately and returning path and content.
  - **Depends On:** None.
  - **AC Ref:** AC 2.1 - System must read content from individual files.

- [x] **Create directory reader utility function**
  - **Action:** Implement a recursive directory reading function that traverses directories and collects file contents.
  - **Depends On:** Create context file reader utility function.
  - **AC Ref:** AC 2.2 - System must recursively read directory contents.

- [x] **Implement .gitignore-based filtering logic**
  - **Action:** Add logic to parse .gitignore files and use their patterns to filter files and directories during traversal.
  - **Depends On:** Create directory reader utility function, Add dependency for gitignore parsing.
  - **AC Ref:** AC 2.3 - System must ignore files and directories according to .gitignore rules.

- [x] **Add detection and handling of binary files**
  - **Action:** Implement logic to detect binary files and skip them (with warning).
  - **Depends On:** Create context file reader utility function.
  - **AC Ref:** AC 2.4 - System must handle binary files appropriately.

- [x] **Add max file size limit checks**
  - **Action:** Implement checks for file size and skip files exceeding limit (with warning).
  - **Depends On:** Create context file reader utility function.
  - **AC Ref:** AC 2.5 - System must implement size limits for individual files.

- [x] **Create master readContextPaths function**
  - **Action:** Implement the main `readContextPaths` function that handles both files and directories, returning a combined array of path/content pairs.
  - **Depends On:** Create context file reader utility function, Create directory reader utility function, Implement .gitignore-based filtering logic.
  - **AC Ref:** AC 2.6 - System must provide a unified API for reading both files and directories.

## Input Formatting

- [x] **Create formatCombinedInput function**
  - **Action:** Implement function to combine prompt content with context files using the defined formatting strategy.
  - **Depends On:** None.
  - **AC Ref:** AC 3.1 - System must format combined content in a way LLMs can understand context separation.

- [x] **Modify _processInput helper**
  - **Action:** Update `_processInput` in `src/workflow/runThinktankHelpers.ts` to accept contextPaths, call readContextPaths, and combine content.
  - **Depends On:** Create master readContextPaths function, Create formatCombinedInput function.
  - **AC Ref:** AC 3.2 - Workflow must integrate context reading with prompt processing.

## Workflow Orchestration

- [x] **Update ProcessInputResult interface**
  - **Action:** Update interface in `runThinktankTypes.ts` to reflect combined prompt+context content.
  - **Depends On:** Modify _processInput helper.
  - **AC Ref:** AC 3.2 - Workflow must integrate context reading with prompt processing.

- [x] **Pass contextPaths to _processInput**
  - **Action:** Modify the call to `_processInput` in `runThinktank` to pass contextPaths from options.
  - **Depends On:** Update RunOptions interface, Modify _processInput helper.
  - **AC Ref:** AC 3.3 - runThinktank must pass contextPaths from options to input processing.

- [x] **Update ExecuteQueriesParams interface**
  - **Action:** Update interface to accept combined prompt+context content.
  - **Depends On:** Update ProcessInputResult interface.
  - **AC Ref:** AC 3.4 - Query execution must handle combined prompt+context.

- [x] **Modify call to _executeQueries**
  - **Action:** Update the call in `runThinktank` to pass combined content from inputResult.
  - **Depends On:** Update ExecuteQueriesParams interface.
  - **AC Ref:** AC 3.4 - Query execution must handle combined prompt+context.

## Testing - Unit Tests

- [x] **Create tests for context file reader utility**
  - **Action:** Write unit tests for the file reading function with various scenarios.
  - **Depends On:** Create context file reader utility function.
  - **AC Ref:** AC 5.1 - File reading functionality has test coverage.

- [x] **Create tests for directory reader utility**
  - **Action:** Write unit tests for recursive directory traversal with various scenarios.
  - **Depends On:** Create directory reader utility function.
  - **AC Ref:** AC 5.2 - Directory traversal functionality has test coverage.

- [x] **Create tests for .gitignore-based filtering logic**
  - **Action:** Write unit tests for .gitignore parsing and pattern matching against file paths.
  - **Depends On:** Implement .gitignore-based filtering logic.
  - **AC Ref:** AC 5.3 - Filtering logic has test coverage.

- [x] **Create tests for binary file detection**
  - **Action:** Write unit tests for binary file detection logic.
  - **Depends On:** Add detection and handling of binary files.
  - **AC Ref:** AC 5.4 - Binary file handling has test coverage.

- [x] **Create tests for formatCombinedInput**
  - **Action:** Write unit tests for the formatting function with various scenarios.
  - **Depends On:** Create formatCombinedInput function.
  - **AC Ref:** AC 5.5 - Formatting logic has test coverage.

- [x] **Create tests for _processInput changes**
  - **Action:** Update or create tests for the modified _processInput helper.
  - **Depends On:** Modify _processInput helper.
  - **AC Ref:** AC 5.6 - Input processing has test coverage.

- [x] **Create tests for CLI command changes**
  - **Action:** Write unit tests for the updated run command to verify contextPaths parsing.
  - **Depends On:** Update action handler to receive contextPaths.
  - **AC Ref:** AC 5.7 - CLI command has test coverage.

## Testing - Integration Tests

- [x] **Create integration tests for runThinktank workflow**
  - **Action:** Write tests that verify the workflow correctly passes context through the entire pipeline.
  - **Depends On:** Pass contextPaths to _processInput, Modify call to _executeQueries.
  - **AC Ref:** AC 6.1 - Integration tests verify workflow correctly handles context.

- [x] **Create integration tests for various path combinations**
  - **Action:** Write tests for different combinations of file/directory paths.
  - **Depends On:** Create integration tests for runThinktank workflow.
  - **AC Ref:** AC 6.2 - Integration tests cover various path combinations.

## Testing - E2E Tests

- [x] **Create E2E tests for CLI usage with context files**
  - **Action:** Create tests that run the actual CLI with file context arguments.
  - **Depends On:** All workflow components complete.
  - **AC Ref:** AC 7.1 - E2E tests verify CLI works with file context.

- [ ] **Create E2E tests for CLI usage with directory context**
  - **Action:** Create tests that run the actual CLI with directory context arguments.
  - **Depends On:** All workflow components complete.
  - **AC Ref:** AC 7.2 - E2E tests verify CLI works with directory context.

- [ ] **Create E2E tests for CLI usage with mixed context paths**
  - **Action:** Create tests that run the actual CLI with mixed file and directory arguments.
  - **Depends On:** All workflow components complete.
  - **AC Ref:** AC 7.3 - E2E tests verify CLI works with mixed context paths.

- [ ] **Create E2E tests for edge cases**
  - **Action:** Create tests for edge cases like non-existent paths, paths with spaces, etc.
  - **Depends On:** All workflow components complete.
  - **AC Ref:** AC 7.4 - E2E tests verify CLI handles edge cases.

## Documentation

- [ ] **Update CLI help text**
  - **Action:** Update the help text in `src/cli/commands/run.ts` with contextPaths parameter details.
  - **Depends On:** Update `commander` definition in run command.
  - **AC Ref:** AC 8.1 - CLI help text includes contextPaths parameter.

- [ ] **Update README with context path usage**
  - **Action:** Add documentation to README.md explaining how to use the context paths feature.
  - **Depends On:** All feature implementation complete.
  - **AC Ref:** AC 8.2 - README includes context paths feature documentation.

- [ ] **Document context formatting strategy**
  - **Action:** Add documentation explaining how context is formatted and presented to LLMs.
  - **Depends On:** Create formatCombinedInput function.
  - **AC Ref:** AC 8.3 - Documentation explains context formatting.
