# TODO

## Test Migration Completion

- [x] **Enable skipped tests in virtualFsUtils.test.ts**
  - **Action:** Review, update, and un-skip tests for the `addVirtualGitignoreFile` function to ensure proper validation of its behavior, including error handling and edge cases.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1, AC 2.2.

- [x] **Fix virtualFsUtils core function tests**
  - **Action:** Update the implementation of `addVirtualGitignoreFile` in virtualFsUtils.ts to address any issues found during testing, ensure proper path normalization, and handle edge cases.
  - **Depends On:** Enable skipped tests in virtualFsUtils.test.ts
  - **AC Ref:** AC 1.1, AC 2.2.

- [x] **Enable skipped tests in gitignoreFilterIntegration.test.ts**
  - **Action:** Review, update, and un-skip integration tests that verify the gitignore filter works correctly with directory traversal, focusing on real-world file patterns and structures.
  - **Depends On:** Fix virtualFsUtils core function tests
  - **AC Ref:** AC 2.3, AC 2.4.

- [x] **Enable skipped tests in gitignoreFiltering.test.ts**
  - **Action:** Update and un-skip tests for the gitignore pattern filtering to ensure correct behavior with the virtual filesystem approach.
  - **Depends On:** Fix virtualFsUtils core function tests
  - **AC Ref:** AC 2.3.

- [x] **Enable skipped tests in gitignoreFilteringIntegration.test.ts**
  - **Action:** Review, update, and un-skip integration tests that validate the gitignore filtering logic works correctly in real-world scenarios.
  - **Depends On:** Enable skipped tests in gitignoreFiltering.test.ts
  - **AC Ref:** AC 2.3, AC 2.4.

- [x] **Enable skipped tests in readDirectoryContents.test.ts**
  - **Action:** Update and un-skip tests for the `readDirectoryContents` function to ensure it correctly interacts with the gitignore filtering mechanism using the virtual filesystem.
  - **Depends On:** Enable skipped tests in gitignoreFilteringIntegration.test.ts
  - **AC Ref:** AC 2.3, AC 2.4.
  - **Note:** The core gitignore functionality tests have been enabled. Some path-handling edge cases in the directory traversal tests still need refinement in a separate task but the core integration with gitignore filtering is now tested.

- [x] **Document limitations for complex gitignore patterns**
  - **Action:** Investigate limitations with complex gitignore pattern testing identified in the skipped test in gitignoreFiltering.test.ts. If the limitation can't be resolved, document it clearly in the test and consider workarounds.
  - **Depends On:** Enable skipped tests in gitignoreFiltering.test.ts
  - **AC Ref:** AC 2.3, AC 2.5.

## Test Code Cleanup

- [x] **Remove console.log statements from test code**
  - **Action:** Identify and remove all `console.log` statements from test files, particularly in gitignoreFiltering.test.ts and gitignoreUtils.test.ts, to ensure clean test output.
  - **Depends On:** None.
  - **AC Ref:** AC 3.1.

- [x] **Complete removal of mockGitignoreUtils references**
  - **Action:** Remove the remaining `jest.mock('../gitignoreUtils')` in readContextPaths.test.ts and update the test to use the virtual filesystem approach consistently.
  - **Depends On:** None.
  - **AC Ref:** AC 3.1.

- [x] **Create standardized path handling helper**
  - **Action:** Develop a shared helper function for path normalization to ensure consistent handling of leading slashes and path separators across all tests.
  - **Depends On:** None.
  - **AC Ref:** AC 3.2.

- [x] **Refactor tests to use standardized path handling**
  - **Action:** Update all test files to use the new path normalization helper function, ensuring consistent behavior across the test suite.
  - **Depends On:** Create standardized path handling helper
  - **AC Ref:** AC 3.2.

- [x] **Create reusable setup for filesystem mocking**
  - **Action:** Extract common patterns for setting up filesystem mocks into shared helper functions to reduce duplication across test files.
  - **Depends On:** None.
  - **AC Ref:** AC 3.3.

- [x] **Create reusable cache clearing mechanism**
  - **Action:** Implement a consistent approach for clearing the gitignore cache in beforeEach hooks across all tests that use gitignore functionality.
  - **Depends On:** None.
  - **AC Ref:** AC 3.3.

- [ ] **Refactor complex mocking patterns**
  - **Action:** Identify and simplify complex mocking patterns that use `Object.defineProperty` or mock the function being tested, replacing them with cleaner approaches using the virtual filesystem setup.
  - **Depends On:** Create reusable setup for filesystem mocking
  - **AC Ref:** AC 3.4.

## Documentation Updates

- [ ] **Create missing documentation files**
  - **Action:** Create or restore essential documentation files that were removed during the refactoring, especially TESTING.md if needed.
  - **Depends On:** None.
  - **AC Ref:** AC 4.1.

- [ ] **Fix broken links in README.md**
  - **Action:** Update all documentation links in README.md to point to the correct locations after the documentation reorganization.
  - **Depends On:** Create missing documentation files
  - **AC Ref:** AC 4.1.

- [ ] **Create testing guide for virtual filesystem approach**
  - **Action:** Document the new virtual filesystem testing approach, including best practices, common patterns, and examples to help future contributors.
  - **Depends On:** Complete all test migration and cleanup tasks
  - **AC Ref:** AC 4.2.

## Final Verification

- [ ] **Run full test suite**
  - **Action:** Ensure all tests are passing with the completed refactoring, with no skipped tests or temporary mocks remaining.
  - **Depends On:** Enable all skipped tests, Complete removal of mockGitignoreUtils references
  - **AC Ref:** All ACs.

- [ ] **Verify documentation coherence**
  - **Action:** Review all documentation to ensure it presents a coherent and accurate picture of the codebase after the refactoring.
  - **Depends On:** Fix broken links in README.md, Create testing guide for virtual filesystem approach
  - **AC Ref:** AC 4.1, AC 4.2.

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Assumption:** The Acceptance Criteria referenced in the tasks (AC 1.1, AC 2.3, etc.) are implied from the code review, as they are not explicitly defined in the CODE_REVIEW.md file.
  - **Context:** The code review does not explicitly list Acceptance Criteria with identifiers, so I've derived them from the issues and solutions in the review.

- [ ] **Assumption:** The limitations with complex gitignore patterns in `gitignoreFiltering.test.ts` are related to the `memfs` or `ignore` library capabilities.
  - **Context:** The code review mentions "Test Limitations for Complex Patterns" but doesn't explain the underlying technical limitation that causes the test to be skipped.

- [ ] **Assumption:** All skipped tests are intended to be enabled eventually, not permanently skipped.
  - **Context:** The code review recommends "Un-skip and fix these tests" but doesn't clarify if there are any tests that should remain skipped due to unsolvable limitations.

- [ ] **Assumption:** The reorganized documentation structure is intended to have documentation files primarily in the docs/ directory rather than the root.
  - **Context:** The review notes issues with documentation migration but doesn't explicitly state the intended final structure.