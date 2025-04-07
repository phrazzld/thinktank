# TODO

## Enhance Filesystem Utilities
- [x] **Add hidden file support to virtualFsUtils.ts**
  - **Action:** Ensure virtualFsUtils.ts can properly create and handle hidden files like .gitignore in the virtual filesystem.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1.

- [x] **Implement or integrate addVirtualGitignoreFile functionality**
  - **Action:** Review the existing implementation of addVirtualGitignoreFile and either enhance it or integrate its logic into virtualFsUtils.ts for a unified approach.
  - **Depends On:** Add hidden file support to virtualFsUtils.ts.
  - **AC Ref:** AC 1.1.

## Refactor Gitignore-Related Tests
- [x] **Remove mock dependencies from gitignore tests**
  - **Action:** Identify all tests that use jest.mock('../gitignoreUtils') and remove these mocks along with any imports from mockGitignoreUtils.
  - **Depends On:** Implement or integrate addVirtualGitignoreFile functionality.
  - **AC Ref:** AC 2.1.

- [x] **Import actual gitignoreUtils functions**
  - **Action:** Replace mock imports with actual functions from src/utils/gitignoreUtils in all affected test files.
  - **Depends On:** Remove mock dependencies from gitignore tests.
  - **AC Ref:** AC 2.1.

- [ ] **Setup virtual .gitignore file creation in tests**
  - **Action:** Implement beforeEach hooks to create virtual filesystem with .gitignore files for each test, following the pattern in the example.
  - **Depends On:** Implement or integrate addVirtualGitignoreFile functionality.
  - **AC Ref:** AC 2.2.

- [ ] **Update test assertions to use actual implementation**
  - **Action:** Modify test assertions to test against the actual gitignoreUtils implementation using the virtual filesystem.
  - **Depends On:** Setup virtual .gitignore file creation in tests.
  - **AC Ref:** AC 2.3.

- [ ] **Add cache clearing logic in tests**
  - **Action:** Ensure gitignoreUtils.clearIgnoreCache() is called in the test setup if the implementation uses caching.
  - **Depends On:** Setup virtual .gitignore file creation in tests.
  - **AC Ref:** AC 2.2.

## Cleanup
- [ ] **Remove mockGitignoreUtils.ts**
  - **Action:** Once all tests have been refactored to use the virtual filesystem approach, delete the mockGitignoreUtils.ts file.
  - **Depends On:** Update test assertions to use actual implementation.
  - **AC Ref:** AC 3.1.

- [ ] **Run all tests to verify refactoring**
  - **Action:** Run the full test suite to ensure all refactored tests pass and no regressions were introduced.
  - **Depends On:** Remove mockGitignoreUtils.ts.
  - **AC Ref:** All ACs.

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS
- [ ] **Assumption:** Existing virtualFsUtils.ts already has most of the needed functionality but may need enhancement for hidden files.
  - **Context:** Step 1 mentions "Ensure virtualFsUtils.ts can create hidden files" which implies the file exists but may need enhancement.

- [ ] **Assumption:** The function clearIgnoreCache() exists in gitignoreUtils.
  - **Context:** Step 2 example code mentions calling gitignoreUtils.clearIgnoreCache(), but its existence isn't explicitly confirmed.