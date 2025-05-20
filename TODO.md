# Todo

## Tooling & Developer Environment
- [x] **T001 · Feature · P0: create tools.go to pin Go tool dependencies**
    - **Context:** PLAN.md "Tool Installation & Pinning"
    - **Action:**
        1. Add `tools.go` in project root or `/tools` directory listing all required tool imports for `go install`.
        2. Run `go mod tidy` to ensure dependencies are tracked in `go.mod`/`go.sum`.
    - **Done‑when:**
        1. `tools.go` exists, imports all required tool modules, and is tracked by Go modules.
    - **Verification:**
        1. Run `go install` on each tool import path and verify installation.
    - **Depends‑on:** none

- [x] **T002 · Feature · P1: add Makefile target for Go tool installation**
    - **Context:** PLAN.md "Tool Installation & Pinning"
    - **Action:**
        1. Create (or update) `Makefile` with `tools` target that runs `go install` for each tool from `tools.go`.
        2. Comment `Makefile` targets for clarity.
    - **Done‑when:**
        1. `make tools` installs all required CLI tools at pinned versions.
    - **Verification:**
        1. Run `make tools` in a fresh shell; all tools become available in `$GOPATH/bin` or `$PATH`.
    - **Depends‑on:** [T001]

- [x] **T003 · Feature · P2: document local tooling setup in README.md and/or CONTRIBUTING.md**
    - **Context:** PLAN.md "Tool Installation & Pinning", "Documentation Updates"
    - **Action:**
        1. Add clear instructions for running `make tools` and setting up Go toolchain.
        2. Reference `tools.go` and required Go version.
    - **Done‑when:**
        1. New developers can follow docs to set up required tooling with no missing steps.
    - **Verification:**
        1. Fresh clone, follow setup steps, and verify all tools are installed.
    - **Depends‑on:** [T002]

## Commit Message Enforcement & Hooks
- [x] **T004 · Feature · P0: add go-conventionalcommits to pre-commit hooks**
    - **Context:** "Commit Message Enforcement"
    - **Action:**
        1. Update the project's `.pre-commit-config.yaml` file to add a local hook for the commit-msg stage.
        2. Configure the hook to run `go-conventionalcommits validate --message-file .git/COMMIT_EDITMSG`
        3. Ensure the hook is set to run at the `commit-msg` stage.
    - **Done‑when:**
        1. The `.pre-commit-config.yaml` file includes the new hook configuration.
        2. The pre-commit hook is active and operating at the commit-msg stage.
    - **Verification:**
        1. Run `pre-commit install --hook-type commit-msg` to ensure the hook is installed.
        2. Attempt a commit to verify the hook is triggered.
    - **Depends‑on:** [T002]

- [x] **T005 · Feature · P0: verify pre-commit hook installation**
    - **Context:** "Commit Message Enforcement"
    - **Action:**
        1. Ensure pre-commit is properly installed and configured in the project.
        2. Verify the commit-msg hook is correctly installed in the Git hooks directory.
    - **Done‑when:**
        1. Running `pre-commit install --hook-type commit-msg` completes successfully.
        2. The `.git/hooks/commit-msg` file exists and contains pre-commit execution code.
    - **Verification:**
        1. Run `pre-commit info` to confirm all hooks are correctly installed.
    - **Depends‑on:** [T004]

- [x] **T006 · Feature · P2: document pre-commit setup in developer documentation**
    - **Context:** "Documentation Updates"
    - **Action:**
        1. Update `README.md` and/or `CONTRIBUTING.md` to explain the pre-commit hook setup.
        2. Include instructions for developers to run `pre-commit install --hook-type commit-msg` after cloning.
        3. Document the Conventional Commits requirement and link to the specification.
    - **Done‑when:**
        1. Documentation clearly explains pre-commit hook installation for commit message validation.
        2. Examples of valid and invalid commit messages are provided.
    - **Verification:**
        1. Review documentation for clarity and completeness.
    - **Depends‑on:** [T005]

