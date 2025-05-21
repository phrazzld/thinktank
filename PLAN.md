# Remediation Plan – Sprint 1

## Executive Summary

This plan targets critical and high-severity issues identified in the code review, focusing on reproducibility, CI/CD reliability, security, and developer workflow consistency. Blocker issues are prioritized by impact and dependency, with quick wins delivered first to unlock parallel work. Each remedy is mapped directly to review findings and the project's Development Philosophy, ensuring alignment with core values and long-term maintainability.

## Strike List

| Seq | CR‑ID | Title                                               | Effort | Owner       |
|-----|-------|-----------------------------------------------------|--------|-------------|
| 1   | cr‑01 | Standardize and Pin Tool Versions                   | m      | devops      |
| 2   | cr‑02 | Upgrade and Unify golangci-lint and Lint Config     | m      | devops      |
| 3   | cr‑03 | Unify Commit Message Validation (commitlint)        | m      | devops      |
| 4   | cr‑04 | Add Automated Vulnerability Scanning to CI          | s      | devops      |
| 5   | cr‑05 | Enforce Robust Pre-commit Hook Installation         | s      | devops      |
| 6   | cr‑06 | Audit Test Mock Concurrency for Race Conditions     | m      | dev         |
| 7   | cr‑07 | Enforce and Raise Coverage Thresholds, E2E in CI    | m      | dev         |
| 8   | cr‑08 | Consolidate API Key Resolution Logic                | s      | dev         |
| 9   | cr‑09 | Clarify Local Commitlint Config & Remove No-ops     | s      | devops      |
| 10  | cr‑10 | (None)                                              | -      | -           |

## Detailed Remedies

---

### cr‑01 Standardize and Pin Tool Versions

- **Problem:** Tooling versions for critical CI/dev tools are not pinned or consistent across Makefile, tools.go, workflows, or package.json.
- **Impact:** Builds are non-reproducible, "works on my machine" errors occur, and CI can break unexpectedly.
- **Chosen Fix:** Pin *every* tooling version (Go CLI, JS, Actions) in code, workflows, and documentation.
- **Steps:**
  1. Audit all Go CLI tools (`golangci-lint`, `svu`, `git-chglog`, `goreleaser`, etc.) and JS tools (`commitizen`, etc.).
  2. Update `tools.go` to use correct import paths (e.g., `github.com/caarlos0/svu/v3`) and only for installable CLIs.
  3. Update `Makefile` so `go install` uses versioned import paths, *never* `@latest`.
  4. Update `.github/workflows/*` to use pinned versions for `go install`, GitHub Actions (e.g., `actions/checkout@v4.1.7`), and JS/Node tools.
  5. Update `package.json` to pin all dev dependencies.
  6. Update `CONTRIBUTING.md` to list all required tool versions and describe installation.
- **Done‑When:** Running `make tools`/CI installs exact tool versions everywhere; `CONTRIBUTING.md` is accurate; all builds are reproducible.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** m

---

### cr‑02 Upgrade and Unify golangci-lint and Lint Config

- **Problem:** golangci-lint is extremely outdated in CI/pre-commit, config is inconsistent/minimal, and versions are not aligned.
- **Impact:** Linting is weak, security/quality gates are bypassed, and confusion for devs.
- **Chosen Fix:** Upgrade to latest golangci-lint, unify version in `go.mod`, Makefile, `.pre-commit-config.yaml`, workflows, and expand linter set.
- **Steps:**
  1. Choose a recent stable golangci-lint version (e.g., `v1.59.1`).
  2. Update all install commands and config references to use this version.
  3. Replace direct `curl | sh` installs in workflows with pinned `go install`.
  4. In `.golangci.yml`, enable a comprehensive linter set (`errcheck`, `staticcheck`, `gosec`, `unused`, `stylecheck`, `gocritic`, `govet`, etc.).
  5. Remove any package-specific disables unless strictly justified.
  6. Update `tools.go` and document in `CONTRIBUTING.md`.
