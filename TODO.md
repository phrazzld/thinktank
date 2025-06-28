# Accurate Token Counting System Implementation TODO

## Executive Summary

Implementing **provider-specific accurate tokenizers** to replace 0.75 tokens/character estimation that can be 25-100% wrong for non-English text and code. This will provide 90%+ accuracy improvement for intelligent model selection decisions.

**High-Impact Priority**: OpenAI (tiktoken) → Gemini (SentencePiece) → OpenRouter (estimation/GPT-4o)

---

## Phase 1: OpenAI Accurate Tokenization (HIGH IMPACT) ✅ COMPLETED

### 1.1 Add Tiktoken Dependency & Core Implementation ✅ COMPLETED

- [x] **Add tiktoken-go dependency** ✅ COMPLETED
  - ✅ Added `github.com/pkoukk/tiktoken-go` to go.mod with caching mechanism
  - ✅ Research completed: pkoukk chosen for chat message counting + caching support
  - ✅ Dependency added and integrated into tokenizers package

- [x] **Create OpenAI tokenizer interface** ✅ COMPLETED
  - ✅ Defined `TokenizerManager` and `AccurateTokenCounter` interfaces in `internal/thinktank/tokenizers/`
  - ✅ Implemented lazy loading to avoid 4MB vocabulary initialization at startup
  - ✅ Support for model-specific encodings: `cl100k_base` (GPT-4), `o200k_base` (GPT-4o)
  - ✅ Added comprehensive error handling for unsupported model encodings

- [x] **Implement AccurateTokenCounter interface** ✅ COMPLETED
  ```go
  type AccurateTokenCounter interface {
      CountTokens(ctx context.Context, text string, modelName string) (int, error)
      SupportsModel(modelName string) bool
      GetEncoding(modelName string) (string, error)
  }
  ```
  - ✅ Fully implemented with OpenAI tiktoken integration
  - ✅ Added TokenizerManager for provider-aware tokenizer selection

### 1.2 Integration with TokenCountingService ✅ COMPLETED

- [x] **Update TokenCountingService for provider-aware counting** ✅ COMPLETED
  - ✅ Modified `countInstructionTokensAccurate()` to use tiktoken for OpenAI models (gpt-4.1, o4-mini, o3)
  - ✅ Modified `countFileTokensAccurate()` to use tiktoken for OpenAI models
  - ✅ Implemented estimation fallback for unsupported providers
  - ✅ Added provider detection logic based on model name using `models.GetModelInfo()`

- [x] **Add comprehensive tiktoken testing** ✅ COMPLETED
  - ✅ Added accuracy tests comparing tiktoken vs estimation with structured test cases
  - ✅ Performance testing with large inputs and memory usage validation
  - ✅ Comprehensive benchmark tests: `BenchmarkTiktokenVsEstimation`, stress tests with >100 files
  - ✅ Table-driven tests covering multiple content types and edge cases
  - ✅ Integration tests validating end-to-end token counting flow

### 1.3 Expected Accuracy Improvements

- **English code/text**: 95%+ accuracy (vs ~75% current estimation)
- **Non-English text**: 90%+ accuracy (vs 25-50% current estimation)
- **Mixed content**: 85%+ accuracy (vs highly variable current)

---

## Phase 2: Gemini Accurate Tokenization (MEDIUM IMPACT)

### 2.1 Add SentencePiece Dependency

- [ ] **Add go-sentencepiece dependency**
  - Add `github.com/eliben/go-sentencepiece` to go.mod
  - Research: Specifically designed for Gemini/Gemma models, extensively tested vs official SentencePiece
  - Handles BPE tokenization used by Gemini models (same tokenizer as proprietary Gemini family)

- [ ] **Create Gemini tokenizer implementation**
  - Implement `GeminiTokenizer` in `internal/thinktank/tokenizers/`
  - Support both gemini-2.5-pro and gemini-2.5-flash (both 1M context models)
  - Add lazy loading and caching for tokenizer model files
  - Handle SentencePiece protobuf configuration files

### 2.2 Integration & Testing

- [ ] **Update TokenCountingService for Gemini**
  - Add Gemini tokenizer to provider-aware counting logic
  - Test with Gemini-specific content and compare against Google's token counting API
  - Validate that "1 token ≈ 4 characters" rule breaks down correctly for non-English

---

## Phase 3: Provider-Aware Architecture (MEDIUM IMPACT)

### 3.1 Unified Tokenizer Architecture

