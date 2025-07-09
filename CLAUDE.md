# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

* **Build:** `go build ./...`
* **Run Tests:** `go test ./...`
* **Coverage Check:** `./scripts/check-coverage.sh` (required 90% threshold)
* **Race Detection:** `go test -race ./...` (required before committing)
* **Lint:** `golangci-lint run ./...` (fix all violations)
* **Vulnerability Scan:** `govulncheck -scan=module`

## Code Quality Requirements

* **TDD:** Write tests first, then implement
* **Direct Function Testing:** Use dependency injection over subprocess tests
* **90% Test Coverage:** CI fails below 90% - use coverage scripts
* **No Error Suppression:** Fix all `golangci-lint` violations, never ignore with `_`
* **Dual-Output Logging:** ConsoleWriter for users + structured JSON for debugging

## Mandatory Testing Patterns

**✅ Preferred: Direct function testing with dependency injection**
```go
type RunConfig struct {
    Context    context.Context
    Logger     logutil.LoggerInterface
    FileSystem FileSystemInterface
}

func Run(cfg *RunConfig) *RunResult {
    // Business logic - returns result instead of os.Exit()
}
```

**❌ Avoid: Subprocess testing**
```go
cmd := exec.Command("./binary", "--flag", "value")
// Fragile, slow, hard to debug
```

## Repository-Specific Workflows

### Adding New Models
1. Edit `internal/models/models.go` - add to `ModelDefinitions` map
2. Run `go test ./internal/models` and `go test ./...`
3. Test integration: `go run cmd/thinktank/main.go --model new-model --dry-run`

### Using thinktank for Analysis
When stuck on complex problems:
```bash
thinktank --instructions temp_instructions.txt ./path/to/code
```

### Token Counting Testing
* Test tiktoken (OpenRouter uses tiktoken-o200k for all models)
* Include performance benchmarks for >100 files, >1MB
* Validate against OpenAI tokenizer playground for reference

### API Key Testing (Post-OpenRouter Consolidation)
* **Single API Key**: All tests use `OPENROUTER_API_KEY` only
* **Environment Isolation**: Always save/restore original API keys in tests
* **Test Helpers**: Use `setupTestEnvironment()` for consistent env management
* **Mock Keys**: Use `"test-openrouter-key"` or `"sk-or-test-key"` format
* **Skip Pattern**: `t.Skip("OPENROUTER_API_KEY not set")` for integration tests
* **Security**: Never use production keys in tests - use test-prefixed keys

## Critical Constraints

* **Conventional Commits:** Required for automated versioning
* **No Secrets:** Use env vars only
* **Security:** `govulncheck` hard-fails on ANY vulnerability
* **Structured Results:** Return `RunResult` from main logic, not `os.Exit()`

## Pre-commit Hooks

* **Installation:** `pre-commit install` (required for development)
* **Timeout Limits:** Hooks have aggressive timeouts to prevent hanging
  - golangci-lint: 4 minutes (with --fast flag)
  - go-build-check: 2 minutes
  - go-coverage-check: 8 minutes (with intelligent caching)
  - fast-tokenizer-tests: 1 minute
* **Performance Features:**
  - Documentation-only changes skip coverage checks automatically
  - Aggressive caching based on content hashes
  - Parallel test execution where possible
  - Timeout recovery with cached fallbacks
* **Troubleshooting:** `./scripts/precommit-troubleshoot.sh` for performance issues
* **Emergency Skip:** `git commit --no-verify` (use sparingly)

## Function Organization & Testing Patterns

### Carmack-Style Function Extraction (2025-07-08 Refactoring)

Following John Carmack's incremental refactoring philosophy, the codebase underwent systematic function extraction to separate I/O operations from business logic. This approach achieved 90.4% test coverage and 35-70% performance improvements.

**✅ Pure Function Extraction Pattern**
```go
// Before: Mixed I/O and business logic
func processFile(path string) error {
    data, err := os.ReadFile(path)  // I/O operation
    if err != nil {
        return err
    }

    result := calculateStatistics(data)  // Business logic
    fmt.Printf("Results: %v\n", result)  // I/O operation
    return nil
}

// After: Pure business logic extracted
func CalculateFileStatistics(content []byte) FileStats {
    // Pure business logic - no I/O, fully testable
    return FileStats{...}
}

func ReadFileContent(path string) ([]byte, error) {
    // Pure I/O operation
    return os.ReadFile(path)
}
```

**✅ Function Decomposition Pattern**
```go
// Before: Large function (370+ LOC)
func Execute(config Config) error {
    // Setup, validation, processing, output - all mixed
}

// After: Focused functions (<100 LOC each)
func gatherProjectFiles(config Config) error      // Setup phase
func processFiles(config Config) error            // Processing phase
func generateOutput(config Config) error          // Generation phase
func writeResults(config Config) error            // Output phase
```

**✅ Testing Pattern for Extracted Functions**
```go
// Table-driven tests for pure functions
func TestCalculateFileStatistics(t *testing.T) {
    tests := []struct {
        name     string
        content  []byte
        expected FileStats
    }{
        {"go file", []byte("package main\n"), FileStats{Lines: 1, Type: "go"}},
        {"empty file", []byte(""), FileStats{Lines: 0, Type: "unknown"}},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := CalculateFileStatistics(tt.content)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Key Refactoring Outcomes

**Performance Improvements:**
- Token counting: 35-70% faster across all file sizes
- Memory allocation: No regressions, improved in most cases
- Function call overhead: Minimal impact despite decomposition

**Code Organization:**
- `internal/logutil/formatting.go` - Pure formatting functions (210 LOC)
- `internal/fileutil/filtering.go` - Pure filtering/statistics functions (469 LOC)
- Main functions decomposed: `Execute()` (370→27 LOC), `Main()` (120→40 LOC)

**Testing Coverage:**
- Overall coverage: 83.6% → 90.4% (exceeds 90% target)
- Pure functions: 95-100% coverage (no mocking required)
- Integration tests: Verify behavioral equivalence post-refactoring

**Function Size Compliance:**
- All functions now <100 LOC (Carmack principle)
- Clear separation of concerns: I/O vs business logic
- Improved testability and maintainability

### Implementation Guidelines

When extracting functions, follow this proven pattern:
1. **Identify Pure Logic**: Extract business logic with no I/O dependencies
2. **Separate I/O Operations**: Create focused I/O functions
3. **Decompose Large Functions**: Break >100 LOC into focused phases
4. **Test Extracted Functions**: Use table-driven tests for pure functions
5. **Validate Behavior**: Ensure identical behavior post-refactoring

## Reference

* Architecture: `docs/leyline/` for development philosophy
* Testing: Focus on integration tests for critical paths only
* Coverage: `./scripts/check-package-coverage.sh` for per-package analysis
