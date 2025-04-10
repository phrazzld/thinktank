# Add unit tests for file existence and readability checks

**Completed:** April 9, 2025

## Task Description
- **Action:** Create unit tests that verify error handling when the file doesn't exist or is unreadable.
- **Depends On:** Improve error messages for task file validation
- **AC Ref:** Error handling tests (Testing Strategy)

## Implementation Details

The implementation added comprehensive unit tests for the `readTaskFromFile` function, focusing on validating all error handling scenarios:

1. File not found (ErrTaskFileNotFound)
2. File path is a directory (ErrTaskFileIsDir)
3. File is empty (ErrTaskFileEmpty)
4. File contains only whitespace (ErrTaskFileEmpty)
5. File lacks read permissions (ErrTaskFileReadPermission)
6. Successful file reading
7. Relative path handling

### Approach
The tests use the Go testing package's subtest feature to organize tests by error condition. Each test creates test files or directories as needed, controls their permissions, and verifies that the correct error type is returned.

### Key Implementation Features
- Temporary test directories and files that are cleaned up after tests
- Proper error checking using Go's errors.Is() to ensure the correct sentinel errors are being returned
- Smart handling of platform-specific limitations (like file permission tests)
- Test skip logic for cases that might not be testable in all environments
- Comprehensive validation of both success and error paths

### Challenges
- Testing permission errors is challenging in a cross-platform way, so the implementation includes skip logic when permissions can't be changed
- Making sure tests clean up properly even when they fail

### Test Coverage
The tests verify that:
1. All error conditions in the readTaskFromFile function are properly detected
2. Each error condition returns the appropriate sentinel error
3. Relative paths are properly handled and converted to absolute paths
4. The function works correctly with valid files and returns the expected content