# Git Hooks

This directory contains templates for Git hooks that can be used to enforce quality standards before commits.

## Pre-commit Hook

The pre-commit hook runs automatically before each commit and performs the following checks:

1. **Code formatting**: Formats all Go files using `go fmt`
2. **Linting**: Runs `golangci-lint` to catch common issues
3. **Build verification**: Ensures the code builds without errors
4. **Quick tests**: Runs the fast unit tests
5. **Large file detection**: Warns about Go files exceeding 1000 lines, encouraging refactoring

## Installation

To use these hooks, you can either:

### Option 1: Install directly

```bash
# From the project root
cp hooks/pre-commit .git/hooks/
chmod +x .git/hooks/pre-commit
```

### Option 2: Use symlinks

```bash
# From the project root
ln -sf "$(pwd)/hooks/pre-commit" .git/hooks/pre-commit
```

## Requirements

- The pre-commit hook requires `golangci-lint` to be installed.
- Install with: `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`

## Skipping Hooks

In rare cases when you need to bypass the hooks:

```bash
git commit --no-verify -m "Your message"
```

## Troubleshooting

If you encounter issues with the hooks:

1. Ensure they are executable: `chmod +x .git/hooks/pre-commit`
2. Check that all required tools are installed
3. Try running the commands manually to identify specific errors