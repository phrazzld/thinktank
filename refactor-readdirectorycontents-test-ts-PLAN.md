# Refactor readDirectoryContents.test.ts

## Task Goal
Replace direct fs mocks in the readDirectoryContents.test.ts file with calls to the new utility functions from mockFsUtils.ts and replace gitignore mocks with mockGitignoreUtils.ts.

## Implementation Approach
1. Analyze the existing readDirectoryContents.test.ts file to identify all instances where:
   - fs module mocks are directly set up
   - gitignore utility mocks are directly set up
2. Import the necessary mock utilities from mockFsUtils.ts and mockGitignoreUtils.ts
3. Replace each direct fs mock with the corresponding utility function calls (mockStat, mockReaddir)
4. Replace each direct gitignore mock with the corresponding utility function call (mockShouldIgnorePath)
5. Implement proper setup and teardown in test beforeEach/afterEach hooks using resetMockFs() and resetMockGitignore()
6. Maintain all existing test cases and behavior while using the new mock utilities

## Reasoning
This approach follows the pattern established in previously refactored tests while ensuring all tests continue to function correctly. Using the shared mock utilities provides several benefits:

1. **Consistency**: All test files will use the same approach for mocking filesystem and gitignore operations
2. **Improved readability**: The mock setup will be more declarative and easier to understand
3. **Reduced duplication**: Common mock patterns are consolidated in utility functions
4. **Better maintainability**: Changes to mock implementation only need to be made in one place
5. **Simplified test setup**: More streamlined test preparation with less boilerplate code

This approach maintains the original test behaviors and assertions while modernizing the implementation to align with the project's test utility structure.