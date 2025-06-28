# Dependency Updates Configuration

This document describes the automated dependency update system implemented for this repository.

## Overview

The repository uses Dependabot for automatic dependency updates with intelligent auto-merge capabilities for secure, low-risk updates.

## Configuration Files

### `.github/dependabot.yml`
- Configures Dependabot to check for Go module updates weekly
- Groups patch/minor updates together for easier review
- Separates major updates for careful manual review
- Labels PRs appropriately for filtering and automation

### `.github/workflows/dependency-updates.yml`
- Runs full quality gate suite on all Dependabot PRs
- Implements intelligent auto-merge for eligible updates
- Provides detailed status reporting and audit trails

## Auto-merge Criteria

### Eligible for Auto-merge
Updates that meet ALL of the following criteria are automatically merged:

1. **Source**: Created by Dependabot
2. **Update Type**: Patch version updates (`x.y.z` → `x.y.z+1`) OR security updates
3. **Quality Gates**: All quality gates must pass:
   - ✅ Code quality (lint, format, vet)
   - ✅ Security scans (secrets, licenses, SAST)
   - ✅ Test suite (unit, integration, E2E)
   - ✅ Build verification
   - ✅ Coverage thresholds

### Requires Manual Review
- **Major updates** (`x` → `x+1`): Always require manual review
- **Minor updates** (`x.y` → `x.y+1`): Require manual review (configurable)
- **Updates with failing quality gates**: Cannot be auto-merged

## Repository Setup Requirements

To enable auto-merge functionality, ensure the following repository settings are configured:

### 1. Enable Auto-merge (Required)
```
Settings → General → Pull Requests → Allow auto-merge
```

### 2. Branch Protection Rules (Recommended)
Configure branch protection for `master` branch:

```
Settings → Branches → Add rule for "master":
☑️ Require status checks to pass before merging
☑️ Require branches to be up to date before merging
☑️ Status checks that are required:
   - lint
   - vulnerability-scan
   - test
   - secret-scan
   - license-scan
   - sast-scan
☑️ Require conversation resolution before merging
☑️ Include administrators (recommended)
```

### 3. Repository Permissions
The workflow requires the following permissions (already configured):
- `contents: write` - For merging PRs
- `pull-requests: write` - For managing PR labels and comments
- `checks: read` - For checking CI status
- `actions: read` - For workflow coordination

## Security Model

### Security Update Priority
- **Security patches** are automatically merged regardless of semantic version level
- Security updates are identified by:
  - Dependabot security advisory mentions
  - CVE references in PR description
  - Security-related labels applied by Dependabot

### Audit Trail
Every auto-merge action creates a comprehensive audit trail including:
- Dependency name and version change
- Update type classification (patch/minor/major/security)
- Quality gate results
- Auto-merge decision rationale
- Timestamp and workflow run links

### Quality Gate Enforcement
Auto-merge is **never** allowed if any quality gate fails:
- Code fails to compile
- Tests fail
- Security vulnerabilities detected
- Coverage drops below threshold
- Linting errors present

## Monitoring and Troubleshooting

### Workflow Status
Monitor dependency updates via:
- GitHub Actions tab: `Dependency Updates` workflow runs
- PR labels: `auto-merge-eligible`, `requires-review`
- PR comments: Automated status reports

### Common Issues

#### Auto-merge Not Working
1. **Check repository settings**: Ensure auto-merge is enabled in repository settings
2. **Verify branch protection**: Status checks must be required for auto-merge to work
3. **Review workflow logs**: Check for specific error messages in workflow output
4. **Check permissions**: Workflow must have write permissions for contents and PRs

#### False Security Classification
If a regular update is incorrectly identified as a security update:
1. Remove the `security-update` label from the PR
2. The workflow will re-evaluate auto-merge eligibility
3. Consider updating the security detection logic in the workflow

#### Quality Gates Failing
If legitimate updates fail quality gates:
1. Review the specific failing gate in the workflow logs
2. Fix the underlying issue (update tests, fix security findings, etc.)
3. The PR will automatically re-evaluate for auto-merge after fixes

### Manual Override
Administrators can force-merge any Dependabot PR using:
```bash
# Trigger workflow with force merge option
gh workflow run dependency-updates.yml -f force_merge=true
```

## Configuration Customization

### Adjusting Auto-merge Criteria
To modify what updates are eligible for auto-merge, edit the logic in
`.github/workflows/dependency-updates.yml` in the `analyze-dependabot-pr` job:

```yaml
# Example: Allow minor updates for auto-merge
canAutoMerge = isSecurityPatch || updateType === 'patch' || updateType === 'minor';
```

### Changing Update Frequency
Modify the schedule in `.github/dependabot.yml`:

```yaml
schedule:
  interval: "daily"  # or "monthly"
  day: "tuesday"     # for weekly updates
```

### Adding Dependencies to Ignore List
Add problematic dependencies to the ignore list in `.github/dependabot.yml`:

```yaml
ignore:
  - dependency-name: "problematic-package"
    update-types: ["version-update:semver-major"]
```

## Security Considerations

- **Minimal trust model**: Auto-merge only applies to low-risk updates that pass all quality gates
- **Audit transparency**: Every auto-merge action is logged and traceable
- **Fail-safe defaults**: When in doubt, the system requires manual review
- **Quality gate enforcement**: Security scans must pass before any merge
- **Branch protection**: Repository settings enforce additional review requirements

## Maintenance

### Weekly Review
- Review auto-merged PRs for any unexpected issues
- Check for any Dependabot PRs that failed quality gates
- Monitor security updates for successful deployment

### Monthly Review
- Review ignored dependencies for continued relevance
- Assess auto-merge criteria for effectiveness
- Update security detection patterns if needed

### Quarterly Review
- Evaluate update frequency and timing
- Review repository access permissions
- Update documentation for any process changes
