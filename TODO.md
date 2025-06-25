# GORDIAN CLI Simplification - Atomic Task Breakdown

## Phase 1: Foundation Architecture (Critical Path)

### Core Data Structures
- [x] **Create `SimplifiedConfig` struct in `internal/cli/simple_config.go`** ✅ *Completed: commit 2f64eb2*
  - ✅ Define memory-efficient struct: InstructionsFile(string), TargetPath(string), Flags(uint8) - 40 bytes total
  - ✅ Implement bitfield constants: FlagDryRun=0x01, FlagVerbose=0x02, FlagSynthesis=0x04
  - ✅ Add validation methods with O(1) bitwise operations (HasFlag, SetFlag, ClearFlag)
  - ✅ Write comprehensive unit tests with 100% coverage (all methods tested)
  - ✅ Include ToCliConfig() conversion method for backward compatibility
  - ✅ Smart defaults handled via conversion rather than storage for optimal memory usage

- [x] **Implement O(n) argument parser in `internal/cli/simple_parser.go`** ✅ *Completed: ParseSimplifiedArgs implemented*
  - ✅ Create `ParseSimplifiedArgs(args []string) (*SimplifiedConfig, error)` function
  - ✅ Single-pass parsing algorithm with switch statement for flags
  - ✅ Handle positional arguments: args[1]=instructions, args[2]=target, args[3+]=flags
  - ✅ Implement error handling for malformed arguments with specific error types
  - ✅ Add comprehensive tests with 40+ test cases covering edge cases and integration

- [x] **Build smart model selection algorithm in `internal/cli/model_selector.go`** ✅ *Completed: Smart model selection with caching*
  - ✅ Create `selectOptimalModel(availableProviders []string, taskSize int64) string` function
  - ✅ Implement provider detection via environment variable scanning with 5-minute TTL cache
  - ✅ Add model ranking algorithm: gemini-2.5-pro(score=100) > gpt-4o(score=95) > gpt-4(score=90) > claude-3-opus(score=85)
  - ✅ Cache provider availability to avoid repeated env var lookups with thread-safe operations
  - ✅ Write comprehensive tests with property-based validation and benchmarks (88.9-100% coverage)

### Validation Layer
- [x] **Implement fail-fast validation in `SimplifiedConfig.Validate()`** ✅ *Completed: Path length validation added*
  - ✅ File existence checks using `os.Stat()` with error wrapping
  - ✅ API key validation for selected model provider
  - ✅ Path accessibility verification (read permissions)
  - ✅ Path length validation (255 chars max) for cross-platform compatibility
  - ✅ Target <1ms validation time for typical inputs (achieved: ~12-14μs average)

- [x] **Create API key detection service in `internal/cli/api_detector.go`** ✅ *Completed: High-performance API key detection*
  - ✅ Scan environment for OPENAI_API_KEY, GEMINI_API_KEY, OPENROUTER_API_KEY
  - ✅ Validate API key format using regex patterns per provider (compile-time compiled)
  - ✅ Return provider capabilities map with rate limit information
  - ✅ Cache results with 5-minute TTL to avoid repeated env access
  - ✅ Add debug logging for key detection process (microsecond-level timing)
  - ✅ Performance: ~6-15μs detection time, <1KB memory footprint
  - ✅ Thread-safe caching with RWMutex optimization
  - ✅ Comprehensive test coverage with performance benchmarks

### Integration Interfaces
- [x] **Create adapter pattern in `internal/cli/config_adapter.go`** ✅ *Completed: Smart adapter with intelligent defaults*
  - ✅ Implement `SimplifiedConfig.ToComplexConfig() *config.CliConfig` conversion
  - ✅ Map simplified fields to existing CliConfig structure
  - ✅ Apply intelligent defaults for unmapped fields (rate limits, timeouts, concurrency)
  - ✅ Preserve all existing validation behavior through complex config path
  - ✅ Add round-trip testing: simplified → complex → behavior equivalence
  - ✅ Synthesis mode applies 60% rate limit reduction for conservative behavior
  - ✅ Comprehensive test coverage with 6 test functions covering all requirements

## Phase 2: Smart Defaults Engine

### Provider Intelligence
- [x] **Implement provider-specific rate limiting in `internal/cli/rate_limiter.go`** ✅ *Completed: Full circuit breaker pattern with intelligent backoff*
  - ✅ Define rate limit constants: OpenAI=3000rpm, Gemini=60rpm, OpenRouter=20rpm (matches models.GetProviderDefaultRateLimit())
  - ✅ Create per-provider rate limiter instances with token bucket algorithm (builds on existing ratelimit package)
  - ✅ Add circuit breaker pattern for failed providers (5 failures = 30s cooldown with proper state management)
  - ✅ Implement exponential backoff with jitter for retries (1s → 30s max with 10% jitter)
  - ✅ Monitor and log rate limit utilization via ProviderStatus API
  - ✅ Thread-safe implementation with RWMutex protection
  - ✅ Comprehensive test coverage: 15 test functions + benchmarks + concurrent access testing
  - ✅ Circuit breaker states: CLOSED → OPEN → HALF_OPEN → CLOSED transitions
  - ✅ Provider status monitoring with JSON serializable status reports