## Semantic Versioning (svu) Integration
- [x] **T007 · Feature · P1: integrate svu for version calculation in CI**
    - **Context:** PLAN.md "Semantic Version Utility (svu) Integration"
    - **Action:**
        1. Ensure `svu` is installed and available in CI.
        2. Update CI release workflow to run `svu next` and export result as environment variable.
    - **Done‑when:**
        1. CI workflow determines next semver version using commit history.
    - **Verification:**
        1. CI logs show correct version calculation for various commit scenarios.
    - **Depends‑on:** [T002]

## Changelog Generation
- [x] **T008 · Feature · P1: create .chglog/config.yml with commit type mapping**
    - **Context:** PLAN.md "Changelog Generation (git-chglog)"
    - **Action:**
        1. Create `.chglog/config.yml` specifying commit types, filters, and section mappings per project conventions.
    - **Done‑when:**
        1. Config file exists and matches desired changelog structure.
    - **Verification:**
        1. Run `git-chglog` locally and verify grouping/sections in output.
    - **Depends‑on:** none

- [x] **T009 · Feature · P1: create .chglog/CHANGELOG.tpl.md template**
    - **Context:** PLAN.md "Changelog Generation (git-chglog)"
    - **Action:**
        1. Add a markdown template to `.chglog/CHANGELOG.tpl.md` for changelog formatting.
    - **Done‑when:**
        1. `git-chglog` produces changelogs matching the template.
    - **Verification:**
        1. Generate a changelog and review output for formatting/section headers.
    - **Depends‑on:** [T008]

## Release Automation (Goreleaser)
- [x] **T010 · Feature · P0: create .goreleaser.yml for build and release**
    - **Context:** PLAN.md "Release Automation (Goreleaser)"
    - **Action:**
        1. Write `.goreleaser.yml` configuring builds, changelog integration, and GitHub release process.
        2. Add comments explaining non-obvious config aspects.
    - **Done‑when:**
        1. File is present and passes `goreleaser check`.
    - **Verification:**
        1. Run `goreleaser check` locally and in CI; confirm no errors.
    - **Depends‑on:** [T002], [T009]

- [x] **T011 · Feature · P1: configure goreleaser to use pre-generated changelog**
    - **Context:** PLAN.md "Release Automation (Goreleaser)"
    - **Action:**
        1. Update `.goreleaser.yml` so changelog is sourced from file generated by `git-chglog` (not auto-generated by goreleaser).
    - **Done‑when:**
        1. Released GitHub releases include correct, template-driven changelog content.
    - **Verification:**
        1. Dry-run release produces release notes matching expected changelog.
    - **Depends‑on:** [T010]

## CI/CD Pipeline Integration
- [x] **T012 · Feature · P0: implement .github/workflows/release.yml for CI/CD**
    - **Context:** PLAN.md "CI/CD Pipeline Integration"
    - **Action:**
        1. Create `release.yml` workflow with jobs for lint/test/build and release, per plan.
        2. Ensure correct triggers (`main` branch, `v*` tags), environment setup, tool installation, versioning, changelog, and release steps.
    - **Done‑when:**
        1. Workflow runs end-to-end on push/tag and passes all steps.
    - **Verification:**
        1. Test workflow on PR and `main` branch; verify builds, versioning, changelog, and (in dry run) release steps.
    - **Depends‑on:** [T002], [T007], [T009], [T011], [T010]

- [x] **T013 · Feature · P1: add goreleaser release --snapshot to PR/main CI runs**
    - **Context:** PLAN.md "Testing Strategy", "CI/CD Pipeline Integration"
    - **Action:**
        1. Update CI workflow to run `goreleaser release --snapshot` on PRs and non-tag main pushes for pipeline validation.
    - **Done‑when:**
        1. Snapshot releases run on every PR/main push (without tagging/publishing).
    - **Verification:**
        1. CI logs show dry-run/snapshot release output.
    - **Depends‑on:** [T012]

