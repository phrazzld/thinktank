# Task: Define FileMeta Struct (`internal/fileutil/fileutil.go`)

## Goal
Define a new struct in the `internal/fileutil` package that can represent a file with its path and content. This struct will be used as part of the refactoring effort to simplify the application's interface by moving from a template-based approach to an instructions-based approach.

## Implementation Approach
Add a simple struct definition near the top of the `fileutil.go` file:

```go
// FileMeta represents a file with its path and content.
type FileMeta struct {
    Path    string
    Content string
}
```

## Key Reasoning
1. **Simplicity**: The struct is intentionally minimal, containing only the essential fields needed to represent a file for context gathering purposes.
2. **Separation of Concerns**: By using this struct, we can clearly separate the file content gathering logic from the formatting logic, which will be moved to a dedicated prompt stitching function.
3. **Alignment with Design**: This implementation directly matches the requirements specified in the TODO list and aligns with the overall refactoring plan to move away from templates.