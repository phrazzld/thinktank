# Refactor gitignoreFilteringIntegration.test.ts

## Task Goal
Replace direct fs and gitignore mocks in the gitignoreFilteringIntegration.test.ts file with calls to the new utility functions from mockFsUtils.ts and mockGitignoreUtils.ts.

## Implementation Approach
1. Analyze the existing gitignoreFilteringIntegration.test.ts file to identify:
   - Direct fs module mocks that need to be replaced with mockFsUtils functions
   - Direct gitignore utility mocks that need to be replaced with mockGitignoreUtils functions
2. Import the necessary mock utilities from both mockFsUtils.ts and mockGitignoreUtils.ts
3. Replace the current test setup with proper calls to:
   - `resetMockFs()` and `setupMockFs()` from mockFsUtils.ts
   - `resetMockGitignore()` and `setupMockGitignore()` from mockGitignoreUtils.ts
4. Replace direct fs mocks with appropriate mockFsUtils functions:
   - Replace `mockedFs.access` with `mockAccess`
   - Replace `mockedFs.stat` with `mockStat`
   - Replace `mockedFs.readdir` with `mockReaddir`
   - Replace `mockedFs.readFile` with `mockReadFile`
5. Replace direct gitignore mocks with the appropriate mockGitignoreUtils functions:
   - Use `mockedGitignoreUtils.shouldIgnorePath.mockImplementation()` instead of direct mocking
6. Ensure all three test cases maintain the same behavior:
   - "should filter files based on gitignore patterns"
   - "should handle nested .gitignore files with different patterns"
   - "should respect negated patterns"
7. Maintain test assertions and expectations as they are

## Reasoning
This approach follows the established pattern that has been successfully applied in the other recently refactored test files (including the similar gitignoreFilterIntegration.test.ts) while ensuring all tests continue to function correctly. The refactoring will make the tests more maintainable and consistent with the rest of the codebase.

The main benefits of this refactoring include:

1. **Consistency**: All test files will use the same approach for mocking filesystem and gitignore operations
2. **Improved readability**: The mock setup will be more declarative and easier to understand with explicit function calls
3. **Reduced duplication**: Common mock patterns are consolidated in utility functions
4. **Better maintainability**: Changes to mock implementation only need to be made in one place (the utility files)
5. **Simplified test setup**: More streamlined test preparation with less boilerplate code

This test file tests the integration between filesystem operations and gitignore filtering, focusing on complex scenarios like nested .gitignore files and negated patterns. By using the dedicated mock utilities for both concerns, the test will be more focused on testing the integration behavior rather than on setting up complex mocks.