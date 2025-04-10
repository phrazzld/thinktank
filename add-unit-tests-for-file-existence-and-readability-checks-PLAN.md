# Add unit tests for file existence and readability checks

## Task Goal
Create comprehensive unit tests to verify that the application correctly handles error conditions when working with task files, specifically when the file doesn't exist or is unreadable.

## Chosen Implementation Approach
We'll add unit tests that verify all the error handling cases in the `readTaskFromFile` function, which has been enhanced with specific error types:
- `ErrTaskFileNotFound` - When the file doesn't exist
- `ErrTaskFileReadPermission` - When the file exists but can't be read due to permissions
- `ErrTaskFileIsDir` - When the path points to a directory instead of a file
- `ErrTaskFileEmpty` - When the file exists but is empty or only contains whitespace

The unit tests will utilize temporary test files and mock file operations to simulate these error conditions in a controlled manner.

## Implementation Reasoning
This approach is the most suitable because:

1. It directly tests the specific error conditions defined in the codebase by the sentinel error variables.
2. It focuses on the `readTaskFromFile` function which is the core function for file operations related to task files.
3. It uses the standard Go testing tools and temporary files, making the tests reliable and repeatable.
4. It avoids relying on external resources or actual file system permissions, making the tests more portable and reliable.
5. By using Go's testing package features like subtests, we can keep the tests organized and focused on specific error conditions.

## Alternative Approaches Considered
1. **Integration Tests**: We could test these error conditions through integration tests that run the full application with various file scenarios. However, this would be more complex to set up and slower to run than targeted unit tests.

2. **Mocking the OS Package**: We could use advanced mocking to replace the entire `os` package functions. This would allow for complete control but would make the tests more complex and potentially fragile.

3. **Manual File Creation**: We could manually create files with the specific error conditions, but this would be less reliable and might not work consistently across different operating systems.