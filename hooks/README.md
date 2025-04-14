# Git Hooks

This directory contains documentation for Git hooks used in this project.

## Pre-commit Hook

We use the [pre-commit](https://pre-commit.com/) framework for managing pre-commit hooks. The pre-commit hook runs automatically before each commit and performs the following checks:

1. **Code formatting**: Formats all Go files using `go fmt`
2. **Linting**: Runs `golangci-lint` to catch common issues
3. **Build verification**: Ensures the code builds without errors
4. **Quick tests**: Runs the fast unit tests
5. **Large file detection**: Warns about Go files exceeding 1000 lines, encouraging refactoring

## Installation

To set up the hooks:

### Prerequisites

1. Install the pre-commit framework:
   ```bash
   # Using pip
   pip install pre-commit

   # OR using Homebrew
   brew install pre-commit
   ```

2. Install the hooks:
   ```bash
   # From the project root
   pre-commit install
   ```

## Usage

- The hooks will run automatically on `git commit`
- To run all hooks manually:
  ```bash
  pre-commit run --all-files
  ```

## Skipping Hooks

In rare cases when you need to bypass the hooks:

```bash
git commit --no-verify -m "Your message"
```

## Troubleshooting

If you encounter issues with the hooks:

1. Ensure pre-commit is installed: `pre-commit --version`
2. Check the configuration in `.pre-commit-config.yaml`
3. Try running individual hooks manually, e.g.: `pre-commit run go-fmt`