- [x] **T014 · Feature · P1: ensure CI fails on invalid commit messages**
    - **Context:** PLAN.md "Error & Edge-Case Strategy", "CI/CD Pipeline Integration"
    - **Action:**
        1. Add CI step to validate all commit messages in the push using `go-conventionalcommits`.
    - **Done‑when:**
        1. CI fails if any commit message is not Conventional Commit compliant.
    - **Verification:**
        1. Push with invalid message; CI run fails fast.
    - **Depends‑on:** [T012]

## Documentation & Communication
- [x] **T015 · Feature · P2: update README.md with commit conventions and release process**
    - **Context:** PLAN.md "Documentation Updates"
    - **Action:**
        1. Add clear section on Conventional Commits, with examples, and outline automated release process.
    - **Done‑when:**
        1. README.md contains examples, links to spec, and release automation explanation.
    - **Verification:**
        1. Review for clarity and completeness.
    - **Depends‑on:** none

- [x] **T016 · Feature · P2: update CONTRIBUTING.md with commit message and tooling policy**
    - **Context:** PLAN.md "Documentation Updates"
    - **Action:**
        1. Document enforced commit format, tooling install steps, and policy on hook/CI bypass.
    - **Done‑when:**
        1. CONTRIBUTING.md covers all relevant policies for new contributors.
    - **Verification:**
        1. Review document for clarity and policy coverage.
    - **Depends‑on:** none

- [x] **T017 · Feature · P2: document automated changelog generation and manual first entry**
    - **Context:** PLAN.md "Documentation Updates"
    - **Action:**
        1. Add note in README.md or CHANGELOG.md header about changelog being auto-generated.
        2. Create initial changelog entry if required for migration.
    - **Done‑when:**
        1. Users understand changelog is managed by automation.
    - **Verification:**
        1. Review for visibility and clarity.
    - **Depends‑on:** [T009]

## Initial Tagging & Migration
- [x] **T018 · Feature · P0: perform initial semantic version tag if needed**
    - **Context:** PLAN.md "Initial Tagging & Migration"
    - **Action:**
        1. If no semantic tags exist, decide initial version and create/tag in git (e.g., `v1.0.0`).
    - **Done‑when:**
        1. Project has a valid starting semantic version tag.
    - **Verification:**
        1. `git tag --list` shows correct tag; `svu current` reports correct version.
    - **Depends‑on:** none

## Testing & Edge-Case Verification
- [x] **T019 · Test · P1: verify pre-commit hook blocks non-conventional commits locally**
    - **Context:** PLAN.md "Testing Strategy"
    - **Action:**
        1. Attempt valid and invalid commits; observe blocking behavior.
    - **Done‑when:**
        1. Non-conventional commits are blocked; valid ones succeed.
    - **Verification:**
        1. Test multiple commit scenarios.
    - **Depends‑on:** [T005]

- [x] **T020 · Test · P1: verify svu version calculation for major/minor/patch/prerelease**
    - **Context:** PLAN.md "Testing Strategy"
    - **Action:**
        1. Craft commit histories with various commit types and run `svu next`.
    - **Done‑when:**
        1. `svu next` recommends correct version for each scenario.
    - **Verification:**
        1. Table of inputs/outputs for test cases.
    - **Depends‑on:** [T007]

- [x] **T021 · Test · P1: verify git-chglog changelog output for all commit types**
    - **Context:** PLAN.md "Testing Strategy"
    - **Action:**
        1. Generate changelog for sample history covering all configured types.
    - **Done‑when:**
        1. All relevant commit types appear in correct sections in output.
    - **Verification:**
        1. Manual review of generated markdown.
    - **Depends‑on:** [T009]

