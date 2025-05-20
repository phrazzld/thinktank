# CI Resolution Plan - Modified Approach

## 1. Root Issue Analysis

Based on the CI failures and the user's direction, we need to implement a solution that validates commit messages without requiring changes to historical commits. This requires a modified approach from the original CI resolution plan.

### Problem Statement

The current CI pipeline is failing because it's validating all commits in the history against the Conventional Commits standard, including commits that were made before this standard was adopted in the project.

### Core Principles for the Modified Approach

1. **Preserve History**: We will not alter or rewrite any historical commit messages.
2. **Forward-Looking Enforcement**: We'll implement validation that only applies to new commits made after the policy was established.
3. **Consistent Standards**: While exempting old commits, we'll maintain strict standards for all new commits.

## 2. Implementation Strategy

### Phase 1: Define a Baseline Reference Point

1. **Identify the Baseline Commit**:
   - Determine the commit where the Conventional Commits policy was officially adopted
   - This could be when the commitlint configuration was added or when the policy was documented
   - Use this commit's SHA or date as the reference point for validation

2. **Configure Commitlint to Use the Baseline**:
   - Update the validation configuration to only check commits after the baseline
   - Example approach: Use `--from=<baseline-commit>` parameter with commitlint
   - Ensure consistent baseline reference across all validation mechanisms

### Phase 2: Update CI Workflow

1. **Modify the Validation Step**:
   - Update `.github/workflows/release.yml` to use the baseline approach
   - Add clear comments documenting the baseline commit and rationale
   - Configure the validation to skip commits before the baseline

2. **Enhance Error Messages**:
   - Ensure CI failure messages clarify that only new commits are being validated
   - Include guidance on how to fix non-compliant commits

### Phase 3: Update Local Tools and Documentation

1. **Configure Pre-commit Hooks**:
   - Update the commit-msg hook to use the same baseline approach
   - Ensure developers aren't blocked by historical non-conforming commits

2. **Create Developer Tools**:
   - Develop scripts that help developers validate their branches using the baseline approach
   - Create tools for checking PR compliance before submission

3. **Update Documentation**:
   - Clearly document the baseline policy in all relevant guides
   - Explain when and why the policy was adopted
   - Provide examples of compliant commits

## 3. Implementation Steps and Tasks

The specific tasks for implementation are defined in the updated TODO.md file and include:

1. Update CI workflow to validate only new commits (T049)
2. Implement automated hook installation with baseline awareness (T050)
3. Create validation scripts for local use (T051)
4. Add pre-push hook with baseline exclusion (T052)
5. Create repository commit template (T053)
6. Enhance CI workflow with better error messages (T054)
7. Implement guided commit creation with Commitizen (T055)
8. Create documentation about conventional commits with baseline policy (T056)

## 4. Verification Strategy

Each implementation task includes specific verification steps, but generally we will:

1. Test that CI properly validates new commits while ignoring historical ones
2. Ensure local tools provide consistent validation with the CI pipeline
3. Verify that documentation clearly explains the approach
4. Test with branches containing both old non-conforming and new commits

## 5. Benefits of This Approach

1. **Non-Disruptive**: Preserves git history without requiring rebase or history rewriting
2. **Clear Standards**: Maintains high standards for new code while being pragmatic about legacy code
3. **Developer-Friendly**: Provides tools and documentation to help developers comply with the standards
4. **Sustainable**: Creates a foundation for long-term commit message quality
