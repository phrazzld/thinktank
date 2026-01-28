# CLAUDE

## Purpose

CLI tool that analyzes codebases using multiple LLMs via OpenRouter. Sends instructions + code to models, saves responses, optionally synthesizes.

## Architecture Map

```
cmd/thinktank/main.go          → CLI entry, flag parsing
internal/cli/run.go:50         → Run() orchestrates execution
internal/thinktank/            → Core processing logic
  orchestrator/orchestrator.go → Model coordination
  modelproc/processor.go       → Individual model calls
internal/providers/openrouter/ → API client
internal/models/models.go      → Model definitions (ADD NEW MODELS HERE)
internal/logutil/              → Console output + TUI components
internal/fileutil/             → File filtering, gitignore
```

**Start here:** `internal/cli/run.go:50` - the `Run()` function shows the full execution flow.

## Run & Test

```bash
# Build & run
go build ./... && thinktank instructions.txt ./src --dry-run

# Test (race detection required before commit)
go test -race ./...

# Coverage (79% minimum)
./scripts/check-coverage.sh

# Lint (fix all violations)
golangci-lint run ./...

# Vulnerability scan (hard fail)
govulncheck -scan=module
```

**Required env:** `OPENROUTER_API_KEY` (single key for all models)

## Quality & Pitfalls

### Definition of Done
- Tests pass with `-race`
- Coverage ≥79%
- golangci-lint clean
- Conventional commit message

### Critical Invariants
- **No error suppression**: Never ignore errors with `_`. Fix them.
- **TDD**: Write tests first, then implement
- **Direct function testing**: Use dependency injection, not subprocess tests
- **Dual logging**: ConsoleWriter for users, structured JSON for debugging

### Adding New Models
1. Edit `internal/models/models.go` → add to `ModelDefinitions` map
2. Run `go test ./internal/models && go test ./...`
3. Test: `thinktank instructions.txt ./src --dry-run`

### TUI/Terminal Output (internal/logutil)
- Use `runewidth.StringWidth()` for alignment, not `len()` (Unicode width ≠ bytes)
- Use `AdaptiveColor{Light: dark, Dark: light}` for accessibility
- Provide ASCII fallbacks when `!isInteractive` (CI/logs)
- Structs with goroutines need `Stop()` method + `t.Cleanup(obj.Stop)` in tests
- Render methods are lock-free; caller owns serialization (consoleWriter)

### Pre-commit Hooks
```bash
pre-commit install  # One-time setup
```
Timeouts: lint 4min, coverage 8min. Skip with `--no-verify` (sparingly).

## References

- [README.md](README.md) - Usage, CLI flags, model list
- [docs/DEVELOPMENT.md](docs/DEVELOPMENT.md) - License policy, detailed setup
- [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) - Common issues
- [docs/leyline/](docs/leyline/) - Development philosophy (tenets + bindings)
- [AGENTS.md](AGENTS.md) - Coding style, testing guidelines