- [x] **T022 · Test · P1: verify full release pipeline (CI) in dry-run mode**
    - **Context:** PLAN.md "Testing Strategy", "E2E (Release Process)"
    - **Action:**
        1. Run workflow for PRs and main pushes with `goreleaser --snapshot`, ensuring all steps pass and outputs are correct.
    - **Done‑when:**
        1. CI passes and produces expected logs/artifacts (without publishing).
    - **Verification:**
        1. Review CI logs and artifacts; confirm no real releases occur.
    - **Depends‑on:** [T013]

- [x] **T023 · Test · P2: verify error handling for missing or invalid commit messages in CI**
    - **Context:** PLAN.md "Testing Strategy", "Error & Edge-Case Strategy"
    - **Action:**
        1. Push history with an invalid commit message and observe CI failure.
    - **Done‑when:**
        1. CI fails early and clearly on invalid messages.
    - **Verification:**
        1. Check logs for clear error message.
    - **Depends‑on:** [T014]

## Security & Permissions
- [x] **T024 · Feature · P0: configure GITHUB_TOKEN with correct permissions in CI**
    - **Context:** PLAN.md "Security & Config"
    - **Action:**
        1. Ensure workflow uses `GITHUB_TOKEN` with `contents: write` (and `packages: write` if needed) and does not expose in logs.
    - **Done‑when:**
        1. Release job can create tags/releases but does not leak secrets.
    - **Verification:**
        1. Review workflow and logs for proper handling.
    - **Depends‑on:** [T012]

## CI Resolution Tasks

- [x] **T025 · Bugfix · P0: update go-conventionalcommits installation in ci workflow**
    - **Context:** CI failure - Go Tool Installation Failure (BLOCKING)
    - **Action:**
        1. In `.github/workflows/release.yml`, update the `go-conventionalcommits` installation command to use the path `github.com/leodido/go-conventionalcommits`
        2. Pin the tool version to `v0.12.0` instead of using `@latest`
    - **Done‑when:**
        1. The CI workflow step for installing `go-conventionalcommits` completes successfully using version `v0.12.0`
    - **Verification:**
        1. Trigger the CI pipeline and observe the tool installation logs for successful execution
    - **Depends‑on:** none

- [x] **T026 · Chore · P1: update go-conventionalcommits in tools.go**
    - **Context:** CI failure - Go Tool Installation Failure (BLOCKING)
    - **Action:**
        1. If `tools.go` exists, update its reference to `go-conventionalcommits` to use the path `github.com/leodido/go-conventionalcommits` and version `v0.12.0`
        2. Run `go mod tidy` to reflect changes in `go.mod` and `go.sum`
    - **Done‑when:**
        1. `tools.go` (if present) is updated with the correct path and pinned version
        2. `go mod tidy` completes without errors related to `go-conventionalcommits`
    - **Verification:**
        1. Inspect `tools.go` (if present) and `go.mod` for the correct entries
    - **Depends‑on:** none

- [x] **T027 · Chore · P1: update go-conventionalcommits installation in makefile**
    - **Context:** CI failure - Go Tool Installation Failure (BLOCKING)
    - **Action:**
        1. Update `Makefile` tool installation targets for `go-conventionalcommits` to use the path `github.com/leodido/go-conventionalcommits` and pin to version `v0.12.0`
    - **Done‑when:**
        1. Running the relevant `Makefile` target installs `go-conventionalcommits` version `v0.12.0` successfully
    - **Verification:**
        1. Execute the `Makefile` target locally and verify the installed tool version
    - **Depends‑on:** none

- [x] **T028 · Chore · P1: update contributing.md for go-conventionalcommits installation**
    - **Context:** CI failure - Go Tool Installation Failure (BLOCKING)
    - **Action:**
        1. Update `CONTRIBUTING.md` to instruct users to install `go-conventionalcommits` from `github.com/leodido/go-conventionalcommits` at version `v0.12.0`
    - **Done‑when:**
        1. `CONTRIBUTING.md` accurately reflects the correct installation path and pinned version
    - **Verification:**
        1. Review the updated `CONTRIBUTING.md` for clarity and correctness
    - **Depends‑on:** [T025, T026, T027]

