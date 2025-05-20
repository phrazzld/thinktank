# CI Resolution Plan

## Executive Summary

The CI pipeline for PR #24 (feature/automated-semantic-versioning) failed during the commit message validation step due to two issues:

1. **Critical Error**: One commit in the branch history has an invalid format missing the conventional commit type prefix entirely.
2. **Warning**: The most recent commit has a formatting issue in the footer section where a blank line is missing.

These issues violate the project's Conventional Commits specification requirements, which are enforced through the `wagoid/commitlint-github-action@v5` GitHub Action as part of the CI/CD pipeline. The solution requires fixing the commit history to comply with the project's commit message standards.

## Root Cause Analysis

### Primary Issues

1. **Invalid Commit Message**
   - A commit in the branch history was made without adhering to the Conventional Commits format.
   - The commit is missing both the required type (e.g., feat, fix, docs) and subject.
   - Error: `type may not be empty [type-empty]` and `subject may not be empty [subject-empty]`
   - This suggests the commit message was created without using the project's pre-commit hooks.

2. **Commit Footer Format Warning**
   - The most recent commit (`04d7a4c docs: add PR #24 incident details to ci-troubleshooting guide`) has an improperly formatted footer.
   - Warning: `footer must have leading blank line [footer-leading-blank]`
   - The footer needs to be separated from the commit body by a blank line.

### Contributing Factors

1. **Local Hook Bypass or Missing Installation**
   - The pre-commit hooks that should validate commit messages locally were either:
     - Not installed on the developer's machine
     - Bypassed using `--no-verify` flag (against project policy)
     - Failed to execute properly

2. **Lack of Clear Documentation**
   - While hooks are documented in the project, there might be insufficient emphasis on their mandatory nature.
   - The CI-troubleshooting.md document includes information about commit message validation, but developers might not have reviewed it before committing.

3. **Validation Gap**
   - The local pre-commit hook for commit message validation exists (`conventional-commit-check` in .pre-commit-config.yaml) but only validates the current commit, not the entire history that a PR might include.
   - GitHub Actions validates all commits in PR history, creating a validation gap between local and CI checks.

## Resolution Options

### Option 1: Rewrite Commit History

**Approach**: Use Git's interactive rebase to rewrite the problematic commits.

**Pros**:
- Creates a clean, standards-compliant commit history
- Eliminates all validation errors and warnings
- Aligns with project's strict policy on commit quality

**Cons**:
- Requires force-pushing the branch
- Can create issues for collaborators if they have based work on the current branch state
- More complex and risky than non-history-altering approaches

### Option 2: Fix-up Commits

**Approach**: Add "fix-up" commits that correct the issues without rewriting history.

**Pros**:
- Easier and safer than rewriting history
- No force-push required
- Maintains existing commit timestamps and authorship

**Cons**:
- Doesn't actually fix the non-compliant commits
- Would require modifying the CI pipeline to either:
  - Ignore specific non-compliant commits (which undermines the validation purpose)
  - Only validate the most recent N commits (which creates a permanent validation gap)
- Not aligned with the project's commit quality standards

### Option 3: Squash All Commits

**Approach**: Squash all commits in the PR into a single, compliant commit.

**Pros**:
- Creates a single, clean commit
- Simplifies the commit history
- Eliminates all validation issues

**Cons**:
- Loses detailed commit history within the PR
- Requires force-pushing the branch
- Less granular history for future troubleshooting

## Recommended Action Plan

Based on the project's emphasis on commit quality and the nature of the issues, **Option 1 (Rewrite Commit History)** is recommended. This aligns with the project's standards and creates a clean, compliant history.

### Step-by-Step Implementation

1. **Backup the current branch state**:
   ```bash
   # Create a backup branch
   git checkout feature/automated-semantic-versioning
   git branch feature/automated-semantic-versioning-backup
   ```

2. **Identify the problematic commits**:
   ```bash
   # Show all commits in the branch
   git log --pretty=format:"%h %s" master..HEAD

   # Identify the problematic commit(s) from the output
   ```

3. **Perform interactive rebase to fix commit messages**:
   ```bash
   # Start an interactive rebase from the common ancestor with master
   git rebase -i $(git merge-base master HEAD)
   ```

