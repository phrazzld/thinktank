# TODO

## Fix Critical Testing Issues

- [x] **Fix skipped tests in readDirectoryContents.test.ts**
  - **Action:** Un-skip and fix all tests in readDirectoryContents.test.ts, ensuring they work correctly with the new virtual filesystem approach. Include tests for directory traversal, path handling, and error cases.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 1.

- [ ] **Complete gitignore integration tests**
  - **Action:** Update gitignoreFilterIntegration.test.ts and gitignoreFilteringIntegration.test.ts to fully verify end-to-end behavior using virtual gitignore files rather than mocking intermediate steps.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 2.

- [ ] **Fix readContextPaths.test.ts mock implementation**
  - **Action:** Replace the mock implementation of gitignoreUtils.shouldIgnorePath with the actual integration against virtual filesystem.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 2.

- [ ] **Restore complex gitignore pattern tests**
  - **Action:** Investigate and fix the limitations in testing complex glob patterns in gitignoreFiltering.test.ts. If not fully possible, document the limitations and create alternative tests.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 4.

## Fix Documentation Issues

- [ ] **Update README.md links**
  - **Action:** Fix broken links in README.md, ensuring they point to valid documentation resources.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 3.

- [ ] **Restore or relocate essential documentation**
  - **Action:** Restore or relocate essential documentation that was deleted (TESTING.md, MASTER_PLAN.md, etc.) to maintain accessibility for users and contributors.
  - **Depends On:** None.
  - **AC Ref:** Critical Issue 3.

- [ ] **Create testing approach documentation**
  - **Action:** Create comprehensive documentation on the virtual filesystem testing approach, including examples and best practices for future contributors.
  - **Depends On:** Fix skipped tests in readDirectoryContents.test.ts, Complete gitignore integration tests.
  - **AC Ref:** Recommendation 7.

## Standardize Testing Infrastructure

- [ ] **Implement consistent path normalization**
  - **Action:** Create and consistently use a standard path normalization utility across all tests to handle both Unix and Windows-style paths correctly.
  - **Depends On:** None.
  - **AC Ref:** Additional Issue 5, Recommendation 6.

- [ ] **Standardize mocking approach**
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

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Issue/Assumption:** Specific limitations of virtual filesystem for complex gitignore patterns
  - **Context:** The code review mentions that tests for complex glob patterns were replaced with simpler tests due to limitations, but doesn't specify what these limitations are.

- [ ] **Issue/Assumption:** Extent of remaining mocks
  - **Context:** While the review identifies inconsistent mocking strategies, it's unclear how many files still use the old approach vs. the new virtual filesystem approach.

- [ ] **Issue/Assumption:** Testing cross-platform behavior
  - **Context:** The review mentions inconsistent path handling between Unix and Windows-style paths, but it's unclear if tests need to run on multiple platforms or if normalization utilities are sufficient.

- [ ] **Issue/Assumption:** Documentation organization strategy
  - **Context:** The review mentions deleted documentation files, but doesn't specify whether these should be restored in their original locations or reorganized into a new structure.

- [ ] **Issue/Assumption:** Test isolation requirements
  - **Context:** While the gitignore cache clearing issue is noted, it's unclear if there are other state-related issues that might affect test isolation.