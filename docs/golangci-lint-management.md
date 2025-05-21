# Managing golangci-lint Configuration

This document outlines best practices for updating and managing the golangci-lint configuration in the thinktank project.

## Current Configuration

The project currently uses golangci-lint v2.1.1 with an explicit configuration version set to "2" in `.golangci.yml`.

## Version Management

### Workflow Configuration

The golangci-lint version is specified in two GitHub workflow files:

1. `.github/workflows/ci.yml`
2. `.github/workflows/release.yml`

Both files use the same version installation command:

```yaml
- name: Install golangci-lint and run it directly
  run: |
    # Install golangci-lint v2.1.1 directly
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.1
    # Run golangci-lint directly without using the action to avoid --out-format flag issues
    $(go env GOPATH)/bin/golangci-lint run --timeout=5m
```

### Configuration File

The `.golangci.yml` file contains the linter configuration. The version is specified at the top of the file:

```yaml
# Specify the configuration version explicitly
version: "2"
```

This version must be compatible with the golangci-lint binary version.

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
