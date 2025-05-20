# Contributing to Thinktank

Thank you for your interest in contributing to Thinktank! This document provides detailed information about the development environment setup, coding standards, testing requirements, and workflow processes.

## Table of Contents

- [Development Environment Setup](#development-environment-setup)
  - [Prerequisites](#prerequisites)
  - [Tools Installation](#tools-installation)
  - [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Project Structure](#project-structure)

## Development Environment Setup

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.21+**: The project requires Go version 1.21 or later. [Download Go](https://golang.org/dl/)
- **Git**: For version control. [Download Git](https://git-scm.com/downloads)
- **Make**: Required for running Makefile commands

### Tools Installation

This project uses several tools for development, testing, and code quality. We maintain a `tools.go` file to pin tool dependencies and ensure consistent versions across all development environments.

#### Automatic Installation (Recommended)

The simplest way to install all required tools is through our Makefile:

```bash
# Clone the repository if you haven't already
git clone https://github.com/phrazzld/thinktank.git
cd thinktank

# Install all development tools
make tools
```

This command parses the `tools.go` file and installs all required tools automatically.

#### Manual Installation

If you prefer to install tools manually, you can run:

```bash
# Install tools directly using go install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/caarlos0/svu@latest
go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
```

#### PATH Configuration

Ensure that your `$GOPATH/bin` directory is in your PATH to access the installed tools:

```bash
# Add this to your .bashrc, .zshrc, or equivalent configuration file
export PATH=$PATH:$(go env GOPATH)/bin
```

#### Tool Verification

To verify that all tools are installed correctly:

```bash
# Check that the tools are available in your PATH
golangci-lint --version
govulncheck --help
svu --version
git-chglog --version
```

### Development Workflow

**FIRST TIME SETUP (REQUIRED):**
```bash
# Clone the repository
git clone https://github.com/phrazzld/thinktank.git
cd thinktank

# Install all tools AND git hooks (MANDATORY)
make tools

# OR run the setup script for a more comprehensive setup
./scripts/setup.sh
```

Our project provides several Make commands to streamline development:

```bash
make help           # Display all available commands
make tools          # Install all tools and git hooks (run this first!)
make hooks          # Install only git hooks
make build          # Build the project
make test           # Run all tests
make test-short     # Run short tests (faster)
make lint           # Run the linter
make fmt            # Format code
make coverage       # Run tests with coverage
make cover-report   # Generate HTML coverage report
make clean          # Clean build artifacts
make vendor         # Update vendor directory
```

## Coding Standards

We follow strict coding standards in this project, as detailed in our [Development Philosophy](docs/DEVELOPMENT_PHILOSOPHY.md) document. Key points include:

- Follow Go's official [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Ensure proper error handling is implemented
- Write clear, concise, self-documenting code
- Document the "why" (rationale), not just the "how" (mechanics)
- Maintain high test coverage

## Pre-commit Hooks and Commit Message Standards (MANDATORY)

This project uses pre-commit hooks to ensure code quality and enforce Conventional Commits for consistent commit messages.

**‼️ IMPORTANT: HOOKS ARE REQUIRED ‼️**  
**Pre-commit hooks are ABSOLUTELY MANDATORY for all contributors without exception. The installation of these hooks is a non-negotiable prerequisite for contributing to this project. Bypassing hooks using `--no-verify` or any other method is STRICTLY FORBIDDEN and will result in your contributions being rejected.**

### Automatic Hook Installation (REQUIRED STEP)

As part of the initial setup process, you MUST install the pre-commit hooks. This is automatically done when you run:
```bash
make tools      # Installs all development tools AND git hooks
# OR
make hooks      # Installs only git hooks
# OR
./scripts/setup.sh  # Full development environment setup
```

**Baseline-Aware Commit Validation:**
The setup process configures commit message validation with baseline awareness. This means:
- Commit message validation only applies to commits made after May 18, 2025 (commit `1300e4d`)
- Historical commits before the baseline date are exempt from validation
- This matches how CI validates commits, avoiding false failures for old commits
- All new commits must still follow the Conventional Commits standard

**VERIFICATION:** After installation, verify hooks are properly installed by running:
```bash
pre-commit info
```

You should see confirmation that the hooks are installed for this repository.

**Pre-push Validation:** When you attempt to push commits, a pre-push hook will automatically validate your commits against the conventional commit standard, using the same baseline-aware approach. This means:

- Only commits made after May 18, 2025 (baseline commit `1300e4d`) will be validated
- Historical commits before the baseline date will be skipped
- If any new commits don't follow the conventional commit format, the push will be blocked
- You'll receive clear error messages and fix tips for any invalid commits

**Manual PR Validation:** To manually check your branch for commit message compliance before pushing, run:
```bash
./scripts/validate-pr.sh
```

This script will:
- Only validate commits made after May 18, 2025 (baseline commit `1300e4d`)
- Skip historical commits before the baseline date
- Show detailed information about each commit being checked
- Provide helpful error messages and fix tips for invalid commits

Example usage:
```bash
# Check current branch against master
./scripts/validate-pr.sh

# Check a specific branch against master
./scripts/validate-pr.sh feature/my-feature

# Check a specific branch against a custom base branch
./scripts/validate-pr.sh feature/my-feature main
```

Running this validation before pushing or submitting your PR will help you catch and fix any commit message issues early.

### Manual Hook Installation (ONLY if automatic installation fails)

If for any reason the automatic installation fails, you MUST follow these steps:

1. **Install pre-commit** (if not already installed):
   ```bash
   # Using pip
   pip install pre-commit

   # Or using Homebrew on macOS
   brew install pre-commit
   ```

2. **Install ALL required hooks**:
   ```bash
   # Install pre-commit hooks with automatic fixes
   pre-commit install --install-hooks

   # Install commit message validation hook
   pre-commit install --hook-type commit-msg

   # Install pre-push hooks
   pre-commit install --hook-type pre-push

   # Install post-commit hooks
   pre-commit install --hook-type post-commit
   ```

3. **Verify installation** (REQUIRED):
   ```bash
   pre-commit info
   ```

### Hook Features (Benefits of Mandatory Hooks)
- **Automatic Formatting**: EOF newlines, trailing whitespace, and Go formatting are automatically fixed on commit
- **Commit Message Validation**: Ensures all commits follow Conventional Commits specification
- **Code Quality Checks**: Runs linters and tests before allowing commits
- **Pre-Push Validation**: Validates commit messages before pushing to remote repositories
- **Baseline-Aware Validation**: All validation respects the May 18, 2025 baseline date
- **Prevents CI Failures**: Ensures your code will pass CI checks before pushing

3. **Note on go-conventionalcommits**:

   The project references `github.com/leodido/go-conventionalcommits` v0.12.0 in the tools.go file, but this is a Go library, not a command-line tool. It's used programmatically within the codebase for parsing conventional commit messages. No manual installation is required as it's managed through Go modules.

   For commit message validation, this project uses different tools in pre-commit hooks and CI.

   For more information about Go libraries vs CLI tools, see our [Go Tool Installation Guide](docs/development/tooling.md).

### Commit Message Format

All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This enables automated semantic versioning and changelog generation. **Non-compliant commits will be rejected by both local hooks and CI.**

**Format:** `<type>[optional scope]: <description>`

**Line length limits:**
- Header (first line): Maximum 72 characters
- Body lines: Maximum 100 characters
- Footer lines: Maximum 100 characters

#### Git Commit Template

This project provides a standardized Git commit template to help you follow the conventional commit format. When you run the setup script (`./scripts/setup.sh`), it automatically configures Git to use this template.

To manually set up the template:
```bash
# Configure Git to use our commit template
./scripts/setup-commit-template.sh

# Or directly with Git
git config commit.template .github/commit-template.txt
```

Then, when you run `git commit` (without the `-m` flag), your editor will open with the template pre-filled, including:
- Conventional commit format examples
- Type definitions with descriptions
- Format guidelines for body and footer
- Breaking change notation examples

**Examples of valid commit messages:**
```
feat: add new file processing module
fix(parser): handle null input correctly
docs: update API documentation
refactor!: rename core service interfaces (breaking change)
test(integration): add coverage for edge cases
chore: update dependencies
```

**Examples of invalid commit messages:**
```
Update code            # Missing type prefix
Fixed bug             # Incorrect format, missing type
feat Add new feature  # Missing colon after type
FEAT: add feature     # Type should be lowercase
```

**Common commit types:**
- `feat`: A new feature (triggers minor version bump)
- `fix`: A bug fix (triggers patch version bump)
- `docs`: Documentation changes (no version bump)
- `style`: Code style changes (formatting, missing semicolons, etc.) (no version bump)
- `refactor`: Code refactoring without changing functionality (no version bump)
- `test`: Adding or modifying tests (no version bump)
- `chore`: Maintenance tasks (updating dependencies, build scripts, etc.) (no version bump)

**Breaking changes:** Add `!` after the type (e.g., `feat!:` or `refactor!:`) or include a `BREAKING CHANGE:` footer to trigger a major version bump.

### Custom Git Hooks Path

This project uses a custom Git hooks path at `.commitlint/hooks` for commit message validation. This is configured automatically via the project's Git configuration. If you need to reset this configuration:

```bash
# View the current hooks path
git config core.hooksPath

# Reset to the project's custom hooks path if needed
git config core.hooksPath .commitlint/hooks
```

### Commit Message Validation Tools

This project uses pre-commit hooks with `commitlint` for local commit message validation. The hooks are automatically run when you commit.

### Enforcement Policies

**MANDATORY ENFORCEMENT:** The following policies are strictly enforced without exception:

1. **Absolutely NO bypassing of pre-commit hooks**: Using `--no-verify` or similar methods to bypass hooks is strictly forbidden and considered a **serious violation** of project standards
2. **CI validation is mandatory**: All commits pushed to the repository will be validated by CI workflows
3. **Conventional Commits are required**: Non-compliant commits will block the automated release process and be rejected
4. **All commits must pass validation**: Every commit in a PR or push must be individually compliant
5. **No exceptions without approval**: Any exceptional circumstances requiring hook bypassing MUST be discussed and approved by project maintainers in advance

**VIOLATIONS:** Pull requests containing commits that have bypassed hooks or do not conform to our standards will be rejected. Contributors who repeatedly bypass hooks may lose contribution privileges.

### Troubleshooting Common Issues

**Issue: Commit rejected due to invalid format**
- Solution: Ensure your commit follows the pattern `<type>[optional scope]: <description>`
- Check that the type is lowercase and from the allowed list
- Include a colon and space after the type/scope

**Issue: Line length exceeded**
- Solution: Keep the first line under 72 characters
- Use the commit body (second paragraph) for additional details

**Issue: Commit message validation failure**
- Solution: Ensure your pre-commit hooks are properly installed with `pre-commit install --hook-type commit-msg`
- Check that your commit message follows the Conventional Commits format

**Issue: Breaking change not detected**
- Solution: Use `!` after the type (e.g., `feat!:`) or include `BREAKING CHANGE:` in the footer

### Automated Release Process

Our project uses automated semantic versioning based on commit messages:

1. `feat` commits trigger minor version bumps (1.x.0 → 1.y.0)
2. `fix` commits trigger patch version bumps (1.2.x → 1.2.y)
3. Breaking changes (`feat!` or `BREAKING CHANGE:`) trigger major version bumps (x.y.z → x+1.0.0)
4. Other commit types (`docs`, `test`, `chore`, etc.) don't affect versioning

The CI/CD pipeline automatically:
- Calculates the next version using `svu` based on commit history
- Generates changelogs with `git-chglog`
- Creates releases with `goreleaser`

This automation depends entirely on proper commit message formatting, which is why enforcement is mandatory.

## Testing Requirements

Testing is a core component of our development process:

- **Test-Driven Development (TDD)**: Write tests before implementing features
- **Coverage Requirements**: Maintain at least 75% overall code coverage, with a 90% target
- **Testing Approaches**:
  - Unit tests for isolated logic
  - Integration tests for component interactions
  - End-to-end tests for critical workflows
- **Mocking Policy**: Only mock true external dependencies, never internal collaborators

See the README section on [Code Coverage Requirements](README.md#code-coverage-requirements) for detailed information.

## Pull Request Process

1. Ensure your code passes all tests and linting checks
2. Update documentation as needed
3. Submit a pull request with a clear description of the changes
4. Wait for reviews and address any feedback

## Project Structure

The project follows a standard Go project layout:

- `/cmd/thinktank`: Main application entry point
- `/internal`: Private application code
- `/docs`: Documentation files
- `/scripts`: Utility scripts for development and CI
- `/config`: Configuration files
- `/vendor`: Vendored dependencies

For more details on our project structure and philosophy, see the [Development Philosophy](docs/DEVELOPMENT_PHILOSOPHY.md) document.

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go FAQ](https://golang.org/doc/faq)