- **Done‑When:** Lint version is consistent everywhere, linter config is strict, all lint checks pass/fail as expected in CI and locally.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** m

---

### cr‑03 Unify Commit Message Validation (commitlint)

- **Problem:** Commit message validation uses inconsistent logic: bash regex in CI/workflows, commitlint in pre-commit, custom/no-op ignores in local configs.
- **Impact:** Inconsistent enforcement, reduced trust, "why did CI fail" confusion.
- **Chosen Fix:** Use `commitlint` as the *sole* validation tool everywhere—local, pre-push, CI.
- **Steps:**
  1. Remove all bash/regex commit validation scripts from workflows and repo (`scripts/ci/validate-baseline-commits.sh`).
  2. Adapt `scripts/validate-push.sh` for CI and pre-push, using `commitlint` with baseline awareness.
  3. Ensure `.pre-commit-config.yaml`'s `commit-msg` and `pre-push` hooks use `commitlint` and are baseline-aware.
  4. Remove `.commitlintrc-local.js` if not used; ensure only one local config is active and that ignores are meaningful or removed.
  5. Update `docs/conventional-commits.md` to reflect unified validation and baseline policy.
- **Done‑When:** Only `commitlint` is used for validation everywhere; CI and local validation yield same results; docs are clear.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** m

---

### cr‑04 Add Automated Vulnerability Scanning to CI

- **Problem:** CI/CD lacks any Go dependency vulnerability scanning.
- **Impact:** Potential for critical vulnerabilities to be merged/released undetected.
- **Chosen Fix:** Add `govulncheck` to CI and Release workflows, pinned to a specific version.
- **Steps:**
  1. Add `go install golang.org/x/vuln/cmd/govulncheck@v1.1.4` step to both workflows.
  2. Add `govulncheck ./...` step and fail the build if any Critical/High vulnerabilities are found.
  3. Document in `CONTRIBUTING.md`.
- **Done‑When:** CI fails on new critical/high vulnerabilities; developers are notified via logs.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** s

---

### cr‑05 Enforce Robust Pre-commit Hook Installation

- **Problem:** Setup scripts and Makefile targets for pre-commit hooks are incomplete; some hooks (pre-push) may not be installed; missing or unclear failure modes.
- **Impact:** Hooks may not run for all devs, inconsistent enforcement, increased CI/test failures.
- **Chosen Fix:** Make hook installation explicit, robust, and mandatory for all contributors.
- **Steps:**
  1. Update `scripts/setup.sh` and `Makefile`'s `hooks`/`tools` targets:
      - Explicitly check for `pre-commit` CLI, fail with clear instructions if missing.
      - Run `pre-commit install --install-hooks`, `--hook-type commit-msg`, `--hook-type pre-push`, `--hook-type post-commit`.
  2. Ensure `CONTRIBUTING.md` and setup instructions are strict: `make tools` or `./scripts/setup.sh` is REQUIRED.
  3. Add a CI step to validate `.pre-commit-config.yaml` and/or simulate `pre-commit run --all-files`.
- **Done‑When:** All hooks are installed by default; contributors cannot bypass; CI rejects PRs with missing hooks.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** s

---

### cr‑06 Audit Test Mock Concurrency for Race Conditions

- **Problem:** Only `MockFileWriter` was fixed for race conditions; other test mocks with mutable state may still be unsafe.
- **Impact:** Race detector only sometimes catches issues; tests can be flaky and misleading.
- **Chosen Fix:** Audit all test mocks for shared mutable state; add mutexes where required.
- **Steps:**
  1. Review all test mocks (especially in `*_test.go` files) for shared maps/slices/fields.
  2. Add mutex protection for all shared state that can be accessed from multiple goroutines.
  3. Ensure all test constructors initialize mutexes.
  4. Regularly run `go test -race ./...` locally and in CI.
- **Done‑When:** All test races are fixed; race detector reports clean; CI is stable.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** m

---

