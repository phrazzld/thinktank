# Add logger initialization to main.go

## Goal
Implement code in main.go to initialize the appropriate structured audit logger (FileLogger or NoopLogger) based on the application configuration, ensuring proper integration with the existing logging system.

## Implementation Approach
I'll create a new function called `initAuditLogger` in main.go that will:

1. Check the `AuditLogEnabled` flag in the AppConfig to determine if audit logging is enabled.
2. If enabled:
   - Resolve the log file path (using the configured path or a default based on XDG directories)
   - Create and return a FileLogger instance
3. If disabled:
   - Create and return a NoopLogger instance
4. Add proper error handling for FileLogger creation failures
5. Update the main function to initialize the audit logger early in the application lifecycle
6. Ensure the audit logger is properly closed at application exit using `defer`

Additionally, I'll:
- Modify the `initConfigSystem` function to return both the config manager and app config
- Add structured audit logging for application startup in the main function

## Key Design Decisions

1. **Logger Initialization Timing**: The audit logger will be initialized after the regular logger but before configuration loading to ensure we can log the configuration process itself.

2. **Default Log Path Strategy**: If AuditLogFile is empty, we'll use a default path within the XDG cache directory: `~/.cache/architect/audit.log`.

3. **Error Handling Strategy**: If creating the FileLogger fails, we'll:
   - Log the error using the regular logger
   - Fall back to the NoopLogger to ensure the application can continue running
   - This balances reliability with observability

4. **Clean Resource Management**: The Close method will be called using a defer statement to ensure proper cleanup even if the application exits unexpectedly.

This approach ensures that the audit logging system is properly integrated into the application's lifecycle, while providing sensible defaults and fallbacks that align with the application's existing patterns and error handling strategy.