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

- [ ] **T002 · Feature · P1: add Makefile target for Go tool installation**
    - **Context:** PLAN.md "Tool Installation & Pinning"
    - **Action:**
        1. Create (or update) `Makefile` with `tools` target that runs `go install` for each tool from `tools.go`.
        2. Comment `Makefile` targets for clarity.
    - **Done‑when:**
        1. `make tools` installs all required CLI tools at pinned versions.
    - **Verification:**
        1. Run `make tools` in a fresh shell; all tools become available in `$GOPATH/bin` or `$PATH`.
    - **Depends‑on:** [T001]

- [ ] **T003 · Feature · P2: document local tooling setup in README.md and/or CONTRIBUTING.md**
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
- [ ] **T004 · Feature · P0: add go-conventionalcommits to pre-commit hooks**
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

- [ ] **T005 · Feature · P0: verify pre-commit hook installation**
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

- [ ] **T006 · Feature · P2: document pre-commit setup in developer documentation**
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
- [ ] **T007 · Feature · P1: integrate svu for version calculation in CI**
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
- [ ] **T008 · Feature · P1: create .chglog/config.yml with commit type mapping**
    - **Context:** PLAN.md "Changelog Generation (git-chglog)"
    - **Action:**
        1. Create `.chglog/config.yml` specifying commit types, filters, and section mappings per project conventions.
    - **Done‑when:**
        1. Config file exists and matches desired changelog structure.
    - **Verification:**
        1. Run `git-chglog` locally and verify grouping/sections in output.
    - **Depends‑on:** none

- [ ] **T009 · Feature · P1: create .chglog/CHANGELOG.tpl.md template**
    - **Context:** PLAN.md "Changelog Generation (git-chglog)"
    - **Action:**
        1. Add a markdown template to `.chglog/CHANGELOG.tpl.md` for changelog formatting.
    - **Done‑when:**
        1. `git-chglog` produces changelogs matching the template.
    - **Verification:**
        1. Generate a changelog and review output for formatting/section headers.
    - **Depends‑on:** [T008]

## Release Automation (Goreleaser)
- [ ] **T010 · Feature · P0: create .goreleaser.yml for build and release**
    - **Context:** PLAN.md "Release Automation (Goreleaser)"
    - **Action:**
        1. Write `.goreleaser.yml` configuring builds, changelog integration, and GitHub release process.
        2. Add comments explaining non-obvious config aspects.
    - **Done‑when:**
        1. File is present and passes `goreleaser check`.
    - **Verification:**
        1. Run `goreleaser check` locally and in CI; confirm no errors.
    - **Depends‑on:** [T002], [T009]

- [ ] **T011 · Feature · P1: configure goreleaser to use pre-generated changelog**
    - **Context:** PLAN.md "Release Automation (Goreleaser)"
    - **Action:**
        1. Update `.goreleaser.yml` so changelog is sourced from file generated by `git-chglog` (not auto-generated by goreleaser).
    - **Done‑when:**
        1. Released GitHub releases include correct, template-driven changelog content.
    - **Verification:**
        1. Dry-run release produces release notes matching expected changelog.
    - **Depends‑on:** [T010]

## CI/CD Pipeline Integration
- [ ] **T012 · Feature · P0: implement .github/workflows/release.yml for CI/CD**
    - **Context:** PLAN.md "CI/CD Pipeline Integration"
    - **Action:**
        1. Create `release.yml` workflow with jobs for lint/test/build and release, per plan.
        2. Ensure correct triggers (`main` branch, `v*` tags), environment setup, tool installation, versioning, changelog, and release steps.
    - **Done‑when:**
        1. Workflow runs end-to-end on push/tag and passes all steps.
    - **Verification:**
        1. Test workflow on PR and `main` branch; verify builds, versioning, changelog, and (in dry run) release steps.
    - **Depends‑on:** [T002], [T007], [T009], [T011], [T010]

- [ ] **T013 · Feature · P1: add goreleaser release --snapshot to PR/main CI runs**
    - **Context:** PLAN.md "Testing Strategy", "CI/CD Pipeline Integration"
    - **Action:**
        1. Update CI workflow to run `goreleaser release --snapshot` on PRs and non-tag main pushes for pipeline validation.
    - **Done‑when:**
        1. Snapshot releases run on every PR/main push (without tagging/publishing).
    - **Verification:**
        1. CI logs show dry-run/snapshot release output.
    - **Depends‑on:** [T012]

- [ ] **T014 · Feature · P1: ensure CI fails on invalid commit messages**
    - **Context:** PLAN.md "Error & Edge-Case Strategy", "CI/CD Pipeline Integration"
    - **Action:**
        1. Add CI step to validate all commit messages in the push using `go-conventionalcommits`.
    - **Done‑when:**
        1. CI fails if any commit message is not Conventional Commit compliant.
    - **Verification:**
        1. Push with invalid message; CI run fails fast.
    - **Depends‑on:** [T012]

