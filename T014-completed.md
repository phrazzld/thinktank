# Ticket T014: Update thinktank/registry for context, logging, and LLMError

## Summary of Changes
In this ticket, we've updated the thinktank/registry package to properly use context propagation, context-aware logging, and structured error handling using LLMError. This is part of the "Error Handling and Logging Consistency" epic.

## Specific Changes
1. Updated all method signatures in registry.go to accept a context.Context parameter:
   - GetModel
   - GetProvider
   - GetProviderImplementation

2. Updated all logging calls to use context-aware versions:
   - Replaced Info/Debug/Warn with InfoContext/DebugContext/WarnContext
   - Ensured all logs propagate the correlation ID from the context

3. Replaced error handling with structured LLMError:
   - Used llm.Wrap to wrap errors with proper categorization
   - Added error types for common failure cases (e.g., ErrProviderNotFound)
   - Enhanced error messages with more detailed information for troubleshooting

4. Updated tests to work with the new method signatures:
   - registry_test.go: Added context parameters to all test calls
   - registry_secret_test.go: Added context parameters to all test calls
   - BoundaryMockRegistry: Updated method signatures to include context
   - MockRegistryAPI: Updated method signatures to include context
   - Various test mocks: Updated to return LLMError-wrapped errors

5. Fix for the thinktank package tests:
   - Updated registry_api.go to wrap provider errors as ErrClientInitialization
   - Fixed token limit tests in registry_api_token_limits_test.go
   - Fixed validation tests in registry_api_validation_test.go

## Testing Notes
All tests in the registry and thinktank packages now pass with the updated interface.

Integration tests will need to be updated in a separate ticket to match the new method signatures, but this was out of scope for the current ticket.

## Related Tickets
- Completed: T013 - Add unit tests for provider error translation
- Next: Update integration tests to work with the new context-based registry interface
