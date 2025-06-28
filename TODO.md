# Token Counting System Implementation TODO

## Phase 1: Core Token Counting Infrastructure

- [ ] **Create TokenCountingService struct** in `internal/thinktank/token_counting.go`
  - Define interface with methods: `CountTokensFromContext(ctx, instructions, files) (int, error)`
  - Add method `GetCompatibleModels(estimatedTokens int, availableProviders []string) (compatible, skipped []ModelWithReason)`
  - Include structured result types: `ModelWithReason{Name string, Reason string, ContextWindow int}`
  - Add audit logging for all token counting operations with correlation IDs

- [ ] **Implement accurate token counting algorithm** in TokenCountingService
  - Replace estimation with precise counting: instructions text + file content + formatting overhead
  - Add safety margin calculation (20% buffer for output tokens as per current code)
  - Implement character-to-token conversion using existing `EstimateTokensFromText()` for consistency
  - Add validation: return error if input is empty or context gathering fails

- [ ] **Add comprehensive logging to TokenCountingService**
  - Log total input character count, estimated tokens, and safety margin applied
  - Log each model evaluation: name, context window, compatibility decision, reason if skipped
  - Use structured logging with fields: `{"component": "token_counting", "total_tokens": X, "compatible_models": Y, "skipped_models": Z}`
  - Include correlation ID in all log entries for traceability

## Phase 2: Model Selection Enhancement

- [ ] **Update selectModelsForConfig() in internal/cli/main.go**
  - Replace `models.SelectModelsForInput()` call with new TokenCountingService integration
  - Add context gathering before model selection to get accurate file content for token counting
  - Create GatherConfig from SimplifiedConfig and call contextGatherer.GatherContext()
  - Pass actual token count (not estimate) to model selection logic

- [ ] **Implement model filtering with detailed logging**
  - Log start of model selection: `"Starting model selection with X estimated tokens from Y files"`
  - For each available model: log context window size and compatibility decision
  - For compatible models: `"Model {name} (context: {window}) - COMPATIBLE for {tokens} tokens"`
  - For skipped models: `"Model {name} (context: {window}) - SKIPPED: input too large ({tokens} > {window})"`
  - Log final selection: `"Selected {count} compatible models: {names}"`

- [ ] **Add error handling for token counting failures**
  - If context gathering fails during model selection, fall back to estimation-based selection
  - Log fallback: `"Token counting failed, using estimation: {error}"`
  - Ensure model selection never fails completely due to token counting issues
  - Add integration test covering the fallback scenario

## Phase 3: Orchestrator Logging Enhancement

- [ ] **Enhance processModelsWithErrorHandling() logging**
  - Add log entry at start: `"Processing {count} models with {tokens} total input tokens"`
  - Log each model processing attempt before calling processor.Process()
  - Format: `"Attempting model {name} ({index}/{total}) - context window: {window}, estimated processing time: {time}"`
  - Include model-specific rate limiting information in attempt logs

- [ ] **Add token context to model processing logs**
  - Modify processModelWithRateLimit() to log token efficiency metrics
  - Log: `"Model {name} processing: {input_tokens} input + {output_tokens} output = {total}/{context_window} ({percentage}% utilization)"`
  - Add output token estimation and actual usage tracking if available from LLM responses
  - Log processing completion with token utilization summary

- [ ] **Enhance structured logging output**
  - Add token counting fields to thinktank.log: `{"input_tokens": X, "selected_models": Y, "skipped_models": Z}`
  - Update audit.jsonl with token counting decisions: operation type "model_selection" with token details
  - Ensure all token-related logs include correlation ID for request tracing
  - Add log aggregation summary at end: `"Token counting summary: {input_tokens} total, {compatible_count} compatible models, {skipped_count} skipped"`

## Phase 4: Integration & Error Handling