- [x] **T029 · Refactor · P1: ensure pre-commit configuration covers all file types**
    - **Context:** CI failure - Formatting Violations
    - **Action:**
        1. Review `.pre-commit-config.yaml` to ensure hooks for trailing whitespace and end-of-file newlines cover all relevant file types (e.g., `.md`, `.sh`, `.yml`, Go files)
        2. Add or adjust hook configurations if necessary
    - **Done‑when:**
        1. `.pre-commit-config.yaml` is confirmed or updated to provide comprehensive file type coverage for formatting
    - **Verification:**
        1. Manually inspect the pre-commit configuration file
    - **Depends‑on:** none

- [x] **T030 · Bugfix · P1: apply and commit formatting fixes using pre-commit**
    - **Context:** CI failure - Formatting Violations
    - **Action:**
        1. Run `pre-commit run --all-files` locally to fix all formatting issues (trailing whitespace, missing end-of-file newlines)
        2. Commit the changes applied by the pre-commit hooks
    - **Done‑when:**
        1. `pre-commit run --all-files` reports no further changes needed
        2. All formatting violations are fixed and the changes are committed
    - **Verification:**
        1. Re-run `pre-commit run --all-files` locally to confirm no violations remain
    - **Depends‑on:** [T025, T026, T027, T029]

- [x] **T031 · Test · P0: verify all ci checks pass**
    - **Context:** CI Resolution - Final verification
    - **Action:**
        1. Ensure all preceding fixes (T025-T030) are pushed to the relevant branch
        2. Trigger and monitor the CI pipeline
    - **Done‑when:**
        1. The CI pipeline completes successfully, with all checks (tool installation, formatting, linting, tests, build) passing
    - **Verification:**
        1. Check the CI dashboard/logs for green status on all jobs
    - **Depends‑on:** [T028, T030]

## CI Resolution

- [x] **T032 · Bugfix · P0: remove `go-conventionalcommits` installation from `release.yml`**
    - **Context:** CI Resolution Plan - Phase 1: Remove Invalid Tool Installation; Implementation Steps - Step 1 (Remove `Install go-conventionalcommits` step)
    - **Action:**
        1. Edit `.github/workflows/release.yml`.
        2. Remove the entire `Install go-conventionalcommits` step, including the `go install github.com/leodido/go-conventionalcommits@v0.12.0` command.
    - **Done‑when:**
        1. The `Install go-conventionalcommits` step is completely removed from `.github/workflows/release.yml`.
        2. The CI pipeline no longer attempts to `go install` `go-conventionalcommits`.
    - **Depends‑on:** none

- [x] **T033 · Bugfix · P0: fix missing eof newline in `docs/ci-resolution-status.md`**
    - **Context:** CI Resolution Plan - Phase 2: Fix Formatting Issues; Root Cause Analysis - Issue 2
    - **Action:**
        1. Run `pre-commit run --all-files` locally to fix formatting issues, specifically ensuring `docs/ci-resolution-status.md` has an end-of-file newline.
        2. Commit the formatting fixes.
    - **Done‑when:**
        1. `docs/ci-resolution-status.md` has a trailing newline.
        2. `pre-commit run --all-files` passes locally for `docs/ci-resolution-status.md` without making changes.
        3. The `Lint and Format` CI job passes.
    - **Depends‑on:** none