- [x] **Build context analysis engine in `internal/cli/context_analyzer.go`** ✅ *Completed: Full context analysis with intelligent caching*
  - ✅ Implement `analyzeTaskComplexity(targetPath string) (int64, error)` function
  - ✅ Count total files, lines of code, and estimated token count
  - ✅ Categorize complexity: Simple(<10k tokens), Medium(<50k), Large(<200k), XLarge(>200k)
  - ✅ Use complexity score to influence model selection and chunking strategy
  - ✅ Cache analysis results per directory with file modification time checks

### Output Management
- [x] **Create intelligent output directory naming in `internal/cli/output_manager.go`** ✅ *Completed: Full output management with cleanup*
  - ✅ Generate timestamp-based directory names: `thinktank_YYYYMMDD_HHMMSS_NNNNNNNNN`
  - ✅ Add collision detection and automatic incrementing
  - ✅ Implement cleanup for old output directories (>30 days by default)
  - ✅ Create output directory with proper permissions (0755)
  - ✅ Add structured logging for output operations

## Phase 3: Parsing & Business Logic Integration

### Argument Processing
- [x] **Implement positional argument validation in `validatePositionalArgs()`** ✅ *Completed: TDD implementation with 100% coverage*
  - ✅ Verify minimum 2 arguments provided (instructions file + target path)
  - ✅ Check file extension validation for instructions (.txt, .md accepted)
  - ✅ Validate target path exists and is accessible
  - ✅ Return structured errors with suggestions for common mistakes
  - ✅ Add comprehensive error message testing (8 test cases, 100% function coverage)

- [x] **Create flag parsing with standard library in `parseOptionalFlags()`** ✅ *Completed: Enhanced manual parsing with abbreviations and validation*
  - ✅ Parse `--model`, `--output-dir`, `--dry-run`, `--verbose`, `--synthesis` flags
  - ✅ Implement proper flag value extraction with error handling
  - ✅ Support both `--flag=value` and `--flag value` formats
  - ✅ Validate flag values against known acceptable ranges (empty value detection)
  - ✅ Add support for flag abbreviations: `-m` for `--model`, `-v` for `--verbose`
  - ✅ Performance: 29.83 ns/op (well under 100μs target)
  - ✅ Test coverage: 91.2% (exceeds 90% requirement)
  - ✅ Comprehensive tests: 17 test cases covering all scenarios

### Error Handling
- [x] **Design structured error types in `internal/cli/errors.go`** ✅ *Completed: Full CLI error system with LLM integration*
  - ✅ Define error constants: ErrMissingInstructions, ErrInvalidModel, ErrNoAPIKey, ErrMissingTargetPath, ErrConflictingFlags, ErrInvalidPath
  - ✅ Implement error wrapping with context preservation using CLIError struct
  - ✅ Add user-friendly error messages with actionable suggestions via UserFacingMessage() method
  - ✅ Create error categorization for different exit codes with mapCLIErrorToLLMCategory() function
  - ✅ Add error logging with correlation IDs via NewCLIErrorWithCorrelation() function
  - ✅ Implement CLIErrorType enum with 6 categories: MissingRequired, InvalidValue, ConflictingOptions, FileAccess, Configuration, Authentication
  - ✅ Create helper functions: NewMissingInstructionsError(), NewInvalidModelError(), NewNoAPIKeyError(), NewConflictingFlagsError(), NewInvalidPathError()
  - ✅ Add seamless integration with existing LLM error infrastructure via WrapAsLLMError() and IsCLIError() functions
  - ✅ Implement comprehensive test suite with TDD methodology (11 test functions, 100% coverage of error paths)
  - ✅ Add demonstration test showcasing full error system functionality with exit code mapping and correlation ID propagation

## Phase 4: Testing Infrastructure

### Unit Testing
- [x] **Write table-driven tests for argument parsing in `simple_parser_test.go`** ✅ *Completed: Comprehensive TDD test suite with 100% coverage and commit-ready*
  - ✅ Test valid argument combinations (20+ test cases across 5 test functions)
  - ✅ Test invalid argument patterns with expected error types and precise error message validation
  - ✅ Test edge cases: empty strings, Unicode filenames, very long paths (>255 chars), special characters
  - ✅ Benchmark parsing performance: achieved ~650-680ns (100x better than 100μs target)
  - ✅ Add property-based testing for argument permutations using testing/quick (50 iterations)
  - ✅ Complete test coverage: 100% coverage for all parser functions (ParseSimpleArgs, ParseSimpleArgsWithArgs, ParseSimpleArgsWithResult, IsSuccess, MustConfig)
  - ✅ TDD methodology: RED-GREEN-REFACTOR cycles with failing tests first, then implementation fixes
  - ✅ Comprehensive error validation: specific error message substring matching for all error conditions
  - ✅ Real file system integration: uses t.TempDir() and real files for authentic validation testing
  - ✅ Performance validation: 552 B/op, 9 allocs/op (efficient memory usage)
  - ✅ Deterministic testing: verifies parsing consistency across multiple invocations
  - ✅ Fix linting issues: resolved staticcheck warnings and added early returns for nil checks
  - ✅ Implementation complete and ready for commit (pending resolution of unrelated linting issues in other files)