### cr‑07 Enforce and Raise Coverage Thresholds, E2E in CI

- **Problem:** Coverage thresholds are too low (64%), enforcement is weak, and E2E tests are not run in CI.
- **Impact:** Critical code paths may be untested; regressions go undetected; system reliability is low.
- **Chosen Fix:** Immediately raise overall threshold to 75%, enforce per-package thresholds, and ensure E2E tests run in CI.
- **Steps:**
  1. Update all coverage scripts and CI configs: set overall threshold to 75% (then plan to 90%).
  2. Make per-package coverage checks (e.g., for `internal/thinktank`, `internal/llm`) fail builds if below threshold.
  3. Integrate E2E tests into CI workflow as a required, blocking step.
  4. Track and document a plan to reach 90% coverage (issue backlog).
- **Done‑When:** CI blocks merges below 75%, all critical packages hit their thresholds, E2E runs on every build.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** m

---

### cr‑08 Consolidate API Key Resolution Logic

- **Problem:** API key resolution logic is scattered and unclear; order of precedence and source is implicit.
- **Impact:** Potential for incorrect key usage, hard-to-debug authentication errors, security risk.
- **Chosen Fix:** Centralize API key resolution in a single, documented function with explicit precedence.
- **Steps:**
  1. Refactor API key lookup into a single function that checks environment, then explicit parameter, with clear fallback order.
  2. Document the lookup order and rationale in code and `CONTRIBUTING.md`.
  3. Add unit tests for all resolution paths.
- **Done‑When:** All API key resolution flows through one code path; behavior is clear and testable.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** s

---

### cr‑09 Clarify Local Commitlint Config & Remove No-ops

- **Problem:** Dual commitlint local configs with ineffective/unused ignores; ambiguous which file is authoritative.
- **Impact:** Devs may edit the wrong config, misunderstand baseline, or get inconsistent validation.
- **Chosen Fix:** Pick one local override file, remove the other, and ensure the `ignores` function is either functional or removed.
- **Steps:**
  1. Decide on `.commitlint-with-baseline.js` or `.commitlintrc-local.js` as the single override.
  2. Remove the other, or update setup scripts to manage only the correct one.
  3. Remove or fix the `ignores` function so it matches actual baseline validation behavior.
  4. Update docs to reflect the exact baseline validation policy and how local config is used.
- **Done‑When:** Only one local override is active; configuration is unambiguous; local validation matches CI.
- **Standards Check:**
  - Simplicity: ✔
  - Modularity: ✔
  - Testability: ✔
  - Coding Std: ✔
  - Security: ✔
- **Effort:** s

---

## Standards Alignment

- **Simplicity:** All fixes eliminate accidental complexity (e.g., tool version drift, duplicated validation, ad hoc scripts).
- **Modularity:** Each remedy focuses on one responsibility (tooling, coverage, commit validation), enforcing clear boundaries.
- **Testability:** Every change is validated by automated tests, lint, and CI; configuration is unambiguous and centrally documented.
- **Coding Std:** Aligns fully with the Development Philosophy and Go appendix, especially for automation, dependency management, and error handling.
- **Security:** Automated vulnerability scanning closes a critical gap; API key handling is clarified and made less error-prone.

## Validation Checklist

- [ ] Automated tests green (CI pipeline passes on main and PRs).
- [ ] Static analyzers (golangci-lint) clear, with strict config.
- [ ] All workflows use only pinned tool versions and actions.
- [ ] Pre-commit hooks and commit message validation run for all contributors.
- [ ] Vulnerability scans run in CI and fail on critical/high findings.
- [ ] E2E tests are executed as required CI step.
- [ ] Coverage thresholds are enforced and documented.
- [ ] Manual pen-test (or code review) of API key resolution and error handling passes.
- [ ] No new lint or audit warnings in CI or local runs.
- [ ] Documentation (`CONTRIBUTING.md`, `docs/conventional-commits.md`) is updated and accurate.

---

**End of Plan**
