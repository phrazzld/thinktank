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

## Reference

* Architecture: `docs/leyline/` for development philosophy
* Testing: Focus on integration tests for critical paths only
* Coverage: `./scripts/check-package-coverage.sh` for per-package analysis