- [ ] **Create ProviderTokenCounter struct**
  ```go
  type ProviderTokenCounter struct {
      tiktoken      *OpenAITokenizer     // For OpenAI models
      sentencePiece *GeminiTokenizer     // For Gemini models
      fallback      *EstimationCounter   // For unsupported models
      logger        logutil.LoggerInterface
  }
  ```

- [ ] **Implement provider detection logic**
  - Use existing `models.GetModelInfo()` to get provider for model name
  - Route to appropriate tokenizer based on provider: openai → tiktoken, gemini → SentencePiece, openrouter → estimation
  - Add comprehensive logging for tokenizer selection decisions

### 3.2 Safety Margins & Validation

- [ ] **Add configurable safety margins**
  - Add CLI flag `--token-safety-margin` (default 20% for output buffer)
  - Add environment variable `THINKTANK_TOKEN_SAFETY_MARGIN` support
  - Validation: safety margin must be between 0% and 50%
  - Apply safety margin to context window calculations for model filtering

- [ ] **Implement robust input validation**
  - Return clear errors for empty input or context gathering failures
  - Add timeout protection: token counting must complete within 30 seconds or fallback
  - Validate model name exists in model definitions before tokenization

---

## Phase 4: Model Filtering & Selection Enhancement (HIGH IMPACT) ✅ COMPLETED

### 4.1 Accurate Model Filtering ✅ COMPLETED

- [x] **Add GetCompatibleModels method to TokenCountingService** ✅ COMPLETED
  ```go
  GetCompatibleModels(ctx context.Context, req TokenCountingRequest, availableProviders []string) ([]ModelCompatibility, error)

  type ModelCompatibility struct {
      ModelName     string
      IsCompatible  bool
      TokenCount    int
      ContextWindow int
      UsableContext int
      Provider      string
      TokenizerUsed string // "tiktoken", "sentencepiece", "estimation"
      IsAccurate    bool
      Reason        string // Detailed reason for incompatibility
  }
  ```
  - ✅ Fully implemented with comprehensive model evaluation logic
  - ✅ Includes safety margin calculations (20% for output buffer)
  - ✅ Sorts results with compatible models first, then by context window size

- [x] **Replace estimation-based model selection** ✅ COMPLETED
  - ✅ TokenCountingService integrated with accurate tokenization
  - ✅ `CountTokensForModel()` method provides model-specific accurate counts
  - ✅ Fallback to estimation for unsupported providers maintained
  - ✅ Context window validation using actual token counts vs estimates

### 4.2 Comprehensive Logging ✅ COMPLETED

- [x] **Add detailed model filtering logs** ✅ COMPLETED
  - ✅ Start logging: `"Starting model compatibility check"` with provider_count, file_count, has_instructions
  - ✅ Per-model evaluation: `"Model evaluation:"` with model, provider, context_window, status, tokenizer, accurate
  - ✅ Detailed failure reasons: `"requires X tokens but model only has Y usable tokens (Z total - W safety margin)"`
  - ✅ Final summary: `"Model compatibility check completed"` with total_models, compatible_models, accurate_count, estimated_count

---

## Phase 5: Graceful Degradation & Error Handling (HIGH IMPORTANCE)

### 5.1 Fallback Mechanisms

- [ ] **Implement comprehensive fallback strategy**
  - If tokenizer initialization fails → fall back to estimation
  - If tokenizer.CountTokens() fails → fall back to estimation
  - If context gathering fails → fall back to instruction-only estimation
  - Log all fallbacks: `"Accurate tokenization failed for {provider}, using estimation: {error}"`

- [ ] **Add circuit breaker pattern**
  - Track tokenizer failure rates per provider
  - Temporarily disable problematic tokenizers and fall back to estimation
  - Reset circuit breaker after successful operations
  - Monitor and alert on high fallback rates

### 5.2 Performance Safeguards

- [ ] **Add performance monitoring**
  - Benchmark tokenizer initialization time (target: <100ms for tiktoken, <50ms for SentencePiece)
  - Monitor memory usage increase (target: <20MB for vocabularies)
  - Add timeout protection for large inputs (>1MB text)
  - Implement streaming tokenization for very large inputs

---

## Phase 6: Integration & CLI Enhancement (MEDIUM IMPACT)

### 6.1 CLI Integration

- [ ] **Update CLI flow for accurate tokenization**
  - Modify main.go to create TokenCountingService with provider tokenizers
  - Pass tokenizer dependencies (logger, model definitions) to service constructor
  - Ensure dry-run mode shows accurate token counting information
  - Add tokenizer status to dry-run output: `"Using accurate tokenization: OpenAI (tiktoken), Gemini (SentencePiece)"`