4. **Edit the problematic commits**:
   - For the invalid commit, change `pick` to `reword` (or `r`) in the rebase plan
   - During rebase, provide a properly formatted commit message:
     ```
     feat: implement <specific feature>

     Detailed description of what was changed and why.
     ```
   - For the commit with footer warning, change `pick` to `reword` and ensure there's a blank line between body and footer:
     ```
     docs: add PR #24 incident details to ci-troubleshooting guide

     Added detailed information about the incident and resolution.

     Refs: #24
     ```

5. **Verify the fixed commits**:
   ```bash
   # Verify the commit messages now follow conventional commits format
   git log --pretty=format:"%h %s" master..HEAD

   # Run commit message validation locally
   npx @commitlint/cli --from=master
   ```

6. **Force-push the corrected branch**:
   ```bash
   # Force push the corrected history to the remote
   git push --force-with-lease origin feature/automated-semantic-versioning
   ```

7. **Verify CI pipeline success**:
   - Wait for GitHub Actions to complete
   - Verify the "Validate Commit Messages" step passes

### Implementation Notes

- The force push is necessary when rewriting history. Use `--force-with-lease` rather than `--force` for safety.
- Communicate with any team members who might be working on the same branch to coordinate the history rewrite.
- If the PR is linked to issues via commit messages (e.g., "fixes #123"), ensure these references are preserved during rebase.

## Prevention Measures

To prevent similar issues in the future:

1. **Enforce Pre-commit Hook Installation**:
   - Modify `scripts/setup.sh` to automatically install the pre-commit hooks during project setup
   - Add a pre-push hook that verifies commit message format for all commits being pushed
   - Consider a git template that installs hooks automatically on clone

2. **Enhance Documentation**:
   - Update CONTRIBUTING.md to highlight the critical importance of commit formatting
   - Add explicit warnings about commitlint failures in CI
   - Create a quick reference guide for conventional commits format

3. **Add Local PR Validation**:
   - Create a script that runs the same validation that GitHub Actions will perform
   - Allow developers to verify their branch before pushing:
     ```bash
     ./scripts/validate-pr.sh
     ```

4. **Pre-push Validation**:
   - Add a pre-push hook that validates all new commits before pushing:
     ```yaml
     # In .pre-commit-config.yaml
     -   repo: local
         hooks:
         -   id: validate-commits-for-push
             name: Validate all commits about to be pushed
             entry: bash -c "npx @commitlint/cli --from=origin/master"
             language: system
             stages: [push]
             pass_filenames: false
     ```

5. **Developer Training**:
   - Ensure all team members understand the conventional commits requirements
   - Create internal documentation with examples of good and bad commit messages
   - Hold a brief session on Git best practices for the team

## Long-term Improvements

Beyond addressing the immediate issue, consider these improvements:

1. **Commit Message Template**:
   - Set up a repository-wide commit template with the expected format
   - Include examples and placeholders for type, scope, and description
   - Implement with: `git config commit.template .github/commit-template.txt`

2. **Automated PR Feedback**:
   - Enhance the GitHub Action to provide specific guidance when commits fail validation
   - Add comments to PRs with examples of how to fix common commit message issues

3. **Semantic Release Integration**:
   - Complete the automated semantic versioning workflow (goal of PR #24)
   - Use semantic-release to automatically generate version numbers and changelogs from commit messages
   - This creates tangible benefits from good commit messages, encouraging compliance

4. **Simplified Commit Workflow Tool**:
   - Consider tools like Commitizen that guide developers through creating properly formatted commit messages
   - Integrate with the existing pre-commit setup

## Conclusion

The CI failure in PR #24 stems from commit message formatting issues that violate the project's conventional commits requirements. By rewriting the commit history to fix these issues, we can unblock the pipeline while maintaining high standards for commit quality.

The recommended action plan provides a clear path to resolve the immediate issues, while the prevention measures will help ensure similar problems don't occur in the future. These steps align with the project's emphasis on code quality, automation, and explicit documentation.

Implementation of these recommendations will strengthen the project's commit message validation process end-to-end, from local development through CI/CD pipeline execution.
