# Define StructuredLogger interface

## Goal
Create the StructuredLogger interface in the auditlog package with methods for logging events and cleaning up resources. This interface will establish a contract that all logger implementations must follow.

## Implementation Approach
I'll implement the StructuredLogger interface in the `logger.go` file with two primary methods:

1. `Log(event AuditEvent)` - For writing structured log entries to the destination
2. `Close() error` - For cleaning up resources when the logger is no longer needed

The interface will be kept minimal to make it easy to implement and maintain. This is important because we'll need multiple implementations (FileLogger and NoopLogger) that satisfy this interface.

I'll also add detailed documentation explaining the responsibilities and expected behavior of each method to ensure consistent implementation across concrete loggers.

## Reasoning
I selected this approach for the following reasons:

1. **Minimalist Design** - By keeping the interface small (just two methods), we ensure it's easy to implement and maintain, which follows Go's philosophy of interface design.

2. **Separation of Concerns** - The interface focuses solely on logging responsibility, separating it from specific implementation details like file handling or JSON formatting.

3. **Dependency Inversion** - Defining the interface first allows the application to depend on abstractions rather than concrete implementations, making it easier to test and more flexible in the future.

4. **Future Extensibility** - If needed, we can add more methods to the interface in the future without breaking existing code, as long as we maintain the existing method signatures.

5. **Consistency with Plan** - This approach aligns with the overall structured logging plan that specifies a common interface for different logger implementations.