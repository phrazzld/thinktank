# Git Hooks

This directory contains documentation for Git hooks used in this project.

## Pre-commit Hook

We use the [pre-commit](https://pre-commit.com/) framework for managing pre-commit hooks. The pre-commit hook runs automatically before each commit and performs the following checks:

1. **Code formatting**: Formats all Go files using `go fmt`
2. **Linting**: Runs `golangci-lint` to catch common issues
3. **Build verification**: Ensures the code builds without errors
4. **Quick tests**: Runs a subset of unit tests with the `-short` flag, excluding the orchestrator package which has tests that may fail in the pre-commit environment
5. **Large file detection**: Warns about Go files exceeding 1000 lines, encouraging refactoring
6. **Claude AI code review**: Uses Claude to analyze the commit for adherence to our development philosophy

## Claude AI Code Review

The Claude AI code review hook analyzes your changes before each commit to ensure they adhere to our development philosophy and coding standards.

### Purpose and Functionality

The Claude hook:
- Examines the diff of staged changes
- Evaluates the commit message against Conventional Commits specification
- Checks code against our development philosophy principles
- Provides feedback on maintainability, testability, and Go best practices
- Can block commits with significant issues or provide warnings for minor issues

### How It Works

When you make a commit:
1. The hook extracts the staged diff and commit message
2. It sends them to Claude with our development philosophy context
3. Claude analyzes the changes and responds with one of:
   - **PASS**: Everything looks good
   - **WARN**: Minor issues were found (commit proceeds with warnings)
   - **FAIL**: Major issues were found (commit is blocked)
4. Detailed, actionable feedback is displayed for any issues

### Dependencies

The Claude AI hook requires:
- **Claude CLI**: Must be installed and configured
- **Git**: For accessing diff and commit information

### Customizing

The prompt used by Claude is defined in the `.pre-commit-hooks/claude_sanity_check.sh` script and can be modified to adjust the strictness or focus of the review.

## Post-commit Hook

We use a post-commit hook that runs automatically after each successful commit:

1. **Directory Overview Generation**: Runs `glance ./` to generate or update directory documentation

### Purpose and Functionality

The `glance` tool analyzes the project structure and creates `glance.md` files in each directory. These files provide:

- A high-level overview of the directory's purpose
- A summary of the main components
- Important relationships between files
- Design patterns and architectural considerations

This automated documentation has several benefits:
- **Automatic synchronization**: Documentation stays up-to-date with code changes
- **Quick orientation**: New developers can quickly understand the codebase structure
- **Consistent documentation**: Uniform style across all directories
- **Reduced cognitive load**: No need to manually maintain directory overviews

### How It Works

When you make a commit:
1. The pre-commit hooks run first to validate your code
2. After the commit completes successfully, the post-commit hook executes
3. The hook script (`.pre-commit-hooks/run_glance.sh`) invokes `glance ./`
4. Glance analyzes each directory and creates/updates the `glance.md` files
5. These changes are available for your next commit

### Configuration

#### Hook Configuration

The post-commit hook is configured in `.pre-commit-config.yaml` with these key settings:
- `always_run: true` - Ensures it runs on every commit regardless of what files changed
- `pass_filenames: false` - Processes the entire repository, not just changed files
- `stages: [post-commit]` - Specifies when the hook runs

#### Glance Options

The glance tool supports several options that can be added to the hook script if needed:

- `-force` - Regenerates all glance.md files even if they already exist
- `-verbose` - Enables verbose logging for debugging purposes
- `-prompt-file <path>` - Specifies a custom prompt file to use instead of the default

For example, to force regeneration of all documentation, you could modify the hook script:

```bash
# Original command
glance ./

# Modified command to force regeneration
glance -force ./
```

### Dependencies

The post-commit hook requires:

1. **glance**: The directory documentation generator
   - Installation: `go install github.com/phaedrus-dev/glance@latest`
   - Requires Go to be installed
   - Must be available in your PATH (typically `$(go env GOPATH)/bin`)

2. **pre-commit framework**: For hook management
   - Installation: `pip install pre-commit` or `brew install pre-commit`
   - Configured through `.pre-commit-config.yaml`

The `scripts/setup.sh` script attempts to detect and install these dependencies if needed.

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

### One-time Skip

In rare cases when you need to bypass all hooks for a single commit:

```bash
git commit --no-verify -m "Your message"
```

### Selective Skip

To skip only specific hooks while running others:

```bash
# Skip the post-commit glance hook
SKIP=run-glance git commit -m "Your message"

# Skip the Claude AI code review
SKIP=claude-sanity-check git commit -m "Your message"

# Skip multiple hooks
SKIP=run-glance,claude-sanity-check git commit -m "Your message"
```

## Customizing and Maintaining

### Updating Hook Configuration

If you need to modify how the hooks work:

1. Edit `.pre-commit-config.yaml` to change hook parameters or add new hooks
2. Update `.pre-commit-hooks/run_glance.sh` to customize the glance command
3. Run `pre-commit clean && pre-commit install && pre-commit install --hook-type post-commit` to reinstall hooks

### Managing glance.md Files

The generated `glance.md` files:

- Should be committed to the repository
- Will be updated automatically after each commit
- Can be manually regenerated with `glance ./` or `glance -force ./`
- Are useful for documentation but should not be manually edited (as they'll be overwritten)

### Keeping Hooks Up to Date

As the project evolves:

1. Periodically update pre-commit: `pip install --upgrade pre-commit`
2. Update hook revisions in `.pre-commit-config.yaml` when new versions are available
3. Ensure the `scripts/setup.sh` script remains compatible with hook changes

## Troubleshooting

If you encounter issues with the hooks:

### General Hook Issues:
1. Ensure pre-commit is installed: `pre-commit --version`
2. Check the configuration in `.pre-commit-config.yaml`
3. Try running individual hooks manually, e.g.: `pre-commit run go-fmt` or `pre-commit run run-glance --hook-stage post-commit`
4. For unit test issues, you can run the tests directly with: `go test -short ./cmd/thinktank/... ./internal/thinktank/interfaces ./internal/thinktank/modelproc ./internal/thinktank/prompt ./internal/auditlog ./internal/config ./internal/fileutil ./internal/gemini ./internal/integration ./internal/logutil ./internal/ratelimit ./internal/runutil`

### Post-commit Hook Issues:

#### Common Problems

1. **"Command not found" errors**:
   - **Symptom**: Error message like `glance: command not found` or `claude: command not found` in hook output
   - **Causes**:
     - Required tool is not installed
     - Tool is installed but not in PATH
     - Shell environment in hook differs from your interactive shell
   - **Solutions**:
     - For glance: `go install github.com/phaedrus-dev/glance@latest`
     - For Claude CLI: Follow installation instructions at https://docs.anthropic.com/en/docs/agents-and-tools/claude-code/getting-started
     - Add required tools to PATH: `export PATH=$PATH:$(go env GOPATH)/bin`
     - Edit hook scripts to use absolute paths if needed

2. **Claude hook issues**:
   - **Symptom**: Claude hook fails or provides unexpected feedback
   - **Causes**:
     - Claude CLI not properly authenticated
     - API rate limits exceeded
     - Large diffs exceeding context limits
   - **Solutions**:
     - Verify Claude CLI works outside of hooks: `claude -p "hello"`
     - Split very large commits into smaller ones
     - Temporarily skip the hook if needed: `SKIP=claude-sanity-check git commit ...`

3. **"No hook with id 'run-glance' in stage 'post-commit'"**:
   - **Symptom**: Hook doesn't run after commit, no glance.md files generated
   - **Causes**:
     - Hook not properly registered
     - Incorrect stage configuration
   - **Solutions**:
     - Verify `.pre-commit-config.yaml` includes the hook with correct stage
     - Reinstall hooks (see instructions below)

3. **Silent failures**:
   - **Symptom**: Commit succeeds but no glance.md files are created/updated
   - **Causes**:
     - Hook script failing silently
     - Issue with glance tool execution
     - PATH issues in non-interactive shell
   - **Solutions**:
     - Run hook manually with verbose flag (see instructions below)
     - Check for error messages in `.git/logs/hooks/post-commit`
     - Try running glance directly: `glance ./`

#### Diagnostic Steps

1. **Verify tool installations**:
   ```bash
   # Check glance
   which glance  # Should show path to glance
   glance --version  # Should show version information

   # Check Claude CLI
   which claude  # Should show path to claude
   claude -p "hello"  # Should return a response from Claude
   ```

2. **Check hook installations**:
   ```bash
   # Check pre-commit hooks
   ls -la .git/hooks/pre-commit  # Should exist and be executable
   cat .git/hooks/pre-commit  # Should include pre-commit framework code

   # Check post-commit hooks
   ls -la .git/hooks/post-commit  # Should exist and be executable
   cat .git/hooks/post-commit  # Should include pre-commit framework code
   ```

3. **Manually test the hook scripts**:
   ```bash
   # Test glance script
   bash .pre-commit-hooks/run_glance.sh

   # Test Claude script (note: this won't have actual diff/commit message during manual testing)
   bash .pre-commit-hooks/claude_sanity_check.sh

   # Or run through pre-commit framework with verbose output
   PRE_COMMIT_COLOR=never pre-commit run run-glance --hook-stage post-commit -v
   PRE_COMMIT_COLOR=never pre-commit run claude-sanity-check -v
   ```

4. **Check PATH configuration**:
   ```bash
   # Check your interactive shell PATH
   echo $PATH

   # Check the PATH available during hook execution by modifying run_glance.sh
   # Temporarily add this to the top of the script:
   # echo "PATH during hook execution: $PATH" > /tmp/hook-path.log
   ```

#### Reinstallation Steps

If you need to completely reinstall the hooks:

```bash
# Remove existing hooks
pre-commit clean

# Install pre-commit hooks
pre-commit install

# Install post-commit hooks
pre-commit install --hook-type post-commit

# Verify installation
ls -la .git/hooks/post-commit
```

#### Environment Variables

These environment variables affect hook behavior:

- `PRE_COMMIT_COLOR=never`: Disables color in pre-commit output (useful for logs)
- `SKIP=run-glance`: Skips the glance hook for a single commit
- `PATH`: Affects where the shell looks for commands (ensure it includes glance location)
