# TODO

## Fix Critical Testing Issues

- [x] **Fix skipped tests in readDirectoryContents.test.ts**
  - **Action:** Un-skip and fix all tests in readDirectoryContents.test.ts, ensuring they work correctly with the new virtual filesystem approach. Include tests for directory traversal, path handling, and error cases.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 1.

- [x] **Complete gitignore integration tests**
  - **Action:** Update gitignoreFilterIntegration.test.ts and gitignoreFilteringIntegration.test.ts to fully verify end-to-end behavior using virtual gitignore files rather than mocking intermediate steps.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 2.

- [x] **Fix readContextPaths.test.ts mock implementation**
  - **Action:** Replace the mock implementation of gitignoreUtils.shouldIgnorePath with the actual integration against virtual filesystem.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 2.

- [x] **Restore complex gitignore pattern tests**
  - **Action:** Investigate and fix the limitations in testing complex glob patterns in gitignoreFiltering.test.ts. If not fully possible, document the limitations and create alternative tests.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 4.

## Fix Documentation Issues

- [x] **Update README.md links**
  - **Action:** Fix broken links in README.md, ensuring they point to valid documentation resources.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 3.

## Standardize Testing Infrastructure

- [x] **Implement consistent path normalization**
  - **Action:** Create and consistently use a standard path normalization utility across all tests to handle both Unix and Windows-style paths correctly.
  - **Depends On:** None.
  - **AC Ref:** Additional Issue 5, Recommendation 6.

- [x] **Standardize mocking approach**
  - **Action:** Choose a consistent approach between mockFactories.ts and manual mocks in src/utils/__mocks__/, refactoring tests to follow the chosen pattern.
  - **Depends On:** None.
  - **AC Ref:** Additional Issue 6, Recommendation 5.

- [ ] **Create shared test setup helpers**
  - **Action:** Develop and standardize helper functions for common test setup patterns to reduce code duplication and improve maintainability.
  - **Depends On:** Standardize mocking approach, Implement consistent path normalization.
  - **AC Ref:** Recommendation 5.

- [ ] **Ensure consistent gitignore cache clearing**
  - **Action:** Add gitignoreUtils.clearIgnoreCache() to all relevant beforeEach hooks in test files to prevent test interdependencies.
  - **Depends On:** None.
  - **AC Ref:** Additional Issue 8.

## Code Cleanup

- [ ] **Remove debugging code from tests**
  - **Action:** Remove all console.log statements and other debugging code from test files to reduce noise during test execution.
  - **Depends On:** None.
  - **AC Ref:** Additional Issue 7.

- [ ] **Replace jest.spyOn with memfs helpers**
  - **Action:** Refactor tests to use memfs helpers like createVirtualFs and createFsError instead of jest.spyOn where appropriate.
  - **Depends On:** Standardize mocking approach.
  - **AC Ref:** Additional Issue 6.

## Test Implementation Completion

- [ ] **Fix remaining skipped tests in other files**
  - **Action:** Identify and fix all other skipped tests across the codebase, ensuring complete test coverage.
  - **Depends On:** Fix skipped tests in readDirectoryContents.test.ts.
  - **AC Ref:** Critical Issue 1, Recommendation 2.

- [ ] **Verify test coverage**
  - **Action:** Run test coverage analysis to ensure adequate coverage of all components, especially those affected by the refactoring.
  - **Depends On:** Fix skipped tests in readDirectoryContents.test.ts, Complete gitignore integration tests, Fix remaining skipped tests in other files.
  - **AC Ref:** Recommendation 2.
