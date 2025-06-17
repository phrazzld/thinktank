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
* **Testing:** Write tests first (TDD); use table-driven tests; focus on integration testing
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
* **Cross-Package Testing:** Focus on robust integration tests over unit tests
* **Test Coverage:** Maintain 90% or higher code coverage for all packages
  * CI will fail if overall coverage drops below 90%
  * Use coverage scripts to identify coverage gaps before committing
* **Do not add your signature to commit messages**

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
