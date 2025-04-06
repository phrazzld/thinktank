# Implement filesystem access mock helpers

## Task Goal
Add helper functions to mock `fs.access` for specific paths with success/error responses in the `mockFsUtils.ts` file. This will allow tests to specify which file paths should be accessible and which should trigger errors with specific error codes.

## Implementation Approach
I'll implement a new function called `mockAccess` that will allow tests to:
1. Specify file paths or path patterns (using strings or RegExp) to mock
2. Configure whether access to those paths should be allowed or denied
3. Customize error codes and messages for denied access

The implementation will follow these steps:
1. Create a registry of path patterns and their access configurations
2. Override the default `fs.access` mock to check if the requested path matches any patterns
3. If a match is found, return success or error based on the configuration
4. If no match is found, fall back to the default behavior

## Key Reasoning
I've chosen this approach for the following reasons:

1. **Pattern-based matching**: By supporting both exact path strings and regular expressions, we provide flexibility for tests to match paths precisely or using patterns (e.g., all paths in a specific directory).

2. **Registry-based implementation**: This allows multiple patterns to be registered and checked in sequence, supporting complex test scenarios with different behaviors for different paths.

3. **Compatibility with existing code**: The implementation will maintain compatibility with the existing `setupMockFs` function, ensuring tests can combine both global defaults and path-specific behaviors.

4. **Type safety**: Using the already defined interface `MockAccessFunction` ensures we implement a function with the correct parameter types and return values.

5. **Test maintenance**: This approach will make tests more maintainable by providing a clear, consistent API for mocking filesystem access behavior.