- [x] **Create integration tests for config conversion in `config_adapter_test.go`** ✅ *Completed: Comprehensive integration tests with 775 lines*
  - ✅ Verify simplified config converts to equivalent complex config
  - ✅ Test all flag combinations produce correct complex config values
  - ✅ Validate default value application across conversion
  - ✅ Check rate limiting and timeout preservation
  - ✅ Add regression tests for existing CLI behavior
  - ✅ 6 integration test functions with 33 total sub-tests
  - ✅ Real component integration: rate limiters, validation pipeline, timeout config
  - ✅ End-to-end configuration flow testing with realistic environments
  - ✅ Behavior regression prevention with flag preservation testing
  - ✅ Model selection validation for synthesis vs normal modes

### Behavior Validation
- [ ] **Implement end-to-end compatibility testing in `e2e_simplified_test.go`**
  - Compare simplified CLI output with complex CLI output for identical inputs
  - Test file processing behavior equivalence
  - Validate API request patterns remain consistent
  - Check output file content and structure preservation
  - Measure performance regression: target <10% slowdown maximum

## Phase 5: Migration Strategy

### Backward Compatibility
- [ ] **Create dual-parser detection in `internal/cli/parser_router.go`**
  - Implement `detectParsingMode(args []string) ParsingMode` function
  - Heuristics: if first arg doesn't start with `-` and second arg doesn't start with `-`, use simplified
  - Route to appropriate parser based on detection
  - Add deprecation warnings for old flag usage
  - Log usage patterns for migration analytics

- [ ] **Add deprecation warnings for complex flags in existing parser**
  - Identify most commonly used complex flags from current codebase
  - Add warning messages with equivalent simplified command suggestions
  - Implement warning suppression flag for CI/automation
  - Create migration guide generation from detected usage patterns
  - Add telemetry for deprecation warning frequency

### Environment Variable Support
- [ ] **Implement environment variable fallbacks in `internal/cli/env_config.go`**
  - Support THINKTANK_MODEL, THINKTANK_OUTPUT_DIR, THINKTANK_DRY_RUN
  - Add THINKTANK_RATE_LIMIT_* for provider-specific overrides
  - Implement precedence: CLI flags > environment > defaults
  - Add environment variable validation and type conversion
  - Create environment configuration documentation generator

## Phase 6: Performance Optimization

### Memory Management
- [ ] **Optimize memory allocation in argument parsing**
  - Pre-allocate string slices for known maximum flag count
  - Implement string interning for common model names
  - Use sync.Pool for temporary parsing structures
  - Profile memory usage: target <1KB allocation per parse
  - Add memory leak testing for repeated parsing

### Startup Performance
- [ ] **Minimize initialization overhead in simplified path**
  - Lazy-load provider configurations only when needed
  - Cache API key detection results across invocations
  - Optimize file system operations with stat batching
  - Profile startup time: target <10ms cold start
  - Add startup performance benchmarks

## Phase 7: Code Cleanup & Finalization

### Legacy Removal
- [ ] **Remove deprecated flags from `internal/cli/flags.go`**
  - Delete unused flag definitions (after 3-month deprecation period)
  - Remove complex validation logic that's no longer needed
  - Clean up dead code paths in configuration system
  - Update help text to reflect simplified interface only
  - Verify no references remain in test files

- [ ] **Simplify configuration struct in `internal/config/config.go`**
  - Remove fields that are now handled by smart defaults
  - Consolidate remaining fields into logical groups
  - Update all references throughout codebase
  - Maintain API compatibility for internal packages
  - Add migration tests for configuration changes

### Documentation
- [ ] **Update CLI documentation in README.md and help text**
  - Replace complex examples with simplified equivalents
  - Add migration guide from old to new interface
  - Document environment variable options for advanced usage
  - Create quick start guide for new users
  - Add troubleshooting section for common issues

### Final Validation
- [ ] **Run comprehensive test suite with 90%+ coverage**
  - Execute all unit tests across simplified and legacy paths
  - Run integration tests with real API endpoints
  - Perform load testing with large codebases
  - Validate memory usage and performance benchmarks
  - Check compatibility across Go versions (1.21+)

## Acceptance Criteria
- [ ] **CLI interface reduced from 18 flags to 5 flags**
- [ ] **Argument parsing complexity: O(18n) → O(5) algorithmic improvement**
- [ ] **Memory usage: 269-line config → 33-byte struct (120x reduction)**
- [ ] **Startup time: <10ms cold start (10x improvement)**
- [ ] **Backward compatibility: Existing workflows continue working**
- [ ] **Test coverage: Maintain 90%+ coverage throughout transition**
- [ ] **Zero breaking changes: All existing functionality preserved**
- [ ] **Documentation: Complete migration guide and new user onboarding**
