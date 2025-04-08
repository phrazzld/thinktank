# TODO

## Decouple Dependencies with Interfaces

- [x] **Create Interface definitions for external dependencies**
  - **Action:** Define interfaces for LLMClient, ConfigManagerInterface, and FileSystem in a new `src/core/interfaces.ts` file.
  - **Depends On:** None.
  - **AC Ref:** AC 2

- [x] **Create FileSystem interface implementation**
  - **Action:** Implement a concrete FileSystem interface that wraps the existing file operations in fileReader.ts.
  - **Depends On:** Create Interface definitions for external dependencies.
  - **AC Ref:** AC 2, AC 3

- [x] **Create LLMClient interface implementation**
  - **Action:** Create a concrete implementation of LLMClient that wraps the existing provider logic.
  - **Depends On:** Create Interface definitions for external dependencies.
  - **AC Ref:** AC 2, AC 3

- [x] **Create ConfigManager interface implementation**
  - **Action:** Implement a concrete ConfigManagerInterface that wraps the existing configManager functionality.
  - **Depends On:** Create Interface definitions for external dependencies.
  - **AC Ref:** AC 2, AC 3

- [x] **Refactor _executeQueries to use dependency injection**
  - **Action:** Modify the _executeQueries function to accept and use the LLMClient interface instead of making direct API calls.
  - **Depends On:** Create LLMClient interface implementation.
  - **AC Ref:** AC 3, AC 4

- [x] **Update unit tests for _executeQueries**
  - **Action:** Modify the executeQueriesHelper.test.ts to use mock implementations of the LLMClient interface.
  - **Depends On:** Refactor _executeQueries to use dependency injection.
  - **AC Ref:** AC 4

- [x] **Refactor _setupWorkflow to use ConfigManager interface**
  - **Action:** Modify the _setupWorkflow function to accept and use the ConfigManagerInterface instead of direct config operations.
  - **Depends On:** Create ConfigManager interface implementation.
  - **AC Ref:** AC 3, AC 4

- [x] **Update unit tests for _setupWorkflow**
  - **Action:** Modify the setupWorkflowHelper.test.ts to use mock implementations of the ConfigManagerInterface.
  - **Depends On:** Refactor _setupWorkflow to use ConfigManager interface.
  - **AC Ref:** AC 4

- [x] **Refactor _processInput to use FileSystem interface**
  - **Action:** Modify the _processInput function to use the FileSystem interface instead of direct file operations.
  - **Depends On:** Create FileSystem interface implementation.
  - **AC Ref:** AC 3, AC 4

- [x] **Update unit tests for _processInput**
  - **Action:** Update the processInputHelper.test.ts to use mock implementations of the FileSystem interface.
  - **Depends On:** Refactor _processInput to use FileSystem interface.
  - **AC Ref:** AC 4

- [x] **Refactor _processOutput to use FileSystem interface**
  - **Action:** Modify the _processOutput function to use the FileSystem interface for writing output files.
  - **Depends On:** Create FileSystem interface implementation.
  - **AC Ref:** AC 3, AC 4

- [x] **Change default query timeout (in milliseconds) to five minutes**
  - **Action:** Created DEFAULT_QUERY_TIMEOUT_MS constant in src/core/constants.ts and updated all occurrences of 120000 to use this constant
  - **Depends On:** Nothing
  - **AC Ref:** None

- [x] **Update unit tests for _processOutput**
  - **Action:** Update the processOutputHelper.test.ts to use mock implementations of the FileSystem interface.
  - **Depends On:** Refactor _processOutput to use FileSystem interface.
  - **AC Ref:** AC 4

- [x] **Integrate interfaces with runThinktank workflow**
  - **Action:** Updated the main runThinktank function to instantiate and inject the concrete implementations of the interfaces.
  - **Depends On:** All refactoring of helper functions.
  - **AC Ref:** AC 1, AC 3, AC 4

- [ ] **Update runThinktank.test.ts to use interface mocks**
  - **Action:** Modify the runThinktank tests to use mock implementations of all three interfaces.
  - **Depends On:** Integrate interfaces with runThinktank workflow.
  - **AC Ref:** AC 4
