# Implement filesystem readdir mock helpers

## Task Goal
Add helper functions to mock `fs.readdir` for specific directories with custom file lists or errors in the `mockFsUtils.ts` file. This will allow tests to specify which directory paths should return specific lists of files/subdirectories or throw specific errors when queried.

## Implementation Approach
I'll implement a new function called `mockReaddir` following the consistent pattern established by the recently implemented `mockAccess`, `mockReadFile`, and `mockStat` functions. This approach will:

1. Store path-specific rules in a registry (array) that maps directory paths or patterns to file lists or errors
2. Modify the existing `fs.readdir` mock implementation to check for matching patterns
3. Return the configured file list or error for matched paths
4. Fall back to default behavior for paths that don't match any specific rule

The implementation will support both exact path matching and regular expression pattern matching, allowing tests to:
- Configure specific directories to return exact lists of files and subdirectories
- Configure whole directory trees (via regex) to return specific file lists or throw errors
- Support both string arrays and Dirent-like objects for more complex directory entry information

## Key Reasoning

1. **Consistency with existing patterns**: Following the same pattern established by previous mock helpers ensures a consistent API, making the utilities more intuitive for developers.

2. **Directory content configuration**: By allowing arrays of file/directory names to be specified, tests can easily simulate complex directory structures.

3. **Pattern-based directory matching**: Supporting regex patterns allows tests to mock entire directory trees with a single rule (e.g., all subdirectories of a certain path).

4. **Error simulation**: Supporting error responses allows tests to verify error handling code paths, such as handling permission errors or non-existent directories.

5. **Default fallback**: Using a default configuration (empty directory) with the ability to override it on a path-by-path basis provides flexibility for various test scenarios.

6. **Type safety**: Using the predefined `MockReaddirFunction` interface ensures type safety and maintains consistency with the API design.

7. **Priority-based rules**: Adding new rules to the beginning of the registry ensures that more recently added rules take precedence, allowing tests to override previously configured behavior.