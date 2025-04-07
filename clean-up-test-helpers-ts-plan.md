# Clean Up test-helpers.ts

## Goal
The goal of this task is to clean up the `test-helpers.ts` file by removing any functions related to the old filesystem mocking strategy. This is part of Phase 1 of the test suite refactoring plan, which aims to eliminate all dependencies on the legacy `mockFsUtils.ts` approach.

## Implementation Approach
After analyzing the `test-helpers.ts` file and examining usage patterns across the codebase, I plan to:

1. **Remove Unnecessary Comments**: Clean up any comments that still reference the old mocking strategy or `mockFsUtils.ts`.

2. **Retain Useful Utility Functions**: Keep the generic utility functions that are still valuable and not specifically tied to the old mocking strategy:
   - `wait(ms)`: This is a useful utility for timeouts in tests that should be kept.
   - `promisify(value)`: Keep this function since it's a useful utility for working with promises in tests.

3. **Update File Documentation**: Update the file documentation to reflect its current purpose as a general test utility provider rather than focusing on the legacy mocking strategy.

4. **Keep createFsError Re-Export**: The file already correctly re-exports `createFsError` from `virtualFsUtils.ts` (which was part of the previous task), so this should be kept as is.

5. **Run Tests**: Verify that all tests still pass after these changes.

## Reasoning
This approach is selected because:

1. **No Direct Dependencies Found**: After thorough searching, we didn't find any test files still importing `createTestSafeError` from `test-helpers.ts`. The comment already mentions it has been deprecated and removed.

2. **Utility Functions Are Still Valuable**: The `wait` and `promisify` functions are generic test utilities not tied to the old mocking strategy. They provide valuable functionality for tests (timing and promise handling). The code search showed imports of `wait` from this file across many test files.

3. **Minimal Risk**: By making only the necessary changes to comments and documentation while preserving the useful utility functions, we minimize the risk of breaking any tests. This approach maintains backward compatibility while still fulfilling the cleanup task.

4. **Clean Cut Migration**: The approach supports the clean migration from the legacy `mockFsUtils.ts` to the new `virtualFsUtils.ts` approach while ensuring that utility functions that are still useful remain available.

The changes are minimal since most of the cleanup was already done in the previous task (replacing `mockFsUtils.ts` import with `virtualFsUtils.ts` import for `createFsError`). This task is primarily about cleaning up comments and documentation to complete the migration.