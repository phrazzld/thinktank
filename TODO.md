# TODO

## Filesystem Testing Improvements

- [x] **Fix Unrestored Jest Spies**
  - **Action:** Add `jest.restoreAllMocks()` in `afterEach` blocks or individual `mockRestore()` calls where spies are used but not properly restored.
  - **Depends On:** None
  - **AC Ref:** Identified Issue - Unrestored Jest Spies

- [x] **Replace Hardcoded Paths**
  - **Action:** Search for and replace all Unix-style hardcoded paths with `path.join()` to ensure cross-platform compatibility.
  - **Depends On:** None
  - **AC Ref:** Identified Issue - Hardcoded Paths

- [ ] **Standardize Virtual FS Setup**
  - **Action:** Review test files, replace direct FS calls with `createVirtualFs` helper consistently across tests.
  - **Depends On:** None
  - **AC Ref:** Identified Issue - Inconsistent Virtual FS Setup

- [ ] **Add Missing Newlines**
  - **Action:** Configure editor/formatter to ensure consistent file endings with newline characters.
  - **Depends On:** None
  - **AC Ref:** Identified Issue - Missing Newlines

- [ ] **Improve Type Safety**
  - **Action:** Replace unsafe type assertions with more type-safe approaches or helper functions.
  - **Depends On:** None
  - **AC Ref:** Identified Issue - Unsafe Type Assertions

- [ ] **Expand Test Coverage**
  - **Action:** Add additional tests for edge cases, partial failures, and concurrent operations in output-directory tests.
  - **Depends On:** None
  - **AC Ref:** Identified Issue - Limited Test Coverage

- [x] **Fix Console Mock Pollution**
  - **Action:** Review and update console mocks to use `jest.spyOn` and restore in `afterEach` blocks.
  - **Depends On:** Fix Unrestored Jest Spies
  - **AC Ref:** Identified Issue - Console Mock Pollution

- [ ] **Remove Residual Legacy References**
  - **Action:** Search for and replace any remaining references to legacy mocking utilities.
  - **Depends On:** None
  - **AC Ref:** Identified Issue - Residual Legacy References

- [ ] **Create Centralized Mock Setup**
  - **Action:** Develop a shared Jest setup file for common mocks to reduce duplication.
  - **Depends On:** Fix Unrestored Jest Spies, Fix Console Mock Pollution
  - **AC Ref:** Best Practice - Centralized Mock Setup