- [x] **T034 · Feature · P1: implement `commitlint-github-action` for commit validation in `release.yml`**
    - **Context:** CI Resolution Plan - Phase 1 (Recommended Option B) & Phase 3; Implementation Steps - Step 1 (Replace validation step)
    - **Action:**
        1. Edit `.github/workflows/release.yml`.
        2. Replace the existing `Validate Commit Messages` step (that previously used `go-conventionalcommits validate`) with the `wagoid/commitlint-github-action@v5` configuration.
        3. Ensure the new step includes `if: github.event_name == 'pull_request'` and `with: configFile: .commitlintrc.yml`.
    - **Done‑when:**
        1. The `Validate Commit Messages` step in `.github/workflows/release.yml` uses `wagoid/commitlint-github-action@v5`.
        2. Commit message validation is performed by the new GitHub Action on pull requests.
    - **Verification:**
        1. Push a commit with an invalid message to a PR; verify the CI step fails.
        2. Push a commit with a valid message to a PR; verify the CI step passes.
    - **Depends‑on:** [T035]

- [x] **T035 · Feature · P1: create and configure `.commitlintrc.yml` for conventional commits**
    - **Context:** CI Resolution Plan - Phase 3: Implement Proper Commit Validation - Action 2
    - **Action:**
        1. Create a new file named `.commitlintrc.yml` in the repository root.
        2. Populate `.commitlintrc.yml` with a standard configuration that enforces conventional commit message standards (e.g., extending `@commitlint/config-conventional`).
    - **Done‑when:**
        1. `.commitlintrc.yml` exists in the repository root.
        2. The file contains a valid YAML configuration for `commitlint`.
    - **Verification:**
        1. Run `npx commitlint --edit` against a sample commit message locally to test the configuration.
    - **Depends‑on:** none

- [x] **T036 · Test · P1: verify all CI checks pass on feature branch after fixes**
    - **Context:** CI Resolution Plan - Phase 4: Verify Resolution
    - **Action:**
        1. Ensure changes from T032, T033, T034, and T035 are pushed to the designated feature branch.
        2. Monitor the CI execution for this branch/PR on GitHub.
    - **Done‑when:**
        1. All CI workflows (including linting, testing, build, formatting, and commit message validation) complete successfully.
        2. All checks are green on the GitHub PR.
    - **Verification:**
        1. Review the GitHub Actions logs for the feature branch to confirm no errors related to the resolved issues.
    - **Depends‑on:** [T032, T033, T034]

## Prevention Measures

- [x] **T037 · Chore · P2: document pre-commit hook requirement and setup in `CONTRIBUTING.md`**
    - **Context:** CI Resolution Plan - Prevention Measures: 1. Enforce Pre-commit Hooks
    - **Action:**
        1. Update `CONTRIBUTING.md` to add a section explaining the requirement to use pre-commit hooks.
        2. Include clear, step-by-step instructions for developers to install `pre-commit` and set up the project's hooks locally (e.g., `pre-commit install`).
    - **Done‑when:**
        1. `CONTRIBUTING.md` contains a clear section on mandatory pre-commit hook usage and setup.
    - **Depends‑on:** none

- [x] **T038 · Chore · P3: create script to verify local pre-commit hook installation**
    - **Context:** CI Resolution Plan - Prevention Measures: 1. Enforce Pre-commit Hooks
    - **Action:**
        1. Develop a simple shell script (e.g., `scripts/verify-hooks.sh`) that checks if pre-commit hooks are installed in the local `.git/hooks` directory.
        2. If not installed, the script should prompt the user with instructions or a reference to `CONTRIBUTING.md`.
    - **Done‑when:**
        1. The verification script is created and executable.
        2. The script correctly identifies if hooks are installed and provides guidance if they are not.
    - **Verification:**
        1. Run the script in a fresh clone without hooks installed to check its output.
        2. Run the script after installing hooks to confirm it reports correctly.
    - **Depends‑on:** [T037]

- [x] **T039 · Chore · P2: configure mandatory reviews for CI workflow changes**
    - **Context:** CI Resolution Plan - Prevention Measures: 2. CI Configuration Reviews
    - **Action:**
        1. Configure repository settings (e.g., via GitHub branch protection rules and/or `CODEOWNERS` file).
        2. Ensure that changes to files within `.github/workflows/` require at least one review and approval from a designated team member or group before merging.
    - **Done‑when:**
        1. Pull requests modifying files in `.github/workflows/` are blocked from merging without the required approvals.
    - **Verification:**
        1. Create a test PR with a change to a workflow file and attempt to merge without approval to confirm the restriction.
    - **Depends‑on:** none

