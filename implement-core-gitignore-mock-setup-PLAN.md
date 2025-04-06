# Implement core gitignore mock setup

## Task Goal
Implement the core setup and reset functions in the `mockGitignoreUtils.ts` file, which will provide the foundation for mocking gitignore-related utilities in tests. This task builds upon the interfaces defined in the previous task and focuses on implementing the `resetMockGitignore` and `setupMockGitignore` functions.

## Implementation Approach
I'll enhance the existing `mockGitignoreUtils.ts` file by implementing the following:

1. **Registry Data Structures**:
   - Create rule registries (similar to the filesystem mock utilities) to store custom behaviors for both `shouldIgnorePath` and `createIgnoreFilter` functions
   - These registries will store pattern-specific rules that will override default behaviors

2. **`resetMockGitignore` Function**:
   - Clear all mock settings and rule registries
   - Reset the mock implementations of `shouldIgnorePath` and `createIgnoreFilter`
   - Reset any other mock state, such as call counts and arguments

3. **`setupMockGitignore` Function**:
   - Configure default behavior for the gitignore mock utilities based on provided configuration
   - Set up mock implementations of `shouldIgnorePath` and `createIgnoreFilter` that check rule registries before applying default behavior
   - Handle various use cases, including default ignore patterns and custom behavior

4. **Default Configurations**:
   - Define sensible default values for all configuration options
   - Ensure these defaults align with typical gitignore behavior in a test environment

## Key Reasoning

1. **Consistent mock architecture**: Following the same pattern as the filesystem mock utilities ensures a consistent experience for developers, making the test utilities more intuitive to use together.

2. **Registry-based implementation**: Using rule registries for path patterns provides flexibility to handle both global defaults and path-specific behaviors, which is essential for complex test scenarios.

3. **Clear state management**: Implementing a thorough reset function prevents test pollution and ensures test isolation, which is critical for reliable test suites.

4. **Configurable defaults**: Making the default behavior configurable through the setup function allows tests to establish baseline behavior without specifying rules for every path.

5. **Preparation for specific mocking functions**: This core implementation lays the groundwork for implementing the more specific mocking functions (`mockShouldIgnorePath` and `mockCreateIgnoreFilter`) in subsequent tasks.

6. **Gradual implementation**: By focusing on the core setup and reset functionality first, we ensure a solid foundation before moving on to more complex mocking behaviors.