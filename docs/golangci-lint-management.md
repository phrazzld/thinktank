# Managing golangci-lint Configuration

This document outlines best practices for updating and managing the golangci-lint configuration in the thinktank project.

## Current Configuration

The project currently uses golangci-lint v2.1.5 with an expanded set of linters configured in `.golangci.yml`. The configuration enables a comprehensive set of linters to catch a wide range of potential issues.

## Version Management

### Workflow Configuration

The golangci-lint version is specified in two GitHub workflow files:

1. `.github/workflows/ci.yml`
2. `.github/workflows/release.yml`

Both files use the same version installation command:

```yaml
- name: Install golangci-lint and run it directly
  run: |
    # Install golangci-lint v2.1.5 directly using go install
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.5
    # Run golangci-lint directly without using the action to avoid --out-format flag issues
    $(go env GOPATH)/bin/golangci-lint run --timeout=5m
```

This approach using `go install` is preferred over the curl-based installation method for better security and consistency with other Go tools.

### Configuration File

The `.golangci.yml` file contains a comprehensive linter configuration with an expanded set of linters:

```yaml
run:
  timeout: 5m
  tests: true
  skip-dirs:
    - vendor/

# Linter selection
linters:
  disable-all: true
  enable:
    # Critical linters - always enabled
    - errcheck      # Detect unchecked errors
    - gosimple      # Simplify code
    - govet         # Go vet examines Go source code for suspicious constructs
    - ineffassign   # Detect ineffectual assignments
    - staticcheck   # Go static analysis tool
    - unused        # Check for unused constants, variables, functions and types

    # Plus many additional linters for code quality
    # ...
```

The configuration includes multiple categories of linters, from essential code correctness checkers to style and security linters. We've also configured specific exclusions for test files to prevent common false positives in test code.

## Updating golangci-lint

When updating golangci-lint, follow these steps:

1. **Research Compatible Versions**
   - Check the [golangci-lint releases](https://github.com/golangci/golangci-lint/releases) for the latest stable version
   - Review the changelog for compatibility concerns
   - Verify configuration format compatibility between versions

2. **Update Local Development Environment**
   - Install the new version locally:
     ```bash
     curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin vX.Y.Z
     ```
   - Run locally to test for compatibility issues:
     ```bash
     golangci-lint run --timeout=5m
     ```

3. **Update Configuration Files**
   - Update `.golangci.yml` if needed to match the new version requirements
   - Ensure the configuration version is compatible with the binary version
   - Test your changes locally before committing

4. **Update Workflow Files**
   - Update the version in BOTH `.github/workflows/ci.yml` and `.github/workflows/release.yml`
   - Ensure both workflows use the same version

5. **Run Tests**
   - Verify that the updated linter works correctly with existing code
   - Fix any new warnings or errors detected by the updated linter

## Troubleshooting Common Issues

### Version Mismatch Between Config and Binary

**Symptoms:**
- Error messages about configuration format incompatibility
- Unexpected linter behavior

**Resolution:**
- Ensure the `version` field in `.golangci.yml` matches the major version of golangci-lint
- For golangci-lint v2.x.x, use `version: "2"`

### Inconsistent Versions Between Workflows

**Symptoms:**
- CI passes but Release workflow fails
- Different linting results in different environments

**Resolution:**
- Synchronize versions between all workflow files
- Use the same installation commands and configuration in all workflows

### New Linters Causing Failures

**Symptoms:**
- After upgrading, builds fail due to new linter rules

**Resolution:**
- Temporarily disable new failing linters using the `enable`/`disable` list
- Gradually fix code to comply with new rules
- Re-enable linters once code is compliant

## Best Practices

1. **Version Pinning**
   - Always pin the exact version (e.g., v2.1.1, not v2)
   - Document the version in both workflow files and configuration

2. **Consistency**
   - Maintain consistency between local, CI, and release workflows
   - Use the same configuration and version everywhere

3. **Gradual Updates**
   - For major version upgrades, consider a phased approach
   - Test thoroughly before committing changes

4. **Documentation**
   - When changing linter configuration, document the reason in the commit message
   - Update this guide when making significant changes to the linting process

## Testing Configuration Changes

To test golangci-lint configuration changes before committing:

1. Make your changes to `.golangci.yml`
2. Run the linter locally:
   ```bash
   golangci-lint run --timeout=5m
   ```
3. Verify the results align with expectations
4. Fix any issues before committing

## Emergency Rollback Procedure

If a golangci-lint update causes critical CI failures:

1. Revert to the previous known-good version in workflow files
2. Restore the previous configuration file
3. Document the issue for future reference
