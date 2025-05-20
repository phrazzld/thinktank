# CI Resolution Plan for PR #24

## Problem Statement

PR #24 is failing CI checks due to two main issues:

1. **Unsupported Configuration Parameter**:
   - The CI workflow uses `wagoid/commitlint-github-action@v5` with an unsupported `fromRef` parameter
   - This parameter was intended to implement baseline validation but is not recognized by the action

2. **Invalid Commit Messages**:
   - Some commits in the PR history don't follow the conventional commits standard
   - One commit is missing a type prefix entirely
   - Other commits have formatting issues with body length and footer spacing

## Approach Overview

We'll implement a two-part solution:

1. **Fix Immediate Configuration Issue**:
   - Remove the unsupported `fromRef` parameter
   - Add clear documentation in workflow comments and error messages about baseline validation policy

2. **Document Baseline Validation Policy**:
   - Since we cannot use the `fromRef` parameter as intended, we'll clearly document the baseline validation policy
   - Add explicit messaging that commits before baseline date (May 18, 2025) are exempt from validation
   - Add information to CI error messages to guide reviewers

## Implementation Steps

### 1. Fix GitHub Action Configuration

Update `.github/workflows/release.yml` to remove the unsupported parameter:

```yaml
- name: Validate Commit Messages (baseline-aware)
  if: github.event_name == 'pull_request'
  uses: wagoid/commitlint-github-action@v5
  with:
    configFile: .commitlintrc.yml
    # Remove the unsupported fromRef parameter
    helpURL: https://github.com/phrazzld/thinktank/blob/master/docs/conventional-commits.md
    failOnWarnings: false
    failOnErrors: true
  env:
    # Add a custom message that will appear in GitHub checks UI when validation fails
    GITHUB_FAILED_MESSAGE: |
      ❌ Some commit messages don't follow the Conventional Commits standard.

      Note: Only commits made AFTER May 18, 2025 (baseline: 1300e4d) are validated.

      Please fix any invalid commit messages. See docs/conventional-commits.md for:
      - Our baseline validation policy
      - Commit message format requirements
      - Instructions for fixing commit messages
```

### 2. Update Commit Messages

Fix the recent commits with formatting issues:

1. For `docs: add PR #24 incident details to ci-troubleshooting guide`:
   - Add blank line before any footer sections
   - Use `git commit --amend` to fix the most recent commit

2. For `feat(ci): update commit validation to use baseline commit`:
   - Ensure body lines are under 100 characters
   - Split long lines or reword as needed
   - Use interactive rebase to fix this commit

### 3. Document Baseline Policy

Ensure `docs/conventional-commits.md` clearly explains our baseline policy:

```markdown
> **⚠️ Important:** This project uses a baseline validation policy. Commit messages are only validated **after** our baseline commit:
>
> **Baseline Commit:** `1300e4d675ac087783199f1e608409e6853e589f` (May 18, 2025)
>
> This allows us to preserve git history while enforcing standards for all new development.
```

### 4. Update PR Description

Add a note to the PR description for reviewers:

```markdown
## Note on Commit Messages

This PR includes commits made before our baseline date (May 18, 2025) that don't follow the conventional commits standard. Only commits made after the baseline commit `1300e4d` are required to follow the standard.

The CI configuration has been updated to document this policy, but technical limitations prevent automatic exclusion of pre-baseline commits from validation.

Please focus review on commits made after May 18, 2025.
```

## Future Improvements

Once PR #24 is merged, we'll implement a more robust solution:

1. **Script-Based Validation**:
   - Create a custom script that only validates commits after the baseline date
   - Configure the workflow to use this script instead of direct commitlint validation

2. **Extend commitlint-github-action**:
   - Consider forking and extending the action to support the `fromRef` parameter
   - Submit a PR to the original action to add this functionality

3. **Pre-Push Validation**:
   - Enhance the pre-push hook to validate commits against the baseline policy
   - This will catch issues before they reach CI

## Expected Outcome

After implementing these changes:

1. CI checks will still identify the non-compliant historical commits
2. Documentation and clear error messages will guide reviewers to understand the baseline policy
3. PR can be approved and merged despite historical commit message issues
4. Future PRs will benefit from clearer guidance on commit message standards

## Assigned Tasks

1. Fix GitHub Action configuration (remove unsupported parameter)
2. Fix formatting in recent commits
3. Enhance documentation of baseline policy
4. Update PR description with relevant information
5. Plan for future improvements to implement technical baseline validation
