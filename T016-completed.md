# Ticket T016: Update thinktank/orchestrator for context, logging, and error aggregation

## Summary of Changes
In this ticket, we've updated the thinktank/orchestrator package to properly use context propagation, context-aware logging, and structured error handling using LLMError. This is part of the "Error Handling and Logging Consistency" epic and follows the updates to the registry and modelproc packages.

## Specific Changes

1. **Added error categorization helpers**:
   - Created `CategorizeOrchestratorError` function to map orchestrator errors to LLM error categories
   - Implemented `WrapOrchestratorError` helper for consistent error wrapping
   - Ensured proper error categorization based on error type and context

2. **Updated logging to use context-aware versions**:
   - Replaced all `Info`/`Debug`/`Error` calls with `InfoContext`/`DebugContext`/`ErrorContext`
   - Updated `buildPrompt` method to accept context parameter
   - Ensured all logs propagate correlation ID from context

3. **Enhanced error handling with structured LLMError**:
   - Replaced standard error wrapping with `llm.Wrap` to provide proper categorization
   - Used appropriate error categories for different types of failures:
     - `CategoryRateLimit` for rate limiting issues
     - `CategoryContentFiltered` for safety-blocked content
     - `CategoryAuth` for authentication failures
     - `CategoryNetwork` for connectivity issues
     - `CategoryServer` for file system errors
     - `CategoryCancelled` for cancelled operations
   - Improved error aggregation in `handleProcessingOutcome` to prioritize errors by severity

4. **Updated method signatures**:
   - Added context parameter to `buildPrompt` method
   - Updated error handling in `processModelWithRateLimit`
   - Enhanced error aggregation in synthesis service

5. **Updated supporting components**:
   - Updated synthesis_service.go with proper error categorization
   - Updated output_writer.go to use WrapOrchestratorError

## Testing Notes
All tests in the orchestrator package have been verified to pass with the updated interfaces. The structured error handling in the orchestrator package now provides proper categorization of errors for better decision making in calling code.

## Related Tickets
- Completed: T014 - Update thinktank/registry for context, logging, and LLMError
- Completed: T015 - Update thinktank/modelproc for context, logging, and LLMError
- Next: T017 - Update auditlog for LoggerInterface and correlation ID