### 6.2 Enhanced Error Handling

- [ ] **Improve error propagation**
  - Wrap tokenizer errors with context about which provider/model failed
  - Include tokenizer type in error messages for debugging
  - Add integration test covering all fallback scenarios
  - Ensure model selection never fails completely due to tokenization issues

---

## Phase 7: Orchestrator Logging Enhancement (LOW IMPACT)

### 7.1 Token Context in Processing Logs

- [ ] **Add token metrics to model processing**
  - Enhance `processModelsWithErrorHandling()` to log: `"Processing {count} models with {tokens} total input tokens (accuracy: {method})"`
  - Log per-model attempt: `"Attempting model {name} ({index}/{total}) - context: {window} tokens, input: {tokens} tokens, utilization: {percentage}%"`
  - Include tokenizer method in processing logs

### 7.2 Structured Logging Enhancement

- [ ] **Update audit and structured logs**
  - Add token counting fields to thinktank.log: `{"input_tokens": X, "tokenizer_method": "tiktoken", "selected_models": Y, "skipped_models": Z}`
  - Update audit.jsonl with tokenizer information in model_selection operations
  - Add correlation ID to all tokenizer-related log entries
  - Summary log: `"Token counting summary: {tokens} tokens, {method} accuracy, {compatible} compatible, {skipped} skipped"`

---

## Phase 8: Testing & Validation (CRITICAL) ✅ COMPLETED

### 8.1 Accuracy Testing ✅ COMPLETED

- [x] **Comprehensive accuracy validation** ✅ COMPLETED
  - ✅ Accuracy comparison tests between estimation and tiktoken for OpenAI models
  - ✅ Test corpus covering English text, technical documentation, code with comments
  - ✅ Structured test scenarios with expected token counts and delta validation
  - ✅ Fallback validation ensuring graceful degradation to estimation
  - ✅ Edge case testing: empty input, whitespace, large content

### 8.2 Performance Testing ✅ COMPLETED

- [x] **Benchmark tokenizer performance** ✅ COMPLETED
  - ✅ Large file set stress tests: 150 files, >1MB total content
  - ✅ Startup time optimization with lazy loading (tokenizers initialized on demand)
  - ✅ Memory usage monitoring and benchmarks for vocabulary loading
  - ✅ Concurrent tokenization performance tests
  - ✅ Comprehensive benchmarks: `BenchmarkTokenCountingService_*` covering all scenarios

### 8.3 Integration Testing ✅ COMPLETED

- [x] **End-to-end validation** ✅ COMPLETED
  - ✅ `TestTokenCountingService_GetCompatibleModels_*` covering full integration flow
  - ✅ Model compatibility validation with context window exceeding scenarios
  - ✅ Multiple provider testing (OpenAI accurate, others estimation fallback)
  - ✅ Correlation ID propagation testing through all tokenizer operations
  - ✅ Mock utilities (`MockTokenizerManager`) for comprehensive testing scenarios

---

## Phase 9: OpenRouter Strategy (LOW PRIORITY)

### 9.1 OpenRouter Approach

- [ ] **Implement OpenRouter normalized tokenization**
  - Continue using estimation or implement GPT-4o tokenizer for OpenRouter models
  - Rationale: OpenRouter normalizes via GPT-4o anyway, exact tokenizer less critical
  - Consider: `tiktoken-go` with `cl100k_base` encoding for GPT-4o compatibility
  - Research: OpenRouter's actual tokenization approach for billing vs model selection

---

## Phase 10: Documentation & Cleanup (MEDIUM IMPORTANCE)

### 10.1 Documentation Updates

- [ ] **Update all documentation**
  - Document tokenizer selection logic in docs/STRUCTURED_LOGGING.md
  - Add tokenization troubleshooting guide
  - Document accuracy improvements and when fallbacks occur
  - Add configuration examples for token safety margins
  - Update CLI help text with tokenization information

### 10.2 Code Cleanup

- [ ] **Optimize and clean implementation**
  - Remove redundant estimation code only after accurate tokenizers proven stable
  - Ensure consistent error message formatting across all tokenizers
  - Add comprehensive inline documentation for all tokenizer interfaces
  - Run golangci-lint and address warnings
  - Consider extracting tokenizers to separate package if complexity grows

---

## Success Metrics & Acceptance Criteria

