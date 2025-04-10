# TODO

## Structured Logger Creation

- [x] **Task Title:** Create basic package structure for auditlog
  - **Action:** Create the new `internal/auditlog` directory and set up initial package files.
  - **Depends On:** None.
  - **AC Ref:** AC 1.1, AC 1.3

- [x] **Task Title:** Define AuditEvent and ErrorDetails structs
  - **Action:** Implement the AuditEvent struct with appropriate fields and JSON tags in `event.go`.
  - **Depends On:** Create basic package structure for auditlog
  - **AC Ref:** AC 1.1, AC 2.1, AC 2.2

- [x] **Task Title:** Implement helper function for creating events
  - **Action:** Create the NewAuditEvent function to simplify event creation with proper defaults.
  - **Depends On:** Define AuditEvent and ErrorDetails structs
  - **AC Ref:** AC 2.1

- [x] **Task Title:** Define StructuredLogger interface
  - **Action:** Create the interface with Log and Close methods in `logger.go`.
  - **Depends On:** Define AuditEvent and ErrorDetails structs
  - **AC Ref:** AC 1.1, AC 1.3

- [ ] **Task Title:** Implement FileLogger struct
  - **Action:** Create the FileLogger struct that satisfies the StructuredLogger interface.
  - **Depends On:** Define StructuredLogger interface
  - **AC Ref:** AC 1.1, AC 1.2, AC 2.1

- [ ] **Task Title:** Implement file path management in NewFileLogger
  - **Action:** Add directory creation and proper file opening in the constructor.
  - **Depends On:** Implement FileLogger struct
  - **AC Ref:** AC 1.2, AC 4.1

- [ ] **Task Title:** Implement thread-safe Log method
  - **Action:** Add mutex locking and proper error handling in the Log method.
  - **Depends On:** Implement FileLogger struct
  - **AC Ref:** AC 2.1, AC 4.1, AC 4.2

- [ ] **Task Title:** Implement file closing in Close method
  - **Action:** Ensure proper cleanup of file resources.
  - **Depends On:** Implement FileLogger struct
  - **AC Ref:** AC 4.1

- [ ] **Task Title:** Implement NoopLogger for disabled logging
  - **Action:** Create a no-operation logger implementation to use when logging is disabled.
  - **Depends On:** Define StructuredLogger interface
  - **AC Ref:** AC 3.2, AC 4.1

## Configuration System Integration

- [ ] **Task Title:** Add audit log configuration fields to AppConfig
  - **Action:** Update the AppConfig struct in `config.go` to include AuditLogEnabled and AuditLogFile fields.
  - **Depends On:** None
  - **AC Ref:** AC 3.1, AC 3.2

- [ ] **Task Title:** Update DefaultConfig with audit log defaults
  - **Action:** Add default values for audit logging configuration in the DefaultConfig function.
  - **Depends On:** Add audit log configuration fields to AppConfig
  - **AC Ref:** AC 3.1, AC 3.3

- [ ] **Task Title:** Update example_config.toml with audit log options
  - **Action:** Add documentation and examples for the audit log configuration options.
  - **Depends On:** Add audit log configuration fields to AppConfig
  - **AC Ref:** AC 3.1, AC 3.2

- [ ] **Task Title:** Update setViperDefaults in loader.go
  - **Action:** Ensure new configuration fields have proper defaults in the Viper configuration.
  - **Depends On:** Update DefaultConfig with audit log defaults
  - **AC Ref:** AC 3.1, AC 3.3

## Application Flow Integration

- [ ] **Task Title:** Add logger initialization to main.go
  - **Action:** Implement the code to initialize the appropriate logger based on configuration.
  - **Depends On:** Implement FileLogger struct, Implement NoopLogger for disabled logging, Add audit log configuration fields to AppConfig
  - **AC Ref:** AC 1.2, AC 3.2, AC 4.1

- [ ] **Task Title:** Implement file path resolution for relative paths
  - **Action:** Add logic to resolve relative paths to appropriate locations using XDG standards.
  - **Depends On:** Add logger initialization to main.go
  - **AC Ref:** AC 1.2, AC 3.3

- [ ] **Task Title:** Add application start/end logging
  - **Action:** Log application startup and shutdown events.
  - **Depends On:** Add logger initialization to main.go
  - **AC Ref:** AC 2.1, AC 5.1

- [ ] **Task Title:** Update key function signatures to accept structured logger
  - **Action:** Modify functions like clarifyTaskDescriptionWithPromptManager, generateAndSavePlanWithPromptManager, etc.
  - **Depends On:** Add logger initialization to main.go
  - **AC Ref:** AC 5.1

- [ ] **Task Title:** Add configuration loading logging
  - **Action:** Log configuration loading events and results.
  - **Depends On:** Update key function signatures to accept structured logger
  - **AC Ref:** AC 2.1, AC 5.1

- [ ] **Task Title:** Add task clarification process logging
  - **Action:** Add structured logging for the task clarification workflow.
  - **Depends On:** Update key function signatures to accept structured logger
  - **AC Ref:** AC 2.1, AC 5.1

- [ ] **Task Title:** Add context gathering logging
  - **Action:** Log details about the context gathering process, including file counts and statistics.
  - **Depends On:** Update key function signatures to accept structured logger
  - **AC Ref:** AC 2.1, AC 5.1

- [ ] **Task Title:** Add token counting operation logging
  - **Action:** Log token usage information.
  - **Depends On:** Update key function signatures to accept structured logger
  - **AC Ref:** AC 2.1, AC 5.1

