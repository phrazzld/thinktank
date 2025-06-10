# Dependency Updates Guide

This document explains how to update pinned CI tool dependencies in this project.

## Overview

For security and reproducibility, our CI workflows pin specific versions of security tools rather than using `@latest`. This prevents supply chain attacks and ensures consistent builds across environments.

## Pinned Dependencies

### CI Workflow (.github/workflows/ci.yml)

- **govulncheck**: `golang.org/x/vuln/cmd/govulncheck@v1.0.4`
  - Purpose: Go vulnerability scanning
  - Current version: v1.0.4
  - Location: Line 241

### Security Gates Workflow (.github/workflows/security-gates.yml)

- **go-licenses**: `github.com/google/go-licenses@v1.6.0`
  - Purpose: License compliance checking
  - Current version: v1.6.0
  - Location: Line 144

- **gosec**: `github.com/securego/gosec/v2/cmd/gosec@v2.20.0`
  - Purpose: Static Application Security Testing (SAST)
  - Current version: v2.20.0
  - Location: Line 213

## Update Process

### 1. Check for New Versions

Before updating, check the official repositories for new releases:

```bash
# Check govulncheck releases
go list -m -versions golang.org/x/vuln/cmd/govulncheck

# Check go-licenses releases
curl -s https://api.github.com/repos/google/go-licenses/releases/latest | jq -r .tag_name

# Check gosec releases
curl -s https://api.github.com/repos/securego/gosec/releases/latest | jq -r .tag_name
```

### 2. Test Locally

Before updating CI, test the new versions locally:

```bash
# Test govulncheck
go install golang.org/x/vuln/cmd/govulncheck@vX.Y.Z
govulncheck -scan=module ./...

# Test go-licenses
go install github.com/google/go-licenses@vX.Y.Z
go-licenses csv .

# Test gosec
go install github.com/securego/gosec/v2/cmd/gosec@vX.Y.Z
gosec -severity medium ./...
```

### 3. Update Workflow Files

Update the version strings in the workflow files:

1. Edit `.github/workflows/ci.yml` and update the govulncheck version
2. Edit `.github/workflows/security-gates.yml` and update go-licenses and gosec versions
3. Update this documentation with the new versions

### 4. Update This Documentation

When updating versions, remember to:

1. Update the "Pinned Dependencies" section above with new version numbers
2. Update the "Current version" entries
3. Update the line numbers if they've changed

### 5. Test in CI

Create a pull request with the changes and verify that:

1. All CI jobs pass successfully
2. Tools run without errors
3. No new vulnerabilities or issues are introduced by the tool updates

## Security Considerations

- Always verify tool integrity when updating versions
- Check release notes for breaking changes
- Test thoroughly before merging dependency updates
- Consider using checksums for additional verification when available

## Rollback Procedure

If a new version causes issues:

1. Revert the version changes in workflow files
2. Create a hotfix PR with the rollback
3. Investigate the issue and plan a proper fix
4. Update documentation if the rollback becomes permanent

## Schedule

Review and update dependencies quarterly or when:

- Security vulnerabilities are reported in current versions
- New features are needed from updated tools
- Current versions become deprecated or unsupported