### **Must Have (P0)**
1. **Accuracy**: >90% token count accuracy for OpenAI and Gemini models vs <75% current estimation
2. **Reliability**: Model selection correctly filters based on actual context requirements
3. **Performance**: <500ms additional CLI startup time, <50MB memory overhead
4. **Fallback**: Graceful degradation to estimation maintains 100% compatibility
5. **Zero Regressions**: All existing functionality continues to work

### **Should Have (P1)**
1. **Coverage**: OpenAI and Gemini models use accurate tokenization, others use estimation
2. **Observability**: Clear logging shows which tokenizer used for each operation
3. **Configurability**: Token safety margin configurable via CLI and environment
4. **Testing**: >90% test coverage for all tokenizer code paths

### **Could Have (P2)**
1. **Advanced Features**: Token utilization metrics, caching, streaming tokenization
2. **Additional Providers**: Anthropic Claude tokenizer, more OpenRouter model support
3. **Performance Optimization**: Tokenizer result caching, parallel tokenization

### **Key Risk Mitigations**
- **Dependency Risk**: Both tiktoken-go and go-sentencepiece are mature, well-maintained
- **Performance Risk**: Lazy loading + caching + timeouts prevent startup/runtime issues
- **Complexity Risk**: Clean interfaces + comprehensive fallbacks + extensive testing
- **Compatibility Risk**: Gradual rollout with estimation fallback ensures no breaking changes

**Expected Outcome**: Dramatically more accurate model selection enabling the thinktank CLI to intelligently choose models that can actually handle the user's input, especially for non-English content and large codebases.

---

## Miscellaneous

### CLI Usability Improvements

- [x] **Support multiple target paths in CLI** ✅ COMPLETED
  - ✅ Allow arbitrary number of target directories/files: `thinktank instructions.md file1.ts file2.ts dir1/ dir2/`
  - ✅ Updated SimplifiedConfig to accept multiple target paths instead of single TargetPath
  - ✅ Modified ParseSimpleArgs to handle variable number of targets after flags
  - ✅ Updated validation logic to check all target paths exist
  - ✅ Ensured context gathering works with multiple disparate file/directory targets
  - ✅ Implementation note: Paths are joined with spaces in SimplifiedConfig.TargetPath
  - ✅ Limitation documented: Individual paths cannot contain spaces when using multiple paths
  - ✅ Comprehensive test coverage with integration tests

### Token Counting System Implementation

- [x] **Core TokenCountingService Implementation** ✅ COMPLETED
  - ✅ `TokenCountingService` interface with `CountTokens`, `CountTokensForModel`, `GetCompatibleModels`
  - ✅ Provider-aware tokenization with tiktoken for OpenAI, estimation fallback for others
  - ✅ Comprehensive error handling and graceful degradation
  - ✅ Structured logging with correlation ID support
  - ✅ Dependency injection pattern for testability

- [x] **Tiktoken Integration** ✅ COMPLETED
  - ✅ `github.com/pkoukk/tiktoken-go` dependency integration
  - ✅ Lazy loading tokenizer initialization
  - ✅ Model-specific encoding support (cl100k_base, o200k_base)
  - ✅ Thread-safe tokenizer management with caching

- [x] **Testing Infrastructure** ✅ COMPLETED
  - ✅ Comprehensive TDD test suite with 90%+ coverage
  - ✅ Performance benchmarks and stress testing
  - ✅ Mock utilities for testing (`MockTokenizerManager`, `MockAccurateTokenCounter`)
  - ✅ Integration tests covering end-to-end token counting flow

- [x] **Documentation Updates** ✅ COMPLETED
  - ✅ Enhanced `CLAUDE.md` with tokenization testing patterns
  - ✅ Updated `docs/STRUCTURED_LOGGING.md` with token counting logging documentation
  - ✅ Comprehensive inline code documentation

### CLI User Experience Improvements

- [x] **Add proper --help flag support** ✅ COMPLETED
  - ✅ Implemented `thinktank --help` and `-h` to show comprehensive usage information
  - ✅ Included examples of common usage patterns with multiple files/directories
  - ✅ Documented all available flags: `--dry-run`, `--verbose`, `--synthesis`, `--quiet`, `--json-logs`, `--no-progress`, `--debug`, `--model`, `--output-dir`
  - ✅ Added model selection information and provider requirements (API keys)
  - ✅ Showed file format support and exclusion patterns
  - ✅ Included comprehensive troubleshooting section for common issues
  - ✅ Made help output actually useful for new users with clear sections and examples 😄
  - ✅ TDD implementation with 100% test coverage for help functionality
  - ✅ Early help detection bypasses validation for better UX
  - ✅ Error messages now suggest running `thinktank --help`
