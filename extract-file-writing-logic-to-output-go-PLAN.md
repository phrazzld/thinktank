# Extract File Writing Logic to output.go Implementation Plan

## Task Description
Extract file writing logic from main.go to cmd/architect/output.go, following the pure refactoring approach where functionality is moved without changing behavior.

## Current Status
- The `saveToFile` function in main.go handles writing content to a file
- A skeleton implementation of output.go exists with stub methods
- The `OutputWriter` interface has been defined but not implemented

## Implementation Approach

### 1. Move `saveToFile` Function to `output.go`
- Migrate the `saveToFile` function from main.go to the `SaveToFile` method in cmd/architect/output.go
- Ensure error handling and logging remain consistent
- Update any imports needed for filepath operations

### 2. Update Main.go References
- Update references to `saveToFile` in main.go to use the new `OutputWriter` interface
- Add "Transitional implementation" comment to the existing `saveToFile` function

### 3. Implementation Details
- The `SaveToFile` method should handle:
  - Converting relative paths to absolute paths
  - Writing content to the file with proper permissions (0644)
  - Logging operations and errors
  - Returning errors rather than calling `Fatal` directly

### 4. Testing Strategy
- Ensure the implementation matches the original behavior
- Verify file writing operations work correctly
- No functional changes should be introduced

## Success Criteria
- The `SaveToFile` method in output.go should be fully implemented
- The original main.go should contain a transitional implementation comment
- Functionality should remain exactly the same as before

## Dependencies
- This task depends on the skeleton cmd/architect files already being created (completed)
- No additional dependencies needed for this specific task