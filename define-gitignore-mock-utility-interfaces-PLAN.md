# Define gitignore mock utility interfaces

## Task Goal
Create TypeScript interfaces for `GitignoreMockConfig` and other types needed by the gitignore mock utility in the `mockGitignoreUtils.ts` file. These interfaces will provide a type-safe foundation for implementing the gitignore mock utilities, which are used to simulate the behavior of gitignore pattern matching in tests.

## Implementation Approach
I'll create a new file `mockGitignoreUtils.ts` in the `src/__tests__/utils/` directory. This file will contain:

1. A `GitignoreMockConfig` interface that defines the configuration options for the gitignore mock utilities, including:
   - Default behavior for unmatched paths
   - Default match patterns
   - Default filter patterns

2. Interfaces for mock functions that will be implemented later:
   - `MockShouldIgnorePathFunction` for the `shouldIgnorePath` function
   - `MockCreateIgnoreFilterFunction` for the `createIgnoreFilter` function

3. Type definitions for the gitignore rule patterns, which could include:
   - Simple string patterns
   - Regular expressions
   - Functions for custom matching logic

The implementation will allow for both global default behavior and path-specific overrides, similar to the file system mock utilities, providing a consistent API style across the test utilities.

## Key Reasoning

1. **Consistency with fs mock utilities**: Following a similar pattern to the filesystem mock utilities ensures consistent API design, making it easier for developers to work with both sets of utilities.

2. **Type safety**: By defining interfaces upfront, we ensure strong typing throughout the implementation phase, catching potential type errors early and providing better IDE support.

3. **Separation of concerns**: Creating separate interfaces for configuration and mock functions follows good software design principles, making the code more maintainable and easier to understand.

4. **Extensibility**: Defining clear interfaces now makes it easier to extend the functionality in the future without breaking existing tests.

5. **Documentation**: The interfaces serve as documentation, helping developers understand how to use the mocking utilities without having to read implementation details.

6. **Preparation for implementation**: By thinking through the interface design first, we'll have a clearer picture of the implementation requirements, potentially avoiding rework later.