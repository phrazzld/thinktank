# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

* **Build:** `go build ./...`
* **Run Tests:** `go test ./...`
* **Run Single Test:** `go test -v -run TestName ./path/to/package`
* **Run E2E Tests:** `./internal/e2e/run_e2e_tests.sh [-v] [-r TestPattern]`
* **Race Detection Testing:**
  * Full suite: `go test -race ./...` (required before committing test changes)
  * Single package: `go test -race ./path/to/package`
  * Repeated testing: `go test -race ./... -count=10` (catch intermittent races)
* **Check Coverage:**
  * Basic: `go test -cover ./...`
  * Detailed: `go test -coverprofile=coverage.out ./...`
  * Verify 90% threshold: `./scripts/check-coverage.sh [threshold]`
  * Per-package report: `./scripts/check-package-coverage.sh [threshold]`
* **Format Code:** `go fmt ./...`
* **Lint Code:** `go vet ./...`
* **Run golangci-lint:** `golangci-lint run ./...` (catches errcheck, staticcheck, and other violations)

## Go Style Guidelines

* **Package Structure:** Package-by-feature, with small focused interfaces
* **Imports:** Group standard library, external, internal imports with blank line separators
* **Error Handling:** Return errors rather than panic; use structured errors with context
* **Naming:** Use clear, descriptive names; camelCase for variables, PascalCase for exports
* **Testing:** Write tests first (TDD); use table-driven tests; prefer direct function testing over subprocess tests
* **Types:** Use strong typing; avoid unnecessary interface{}/any; leverage Go's type system
* **Error Flow:** Explicit error handling; no suppression of errors or linter warnings
* **Comments:** Document *why*, not *how*; code should be self-documenting
* **Validation:** Validate all external input rigorously at system boundaries

## Mandatory Practices

* **Use TDD:** Write tests first, make them fail, then implement code to pass
* **Conventional Commits:** Follow the spec for automated versioning/changelogs
* **Write detailed multiline conventional commit messages**
* **No Secrets in Code:** Use environment variables or designated secret managers
* **Dual-Output Logging:** Use ConsoleWriter for user-facing output and structured logging for debugging
  * ConsoleWriter provides clean progress reporting that adapts to TTY vs CI environments
  * Structured JSON logging maintains comprehensive audit trails with correlation IDs
* **Pre-commit Quality:** All code must pass tests, lint, and format checks
  * Run `golangci-lint run ./...` before committing to catch violations early
  * Fix all errcheck violations - never ignore errors with `_`
* **Direct Function Testing:** Prefer direct function tests with dependency injection over subprocess tests
* **Test Coverage:** Maintain 90% or higher code coverage for all packages
  * CI will fail if overall coverage drops below 90%
  * Use coverage scripts to identify coverage gaps before committing
* **Do not add your signature to commit messages**

## Testing Patterns and Guidelines

### Direct Function Testing (Preferred)

Always prefer direct function testing with dependency injection over subprocess tests:

**✅ Good: Direct function testing**
```go
func TestHandleError(t *testing.T) {
    mockLogger := &testutil.MockLogger{}

    err := handleError(mockLogger, &llm.LLMError{
        Provider: "test",
        Message: "Authentication failed",
        ErrorCategory: llm.CategoryAuth,
    })

    assert.Equal(t, ExitCodeAuthError, err)
    assert.Contains(t, mockLogger.Messages, "Authentication failed")
}
```

**❌ Avoid: Subprocess testing**
```go
func TestHandleError(t *testing.T) {
    cmd := exec.Command("./binary", "--flag", "value")
    output, err := cmd.CombinedOutput()
    // Fragile, slow, hard to debug
}
```

### Dependency Injection Patterns

Extract business logic from main() into testable functions with dependency injection:

**Example: RunConfig/RunResult Pattern**
```go
type RunConfig struct {
    Context         context.Context
    Config          *config.CliConfig
    Logger          logutil.LoggerInterface
    AuditLogger     auditlog.AuditLogger
    FileSystem      FileSystemInterface
    // ... other dependencies
}

func Run(cfg *RunConfig) *RunResult {
    // Business logic with injected dependencies
    // Returns structured result instead of calling os.Exit()
}

func TestRun(t *testing.T) {
    cfg := &RunConfig{
        Context:     context.Background(),
        Config:      &config.CliConfig{DryRun: true},
        Logger:      &testutil.MockLogger{},
        FileSystem:  &testutil.MockFileSystem{},
    }

    result := Run(cfg)
    assert.Equal(t, ExitCodeSuccess, result.ExitCode)
}
```

### When to Use Integration Tests

Use focused integration tests only for critical path validation:

