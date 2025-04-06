# Implement filesystem writeFile mock helpers

## Task Goal
Add helper functions to mock `fs.writeFile` for specific paths with success or error responses in the `mockFsUtils.ts` file. This will allow tests to specify which file write operations should succeed or fail with specific errors.

## Implementation Approach
I'll implement a new function called `mockWriteFile` following the consistent pattern established by the previously implemented mock helper functions (`mockAccess`, `mockReadFile`, `mockStat`, `mockReaddir`, and `mockMkdir`). This approach will:

1. Store path-specific rules in a registry (array) that maps file paths or patterns to success or error results
2. Modify the existing `fs.writeFile` mock implementation to check for matching patterns
3. Return success or the configured error for matched paths
4. Fall back to default behavior for paths that don't match any specific rule

The implementation will support both exact path matching and regular expression pattern matching, allowing tests to:
- Configure specific file paths to succeed or fail on write operations
- Configure patterns of files (via regex) to consistently succeed or fail
- Specify custom error types (e.g., EACCES for permission denied, ENOSPC for disk full)
- Verify file contents being written (optional future enhancement)

## Key Reasoning

1. **Consistency with existing patterns**: Following the same pattern established by previous mock helpers ensures a consistent API, making the utilities more intuitive for developers.

2. **File write scenarios**: By allowing configuration of both success and failure cases, tests can simulate various scenarios like permission errors, disk full errors, or "device read-only" cases.

3. **Pattern-based matching**: Supporting regex patterns allows tests to mock behavior for entire file hierarchies without having to specify each path individually.

4. **Detailed error simulation**: Supporting specific error codes allows tests to verify that code handles different error conditions appropriately (e.g., retrying on transient errors but not on permission errors).

5. **Default fallback**: Using a default configuration (success) with the ability to override it on a path-by-path basis provides flexibility for various test scenarios.

6. **Type safety**: Using the predefined `MockWriteFileFunction` interface ensures type safety and maintains consistency with the API design.

7. **Priority-based rules**: Adding new rules to the beginning of the registry ensures that more recently added rules take precedence, allowing tests to override previously configured behavior.