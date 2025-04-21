# Git Hooks

This directory contains documentation for Git hooks used in this project.

## Pre-commit Hook

We use the [pre-commit](https://pre-commit.com/) framework for managing pre-commit hooks. The pre-commit hook runs automatically before each commit and performs the following checks:

1. **Code formatting**: Formats all Go files using `go fmt`
2. **Linting**: Runs `golangci-lint` to catch common issues
3. **Build verification**: Ensures the code builds without errors
4. **Quick tests**: Runs a subset of unit tests with the `-short` flag, excluding the orchestrator package which has tests that may fail in the pre-commit environment
5. **Large file detection**: Warns about Go files exceeding 1000 lines, encouraging refactoring

## Post-commit Hook

We also use a post-commit hook that runs after each successful commit:

1. **Directory Overview Generation**: Runs `glance ./` to generate directory overview documentation

The glance tool creates `glance.md` files in each directory, providing a high-level overview of the directory's purpose and contents. This documentation is automatically updated after each commit to keep it in sync with the codebase.

### Dependencies

The post-commit hook requires the `glance` tool to be installed and available in your PATH. Without it, the hook will fail. The setup script attempts to detect and install glance if needed.

## Installation

There are two ways to set up the hooks:

### Option 1: Using the Setup Script (Recommended)

Run our setup script, which will check for and install required dependencies (including pre-commit):

```bash
# From the project root
./scripts/setup.sh
```

### Option 2: Manual Installation

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
   pre-commit install  # For pre-commit hooks
   pre-commit install --hook-type post-commit  # For post-commit hooks
   ```

## Usage

- The pre-commit hooks will run automatically on `git commit`
- The post-commit hooks will run automatically after a successful commit
- To run all hooks manually:
  ```bash
  pre-commit run --all-files
  ```
- To run a specific hook manually:
  ```bash
  pre-commit run run-glance --hook-stage post-commit
  ```

## Skipping Hooks

In rare cases when you need to bypass the hooks:

```bash
git commit --no-verify -m "Your message"
```

## Troubleshooting

If you encounter issues with the hooks:

### General Hook Issues:
1. Ensure pre-commit is installed: `pre-commit --version`
2. Check the configuration in `.pre-commit-config.yaml`
3. Try running individual hooks manually, e.g.: `pre-commit run go-fmt` or `pre-commit run run-glance --hook-stage post-commit`
4. For unit test issues, you can run the tests directly with: `go test -short ./cmd/architect/... ./internal/architect/interfaces ./internal/architect/modelproc ./internal/architect/prompt ./internal/auditlog ./internal/config ./internal/fileutil ./internal/gemini ./internal/integration ./internal/logutil ./internal/ratelimit ./internal/runutil`

### Post-commit Hook Issues:
1. Verify glance is installed: `which glance` and `glance --version`
2. Check if the post-commit hook is installed: `ls -la .git/hooks/post-commit`
3. Manually execute the post-commit hook script: `.pre-commit-hooks/run_glance.sh`
4. If you need to reinstall the hooks, run:
   ```bash
   pre-commit clean  # Remove existing hooks
   pre-commit install  # Reinstall pre-commit hooks
   pre-commit install --hook-type post-commit  # Reinstall post-commit hooks
   ```
5. Ensure your PATH includes the directory where glance is installed (typically `$(go env GOPATH)/bin`)

### Debugging Hook Execution:
To get more verbose output from the hooks, set the environment variable `PRE_COMMIT_COLOR=never` when running commands:
```bash
PRE_COMMIT_COLOR=never pre-commit run run-glance --hook-stage post-commit -v
```
