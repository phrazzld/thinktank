# Implement filesystem stat mock helpers

## Task Goal
Add helper functions to mock `fs.stat` for specific paths with custom file/directory stats or errors in the `mockFsUtils.ts` file. This will allow tests to specify which file paths should return specific stats (like file/directory type, size, timestamps) or throw specific errors when queried.

## Implementation Approach
I'll implement a new function called `mockStat` following the consistent pattern established by the recently implemented `mockAccess` and `mockReadFile` functions. This approach will:

1. Store path-specific rules in a registry (array) that maps file paths or patterns to custom stats or errors
2. Modify the existing `fs.stat` mock implementation to check for matching patterns
3. Return the configured stats or error for matched paths
4. Fall back to default behavior for paths that don't match any specific rule

The implementation will support both exact path matching and regular expression pattern matching, allowing tests to:
- Configure specific files to return exact stats (like file size, type, and timestamps)
- Configure whole directories of files (via regex) to have specific stats or throw errors
- Easily configure commonly needed stats combinations (e.g., "is a file", "is a directory", "doesn't exist")

## Key Reasoning

1. **Consistency with existing patterns**: Following the same pattern established by `mockAccess` and `mockReadFile` ensures a consistent API for all mock helpers, making the utilities more intuitive for developers.

2. **Flexible stats configuration**: Stats objects in Node.js have many properties and methods. By using the existing `createStats` utility function, we provide an easy way to create realistic stats objects from partial specifications.

3. **Common use cases**: Most tests only need to specify whether a path is a file, directory, or doesn't exist. The implementation will make these common cases easy while still supporting more complex scenarios.

4. **Type safety**: Using the predefined `MockStatFunction` interface ensures type safety and maintains consistency with the API design.

5. **Error scenarios**: File system operations commonly fail when files don't exist or aren't accessible, so supporting error scenarios is important for comprehensive testing.

6. **Priority-based rules**: Adding new rules to the beginning of the registry ensures that more recently added rules take precedence, allowing tests to override previously configured behavior.