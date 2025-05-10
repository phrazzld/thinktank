# Ticket T015: Update thinktank/modelproc for context, logging, and LLMError

## Summary of Changes
In this ticket, we've updated the thinktank/modelproc package to properly use context propagation, context-aware logging, and structured error handling using LLMError. This is part of the "Error Handling and Logging Consistency" epic.

## Specific Changes
1. **Updated logging to use context-aware versions:**
   - Replaced all `Info`/`Debug`/`Error` calls with `InfoContext`/`DebugContext`/`ErrorContext` in the Process method
   - Added context parameter to the saveOutputToFile method and updated its logging
   - Ensured all logs propagate the correlation ID from the context

2. **Enhanced error handling with structured LLMError:**
   - Replaced standard error wrapping with `llm.Wrap` to provide proper categorization
   - Used appropriate error categories for different types of failures:
     - `CategoryInvalidRequest` for most client-side errors
     - `CategoryContentFiltered` for safety-blocked content
     - `CategoryRateLimit` for rate limiting issues
     - `CategoryInputLimit` for token limit issues
     - `CategoryServer` for file system errors
   - Maintained error tracing with detailed error messages

3. **Improved error message consistency:**
   - Added more detailed error messages with context about the operation
   - Included model name in error messages
   - Propagated original errors for debugging

4. **Updated method signatures:**
   - Added context parameter to saveOutputToFile method
   - Ensured proper context propagation through the call chain

## Testing Notes
All tests in the modelproc and thinktank packages have been verified to pass with the updated interfaces. The tests correctly verify error propagation and categorization.

The integration tests have also been updated to work with the new context-aware interfaces:
- Updated boundary_test_adapter.go with context parameters for GetModelParameters, ValidateModelParameter, GetModelDefinition, and GetModelTokenLimits
- Updated integration_test_mocks.go to support the new context-aware interfaces
- Fixed invalid_synthesis_model_test.go and synthesis_with_failures_test.go to use the updated mocks
- Updated multi_provider_test.go to include context parameters in calls to registry.LoadConfig and registry.RegisterProviderImplementation

## Related Tickets
- Completed: T014 - Update thinktank/registry for context, logging, and LLMError
- Next: T016 - Update thinktank/orchestrator for context, logging, and error aggregation