- [x] **T040 · Chore · P2: update developer documentation on Go tool installation (library vs. CLI)**
    - **Context:** CI Resolution Plan - Prevention Measures: 3. Tool Documentation
    - **Action:**
        1. Add a section to the project's developer documentation (e.g., a new `docs/development/tooling.md` or an existing relevant guide).
        2. Clearly explain the difference between installing Go libraries (typically not `go install` unless they have a `main` package) and Go CLI tools (which use `go install`).
        3. Provide examples of correct installation commands for both types.
    - **Done‑when:**
        1. Developer documentation accurately distinguishes between Go library and CLI tool installation methods.
        2. Examples of correct usage are provided.
    - **Depends‑on:** none

- [x] **T041 · Chore · P3: create CI troubleshooting guide**
    - **Context:** CI Resolution Plan - Prevention Measures: 4. Developer Education
    - **Action:**
        1. Create a new document (e.g., `docs/ci-troubleshooting.md`).
        2. Populate it with common CI failure scenarios encountered in the project (including the `go install` issue and formatting violations) and their diagnostic/resolution steps.
    - **Done‑when:**
        1. A CI troubleshooting guide exists.
        2. The guide covers the issues addressed in this plan and provides actionable advice.
    - **Depends‑on:** none

## CI Resolution Tasks (PR #24)

### Immediate Fixes
- [x] **T042 · Bugfix · P0: add missing trailing newline to docs/ci-troubleshooting.md**
    - **Context:** CI Resolution Plan > Resolution Steps > Priority 1: Fix Formatting Violation
    - **Action:**
        1. Add a trailing newline character to the end of the `docs/ci-troubleshooting.md` file.
        2. Run `pre-commit install` (if not already done) and then `pre-commit run --files docs/ci-troubleshooting.md` to verify the fix locally.
        3. Commit the change with the message "fix: add missing trailing newline to ci-troubleshooting.md" and push to `feature/automated-semantic-versioning`.
    - **Done‑when:**
        1. `docs/ci-troubleshooting.md` has a trailing newline.
        2. Local pre-commit check for `docs/ci-troubleshooting.md` passes.
        3. The "Lint and Format" CI job for PR #24 passes.
    - **Verification:**
        1. Observe that the CI pipeline for PR #24 passes the formatting checks after the push.
    - **Depends‑on:** none

- [x] **T042a · Chore · P0: ensure git hooks are mandatory and auto-installed**
    - **Context:** Missing EOF newlines should never reach CI - hooks exist but weren't installed/run locally
    - **Action:**
        1. Verify `.pre-commit-config.yaml` has `end-of-file-fixer` hook (already present)
        2. Add setup script or Makefile target to automatically install pre-commit hooks
        3. Update CONTRIBUTING.md to emphasize mandatory pre-commit hook installation
        4. Add CI check to verify commits went through pre-commit hooks
        5. Consider adding a git hook installer to project setup/bootstrap process
    - **Done‑when:**
        1. Pre-commit hooks are automatically installed during project setup
        2. Contributors cannot accidentally skip hook installation
        3. CI verifies that commits were made with hooks enabled
    - **Verification:**
        1. Clone repo fresh and run setup process
        2. Verify hooks are automatically installed
        3. Test that commits without hooks are rejected by CI
    - **Depends‑on:** none

- [x] **T043 · Chore · P1: remove leyline docs sync workflow**
    - **Context:** Leyline project is still in development - we'll add this workflow back later
    - **Action:**
        1. Delete the `.github/workflows/leyline.yml` file
        2. Commit the removal with message "chore: remove leyline docs sync workflow (project in development)"
    - **Done‑when:**
        1. Leyline sync workflow is removed from the repository
        2. CI no longer runs the docs sync job
    - **Depends‑on:** none