- [ ] **Task Title:** Add API interaction logging
  - **Action:** Log Gemini API calls with input and output details.
  - **Depends On:** Update key function signatures to accept structured logger
  - **AC Ref:** AC 2.1, AC 5.1

- [ ] **Task Title:** Add error handling logging
  - **Action:** Enhance error logging to include structured details useful for debugging.
  - **Depends On:** Update key function signatures to accept structured logger
  - **AC Ref:** AC 2.1, AC 4.1, AC 5.1

- [ ] **Task Title:** Add plan generation and saving logging
  - **Action:** Log the plan generation process and results.
  - **Depends On:** Update key function signatures to accept structured logger
  - **AC Ref:** AC 2.1, AC 5.1

## Testing Implementation

- [ ] **Task Title:** Create unit tests for AuditEvent
  - **Action:** Test JSON serialization and helper functions for events.
  - **Depends On:** Define AuditEvent and ErrorDetails structs, Implement helper function for creating events
  - **AC Ref:** AC 6.1

- [ ] **Task Title:** Create unit tests for FileLogger creation
  - **Action:** Test logger initialization with various paths and permissions.
  - **Depends On:** Implement FileLogger struct, Implement file path management in NewFileLogger
  - **AC Ref:** AC 6.1

- [ ] **Task Title:** Create unit tests for JSON formatting
  - **Action:** Verify correct JSON line format is written to files.
  - **Depends On:** Implement thread-safe Log method
  - **AC Ref:** AC 1.1, AC 2.2, AC 6.1

- [ ] **Task Title:** Create unit tests for thread safety
  - **Action:** Test concurrent logging operations to verify thread safety.
  - **Depends On:** Implement thread-safe Log method
  - **AC Ref:** AC 4.2, AC 6.1

- [ ] **Task Title:** Create unit tests for error handling
  - **Action:** Test that the logger handles file and JSON errors gracefully.
  - **Depends On:** Implement thread-safe Log method
  - **AC Ref:** AC 4.1, AC 6.1

- [ ] **Task Title:** Create unit tests for NoopLogger
  - **Action:** Verify that the NoopLogger behaves as expected.
  - **Depends On:** Implement NoopLogger for disabled logging
  - **AC Ref:** AC 6.1

- [ ] **Task Title:** Update configuration loading tests
  - **Action:** Add tests for the new configuration fields and defaults.
  - **Depends On:** Add audit log configuration fields to AppConfig, Update DefaultConfig with audit log defaults, Update setViperDefaults in loader.go
  - **AC Ref:** AC 3.1, AC 3.2, AC 3.3, AC 6.1

- [ ] **Task Title:** Create integration tests for logging in application flow
  - **Action:** Test that logs are written correctly during application execution.
  - **Depends On:** Add application start/end logging, Add configuration loading logging, Add API interaction logging, Add error handling logging
  - **AC Ref:** AC 5.1, AC 6.1

- [ ] **Task Title:** Create integration tests for configuration options
  - **Action:** Test behavior with different configuration settings (enabled/disabled, different paths).
  - **Depends On:** Add logger initialization to main.go, Implement file path resolution for relative paths
  - **AC Ref:** AC 3.2, AC 3.3, AC 6.1

- [ ] **Task Title:** Create integration tests for error conditions
  - **Action:** Test that the application continues to function when logging fails.
  - **Depends On:** Add logger initialization to main.go, Add error handling logging
  - **AC Ref:** AC 4.1, AC 6.1

## Documentation

- [ ] **Task Title:** Update code comments for auditlog package
  - **Action:** Ensure comprehensive godoc-style documentation for all types and functions.
  - **Depends On:** Create basic package structure for auditlog, Define StructuredLogger interface, Implement FileLogger struct, Implement NoopLogger for disabled logging
  - **AC Ref:** AC 2.2

- [ ] **Task Title:** Document configuration options
  - **Action:** Update project documentation with information about the new logging capabilities.
  - **Depends On:** Add audit log configuration fields to AppConfig, Update example_config.toml with audit log options
  - **AC Ref:** AC 3.1, AC 3.2

## [!] CLARIFICATIONS NEEDED / ASSUMPTIONS

- [ ] **Issue/Assumption:** Default log file path resolution strategy
  - **Context:** The plan mentions using XDG Cache dir for default locations, but doesn't specify the exact hierarchy or fallback strategy when XDG paths are unavailable.

- [ ] **Issue/Assumption:** Extent of logging in existing functions
  - **Context:** The plan identifies key logging points but doesn't exhaustively list all functions that should be updated to include logging. Assumption is that we'll focus on the main operational points first.

- [ ] **Issue/Assumption:** Error handling strategy for logging failures
  - **Context:** The plan mentions handling errors gracefully but doesn't provide specific guidance on how errors should be reported when logging itself fails.

- [ ] **Issue/Assumption:** Performance impact thresholds
  - **Context:** The plan mentions minimal performance impact but doesn't define specific acceptable thresholds for overhead introduced by logging.

- [ ] **Issue/Assumption:** Secure handling of sensitive data in logs
  - **Context:** The plan doesn't explicitly address how to handle potentially sensitive information (like API keys) in logs. Assumption is that we should not log such information.

- [ ] **Issue/Assumption:** Backward compatibility requirements
  - **Context:** The plan doesn't specify if API changes (function signatures) need special handling for backward compatibility. Assumption is that since this is an internal change, we can modify signatures directly.