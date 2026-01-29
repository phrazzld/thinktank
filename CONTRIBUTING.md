# Contributing to thinktank

## Development Setup

```bash
# Clone and install
git clone https://github.com/misty-step/thinktank.git
cd thinktank
go install

# Set up pre-commit hooks
pre-commit install

# Set API key for integration tests
export OPENROUTER_API_KEY="your-key"
```

## Quality Gates

All changes must pass:

```bash
# Required before commit
go test -race ./...              # Tests with race detection
golangci-lint run ./...          # Lint (fix all violations)
./scripts/check-coverage.sh      # 79% minimum coverage

# Security scan (before release)
govulncheck -scan=module
```

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` new feature
- `fix:` bug fix
- `docs:` documentation only
- `chore:` maintenance, dependencies
- `refactor:` code change that neither fixes a bug nor adds a feature
- `test:` adding or updating tests

## Pull Request Process

1. Create a feature branch from `master`
2. Make changes with tests
3. Run quality gates locally
4. Submit PR with:
   - Clear description of changes
   - Test commands you ran
   - Link to related issue (if any)

## Testing Guidelines

- Write tests first (TDD)
- Use table-driven tests for pure functions
- Prefer dependency injection over subprocess tests
- Integration tests should skip when `OPENROUTER_API_KEY` is not set

## Code Style

- Go idioms, `gofmt` formatted
- Deep modules, small interfaces
- No error suppression with `_`
- See [AGENTS.md](AGENTS.md) for full guidelines

## Adding New Models

Edit `internal/models/models.go`:

```go
"new-model": {
    Provider:        "openrouter",
    APIModelID:      "provider/model-id",
    ContextWindow:   128000,
    MaxOutputTokens: 16000,
    DefaultParams: map[string]interface{}{
        "temperature": 0.7,
    },
},
```

Then run: `go test ./internal/models && go test ./...`
