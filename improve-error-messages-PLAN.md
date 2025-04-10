# Implementation Plan: Improve error messages for task file validation

## Task
Enhance error handling for file existence, readability, and content validation with clear error messages.

## Chosen Approach
I'll implement **Approach 1: Enhance readTaskFromFile to Return Specific Error Types/Values** as recommended in the analysis.

This approach focuses on making `readTaskFromFile` return clearly identifiable error types that `validateInputs` can specifically handle and provide tailored error messages to the user.

### Implementation Steps:
1. Define specific error sentinel values for different failure scenarios
2. Modify `readTaskFromFile` to check for specific conditions and return the corresponding sentinel errors
3. Update `validateInputs` to check for these specific errors and provide clearer, more actionable error messages
4. Add content validation to check for empty task files

### Code Changes:

**1. Define Specific Error Values:**
```go
// Define sentinel errors for task file validation
var (
    ErrTaskFileNotFound      = errors.New("task file not found")
    ErrTaskFileReadPermission = errors.New("task file permission denied")
    ErrTaskFileIsDir          = errors.New("task file path is a directory")
    ErrTaskFileEmpty          = errors.New("task file is empty")
)
```

**2. Enhance `readTaskFromFile` Function:**
```go
func readTaskFromFile(taskFilePath string, logger logutil.LoggerInterface) (string, error) {
    // Check if path is absolute, if not make it absolute
    if !filepath.IsAbs(taskFilePath) {
        cwd, err := os.Getwd()
        if err != nil {
            return "", fmt.Errorf("error getting current working directory: %w", err)
        }
        taskFilePath = filepath.Join(cwd, taskFilePath)
    }

    // Enhanced file existence check with specific errors
    fileInfo, err := os.Stat(taskFilePath)
    if err != nil {
        if os.IsNotExist(err) {
            return "", fmt.Errorf("%w: %s", ErrTaskFileNotFound, taskFilePath)
        }
        if os.IsPermission(err) {
            return "", fmt.Errorf("%w: %s", ErrTaskFileReadPermission, taskFilePath)
        }
        // Generic stat error
        return "", fmt.Errorf("error checking task file status: %w", err)
    }

    // Check if it's a directory
    if fileInfo.IsDir() {
        return "", fmt.Errorf("%w: %s", ErrTaskFileIsDir, taskFilePath)
    }

    // Read file content
    content, err := os.ReadFile(taskFilePath)
    if err != nil {
        if os.IsPermission(err) {
            return "", fmt.Errorf("%w: %s", ErrTaskFileReadPermission, taskFilePath)
        }
        // Generic read error
        return "", fmt.Errorf("error reading task file content: %w", err)
    }

    // Check for empty content
    if len(strings.TrimSpace(string(content))) == 0 {
        return "", fmt.Errorf("%w: %s", ErrTaskFileEmpty, taskFilePath)
    }

    // Return content as string
    return string(content), nil
}
```

**3. Update `validateInputs` Function:**
```go
// In validateInputs function
if config.TaskFile != "" {
    taskContent, err := readTaskFromFile(config.TaskFile, logger)
    if err != nil {
        // Specific error handling
        switch {
        case errors.Is(err, ErrTaskFileNotFound):
            logger.Error("Task file not found. Please check the path: %s", config.TaskFile)
        case errors.Is(err, ErrTaskFileReadPermission):
            logger.Error("Cannot read task file due to permissions. Please check permissions for: %s", config.TaskFile)
        case errors.Is(err, ErrTaskFileIsDir):
            logger.Error("The specified task file path is a directory, not a file: %s", config.TaskFile)
        case errors.Is(err, ErrTaskFileEmpty):
            logger.Error("The task file is empty or contains only whitespace: %s", config.TaskFile)
        default:
            // Generic fallback with more specific underlying error
            logger.Error("Failed to load task file '%s': %v", config.TaskFile, err)
        }
        flag.Usage()
        os.Exit(1)
    }
    
    // Rest of existing code...
}
```

## Reasoning for Choice

I selected Approach 1 for the following reasons:

1. **Balanced Approach**: It enhances error handling significantly without requiring major structural changes to the codebase.

2. **Clear Separation of Concerns**: It maintains a clean separation where `readTaskFromFile` handles file operations and identifies error types, while `validateInputs` handles user-facing messages.

3. **Follows Go Best Practices**: Using sentinel errors and `errors.Is()` for error type checking aligns with standard Go error handling practices.

4. **Testability**: Easier to test both components separately - `readTaskFromFile` can be tested for returning the correct error types, and `validateInputs` can be tested for handling those errors appropriately.

5. **Maintainability**: Adding new validation checks is straightforward - define a new sentinel error, check for the condition in `readTaskFromFile`, and add case handling in `validateInputs`.

6. **User Experience**: Provides clearer, more actionable error messages that will help users identify and resolve issues more quickly.

7. **Improves Validation**: Adds checks for common issues like empty files and directories, which weren't explicitly handled before.

This approach also adds a new check for empty files, which will prevent confusion where a task file exists but doesn't contain any usable content.