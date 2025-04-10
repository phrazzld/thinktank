# Implement FileLogger struct

## Goal
Create the FileLogger struct that satisfies the StructuredLogger interface, enabling structured logging to files in JSON line format. This will be the primary implementation of the StructuredLogger interface that handles writing audit log entries to a file.

## Implementation Approach
I'll implement the FileLogger struct in the `logger.go` file with the following components:

1. **FileLogger struct** with fields:
   - `file *os.File` - The file handle for writing logs
   - `mu sync.Mutex` - For thread-safe operations

2. **Constructor function**:
   - `NewFileLogger(filePath string) (*FileLogger, error)` - Creates a new FileLogger instance for the given file path
   - Basic file path validation and error handling
   - File opening with appropriate flags for append mode

3. **Skeleton implementation** of the StructuredLogger interface methods:
   - `Log(event AuditEvent)` - The basic logging functionality without thread safety or advanced error handling
   - `Close() error` - Basic file resource cleanup

This is a minimal implementation focusing on the core structure and functionality. The subsequent tasks ("Implement file path management in NewFileLogger", "Implement thread-safe Log method", and "Implement file closing in Close method") will enhance this implementation with more robust features.

## Reasoning
I chose this approach for the following reasons:

1. **Staged Implementation** - Creating a basic but functional implementation first allows us to add more sophisticated features incrementally in later tasks, making the code easier to test and debug.

2. **Separation of Concerns** - Breaking the implementation into distinct steps (basic structure, file path management, thread safety, and resource cleanup) aligns with good software engineering practices.

3. **Testing Strategy** - This approach makes it easier to write unit tests for the implementation because we can test each aspect separately.

4. **Alignment with Plan** - The approach aligns with the overall structured logging plan and follows the dependency structure in the TODO list.

5. **Flexibility** - This implementation provides a foundation that can be easily extended or modified as requirements evolve.