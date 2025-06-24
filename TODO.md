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
- [ ] **Implement fail-fast validation in `SimplifiedConfig.Validate()`**
  - File existence checks using `os.Stat()` with error wrapping
  - API key validation for selected model provider
  - Path accessibility verification (read permissions)
  - Model name format validation against known patterns
  - Target <1ms validation time for typical inputs

- [ ] **Create API key detection service in `internal/cli/api_detector.go`**
  - Scan environment for OPENAI_API_KEY, GEMINI_API_KEY, OPENROUTER_API_KEY
  - Validate API key format using regex patterns per provider
  - Return provider capabilities map with rate limit information
  - Cache results with 5-minute TTL to avoid repeated env access
  - Add debug logging for key detection process

### Integration Interfaces
- [ ] **Create adapter pattern in `internal/cli/config_adapter.go`**
  - Implement `SimplifiedConfig.ToComplexConfig() *config.CliConfig` conversion
  - Map simplified fields to existing CliConfig structure
  - Apply intelligent defaults for unmapped fields (rate limits, timeouts)
  - Preserve all existing validation behavior through complex config path
  - Add round-trip testing: simplified → complex → behavior equivalence

## Phase 2: Smart Defaults Engine

### Provider Intelligence
- [ ] **Implement provider-specific rate limiting in `internal/cli/rate_limiter.go`**
  - Define rate limit constants: OpenAI=3000rpm, Gemini=60rpm, OpenRouter=20rpm
  - Create per-provider rate limiter instances with token bucket algorithm
  - Add circuit breaker pattern for failed providers (5 failures = 30s cooldown)
  - Implement exponential backoff with jitter for retries
  - Monitor and log rate limit utilization

- [ ] **Build context analysis engine in `internal/cli/context_analyzer.go`**
  - Implement `analyzeTaskComplexity(targetPath string) (int64, error)` function
  - Count total files, lines of code, and estimated token count
  - Categorize complexity: Simple(<10k tokens), Medium(<50k), Large(<200k), XLarge(>200k)
  - Use complexity score to influence model selection and chunking strategy
  - Cache analysis results per directory with file modification time checks

### Output Management
- [ ] **Create intelligent output directory naming in `internal/cli/output_manager.go`**
  - Generate timestamp-based directory names: `thinktank_YYYYMMDD_HHMMSS_NNNNNNNNN`
  - Add collision detection and automatic incrementing
  - Implement cleanup for old output directories (>30 days by default)
  - Create output directory with proper permissions (0755)
  - Add structured logging for output operations

## Phase 3: Parsing & Business Logic Integration

### Argument Processing
- [ ] **Implement positional argument validation in `validatePositionalArgs()`**
  - Verify minimum 2 arguments provided (instructions file + target path)
  - Check file extension validation for instructions (.txt, .md accepted)
  - Validate target path exists and is accessible
  - Return structured errors with suggestions for common mistakes
  - Add comprehensive error message testing

- [ ] **Create flag parsing with standard library in `parseOptionalFlags()`**
  - Parse `--model`, `--output-dir`, `--dry-run`, `--verbose`, `--synthesis` flags
  - Implement proper flag value extraction with error handling
  - Support both `--flag=value` and `--flag value` formats
  - Validate flag values against known acceptable ranges
  - Add support for flag abbreviations: `-m` for `--model`, `-v` for `--verbose`

### Error Handling
- [ ] **Design structured error types in `internal/cli/errors.go`**
  - Define error constants: ErrMissingInstructions, ErrInvalidModel, ErrNoAPIKey
  - Implement error wrapping with context preservation
  - Add user-friendly error messages with actionable suggestions
  - Create error categorization for different exit codes
  - Add error logging with correlation IDs

## Phase 4: Testing Infrastructure

### Unit Testing
- [ ] **Write table-driven tests for argument parsing in `simple_parser_test.go`**
  - Test valid argument combinations (15+ test cases)
  - Test invalid argument patterns with expected error types
  - Test edge cases: empty strings, special characters, very long paths
  - Benchmark parsing performance: target <100μs for typical inputs
  - Add property-based testing for argument permutations

- [ ] **Create integration tests for config conversion in `config_adapter_test.go`**
  - Verify simplified config converts to equivalent complex config
  - Test all flag combinations produce correct complex config values
  - Validate default value application across conversion
  - Check rate limiting and timeout preservation
  - Add regression tests for existing CLI behavior

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
