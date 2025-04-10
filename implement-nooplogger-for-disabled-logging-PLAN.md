# Implement NoopLogger for disabled logging

## Goal
Create a no-operation (noop) implementation of the StructuredLogger interface that performs no actions when used. This implementation will be used when audit logging is disabled in the application configuration, allowing the application code to use the structured logger without checking if logging is enabled.

## Implementation Approach
I'll implement the NoopLogger in the `logger.go` file with the following components:

1. **NoopLogger struct** - An empty struct with no fields since it doesn't need to maintain any state
2. **Log method** - An implementation that does nothing when called
3. **Close method** - An implementation that does nothing and returns nil
4. **NewNoopLogger function** - A constructor that returns a new NoopLogger instance

This implementation will satisfy the StructuredLogger interface but won't perform any actual logging actions, making it perfect for scenarios when audit logging is disabled in the configuration.

## Reasoning
I chose this approach for the following reasons:

1. **Null Object Pattern** - The NoopLogger follows the Null Object design pattern, which provides a do-nothing implementation of an interface. This allows the application code to use a logger without checking if it's enabled, simplifying the codebase and eliminating null/nil checks.

2. **Performance** - Since the implementation does nothing, it has minimal performance impact when logging is disabled, which is important for maintaining application speed when audit logging isn't needed.

3. **Simplicity** - The NoopLogger is extremely simple to implement and test, requiring minimal code while still providing the necessary functionality.

4. **Consistency with Plan** - This approach aligns with the structured logging plan which specifically mentions a NoopLogger for disabled logging.

5. **Maintainability** - If the StructuredLogger interface changes in the future, the NoopLogger implementation can be easily updated since it's a minimal implementation.