- [ ] **Update CLI flow to use new token counting**
  - Modify main.go to create TokenCountingService instance with proper dependencies
  - Pass context, logger, and contextGatherer to TokenCountingService constructor
  - Update error handling to propagate token counting errors appropriately
  - Ensure dry-run mode shows token counting information without API calls

- [ ] **Add configuration for token counting behavior**
  - Add optional CLI flag `--token-safety-margin` (default 20%) for output buffer
  - Add environment variable `THINKTANK_TOKEN_SAFETY_MARGIN` support
  - Document token counting behavior in help text and configuration
  - Add validation: safety margin must be between 0% and 50%

- [ ] **Implement graceful degradation**
  - If TokenCountingService fails, fall back to existing estimation-based selection
  - Log degradation: `"Token counting service unavailable, using fallback estimation"`
  - Ensure all existing functionality continues to work if new service fails
  - Add circuit breaker pattern to prevent repeated token counting failures

## Phase 5: Testing & Validation

- [ ] **Add unit tests for TokenCountingService**
  - Test accurate token counting with various input sizes and file types
  - Test model compatibility decisions with different context window scenarios
  - Test logging output format and structured log field presence
  - Test error handling: empty input, context gathering failures, invalid models

- [ ] **Add integration tests for model selection flow**
  - Test end-to-end: CLI input → token counting → model selection → logging output
  - Verify models are correctly skipped and logged when input exceeds context window
  - Test with multiple providers and different model combinations
  - Validate correlation ID propagation through all log entries

- [ ] **Add regression tests for existing functionality**
  - Ensure existing model selection behavior unchanged when token counting succeeds
  - Verify dry-run mode continues to work with new token counting
  - Test synthesis flow still works with filtered model list
  - Validate all CLI flags and configuration options still function

- [ ] **Performance validation**
  - Benchmark token counting performance with large file sets (>100 files, >1MB total)
  - Ensure token counting doesn't significantly slow down CLI startup
  - Add timeout protection: token counting must complete within 30 seconds or fallback
  - Memory usage validation: token counting shouldn't increase memory footprint >20%

## Phase 6: Documentation & Cleanup

- [ ] **Update logging documentation**
  - Document new log fields and their meanings in docs/STRUCTURED_LOGGING.md
  - Add examples of token counting log output and interpretation
  - Update troubleshooting guide with token counting failure scenarios
  - Add section on token counting best practices and configuration

- [ ] **Code cleanup and optimization**
  - Remove any redundant token estimation code that's been replaced
  - Ensure consistent error message formatting across token counting features
  - Add comprehensive inline documentation for TokenCountingService public methods
  - Run golangci-lint and address any new warnings introduced

- [ ] **Validation checklist**
  - [ ] All TASK.md requirements implemented: ✓ input token counting, ✓ model filtering, ✓ attempt logging, ✓ skip logging
  - [ ] Logging integration: ✓ modern logging system, ✓ structured logs, ✓ audit file updates
  - [ ] Backward compatibility: ✓ existing CLI behavior preserved, ✓ all tests pass
  - [ ] Performance: ✓ no significant CLI slowdown, ✓ memory usage acceptable
  - [ ] Error handling: ✓ graceful fallbacks, ✓ clear error messages, ✓ no crashes

## Acceptance Criteria

**Must Have:**
1. Input token count accurately calculated from instructions + all processed files
2. Models with insufficient context window logged as skipped with clear reason
3. Each model processing attempt logged with token context information
4. All token counting decisions visible in structured logs (thinktank.log, audit.jsonl)
5. Zero regressions in existing functionality

**Should Have:**
1. Configurable safety margin for token counting
2. Performance impact <500ms additional CLI startup time
3. Graceful fallback to estimation if token counting fails
4. Comprehensive test coverage >90% for new code

**Could Have:**
1. Token utilization metrics in processing logs
2. Advanced token counting algorithms for different file types
3. Token counting cache for repeated operations
4. Integration with external token counting services
