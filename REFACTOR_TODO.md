# TODO

## Setup Test Utilities Structure

- [x] **Create shared test utilities directory**
  - **Action:** Create a new directory at `src/__tests__/utils/` to house shared test helper functions.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1, AC 1.2.

## Filesystem Mock Utilities

- [x] **Define filesystem mock utility interfaces**
  - **Action:** Create TypeScript interfaces for `FsMockConfig` and other types needed by the filesystem mock utility.
  - **Depends On:** Create shared test utilities directory.
  - **AC Ref:** AC 2.1.

- [x] **Implement core filesystem mock setup**
  - **Action:** Create `mockFsUtils.ts` with basic mock setup and reset functions for fs/promises.
  - **Depends On:** Define filesystem mock utility interfaces.
  - **AC Ref:** AC 2.1.

- [x] **Implement filesystem access mock helpers**
  - **Action:** Add helper functions to mock fs.access for specific paths with success/error responses.
  - **Depends On:** Implement core filesystem mock setup.
  - **AC Ref:** AC 2.2, AC 2.3.

- [x] **Implement filesystem readFile mock helpers**
  - **Action:** Add helper functions to mock fs.readFile for specific paths with content or errors.
  - **Depends On:** Implement core filesystem mock setup.
  - **AC Ref:** AC 2.2, AC 2.3.

- [x] **Implement filesystem stat mock helpers**
  - **Action:** Add helper functions to mock fs.stat for specific paths with file/directory stats or errors.
  - **Depends On:** Implement core filesystem mock setup.
  - **AC Ref:** AC 2.2, AC 2.3.

- [x] **Implement filesystem readdir mock helpers**
  - **Action:** Add helper functions to mock fs.readdir for specific directories with file lists or errors.
  - **Depends On:** Implement core filesystem mock setup.
  - **AC Ref:** AC 2.2, AC 2.3.

- [x] **Implement filesystem mkdir mock helpers**
  - **Action:** Add helper functions to mock fs.mkdir for specific paths with success/error responses.
  - **Depends On:** Implement core filesystem mock setup.
  - **AC Ref:** AC 2.2, AC 2.3.

- [x] **Implement filesystem writeFile mock helpers**
  - **Action:** Add helper functions to mock fs.writeFile for specific paths with success/error responses.
  - **Depends On:** Implement core filesystem mock setup.
  - **AC Ref:** AC 2.2, AC 2.3.

## Gitignore Mock Utilities

- [x] **Define gitignore mock utility interfaces**
  - **Action:** Create TypeScript interfaces for `GitignoreMockConfig` and other types needed by the gitignore mock utility.
  - **Depends On:** Create shared test utilities directory.
  - **AC Ref:** AC 3.1.

- [x] **Implement core gitignore mock setup**
  - **Action:** Create `mockGitignoreUtils.ts` with basic mock setup and reset functions for gitignoreUtils.
  - **Depends On:** Define gitignore mock utility interfaces.
  - **AC Ref:** AC 3.1.

- [x] **Implement shouldIgnorePath mock helper**
  - **Action:** Add helper function to mock gitignoreUtils.shouldIgnorePath with custom implementations.
  - **Depends On:** Implement core gitignore mock setup.
  - **AC Ref:** AC 3.2.

- [x] **Implement createIgnoreFilter mock helper**
  - **Action:** Add helper function to mock gitignoreUtils.createIgnoreFilter with custom implementations.
  - **Depends On:** Implement core gitignore mock setup.
  - **AC Ref:** AC 3.2.

## Refactor Test Files

- [x] **Refactor fileReader.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts`.
  - **Depends On:** Implement filesystem access mock helpers, Implement filesystem readFile mock helpers, Implement filesystem stat mock helpers, Implement filesystem mkdir mock helpers, Implement filesystem writeFile mock helpers.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor binaryFileDetection.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts`.
  - **Depends On:** Implement filesystem access mock helpers, Implement filesystem stat mock helpers.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor fileSizeLimit.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts`.
  - **Depends On:** Implement filesystem access mock helpers, Implement filesystem stat mock helpers, Implement filesystem readFile mock helpers.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor readContextFile.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts`.
  - **Depends On:** Implement filesystem access mock helpers, Implement filesystem readFile mock helpers, Implement filesystem stat mock helpers.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor readContextPaths.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts`.
  - **Depends On:** Implement filesystem access mock helpers, Implement filesystem stat mock helpers, Implement filesystem readdir mock helpers, Implement filesystem readFile mock helpers.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor readDirectoryContents.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts` and gitignore mocks with `mockGitignoreUtils.ts`.
  - **Depends On:** Implement filesystem readdir mock helpers, Implement filesystem stat mock helpers, Implement shouldIgnorePath mock helper.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor gitignoreUtils.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts`.
  - **Depends On:** Implement filesystem readFile mock helpers.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor gitignoreFiltering.test.ts**
  - **Action:** Replace direct fs mocks with calls to the new utility functions in `mockFsUtils.ts`.
  - **Depends On:** Implement filesystem readFile mock helpers.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor gitignoreFilterIntegration.test.ts**
  - **Action:** Replace direct fs and gitignore mocks with calls to the new utility functions.
  - **Depends On:** Implement filesystem readFile mock helpers, Implement shouldIgnorePath mock helper, Implement createIgnoreFilter mock helper.
  - **AC Ref:** AC 4.1, AC 4.2.

- [x] **Refactor gitignoreFilteringIntegration.test.ts**
  - **Action:** Replace direct fs and gitignore mocks with calls to the new utility functions.
  - **Depends On:** Implement filesystem readFile mock helpers, Implement shouldIgnorePath mock helper, Implement createIgnoreFilter mock helper.
  - **AC Ref:** AC 4.1, AC 4.2.

## Review and Validation

- [x] **Run test suite to verify refactoring**
  - **Action:** Run the entire test suite to ensure all tests still pass after the refactoring.
  - **Depends On:** All refactor test file tasks.
  - **AC Ref:** AC 5.1.

- [x] **Code review for consistency**
  - **Action:** Review all refactored test files to ensure consistent use of the new utilities and removal of unnecessary direct mocks.
  - **Depends On:** All refactor test file tasks.
  - **AC Ref:** AC 4.2.

- [x] **Document new test utilities**
  - **Action:** Add documentation for the new test utilities in relevant README files or comments within the utility files.
  - **Depends On:** Run test suite to verify refactoring.
  - **AC Ref:** AC 5.2.
