# Repository Guidelines

## Project Structure & Module Organization
- `cmd/thinktank/`: CLI entrypoint and wiring.
- `internal/`: core logic (models, file filtering, logging, processing).
- `docs/`: design notes, ADRs, guides.
- `scripts/`: dev tooling (coverage, pre-commit helpers).
- `config/`, `build/`, `docker/`, `benchmarks/`, `reports/`: support assets and automation.
- Tests live beside code as `*_test.go`.

## Build, Test, and Development Commands
- `go build ./...`: build all packages.
- `go test ./...`: run unit tests.
- `go test -race ./...`: race check (required before commit).
- `golangci-lint run ./...`: lint (fix all violations).
- `./scripts/check-coverage.sh`: enforce 79% minimum coverage.
- `./scripts/check-package-coverage.sh`: per-package coverage drilldown.
- `./scripts/test-local.sh`: full local test run.
- `govulncheck -scan=module`: vulnerability scan (hard fail).
- `pre-commit install`: enable hooks (timeouts enforced).

## Coding Style & Naming Conventions
- Go idioms. `gofmt` output only; tabs for indent.
- Names: `PascalCase` exported, `camelCase` unexported, files `lower_snake.go`.
- Prefer deep modules, small interfaces. Extract pure logic from I/O.
- No error suppression with `_`. Fix lint issues.
- Logging: console-friendly output + structured JSON (see `internal/logutil`).

## Testing Guidelines
- TDD first. Table-driven tests for pure functions.
- Prefer direct function testing with dependency injection. Avoid subprocess tests.
- Use `setupTestEnvironment()`; save/restore env vars.
- Integration tests skip when `OPENROUTER_API_KEY` missing:
  `t.Skip("OPENROUTER_API_KEY not set")`.
- Coverage target: 79% minimum via `./scripts/check-coverage.sh`.

## Commit & Pull Request Guidelines
- Conventional Commits required (`feat:`, `fix:`, `docs:`, `chore:`).
- Keep PRs small, focused. Include:
  - intent summary
  - tests run (commands)
  - linked issue or context
- Update docs with code changes.

## Security & Configuration Notes
- No secrets in repo. Use env vars only (`OPENROUTER_API_KEY`).
- Run `govulncheck` before release-quality changes.
