# SVU Integration Guide

This document describes how the [SVU (Semantic Version Utility)](https://github.com/caarlos0/svu) is integrated and used in the thinktank project.

## Installation

SVU v3.x requires the correct module path when installing with Go:

```bash
go install github.com/caarlos0/svu/v3@v3.2.3
```

Note the `/v3` in the module path, which is required for Go modules with major version 2+.

## Usage in CI/CD

SVU is used in both the CI and Release workflows to calculate the next semantic version based on conventional commits:

1. In `.github/workflows/ci.yml`, SVU is used to:
   - Calculate the next version based on conventional commits
   - Output the version information for logging purposes

2. In `.github/workflows/release.yml`, SVU is used to:
   - Calculate version information for PRs and snapshot builds
   - Determine the appropriate version tag for releases

## Version Calculation Rules

SVU follows these rules when calculating the next version:

1. If no tags exist, start with either v0.0.0 or v0.1.0 depending on commit type
2. For "fix:" commits, increment the patch version (X.Y.Z → X.Y.Z+1)
3. For "feat:" commits, increment the minor version (X.Y.Z → X.Y+1.0)
4. For commits with "!" or "BREAKING CHANGE", increment the major version (X.Y.Z → X+1.0.0)
5. For other commit types (docs, chore, etc.), use the lowest applicable increment

## Testing SVU Integration

The script `scripts/test-svu-versioning.sh` provides automated tests for SVU version calculation rules, ensuring that:

1. Version calculation works as expected for different commit types
2. The correct increments are applied based on conventional commits
3. Breaking changes are correctly identified and processed
4. Multiple commits are evaluated correctly with the highest version increment winning

## Troubleshooting

Common issues with SVU integration:

1. **Installation Errors**: Ensure you use the correct module path with the major version in the path:
   ```
   github.com/caarlos0/svu/v3@v3.2.3
   ```

2. **Version Calculation Issues**:
   - Verify that commit messages follow the Conventional Commits specification
   - Run the test script to verify expected calculation behavior
   - Use `svu current` and `svu next` locally to debug version calculations

3. **CI Integration Issues**:
   - Ensure all workflows use the same SVU version
   - Verify that git history is properly fetched (use `fetch-depth: 0` in checkout action)
   - Check for fetch errors in CI logs
