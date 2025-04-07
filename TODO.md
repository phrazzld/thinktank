# TODO

## Filesystem Testing Strategy Redesign

- [x] **Install memfs library**
  - **Action:** Add memfs to the project's devDependencies using `npm install --save-dev memfs`
  - **Depends On:** None
  - **AC Ref:** AC 1.1, AC 1.2

- [x] **Create virtualFsUtils.ts utility**
  - **Action:** Implement a new utility file that provides functions for creating, manipulating, and accessing an in-memory filesystem using memfs
  - **Depends On:** Install memfs library
  - **AC Ref:** AC 1.3, AC 1.4

- [x] **Remove test-specific logic from production code**
  - **Action:** Remove any checks for `_isTestError` or similar flags from `fileReader.ts` and other production code
  - **Depends On:** None
  - **AC Ref:** AC 1.5

- [x] **Create test migration guide**
  - **Action:** Update `src/__tests__/utils/README.md` with instructions on how to use the new virtualFsUtils instead of mockFsUtils
  - **Depends On:** Create virtualFsUtils.ts utility
  - **AC Ref:** AC 1.6

- [x] **Refactor fileReader.test.ts**
  - **Action:** Replace mockFsUtils usage with the new virtualFsUtils approach, ensuring all tests pass successfully
  - **Depends On:** Create virtualFsUtils.ts utility, Remove test-specific logic from production code
  - **AC Ref:** AC 2.1

- [x] **Refactor readContextFile.test.ts**
  - **Action:** Update tests to use virtualFsUtils instead of mockFsUtils, focusing on properly testing file reading behavior
  - **Depends On:** Create virtualFsUtils.ts utility, Remove test-specific logic from production code
  - **AC Ref:** AC 2.1, AC 2.2

- [x] **Refactor fileSizeLimit.test.ts**
  - **Action:** Update tests to use virtualFsUtils for simulating files of different sizes
  - **Depends On:** Create virtualFsUtils.ts utility, Remove test-specific logic from production code
  - **AC Ref:** AC 2.5

- [x] **Refactor binaryFileDetection.test.ts**
  - **Action:** Update tests to use virtualFsUtils for simulating file operations
  - **Depends On:** Create virtualFsUtils.ts utility, Remove test-specific logic from production code
  - **AC Ref:** AC 2.9

- [x] **Refactor readContextPaths.test.ts**
  - **Action:** Use virtualFsUtils to set up directory structures for testing path reading functionality
  - **Depends On:** Create virtualFsUtils.ts utility, Remove test-specific logic from production code
  - **AC Ref:** AC 2.2

- [x] **Refactor formatCombinedInput.test.ts**
  - **Action:** Review test dependencies and update any filesystem interactions to use the new approach
  - **Depends On:** Create virtualFsUtils.ts utility
  - **AC Ref:** AC 2.6

- [x] **Refactor gitignoreFilterIntegration.test.ts**
  - **Action:** Update tests to ensure proper integration between gitignore filtering and the new filesystem virtualization
  - **Depends On:** Create virtualFsUtils.ts utility, Refactor fileReader.test.ts
  - **AC Ref:** AC 2.4

- [x] **Refactor readDirectoryContents.test.ts**
  - **Action:** Update tests to use virtualFsUtils for testing directory reading, particularly recursive operations
  - **Depends On:** Create virtualFsUtils.ts utility, Refactor gitignoreFilterIntegration.test.ts
  - **AC Ref:** AC 2.3

- [x] **Refactor configManager.test.ts**
  - **Action:** Update tests for configuration loading, saving, and path resolution to use virtualFsUtils
  - **Depends On:** Create virtualFsUtils.ts utility, Remove test-specific logic from production code
  - **AC Ref:** AC 2.7

- [x] **Refactor outputHandler.test.ts**
  - **Action:** Update tests for output directory creation and file writing to use virtualFsUtils
  - **Depends On:** Create virtualFsUtils.ts utility
  - **AC Ref:** AC 2.8

- [x] **Review and refactor remaining tests with FS dependencies**
  - **Action:** Systematically identify and update all remaining tests that use filesystem operations
  - **Depends On:** Create virtualFsUtils.ts utility, Refactor fileReader.test.ts
  - **AC Ref:** AC 2.9

- [x] **Re-enable skipped tests in jest.config.js**
  - **Action:** As tests are successfully refactored, remove them from the testPathIgnorePatterns in jest.config.js
  - **Depends On:** Refactoring of the specific test file to be re-enabled
  - **AC Ref:** AC 3.1

- [x] **Fix failing tests and run full test suite**
  - **Action:** Address any remaining failures in the test suite and ensure all tests pass without worker crashes
  - **Depends On:** Re-enable skipped tests in jest.config.js
  - **AC Ref:** AC 3.2

- [x] **Review test coverage**
  - **Action:** Run test coverage analysis and identify critical gaps in filesystem operation testing
  - **Depends On:** Fix failing tests and run full test suite
  - **AC Ref:** AC 3.3

- [ ] **Update testing documentation**
  - **Action:** Update all documentation to reflect the new filesystem testing approach
  - **Depends On:** All previous tasks completed
  - **AC Ref:** AC 3.4

- [ ] **Consider refactoring mockGitignoreUtils**
  - **Action:** Evaluate if mockGitignoreUtils needs similar simplification as mockFsUtils
  - **Depends On:** Refactor gitignoreFilterIntegration.test.ts
  - **AC Ref:** AC 3.5

- [ ] **Refactor E2E tests**
  - **Action:** Update E2E tests to use temporary directories on the real filesystem instead of mocking
  - **Depends On:** Update testing documentation
  - **AC Ref:** AC 3.6