## Documentation & Communication
- [ ] **T015 · Feature · P2: update README.md with commit conventions and release process**
    - **Context:** PLAN.md "Documentation Updates"
    - **Action:**
        1. Add clear section on Conventional Commits, with examples, and outline automated release process.
    - **Done‑when:**
        1. README.md contains examples, links to spec, and release automation explanation.
    - **Verification:**
        1. Review for clarity and completeness.
    - **Depends‑on:** none

- [ ] **T016 · Feature · P2: update CONTRIBUTING.md with commit message and tooling policy**
    - **Context:** PLAN.md "Documentation Updates"
    - **Action:**
        1. Document enforced commit format, tooling install steps, and policy on hook/CI bypass.
    - **Done‑when:**
        1. CONTRIBUTING.md covers all relevant policies for new contributors.
    - **Verification:**
        1. Review document for clarity and policy coverage.
    - **Depends‑on:** none

- [ ] **T017 · Feature · P2: document automated changelog generation and manual first entry**
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
- [ ] **T018 · Feature · P0: perform initial semantic version tag if needed**
    - **Context:** PLAN.md "Initial Tagging & Migration"
    - **Action:**
        1. If no semantic tags exist, decide initial version and create/tag in git (e.g., `v1.0.0`).
    - **Done‑when:**
        1. Project has a valid starting semantic version tag.
    - **Verification:**
        1. `git tag --list` shows correct tag; `svu current` reports correct version.
    - **Depends‑on:** none

## Testing & Edge-Case Verification
- [ ] **T019 · Test · P1: verify pre-commit hook blocks non-conventional commits locally**
    - **Context:** PLAN.md "Testing Strategy"
    - **Action:**
        1. Attempt valid and invalid commits; observe blocking behavior.
    - **Done‑when:**
        1. Non-conventional commits are blocked; valid ones succeed.
    - **Verification:**
        1. Test multiple commit scenarios.
    - **Depends‑on:** [T005]

- [ ] **T020 · Test · P1: verify svu version calculation for major/minor/patch/prerelease**
    - **Context:** PLAN.md "Testing Strategy"
    - **Action:**
        1. Craft commit histories with various commit types and run `svu next`.
    - **Done‑when:**
        1. `svu next` recommends correct version for each scenario.
    - **Verification:**
        1. Table of inputs/outputs for test cases.
    - **Depends‑on:** [T007]

- [ ] **T021 · Test · P1: verify git-chglog changelog output for all commit types**
    - **Context:** PLAN.md "Testing Strategy"
    - **Action:**
        1. Generate changelog for sample history covering all configured types.
    - **Done‑when:**
        1. All relevant commit types appear in correct sections in output.
    - **Verification:**
        1. Manual review of generated markdown.
    - **Depends‑on:** [T009]

- [ ] **T022 · Test · P1: verify full release pipeline (CI) in dry-run mode**
    - **Context:** PLAN.md "Testing Strategy", "E2E (Release Process)"
    - **Action:**
        1. Run workflow for PRs and main pushes with `goreleaser --snapshot`, ensuring all steps pass and outputs are correct.
    - **Done‑when:**
        1. CI passes and produces expected logs/artifacts (without publishing).
    - **Verification:**
        1. Review CI logs and artifacts; confirm no real releases occur.
    - **Depends‑on:** [T013]

- [ ] **T023 · Test · P2: verify error handling for missing or invalid commit messages in CI**
    - **Context:** PLAN.md "Testing Strategy", "Error & Edge-Case Strategy"
    - **Action:**
        1. Push history with an invalid commit message and observe CI failure.
    - **Done‑when:**
        1. CI fails early and clearly on invalid messages.
    - **Verification:**
        1. Check logs for clear error message.
    - **Depends‑on:** [T014]

## Security & Permissions
- [ ] **T024 · Feature · P0: configure GITHUB_TOKEN with correct permissions in CI**
    - **Context:** PLAN.md "Security & Config"
    - **Action:**
        1. Ensure workflow uses `GITHUB_TOKEN` with `contents: write` (and `packages: write` if needed) and does not expose in logs.
    - **Done‑when:**
        1. Release job can create tags/releases but does not leak secrets.
    - **Verification:**
        1. Review workflow and logs for proper handling.
    - **Depends‑on:** [T012]

## Open Questions / Clarifications
- [ ] **Issue: clarify initial semantic version for migration**
    - **Context:** PLAN.md "Open Questions" #1
    - **Blocking?:** yes

- [ ] **Issue: define prerelease workflow for alpha/beta/rc builds**
    - **Context:** PLAN.md "Open Questions" #2
    - **Blocking?:** no

- [ ] **Issue: document hotfix process for older released versions**
    - **Context:** PLAN.md "Open Questions" #3
    - **Blocking?:** no

- [ ] **Issue: finalize who/what commits CHANGELOG.md updates**
    - **Context:** PLAN.md "Open Questions" #4
    - **Blocking?:** yes

- [ ] **Issue: review sufficiency of make tools / tools.go approach for all developer environments**
    - **Context:** PLAN.md "Open Questions" #5
    - **Blocking?:** no
