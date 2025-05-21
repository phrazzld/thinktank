## CI Resolution Tasks (Phase 2)

- [ ] **T049 · Bugfix · P0: fix commit message format issues in PR #24**
    - **Context:** CI Resolution Plan - Primary Issue 1 & 2: Invalid Commit Message Format
    - **Action:**
        1. Create a backup branch of the current `feature/automated-semantic-versioning` branch.
        2. Use interactive rebase to fix the commit with invalid format (missing type/subject).
        3. Fix the commit with footer warning by ensuring a blank line separates body and footer.
        4. Force-push the corrected branch with `--force-with-lease`.
    - **Done‑when:**
        1. All commits in the PR branch follow the Conventional Commits format.
        2. The CI "Validate Commit Messages" check passes.
    - **Verification:**
        1. Run local validation: `npx @commitlint/cli --from=master`.
        2. Verify the GitHub Actions pipeline passes all checks.
    - **Depends‑on:** none

- [ ] **T050 · Feature · P1: implement automated pre-commit hook installation**
    - **Context:** CI Resolution Plan - Prevention Measures: Enforce Pre-commit Hook Installation
    - **Action:**
        1. Modify `scripts/setup.sh` to automatically install pre-commit hooks during project setup.
        2. Update the `Makefile` to include a target for hook installation.
        3. Add validation to ensure hooks are active during key Git operations.
    - **Done‑when:**
        1. Running `scripts/setup.sh` or the relevant Makefile target automatically installs all required Git hooks.
        2. Developers can't easily bypass hook installation.
    - **Verification:**
        1. Fresh clone the repository, run setup, and verify hooks are installed in `.git/hooks/`.
        2. Attempt to create an invalid commit and confirm it's blocked by the hooks.
    - **Depends‑on:** none

- [ ] **T051 · Feature · P1: implement commit message validation for all branch commits**
    - **Context:** CI Resolution Plan - Prevention Measures: Add Local PR Validation
    - **Action:**
        1. Create a script `scripts/validate-pr-commits.sh` that validates all commits from the branch point against commitlint rules.
        2. Make the script executable and add documentation on its usage.
        3. Configure the script to use the same validation rules as the CI workflow.
    - **Done‑when:**
        1. Developers can run a local command to validate their branch's commit history before pushing.
        2. The validation catches the same issues that would fail in CI.
    - **Verification:**
        1. Create a branch with both valid and invalid commits.
        2. Run the validation script and verify it identifies the same issues that CI would catch.
    - **Depends‑on:** none

- [ ] **T052 · Feature · P2: add pre-push hook for commit message validation**
    - **Context:** CI Resolution Plan - Prevention Measures: Pre-push Validation
    - **Action:**
        1. Update `.pre-commit-config.yaml` to add a pre-push hook that validates all new commits.
        2. Configure the hook to run `npx @commitlint/cli --from=origin/master` or equivalent.
        3. Document the new hook in developer documentation.
    - **Done‑when:**
        1. The pre-push hook is properly configured in the project.
        2. Pushing commits with invalid messages is blocked before reaching remote.
    - **Verification:**
        1. Create a branch with an invalid commit message.
        2. Attempt to push the branch and verify the hook blocks the operation.
    - **Depends‑on:** [T050]

- [ ] **T053 · Chore · P2: create repository-wide git commit template**
    - **Context:** CI Resolution Plan - Long-term Improvements: Commit Message Template
    - **Action:**
        1. Create `.github/commit-template.txt` with the conventional commits format and examples.
        2. Update project setup script to configure Git to use this template.
        3. Document how to use the template in CONTRIBUTING.md.
    - **Done‑when:**
        1. The commit template is created and properly configured.
        2. New developers automatically get the template configured during setup.
    - **Verification:**
        1. Run project setup in a fresh clone and verify `git commit` opens the editor with the template pre-populated.
    - **Depends‑on:** none

- [ ] **T054 · Refactor · P3: enhance CI workflow with better error messages**
    - **Context:** CI Resolution Plan - Long-term Improvements: Automated PR Feedback
    - **Action:**
        1. Review and improve error messages in the GitHub Actions workflow.
        2. Configure the commitlint action to provide more specific guidance on fixing commit message issues.
        3. Consider implementing a custom action or script that adds comments to PRs with specific instructions when validation fails.
    - **Done‑when:**
        1. CI failures provide clear, actionable feedback on how to resolve the issue.
        2. Error messages include examples of correct formatting.
    - **Verification:**
        1. Create a PR with an invalid commit and review the error message quality in the CI logs.
    - **Depends‑on:** none

- [ ] **T055 · Feature · P2: implement commitizen for guided commit message creation**
    - **Context:** CI Resolution Plan - Long-term Improvements: Simplified Commit Workflow Tool
    - **Action:**
        1. Add Commitizen to the project's development dependencies.
        2. Create a configuration file (`.czrc` or similar) tailored to the project's commit standards.
        3. Update documentation to recommend using Commitizen for commit creation.
        4. Add a Makefile target or script for easy access.
    - **Done‑when:**
        1. Commitizen is configured and working with the project's commit standards.
        2. Documentation includes instructions for using the tool.
    - **Verification:**
        1. Use the Commitizen CLI to create a commit and verify it passes all validation checks.
    - **Depends‑on:** none

- [ ] **T056 · Chore · P2: create quick reference guide for conventional commits**
    - **Context:** CI Resolution Plan - Prevention Measures: Enhance Documentation
    - **Action:**
        1. Create a concise reference in `docs/conventional-commits-guide.md`.
        2. Include examples of valid commit messages for all common types (feat, fix, docs, etc.).
        3. Document common pitfalls and how to avoid them.
        4. Link to this guide from CONTRIBUTING.md and README.md.
    - **Done‑when:**
        1. The quick reference guide exists and contains clear, helpful examples.
        2. Main documentation references this guide.
    - **Verification:**
        1. Review the guide for clarity and completeness.
        2. Ask a team member to review for usefulness to new contributors.
    - **Depends‑on:** none
