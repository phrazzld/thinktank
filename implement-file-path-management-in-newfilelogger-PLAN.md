# Implement file path management in NewFileLogger

## Goal
Enhance the `NewFileLogger` constructor to handle advanced file path management, including creating missing directories, validating the path, and ensuring proper file permissions. This will ensure that log files can be reliably created and written to in various environments.

## Implementation Approach
I'll enhance the `NewFileLogger` function in `logger.go` with the following improvements:

1. **Path Validation**: Add validation to check if the provided file path is empty or invalid
2. **Enhanced Directory Creation**: Improve the existing directory creation code to handle various permission scenarios
3. **Atomic File Creation**: Ensure that file opening is atomic when possible to avoid race conditions
4. **Error Wrapping**: Use proper error wrapping for improved diagnostics when things go wrong
5. **Permissions Management**: Set appropriate file permissions that work across different operating systems

The implementation will focus on making the function more robust while maintaining its current simple API. The main enhancement will be better error handling and directory creation with appropriate error messages that help diagnose issues.

## Reasoning
I chose this approach for the following reasons:

1. **Robustness** - The enhanced path management will handle edge cases better, such as permission issues, invalid paths, and concurrent file access.

2. **Diagnostics** - Better error handling with wrapped errors will make it easier to debug issues when they occur in production environments.

3. **Security** - Proper file permissions are important for ensuring that log files are both accessible for writing by the application and protected from unauthorized access.

4. **Compatibility** - The implementation will work across different operating systems while maintaining consistent behavior.

5. **Maintainability** - The code will be structured in a way that makes it clear what's happening at each step, making it easier to maintain and modify in the future.

6. **Performance** - By ensuring directories exist before attempting to open files, we avoid unnecessary error states and retries, improving performance in the common case.

Note that this implementation focuses on local file path management. The broader application-level path resolution (like handling relative paths with XDG standards) is addressed in a separate task called "Implement file path resolution for relative paths".