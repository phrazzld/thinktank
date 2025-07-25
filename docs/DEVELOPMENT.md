# Development Guide

This document describes development practices, tools, and policies for the thinktank project.

## License Policy and Compliance

### Overview

This project enforces strict license compliance to ensure all dependencies use acceptable open-source licenses. License checking is integrated into both the CI pipeline and local development workflow.

### Approved Licenses

The following licenses are approved for use in dependencies:

- **Apache-2.0** - Apache License 2.0
- **BSD-2-Clause** - BSD 2-Clause "Simplified" License
- **BSD-3-Clause** - BSD 3-Clause "New" or "Revised" License
- **MIT** - MIT License
- **ISC** - ISC License (functionally equivalent to MIT)
- **Unlicense** - Public Domain Dedication

### License Checking Tools

#### CI Pipeline (Automated)

License compliance is automatically checked in the CI pipeline as part of the security gates:

- **Tool**: [`go-licenses`](https://github.com/google/go-licenses) v1.6.0
- **Trigger**: Every push and pull request to master branch
- **Behavior**: Hard fail on any forbidden license
- **Reports**: License reports uploaded as CI artifacts (30-day retention)

#### Local Development

Run license checks locally before committing:

```bash
# Basic license check
./scripts/check-licenses.sh

# Verbose output showing all dependencies
./scripts/check-licenses.sh -v

# Generate report without failing (for investigation)
./scripts/check-licenses.sh --report-only --output-file deps.csv
```

#### Pre-commit Hooks

License checking is integrated into pre-commit hooks and runs automatically when `go.mod` or `go.sum` files change:

```bash
# Install pre-commit hooks (one-time setup)
pip install pre-commit
pre-commit install

# Pre-commit hooks will automatically run license check when needed
git commit -m "Update dependencies"  # Will trigger license check if go.mod/go.sum changed
```

### Handling License Violations

When a forbidden license is detected:

1. **Immediate Action**: The build/commit will fail
2. **Investigation**: Review the specific dependency and its license
3. **Resolution Options**:
   - **Replace the dependency** with one that has an approved license
   - **Remove the dependency** if it's not essential
   - **Request approval** to add the license to the allowlist (requires security review)

#### Example: Replacing a Dependency

```bash
# Remove problematic dependency
go mod edit -droprequire example.com/problematic-package

# Add acceptable alternative
go get acceptable.com/alternative-package

# Verify compliance
./scripts/check-licenses.sh

# Update imports in code
# ... make necessary code changes ...

# Commit changes
git add .
git commit -m "Replace problematic dependency with acceptable alternative"
```

### Adding New Licenses to Allowlist

To add a new license to the allowlist (requires approval):

1. **Research**: Verify the license is truly open-source and compatible
2. **Security Review**: Ensure the license doesn't introduce legal or compliance risks
3. **Update Configuration**: Modify both CI and local scripts:
   - `.github/workflows/security-gates.yml` (CI configuration)
   - `scripts/check-licenses.sh` (local script)
4. **Documentation**: Update this document with the new approved license

### Troubleshooting

#### Common Issues

**go-licenses not found**
```bash
# Install go-licenses
go install github.com/google/go-licenses@v1.6.0

# Ensure GOPATH/bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin
```

**Network connectivity issues**
```bash
# Ensure dependencies are downloaded
go mod download

# Try with proxy settings if behind corporate firewall
GOPROXY=https://proxy.golang.org,direct ./scripts/check-licenses.sh
```

**False positives**
- Some packages may not have machine-readable license information
- Manual verification may be required for edge cases
- Report issues to the go-licenses project for improvements

#### Getting Help

- **License Questions**: Contact the security team
- **Tool Issues**: Check the [go-licenses GitHub repository](https://github.com/google/go-licenses)
- **CI Problems**: Review workflow logs in GitHub Actions

## Local Development Workflow

### Quick Setup

```bash
# Install required development tools
make install-tools

# Run quick checks before committing
make quick-check

# Run comprehensive CI simulation before pushing
make ci-check
```

### Make Targets

The project includes a comprehensive Makefile for local development that mirrors the CI pipeline:

| Target | Description | Use Case |
|--------|-------------|----------|
| `make help` | Show all available targets | Getting started |
| `make ci-check` | **Full local CI simulation** | Before important pushes |
| `make quick-check` | Fast checks (format, vet, basic tests) | Before each commit |
| `make pre-push` | Recommended checks before pushing | Daily development |
| `make lint` | All linting (format, vet, golangci-lint, pre-commit) | Code quality focus |
| `make test` | All tests (unit, integration, E2E) | Testing focus |
| `make coverage` | Coverage analysis (90% threshold) | Coverage verification |
| `make security-scan` | Security scans (licenses, secrets) | Security focus |
| `make build` | Build the thinktank binary | Build verification |
| `make fmt` | Format Go code | Code formatting |
| `make install-tools` | Install development dependencies | One-time setup |

### Recommended Development Workflow

1. **Daily Development**:
   ```bash
   # Before committing changes
   make quick-check
   git commit -m "your changes"

   # Before pushing (recommended)
   make pre-push
   git push
   ```

2. **Before Important Pushes**:
   ```bash
   # Full CI simulation (matches GitHub Actions exactly)
   make ci-check
   git push
   ```

3. **Troubleshooting**:
   ```bash
   # Focus on specific areas
   make lint          # Just linting issues
   make test          # Just test failures
   make coverage      # Just coverage problems
   make security-scan # Just security issues
   ```

### Pre-commit Integration

The project uses pre-commit hooks that automatically run when you commit:

- **Fast checks**: File hygiene, Go formatting, basic validation
- **Security**: Secret detection, license compliance
- **Advanced linting**: golangci-lint (same version/config as CI)

**Setup pre-commit hooks**:
```bash
pip install pre-commit
pre-commit install
```

**Manual execution**:
```bash
pre-commit run --all-files  # Run all hooks manually
```

### Development Commands Reference

For additional development commands and detailed CI information, see `./CLAUDE.md`:

- Build and test commands
- Code quality tools (formatting, linting, coverage)
- Pre-commit hook setup
- CI/CD pipeline information

## Code Quality Standards

- **Test Coverage**: Maintain 90%+ coverage across all packages
- **Linting**: All code must pass `golangci-lint` with zero warnings
- **Formatting**: Use `go fmt` for consistent code formatting
- **Dependencies**: Keep dependencies minimal and regularly updated
- **Security**: All commits are scanned for secrets and vulnerabilities

## Related Documentation

- [TESTING.md](./TESTING.md) - Testing practices and infrastructure
- [CLAUDE.md](./CLAUDE.md) - Development commands and CI details
- [Security Documentation](./docs/security/) - Security policies and procedures
