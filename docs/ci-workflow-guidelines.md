# CI Workflow and Tools Guidelines

This document provides guidelines for developers who need to add new GitHub Actions workflows or CI tools to the thinktank project.

## Table of Contents

- [Current CI Structure](#current-ci-structure)
- [Guidelines for Adding New Workflows](#guidelines-for-adding-new-workflows)
- [Guidelines for Adding New CI Tools](#guidelines-for-adding-new-ci-tools)
- [Testing and Validation](#testing-and-validation)
- [Documentation Requirements](#documentation-requirements)
- [Examples](#examples)

## Current CI Structure

### Workflow Files

The project currently has two main GitHub Actions workflow files:

1. **ci.yml** - Main continuous integration workflow
   - Triggered on pushes to `master`, pull requests to `master`, and manual dispatch
   - Contains jobs for linting, testing, building, and version calculation
   - Includes extensive caching for Go modules and build outputs

2. **release.yml** - Release workflow
   - Triggered on pushes to `master`, tags starting with 'v', and pull requests to `master`
   - Contains jobs for CI checks and creating releases
   - Uses GoReleaser for automated release creation

### CI Jobs and Dependencies

The CI workflow contains the following jobs with dependencies:

```
lint → test → build → version
```

The release workflow contains:

```
ci_checks → release
```

### Custom CI Scripts

Custom scripts in `scripts/ci/` directory handle specific CI tasks:

- `validate-baseline-commits.sh` - Validates commit messages following a baseline approach
- `check-defaults.sh` - Ensures default model configuration is consistent
- `check-package-specific-coverage.sh` - Checks code coverage for specific packages
- `verify-secret-tests.sh` - Verifies that secret detection tests run and pass
- `break-secret-test.sh` - Helper script to test secret detection validation

## Guidelines for Adding New Workflows

### When to Create a New Workflow

Create a new workflow when:

- The task is significantly different from existing workflows (e.g., deployment vs. testing)
- The task has different triggering events (e.g., scheduled runs vs. push events)
- The task requires different permissions or environments
- Keeping the task separate improves maintainability and clarity

Extend existing workflows when:

- The task is closely related to existing jobs
- The task follows the same dependency chain as existing jobs
- The task uses similar configurations and settings

### Workflow File Structure and Naming

- **Location**: Place all workflow files in `.github/workflows/`
- **Naming**: Use descriptive names with `.yml` extension (e.g., `security-scan.yml`, `deployment.yml`)
- **Format**: Use YAML format with proper indentation (2 spaces)
- **Structure**: Organize the file with:
  - Clear workflow name at the top
  - Well-defined triggers
  - Logically grouped jobs
  - Explicit job dependencies

### Required Elements

Every workflow file must include:

1. **Name**: A clear, descriptive name for the workflow
   ```yaml
   name: Security Scan
   ```

2. **Triggers**: Specific events that trigger the workflow
   ```yaml
   on:
     push:
       branches: [ master ]
     pull_request:
       branches: [ master ]
     # For scheduled workflows
     schedule:
       - cron: '0 0 * * 0'  # Weekly on Sundays
   ```

3. **Jobs**: Well-defined jobs with:
   - Descriptive names
   - Runner specification
   - Dependencies on other jobs
   - Conditional execution if applicable
   ```yaml
   jobs:
     scan:
       name: Security Scan
       runs-on: ubuntu-latest
       # If dependent on another job
       needs: lint
       # Conditional execution
       if: success()
   ```

4. **Steps**: Clear step names and well-defined actions
   ```yaml
   steps:
     - name: Checkout code
       uses: actions/checkout@v4
       with:
         fetch-depth: 0
   ```

### Best Practices for Workflow Design

1. **Modularity**: Break workflows into logical, independent jobs
2. **Step Reuse**: Avoid duplicating steps across jobs; refactor common steps
3. **Caching**: Implement caching for dependencies, builds, and test results
   ```yaml
   - name: Cache Go build cache
     id: go-cache
     uses: actions/cache@v3
     with:
       path: ~/.cache/go-build
       key: ${{ runner.os }}-go-build-${{ hashFiles('**/*.go') }}
       restore-keys: |
         ${{ runner.os }}-go-build-
   ```
4. **Timeout Limits**: Set appropriate timeout limits for long-running steps
   ```yaml
   - name: Run tests
     run: go test ./...
     timeout-minutes: 10
   ```
5. **Fail Fast**: Use job dependencies and conditional execution to fail fast
6. **Artifact Management**: Upload artifacts with appropriate retention periods
   ```yaml
   - name: Upload coverage report
     uses: actions/upload-artifact@v4
     with:
       name: coverage-report
       path: coverage.out
       retention-days: 14
   ```
7. **Environment Variables**: Use environment variables for reusable values
8. **Permissions**: Set minimum required permissions for security
   ```yaml
   permissions:
     contents: read
     # Add other required permissions
   ```

### Integration with Existing Workflows

1. **Dependency Chain**: Add new jobs to the appropriate place in the dependency chain
2. **Shared Resources**: Be aware of shared resources like artifacts or caches
3. **Naming Conventions**: Use consistent naming for similar jobs across workflows
4. **Version Alignment**: Use the same action versions as existing workflows

## Guidelines for Adding New CI Tools

### Criteria for Adding New Tools

Before adding a new CI tool, ensure it meets these criteria:

1. **Necessity**: The tool addresses a specific need not covered by existing tools
2. **Maintenance**: The tool is actively maintained and has a supportive community
3. **Reliability**: The tool has a proven track record of stability
4. **Performance**: The tool won't significantly slow down the CI pipeline
5. **Compatibility**: The tool works with the project's language, framework, and CI environment

### Version Pinning Requirements

**ALWAYS pin tool versions** to ensure consistent CI behavior:

1. For actions, use exact versions (not `latest` or major/minor version tags):
   ```yaml
   uses: actions/checkout@v4   # Good
   uses: actions/checkout@v4.1.1  # Better
   uses: actions/checkout@main  # BAD - Avoid
   ```

2. For installed tools, specify exact versions:
   ```yaml
   # Good - using exact version
   run: go install github.com/caarlos0/svu@v3.2.3

   # Bad - using latest
   run: go install github.com/caarlos0/svu@latest
   ```

3. Update all occurrences of a tool when updating its version to maintain consistency

### Installation Methods

1. **Go Tools**: Install Go tools with `go install` and exact version:
   ```yaml
   run: go install github.com/caarlos0/svu@v3.2.3
   ```

2. **Direct Download**: Use curl or wget with version-specific URLs:
   ```yaml
   run: |
     curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.1.1
   ```

3. **GitHub Actions**: Use action marketplace with exact versions:
   ```yaml
   uses: goreleaser/goreleaser-action@v5
   with:
     install-only: true
   ```

4. **Package Managers**: Use specific versions with package managers:
   ```yaml
   run: pip install pre-commit==4.2.0
   ```

### Caching Strategies

1. **Tool Caching**: Cache installed tools when possible:
   ```yaml
   - uses: actions/cache@v3
     with:
       path: ${{ runner.tool_cache }}/go
       key: ${{ runner.os }}-go-${{ env.GO_VERSION }}
   ```

2. **Output Caching**: Cache tool outputs to speed up workflows:
   ```yaml
   - name: Cache linter results
     uses: actions/cache@v3
     with:
       path: ~/.cache/golangci-lint
       key: ${{ runner.os }}-golangci-lint-${{ hashFiles('.golangci.yml') }}-${{ hashFiles('**/*.go') }}
   ```

### Security Considerations

1. **Action Authentication**: Be cautious when actions require authentication tokens
2. **Permissions**: Set minimum required permissions for GitHub tokens
3. **Vulnerability Scanning**: Regularly check for vulnerabilities in tools
4. **Script Validation**: Review code in any external scripts before use
5. **Input Validation**: Be cautious about how dynamic inputs are processed by tools

## Testing and Validation

### Testing Workflows Locally

1. **act**: Use [nektos/act](https://github.com/nektos/act) to test GitHub Actions locally
   ```bash
   # Install act
   brew install act

   # Run a specific workflow
   act -W .github/workflows/ci.yml

   # Run a specific job
   act -W .github/workflows/ci.yml -j lint
   ```

2. **Manual Testing**: Test individual scripts or tools directly
   ```bash
   # Test a script
   ./scripts/ci/validate-baseline-commits.sh
   ```

### Ensuring Backward Compatibility

1. **Test with Existing PRs**: Ensure your changes don't break existing processes
2. **Staged Rollout**: Consider running old and new processes in parallel initially
3. **Branching Strategy**: Implement major changes in a separate branch first

### Validating Changes

1. **Create a Test PR**: Create a test PR to verify the workflow changes work as expected
2. **Verify Logs**: Check workflow logs thoroughly for errors or warnings
3. **Edge Cases**: Test edge cases like empty files, large files, or special characters

## Documentation Requirements

### What to Document

When adding new tools or workflows, document:

1. **Purpose**: What the tool or workflow does and why it's needed
2. **Configuration**: How the tool is configured and what options are set
3. **Integration**: How it integrates with existing CI processes
4. **Triggers**: What events trigger the workflow
5. **Failure Handling**: What happens when the tool or workflow fails
6. **Troubleshooting**: Common issues and how to resolve them

### Where to Add Documentation

1. **Code Comments**: Add detailed comments in workflow files and scripts
2. **Dedicated Document**: Create a dedicated document for complex workflows
3. **CI Troubleshooting Guide**: Update the CI troubleshooting guide with new information
4. **README Updates**: Update the README if necessary with new CI information

## Examples

### Example: Adding a New Lint Tool

This example shows how to add a new linting tool called `staticcheck` to the CI pipeline:

1. **Update the workflow file**:

```yaml
# In .github/workflows/ci.yml
- name: Install and run staticcheck
  run: |
    # Install staticcheck at a specific version
    go install honnef.co/go/tools/cmd/staticcheck@v2023.1.6

    # Run staticcheck with a specific config
    staticcheck -checks=all,-ST1000,-ST1005 ./...
  timeout-minutes: 2
```

2. **Add a cache for performance**:

```yaml
- name: Cache staticcheck results
  uses: actions/cache@v3
  with:
    path: ~/.cache/staticcheck
    key: ${{ runner.os }}-staticcheck-${{ hashFiles('**/*.go') }}
```

3. **Update documentation**:

Add information about the new tool to the CI troubleshooting guide and update this document if needed.

### Example: Adding a New Security Scan Workflow

This example shows how to add a new security scanning workflow:

1. **Create a new workflow file**:

```yaml
# .github/workflows/security-scan.yml
name: Security Scan

on:
  # Run weekly
  schedule:
    - cron: '0 0 * * 0'
  # Allow manual trigger
  workflow_dispatch:
  # Run on PRs to master
  pull_request:
    branches: [ master ]

jobs:
  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    permissions:
      contents: read
      security-events: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: 'stable'
          cache: true
          cache-dependency-path: go.sum

      - name: Install govulncheck
        run: go install golang.org/x/vuln/cmd/govulncheck@v1.0.1

      - name: Run vulnerability scan
        run: govulncheck ./...

      - name: Run SAST scan
        uses: github/codeql-action/analyze@v2
        with:
          languages: go
```

2. **Document the new workflow**:

Create a document explaining the purpose, configuration, and usage of the security scan workflow.

---

## Conclusion

Following these guidelines will ensure that new workflows and CI tools integrate seamlessly with the existing CI infrastructure, maintain stability, and provide consistent feedback to developers. Before making changes to the CI system, consult with the team and ensure your changes align with the project's development philosophy and quality standards.
