# Refactor E2E tests

## Goal
Update End-to-End (E2E) tests to use temporary directories on the real filesystem instead of mocking, ensuring more realistic testing scenarios that properly validate interactions with the actual filesystem.

## Implementation Approach

After analyzing the current E2E test implementation and reviewing the project's testing documentation, I'll implement a hybrid approach that combines:

1. **Real Filesystem Operations** - Use actual temporary directories for core filesystem operations
2. **Controlled Test Environment** - Use Node's `os.tmpdir()` for creating isolated test directories
3. **Cleanup Mechanisms** - Ensure proper cleanup of temporary files/directories after tests
4. **Proper Error Handling** - Handle platform-specific behaviors in the tests
5. **Skip Logic for CI Environments** - Add mechanisms to skip tests when appropriate

### Key Changes

1. **Update cli.e2e.test.ts**
   - Remove outdated shouldSkip logic for checking CLI build
   - Enhance temporary directory handling with better cleanup
   - Update test assertions to verify actual file contents on disk
   - Add platform-specific path handling

2. **Update runThinktank.e2e.test.ts**
   - Replace mocked filesystem operations with real operations
   - Enhance test fixtures to use platform-appropriate paths
   - Implement more robust context file testing
   - Add proper error simulation for edge cases

3. **Create Common E2E Test Utilities**
   - Implement shared functions for temporary directory management
   - Add utilities for test file creation and validation
   - Create platform-aware path utilities for cross-OS compatibility

## Reasoning for This Approach

I selected this approach for the following reasons:

1. **Comprehensive Testing**: Testing against real filesystems provides the most realistic validation of application behavior, especially for platform-specific edge cases.

2. **Isolation Without Mocking**: Using temporary directories maintains test isolation while avoiding the limitations of mocks.

3. **Simplicity and Maintainability**: The approach removes complex mocking setups in favor of straightforward filesystem operations, making tests easier to understand and maintain.

4. **Consistency with Project Direction**: The project is moving away from heavy mocking toward more realistic testing approaches, as evidenced by the recent migration to virtualFs for unit tests. This change represents the next logical step in that evolution.

5. **Better CI/CD Integration**: Tests that run against real filesystems are more representative of how the code will behave in production environments, improving confidence in CI/CD pipelines.