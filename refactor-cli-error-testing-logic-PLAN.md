# Refactor CLI error testing logic

## Task Goal
Remove duplicated error handling logic in `cli-error-handling.test.ts` by having tests import and use the actual `handleError` function from `src/cli/index.ts` with properly mocked dependencies.

## Current Implementation Analysis
Looking at the current implementation, I can see:

1. `cli-error-handling.test.ts` contains mock implementations of error handling functions (`mockHandleError`, `wrapMockError`, `addMockGuidance`) that duplicate the logic in the actual `handleError`, `wrapStandardError`, and `addCategorySpecificGuidance` functions in `src/cli/index.ts`.

2. The test file captures output by overriding the `logger.error` method and storing its calls, then reconstituting the output as a string.

3. The mocked functions are almost identical to the actual functions in `/src/cli/index.ts`.

4. The tests themselves validate that error messages are formatted correctly and contain appropriate guidance based on error type.

## Implementation Approach

I will:

1. Import the actual `handleError` function and related utility functions from `src/cli/index.ts` instead of duplicating them.

2. Create a wrapper function for test usage that:
   - Mocks `process.exit` to prevent tests from terminating
   - Captures the output from `logger.error` in the same way as the current implementation
   - Calls the actual `handleError` function
   - Returns the captured output

3. Refactor the tests to use this wrapper instead of the duplicated functions.

4. Update any tests that might rely on specific behavior in the mock implementations.

## Key Benefits

1. **DRY Principle**: Tests will now use the actual error handling logic, ensuring that they're testing what's really in production.

2. **Automatic Updates**: If the error handling logic in `src/cli/index.ts` changes, the tests will automatically test the new behavior.

3. **Better Coverage**: Tests will cover the actual implementation, not duplicated code.

4. **Maintainability**: Changes to error handling will only need to be made in one place.

## Implementation Details

1. Create a test utility function that wraps calls to `handleError` to capture output and prevent process.exit
2. Update the tests to use this wrapper instead of the mock functions
3. Remove the duplicated mock functions 
4. Ensure tests still validate the same functionality and pass

This approach will retain the same test coverage while eliminating duplication and improving maintainability.