### Developer Experience & Standards
- [x] **T044 · Chore · P2: update CONTRIBUTING.md to mandate pre-commit hook installation and usage**
    - **Context:** CI Resolution Plan > Prevention Measures > 1. Mandatory Pre-commit Hooks
    - **Action:**
        1. Modify `CONTRIBUTING.md` to clearly state that pre-commit hook installation (`pre-commit install`) and usage are mandatory for all contributors.
        2. Include concise instructions for installation and running hooks.
        3. Document the policy that bypassing hooks (e.g., with `git commit --no-verify`) is prohibited except in explicitly documented and approved exceptions.
    - **Done‑when:**
        1. `CONTRIBUTING.md` is updated with the mandatory pre-commit hook policy and instructions.
    - **Depends‑on:** none

- [x] **T045 · Chore · P2: review and enhance .pre-commit-config.yaml for comprehensive file coverage**
    - **Context:** CI Resolution Plan > Prevention Measures > 2. Enhanced Pre-commit Configuration
    - **Action:**
        1. Audit the `.pre-commit-config.yaml` file.
        2. Ensure hooks provide coverage for all key file types used in the project (Markdown, YAML, shell scripts, etc.).
        3. Add or update hooks to implement more comprehensive checks as needed (e.g., specific linters, formatters).
    - **Done‑when:**
        1. `.pre-commit-config.yaml` is updated to provide comprehensive checks for relevant file types.
        2. The configuration is tested locally by running `pre-commit run --all-files`.
    - **Depends‑on:** none

### Script & Code Quality
### CI/CD Infrastructure & Process
- [x] **T046 · Refactor · P2: implement fail-fast principle in ci workflows**
    - **Context:** CI Resolution Plan > Prevention Measures > 4. Improved CI Workflow Design
    - **Action:**
        1. Review existing GitHub Actions workflows in `.github/workflows/`.
        2. Configure workflows to cancel or skip subsequent jobs/steps as soon as a critical job/step fails (e.g., using `jobs.<job_id>.if` conditions or workflow-level cancellation).
    - **Done‑when:**
        1. CI workflows are updated to halt or skip non-essential jobs promptly after a failure in a critical path.
    - **Verification:**
        1. Intentionally cause a failure in an early CI job and verify that dependent, non-critical jobs are skipped or the workflow run is cancelled quickly.
    - **Depends‑on:** none

- [ ] **T047 · Chore · P2: establish mandatory reviews for ci workflow changes via CODEOWNERS**
    - **Context:** CI Resolution Plan > Prevention Measures > 4. Improved CI Workflow Design
    - **Action:**
        1. Create or update the `CODEOWNERS` file (e.g., in `.github/CODEOWNERS`).
        2. Add an entry specifying a designated team or individuals as owners for the `.github/workflows/` directory.
    - **Done‑when:**
        1. The `CODEOWNERS` file is configured to automatically request reviews from the specified owners for any pull requests modifying files under `.github/workflows/`.
    - **Verification:**
        1. Open a test PR modifying a workflow file and confirm the designated CODEOWNERS are automatically added as reviewers.
    - **Depends‑on:** none

### Knowledge Management & Documentation
- [ ] **T048 · Chore · P2: update docs/ci-troubleshooting.md with pr #24 incident details and lessons learned**
    - **Context:** CI Resolution Plan > Prevention Measures > 6. Living Documentation
    - **Action:**
        1. Add new entries to `docs/ci-troubleshooting.md` detailing the formatting violation (T042) from PR #24.
        2. For each issue, document the symptoms, root cause, and the resolution applied.
        3. Include any general lessons learned that could help prevent similar issues.
    - **Done‑when:**
        1. `docs/ci-troubleshooting.md` is updated with comprehensive details of the PR #24 CI issues and their fixes.
    - **Depends‑on:** [T042]