**✅ Appropriate for integration tests:**
- Binary builds and executes
- End-to-end flag parsing with real command line
- Critical failure modes (exit codes, error messages)
- File I/O operations with real filesystem

**❌ Not appropriate for integration tests:**
- Business logic testing (use direct function tests)
- Error handling variations (test error functions directly)
- Flag validation (test parsing functions directly)
- API client behavior (use mocks)

**Example: Focused integration test**
```go
func TestCriticalPathIntegration(t *testing.T) {
    binaryPath := buildThinktankBinary(t)

    // Test only critical success/failure paths
    stdout, stderr, exitCode, err := executeBinary(t, binaryPath,
        []string{"--instructions", "test.md", "--dry-run", "src/"},
        tempDir, 30*time.Second)

    // Validate integration points, not detailed business logic
}
```

### Testing Architecture Principles

1. **Extract main() logic**: Move business logic from main() into testable Run() functions
2. **Use dependency injection**: Accept interfaces for all external dependencies
3. **Return structured results**: Return RunResult instead of calling os.Exit()
4. **Mock external dependencies**: Use interfaces for filesystem, logger, API clients
5. **Test business logic directly**: Avoid subprocess execution for logic testing
6. **Keep integration tests focused**: Test only critical integration points

### Test Coverage Strategy

- **Unit Tests**: Test individual functions with mocked dependencies (fast, reliable)
- **Direct Function Tests**: Test main logic flows with dependency injection (fast, thorough)
- **Integration Tests**: Test critical binary execution paths (slower, end-to-end validation)

Aim for 90% coverage primarily through direct function tests, supplemented by focused integration tests.

## Security & Vulnerability Scanning

* **Automated Vulnerability Scanning:** All commits and PRs are automatically scanned for vulnerabilities
  * **Tool:** `govulncheck` (official Go vulnerability scanner)
  * **Scan Level:** Module-level scanning (`-scan=module`) for comprehensive coverage
  * **Frequency:** Every commit and PR to master branch
  * **Failure Behavior:** Hard fail - ANY vulnerability detected fails the build
  * **Reports:** JSON and text formats uploaded as artifacts (30-day retention)
  * **Timeout:** 3-minute maximum execution time with retry logic
* **Manual Vulnerability Check:** `govulncheck -scan=module`
* **Emergency Rollback:** Comment out vulnerability-scan job in `.github/workflows/ci.yml`

## Using the `thinktank` CLI

This repo contains the `thinktank` CLI tool itself, which can analyze code using different LLM models. When working on problems:

1. For complex tasks, outline your approach first
2. If stuck, consider using `thinktank` to get a different perspective
3. Example: `thinktank --instructions temp_instructions.txt ./path/to/relevant/files`

The tool works by creating a temporary file with instructions, then analyzing the specified paths. API keys are pre-configured locally.

## Adding New Models

The thinktank project uses a hardcoded model system for simplicity and reliability. To add a new model:

### Process Overview

1. **Add Model Definition:** Edit `internal/models/models.go` and add an entry to the `ModelDefinitions` map
2. **Run Tests:** Verify all tests pass with `go test ./internal/models` and `go test ./...`
3. **Submit PR:** Follow standard contribution process with conventional commit messages

### Step-by-Step Instructions

1. **Edit the ModelDefinitions map in `internal/models/models.go`:**
   ```go
   "new-model-name": {
       Provider:        "provider-name",    // openai, gemini, or openrouter
       APIModelID:      "api-model-id",     // ID used in API calls
       ContextWindow:   100000,             // Max input + output tokens
       MaxOutputTokens: 50000,              // Max output tokens
       DefaultParams: map[string]interface{}{
           "temperature": 0.7,
           // ... other provider-specific parameters
       },
   },
   ```

2. **Run comprehensive tests:**
   ```bash
   # Test the models package specifically
   go test ./internal/models

   # Run full test suite
   go test ./...

   # Check test coverage
   go test -cover ./internal/models
   ```

3. **Verify integration:**
   ```bash
   # Test with the new model using dry-run
   go run cmd/thinktank/main.go --model new-model-name --dry-run ./README.md
   ```

4. **Submit PR with conventional commit:**
   ```bash
   git add internal/models/models.go
   git commit -m "feat: add support for new-model-name

   - Add new-model-name to ModelDefinitions with provider configuration
   - Includes context window: 100k tokens, max output: 50k tokens
   - Uses standard provider parameters for optimal performance"
   ```

### Adding New Providers

To add a completely new provider (beyond openai, gemini, openrouter):

1. Add models with the new provider name to `ModelDefinitions`
2. Update `GetAPIKeyEnvVar()` function to include the new provider
3. Ensure client creation logic supports the new provider (outside models package)
4. Add comprehensive tests for the new provider

For detailed technical documentation, see `internal/models/README.md`.
