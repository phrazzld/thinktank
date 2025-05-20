# CI Troubleshooting Guide

**Note: This document was created as part of PR #24 to provide troubleshooting guidance for CI issues, particularly around commit message validation using our baseline validation approach.**

This guide provides common troubleshooting steps for CI failures in the thinktank project. When the CI pipeline fails, refer to this guide for diagnostic steps and solutions.

## Common CI Failures

### 1. Commit Message Validation Failures

#### Symptom
```
Commit message validation failed
"<commit message>" does not follow Conventional Commits

✖  subject may not be empty [subject-empty]
✖  type may not be empty [type-empty]
```

#### Diagnosis
The CI pipeline is detecting commit messages that don't follow the Conventional Commits format required by this project.

#### Resolution
1. Follow the pattern: `<type>[optional scope]: <description>`
2. Valid types: `feat`, `fix`, `docs`, `chore`, `test`, `refactor`, `style`, `perf`, `ci`, `build`
3. Use lowercase, no period at end of subject line

**Examples:**
```bash
# Good
git commit -m "feat: add user authentication"
git commit -m "fix(parser): handle edge case in JSON parsing"
git commit -m "docs: update installation instructions"

# Bad
git commit -m "Updated readme"  # No type prefix
git commit -m "FEAT: Add login"  # Uppercase type
git commit -m "fix: bug."  # Period at end
```

#### Baseline Validation Policy

Our project uses a baseline validation policy for commit messages:

1. **Only new commits are validated**: Only commits made after our baseline commit (May 18, 2025, commit hash: `1300e4d675ac087783199f1e608409e6853e589f`) are checked against the Conventional Commits standard.

2. **Historical commits are preserved**: This approach allows us to maintain git history without requiring rebasing or rewriting history.

3. **Troubleshooting baseline validation**:
   - If validation fails, first check if the problematic commit was made after the baseline date
   - If you're getting validation errors for pre-baseline commits, ensure you're using the latest version of our CI workflow
   - You can manually run the validation script locally to check your commits:
     ```bash
     ./scripts/ci/validate-baseline-commits.sh
     ```

4. **Body line length limit**: Lines in the commit body must be less than 100 characters long. For longer text, break it into multiple lines.

For more detailed information on our commit message standards and baseline policy, see [docs/conventional-commits.md](./conventional-commits.md).

### 2. Go Test Failures

#### Symptom
```
FAIL: TestXXX
expected X but got Y
```

#### Diagnosis
Unit tests are failing due to code changes or environment differences.

#### Resolution
1. Run tests locally:
   ```bash
   go test ./...
   go test -v -run=TestSpecificName ./path/to/package
   ```
2. Check if tests depend on specific environment variables
3. Verify test data files are committed
4. Look for race conditions in the tests (run with `-race` flag)
   ```bash
   go test -race ./...
   ```

### 3. Build Failures

#### Symptom
```
go build ./... failed
undefined: <identifier>
cannot find package
```

#### Diagnosis
Code compilation errors, missing dependencies, or import issues.

#### Resolution
1. Run build locally first:
   ```bash
   go build ./...
   ```
2. Check for typos in imports or function names
3. Ensure all dependencies are in `go.mod`:
   ```bash
   go mod tidy
   ```
4. Verify no files were accidentally excluded from commit

## General Troubleshooting Tips

### 1. Check CI Logs
- Click on the failed job in GitHub Actions
- Expand the failed step to see detailed error messages
- Look for the first error, not just the final failure

### 2. Reproduce Locally
Always try to reproduce CI failures locally:
```bash
# Format check
pre-commit run --all-files

# Tests
go test ./...

# Build
go build ./...

# Linting
go vet ./...
```

### 3. Verify Environment
Ensure your local environment matches CI:
```bash
# Check Go version
go version

# Check tool versions
pre-commit --version
```

### 4. Common Quick Fixes

**Before pushing, always run:**
```bash
# Fix formatting
pre-commit run --all-files

# Update dependencies
go mod tidy

# Run tests
go test ./...
```

**Mandatory Pre-Push Checklist:**
- ✅ Pre-commit hooks are installed (`pre-commit install`)
- ✅ Running `pre-commit run --all-files` passes all checks
- ✅ All tests are passing locally (`go test ./...`)
- ✅ Code builds without errors (`go build ./...`)
- ✅ Commit messages follow conventional commit format
- ✅ No sensitive information is being committed

## CI Pipeline Overview

Our CI pipeline includes these main stages:

1. **Checkout Code**
2. **Setup Go Environment**
3. **Commit Message Validation** - Uses our baseline validation approach
4. **Lint and Format** - Runs pre-commit hooks
5. **Run Tests** - Executes all unit and integration tests
6. **Build** - Compiles the application
7. **Calculate Version** - Uses conventional commits to determine the next semantic version

## Getting Help

If you encounter a CI issue not covered here:

1. Check the [GitHub Actions logs](https://github.com/phrazzld/thinktank/actions)
2. Search existing [issues](https://github.com/phrazzld/thinktank/issues)
3. Ask in the development channel
4. Update this guide with the solution once resolved
