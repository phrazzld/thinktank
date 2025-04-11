# Task: Refactor `GatherProjectContext` (`internal/fileutil/fileutil.go`)

**Completed:** April 10, 2025

## Summary
Refactored the `GatherProjectContext` function to use the FileMeta struct instead of directly formatting content strings. This separates content gathering from formatting, which will now be handled by the prompt stitching logic.

## Implementation Details
- Changed the signature of `GatherProjectContext` to return a slice of FileMeta:
  ```go
  func GatherProjectContext(paths []string, config *Config) ([]FileMeta, int, error)
  ```
- Removed the string formatting and XML wrapping in the function
- Refactored `processFile` to add FileMeta instances to a slice instead of writing formatted strings to a builder
- Updated tests to verify the returned FileMeta slice contains the expected files and content

## Next Steps
The next task is to update fileutil tests to ensure they properly test the new FileMeta-based implementation.