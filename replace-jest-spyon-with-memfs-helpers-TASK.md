# Replace jest.spyOn with memfs helpers

## Task Details
- **Action:** Refactor tests to use memfs helpers like createVirtualFs and createFsError instead of jest.spyOn where appropriate.
- **Depends On:** Standardize mocking approach (already completed).
- **AC Ref:** Additional Issue 6.

## Context
The thinktank project is undergoing a standardization of its test approach. The codebase already has memfs virtual filesystem helpers (in `src/__tests__/utils/virtualFsUtils.ts` and `src/__tests__/utils/fsTestSetup.ts`), but some tests are still using `jest.spyOn` to mock filesystem operations. According to the project's testing philosophy, we should minimize mocking and use consistent approaches throughout the codebase.

## Requirements
1. Identify tests that use jest.spyOn to mock filesystem operations
2. Replace these mocks with the existing memfs helpers where appropriate
3. Focus particularly on tests that mock fs/promises.readdir, fs/promises.stat, fs.readFile, etc.
4. Unskip tests that can now be implemented using memfs helpers
5. Ensure all tests pass after refactoring

## Request for Implementation Approaches

Please provide 2-3 different approaches for implementing this task, including:

1. A detailed explanation of each approach
2. The pros and cons of each approach
3. A recommendation for which approach to use, considering the project's testing philosophy (from TESTING_PHILOSOPHY.MD) and best practices

When recommending an approach, please consider:
- Minimizing mocking while ensuring tests are reliable
- Consistency with the project's existing testing patterns
- Impact on test readability and maintainability
- Ability to test edge cases and error scenarios
- Alignment with the "Testability is a Design Goal" principle from TESTING_PHILOSOPHY.MD

Thank you!