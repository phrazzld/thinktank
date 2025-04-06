# Implement filesystem readFile mock helpers

## Task Goal
Add helper functions to mock `fs.readFile` for specific paths with custom content or errors in the `mockFsUtils.ts` file. This will allow tests to specify which file paths should return specific content or throw specific errors when read.

## Implementation Approach
I'll implement a new function called `mockReadFile` following a similar pattern to the recently implemented `mockAccess` function. This approach will:

1. Store path-specific rules in a registry (array) that maps file paths or patterns to content or errors
2. Modify the existing `fs.readFile` mock implementation to check for matching patterns
3. Return the configured content or error for matched paths
4. Fall back to default behavior for paths that don't match any specific rule

The implementation will support both exact path matching and regular expression pattern matching. This allows tests to:
- Configure specific files to return exact content
- Configure whole directories of files (via regex) to return specific content or errors
- Override the default behavior for specific paths

## Key Reasoning

1. **Consistency with existing patterns**: Following the same pattern established by `mockAccess` ensures a consistent API for all mock helpers, making the utilities more intuitive for developers.

2. **Flexibility**: Supporting both exact path matches and regex patterns provides maximum flexibility for test scenarios, allowing both precise and broad mocking.

3. **Priority-based rules**: Adding new rules to the beginning of the registry ensures that more recently added rules take precedence, allowing tests to override previously configured behavior.

4. **Type safety**: Using the predefined `MockReadFileFunction` interface ensures type safety and maintains consistency with the API design.

5. **Compatibility**: The implementation will integrate seamlessly with the existing setup mechanism, allowing tests to combine default behaviors with path-specific overrides.

6. **Error handling**: Supporting both successful file reads and simulated errors provides comprehensive testing capabilities for both happy path and error scenarios.