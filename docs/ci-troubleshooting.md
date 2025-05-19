# CI Troubleshooting Guide

This guide provides common troubleshooting steps for CI failures in the thinktank project. When the CI pipeline fails, refer to this guide for diagnostic steps and solutions.

## Common CI Failures

### 1. Go Tool Installation Issues

#### Symptom
```
Error: failed to install go tools: exit status 1
error obtaining VCS status: exit status 128
package <package> is not a main package
```

#### Diagnosis
This typically occurs when attempting to `go install` a Go library instead of a CLI tool. Go libraries cannot be installed as executables - only packages with a `main` function can.

#### Resolution
1. Verify the package is actually a CLI tool (contains a `main` package)
2. Check if you should be adding it as a dependency in `go.mod` instead
3. If it's a library being used in testing/linting, ensure it's properly referenced in tool configuration files

**Example Issue:** The `go-conventionalcommits` package is a library, not a CLI tool. It should not be installed via `go install`.

### 2. Formatting Violations

#### Symptom
```
Formatting check failed
docs/ci-resolution-status.md: no newline at end of file
*.sh: trailing whitespace
```

#### Diagnosis
Pre-commit hooks detect formatting issues that weren't fixed locally before committing.

#### Resolution
1. Run pre-commit hooks locally:
   ```bash
   pre-commit run --all-files
   ```
2. Accept the automatic fixes
3. Commit the formatting changes:
   ```bash
   git add -A
   git commit -m "fix: apply formatting fixes"
   ```

**Common Issues:**
- Missing EOF newlines in markdown files
- Trailing whitespace in shell scripts or YAML files
- Inconsistent indentation

### 3. Invalid Commit Messages

#### Symptom
```
Commit message validation failed
"<commit message>" does not follow Conventional Commits
```

#### Diagnosis
Commit messages don't follow the Conventional Commits specification required by this project.

#### Resolution
1. Follow the pattern: `<type>[optional scope]: <description>`
2. Valid types: `feat`, `fix`, `docs`, `chore`, `test`, `refactor`, `style`, `perf`
3. Use lowercase, no period at end of subject line

**Examples:**
```bash
# Good
git commit -m "feat: add user authentication"
git commit -m "fix: resolve memory leak in cache handler"
git commit -m "docs: update installation instructions"

# Bad
git commit -m "Updated readme"  # No type prefix
git commit -m "FEAT: Add login"  # Uppercase type
git commit -m "fix: bug."  # Period at end
```

### 4. Build Failures

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

### 5. Test Failures

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
   go test -v -run TestSpecificName ./path/to/package
   ```
2. Check if tests depend on specific environment variables
3. Verify test data files are committed
4. Consider if the test needs updating for new behavior

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

## CI Pipeline Overview

Our CI pipeline includes these main stages:

1. **Checkout Code**
2. **Setup Go Environment**
3. **Install Tools**
4. **Lint and Format** - Runs pre-commit hooks
5. **Run Tests** - Executes all unit and integration tests
6. **Build** - Compiles the application
7. **Release** (on tags) - Creates releases using goreleaser

## Related Documentation

- [Contributing Guide](../CONTRIBUTING.md) - Development setup and standards
- [Go Tool Installation Guide](development/tooling.md) - Go libraries vs CLI tools
- [Pre-commit Configuration](../.pre-commit-config.yaml) - Formatting rules

## Getting Help

If you encounter a CI issue not covered here:

1. Check the [GitHub Actions logs](https://github.com/your-org/thinktank/actions)
2. Search existing [issues](https://github.com/your-org/thinktank/issues)
3. Ask in the development channel
4. Update this guide with the solution once resolved

Remember: CI failures are usually due to issues that can be caught locally. Always run pre-commit hooks and tests before pushing!