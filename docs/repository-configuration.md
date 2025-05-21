# Repository Configuration Guide

This guide documents the necessary repository settings and configurations for the thinktank project.

## Branch Protection Rules

To ensure the integrity of our CI/CD workflows and maintain code quality, the following branch protection rules should be configured by a repository administrator:

### Main/Master Branch Protection

1. **Go to Repository Settings** → **Branches** → **Add rule**
2. **Branch name pattern:** `main` (or `master` if that's your default branch)
3. **Required protections:**
   - ✅ Require a pull request before merging
   - ✅ Require approvals (minimum 1)
   - ✅ Dismiss stale pull request approvals when new commits are pushed
   - ✅ Require review from CODEOWNERS
   - ✅ Require status checks to pass before merging
   - ✅ Require branches to be up to date before merging
   - ✅ Include administrators (recommended for consistency)

### CODEOWNERS Enforcement

The repository includes a `.github/CODEOWNERS` file that automatically requires reviews for sensitive areas:

- **CI/CD Workflows** (`.github/workflows/`): Changes to workflow files require review from designated maintainers
- This prevents unauthorized modifications to build, test, and release processes

### Required Status Checks

Configure the following status checks as required:
- Lint and Format
- Test
- Build
- Any other critical CI jobs

## Access Management

### Teams and Permissions
- **Maintainers**: Write access + ability to approve PRs
- **Contributors**: Write access (create branches, open PRs)
- **CI/CD Admins**: Admin access (configure workflows, manage secrets)

### GitHub Actions Permissions
1. Go to **Settings** → **Actions** → **General**
2. Set workflow permissions:
   - Read and write permissions (for creating releases, tags)
   - Allow GitHub Actions to create and approve pull requests (if needed)

## Security Settings

### Dependabot
1. Enable Dependabot security updates
2. Enable Dependabot version updates for dependencies

### Secret Scanning
1. Enable secret scanning
2. Enable push protection for secrets

### Code Scanning
1. Enable CodeQL analysis
2. Configure for relevant languages (Go, etc.)

## Implementation Checklist

When setting up a new repository or updating configuration:

- [ ] Configure branch protection rules
- [ ] Verify CODEOWNERS file is in place
- [ ] Set up required status checks
- [ ] Configure team permissions
- [ ] Enable security features (Dependabot, secret scanning)
- [ ] Review and adjust GitHub Actions permissions
- [ ] Document any repository-specific deviations

## Notes

- Some settings require repository admin permissions
- Branch protection rules can be tested by creating a test PR
- CODEOWNERS file changes take effect immediately upon merge
- Consider using rulesets for more advanced protection scenarios
