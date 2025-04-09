# TODO

## I/O Separation and Test Infrastructure Improvements

### High Priority

- [x] **Remove Duplicated File Writing Logic**
  - **Action:** Delete `_writeOutputFiles` from `runThinktankHelpers.ts` and ensure all code uses only `io.writeFiles`.
  - **Depends On:** None.
  - **AC Ref:** Effective I/O Separation (Key Strength 1).

### Medium Priority

- [ ] **Refactor ConcreteFileSystem Tests**
  - **Action:** Refactor tests in `src/core/__tests__/FileSystem.test.ts` to use virtual FS (`memfs`) via `test/setup/fs.ts` instead of mocking internal dependency. Test behavior and error wrapping, not just delegation.
  - **Depends On:** None.
  - **AC Ref:** Interface-Based Mocking (Key Strength 4).

- [ ] **Add Tests for I/O Module**
  - **Action:** Create new test file `src/workflow/__tests__/io.test.ts` that properly tests the I/O module by mocking `FileSystem`/`ConsoleLogger`/`UISpinner`.
  - **Depends On:** None.
  - **AC Ref:** Test Coverage for `io.ts` (Opportunity 2).

### Low Priority

- [ ] **Fix Example Tests**
  - **Action:** Address failing assertions in example test files to serve as proper reference implementation.
  - **Depends On:** None.
  - **AC Ref:** Example Tests (Opportunity 3).

- [ ] **Improve Error Handling in ConcreteFileSystem**
  - **Action:** Extract common error wrapping logic in `src/core/FileSystem.ts` into private helper methods within the class to reduce repetition.
  - **Depends On:** None.
  - **AC Ref:** Repetitive Error Wrapping.

- [ ] **Standardize Logger Usage**
  - **Action:** Inject `ConsoleLogger` instead of using singleton `logger` for final summary logging in `src/workflow/runThinktank.ts`. This is to maintain consistency with the preferred dependency injection pattern.
  - **Depends On:** None.
  - **AC Ref:** Inconsistent logger usage.

- [ ] **Update Testing Documentation**
  - **Action:** Update docs in `jest/README.md` and `src/__tests__/utils/README.md` to reflect `test/setup/` as standard, document new factories/helpers, and deprecate old patterns.
  - **Depends On:** All other tasks.
  - **AC Ref:** Testing documentation needs update.
