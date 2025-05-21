# TODO

## Critical Fixes (P0)

### CI/CD & Development Environment Stability
- [x] **T001: Standardize and Pin Tool Versions**
  - **Context:** Found inconsistent tool versions across Makefile, tools.go, and workflows
  - **Action:**
    1. Audit all Go and JS CLI tools and current versions
    2. Update `tools.go` with correct versioned import paths (e.g., `github.com/caarlos0/svu/v3@v3.2.3`)
    3. Update Makefile to use versioned `go install` (never `@latest`)
    4. Pin GitHub Action versions in all workflows (e.g., `actions/checkout@v4.1.7`)
    5. Pin JS/Node tool versions in workflows and `package.json`
    6. Document all versions in `CONTRIBUTING.md`
  - **Done-when:** All tool versions are explicitly pinned and reproducible across environments

- [x] **T002: Upgrade and Unify `golangci-lint` Configuration**
  - **Context:** Using outdated `golangci-lint v2.1.1` with minimal linter set
  - **Action:**
    1. Choose a recent, stable `golangci-lint` version (e.g., `v1.59.1`)
    2. Update all configurations to reference this version
    3. Replace `curl | sh` installs with `go install` in workflows
    4. Expand enabled linters in `.golangci.yml` (errcheck, staticcheck, gosec, etc.)
    5. Remove unjustified linter disables
  - **Done-when:** CI, pre-commit, and local environments use the same recent version with comprehensive linter set

- [ ] **T003: Unify Commit Message Validation on `commitlint`**
  - **Context:** Multiple validation mechanisms (`bash regex` vs. `commitlint` CLI) causing inconsistency
  - **Action:**
    1. Remove bash/regex commit validation scripts
    2. Adapt `scripts/validate-push.sh` for CI with `commitlint`
    3. Update pre-commit hooks to use `commitlint` consistently
    4. Consolidate local config files (`.commitlint-with-baseline.js` or `.commitlintrc-local.js`)
    5. Fix or remove broken `ignores` function
    6. Update docs for unified validation process
  - **Done-when:** `commitlint` is the sole validation tool used consistently in local and CI environments

- [ ] **T004: Add Automated Vulnerability Scanning to CI**
  - **Context:** Missing `govulncheck` step in CI/CD pipeline
  - **Action:**
    1. Add `govulncheck` installation to CI/Release workflows
    2. Configure CI to fail on Critical/High vulnerabilities
    3. Document in `CONTRIBUTING.md`
  - **Done-when:** CI/Release builds fail if Critical/High vulnerabilities are found

- [ ] **T005: Enforce Robust Pre-commit Hook Installation**
  - **Context:** Inconsistent hook installation can lead to CI failures
  - **Action:**
    1. Enhance `scripts/setup.sh` to verify pre-commit CLI installation
    2. Update Makefile to install all required hooks (commit-msg, pre-push, post-commit)
    3. Make hook installation mandatory in documentation
    4. Add CI validation for pre-commit configuration
  - **Done-when:** Pre-commit hooks are consistently installed for all contributors

### Security & Stability
- [ ] **T006: Audit Test Mocks for Race Conditions**
  - **Context:** `MockFileWriter` needed mutex protection, other mocks likely have similar issues
  - **Action:**
    1. Review all test mocks for shared mutable state
    2. Add mutex protection where necessary
    3. Ensure all constructors initialize mutexes correctly
    4. Run `go test -race ./...` to verify
  - **Done-when:** Race detector reports no issues

- [ ] **T007: Fix PR #24 Commit Message Format**
  - **Context:** CI Resolution Tasks - Invalid Commit Message Format in PR #24
  - **Action:**
    1. Create backup branch of current `feature/automated-semantic-versioning`
    2. Use interactive rebase to fix commit with invalid format
    3. Fix commit with footer warning (blank line between body/footer)
    4. Force-push corrected branch with `--force-with-lease`
  - **Done-when:** All commits follow Conventional Commits format and CI passes

## High Priority (P1)

### Testing & Quality
- [ ] **T008: Restore and Enforce Coverage Thresholds**
  - **Context:** Coverage thresholds temporarily reduced to 64%
  - **Action:**
    1. Raise overall threshold to 75% in all scripts and CI
    2. Implement per-package coverage thresholds for critical packages
    3. Integrate E2E tests as required blocking CI step
    4. Document plan to reach 90% coverage
  - **Done-when:** CI fails below 75% overall and critical package thresholds

### Developer Experience
- [ ] **T009: Fix golangci-lint Issues Identified by Expanded Linter Set**
  - **Context:** Expanded linter set in T002 revealed 786 issues that need to be fixed
  - **Action:**
    1. Create a plan to gradually address each category of linter issues
    2. Fix high-priority issues (security, potential bugs) first
    3. Address test-related issues (thelper, tparallel)
    4. Fix code quality issues (funlen, gocyclo, etc.)
  - **Done-when:** All issues are fixed and golangci-lint passes in CI and pre-commit hooks

- [ ] **T010: Implement Automated Pre-commit Hook Installation**
  - **Context:** CI Resolution Tasks - Prevention Measures
  - **Action:**
    1. Modify `scripts/setup.sh` for automatic hook installation
    2. Update Makefile with target for hooks
    3. Add validation during Git operations
  - **Done-when:** Running setup automatically installs all required hooks

- [ ] **T010: Implement Commit Message Validation for Branch Commits**
  - **Context:** CI Resolution Tasks - Prevention Measures
  - **Action:**
    1. Create `scripts/validate-pr-commits.sh` for local validation
    2. Add documentation on usage
    3. Ensure same rules as CI workflow
  - **Done-when:** Developers can validate branch history before pushing

- [ ] **T011: Add Pre-push Hook for Commit Message Validation**
  - **Context:** CI Resolution Tasks - Prevention Measures
  - **Action:**
    1. Update `.pre-commit-config.yaml` for pre-push validation
    2. Configure to run validation from branch point
    3. Document in developer docs
  - **Done-when:** Pushing invalid commits is blocked before reaching remote

- [ ] **T012: Create Repository-wide Git Commit Template**
  - **Context:** CI Resolution Tasks - Long-term Improvements
  - **Action:**
    1. Create `.github/commit-template.txt` with examples
    2. Update setup script to configure Git to use template
    3. Document usage in `CONTRIBUTING.md`
  - **Done-when:** New developers get template configured automatically

## Medium Priority (P2)

### Code Organization & Maintenance
- [ ] **T013: Centralize API Key Resolution Logic**
  - **Context:** API key resolution logic is scattered and complex
  - **Action:**
    1. Create single function for key resolution with clear precedence
    2. Document lookup order in code and docs
    3. Add comprehensive unit tests
  - **Done-when:** All code paths use centralized key resolution

### Developer Tooling
- [ ] **T014: Enhance CI Workflow with Better Error Messages**
  - **Context:** CI Resolution Tasks - Long-term Improvements
  - **Action:**
    1. Improve error messages in GitHub Actions workflow
    2. Configure commitlint for specific guidance
    3. Consider custom action for PR comments
  - **Done-when:** CI failures provide clear, actionable feedback

- [ ] **T015: Implement Commitizen for Guided Commit Creation**
  - **Context:** CI Resolution Tasks - Long-term Improvements
  - **Action:**
    1. Add Commitizen to project dependencies
    2. Create configuration file for project standards
    3. Update documentation for usage
    4. Add Makefile target/script for easy access
  - **Done-when:** Commitizen is configured and working with standards

- [ ] **T016: Create Quick Reference Guide for Conventional Commits**
  - **Context:** CI Resolution Tasks - Prevention Measures
  - **Action:**
    1. Create concise reference in `docs/conventional-commits-guide.md`
    2. Include examples of valid messages for common types
    3. Document pitfalls and solutions
    4. Link from `CONTRIBUTING.md` and `README.md`
  - **Done-when:** Reference guide exists with clear examples

## Long-Term Improvements (P3)

### Core Features & Value Delivery

- [ ] **T017: Develop Flexible Workflow Engine**
  - **Context:** BACKLOG.md Core Features
  - **Action:**
    1. Design composable multi-step AI workflow system
    2. Support chaining operations with context passing
    3. Include built-in common workflows
  - **Done-when:** Users can define and execute custom processing pipelines

- [ ] **T018: Standardize Output Handling & Add JSON Output Mode**
  - **Context:** BACKLOG.md Medium Priority Features
  - **Action:**
    1. Route primary results to stdout, logs to stderr
    2. Make file output optional, defaulting to stdout
    3. Implement `--output-format json` flag
  - **Done-when:** Output handling is consistent and supports machine-readable formats

- [ ] **T019: Enable Custom System Prompts & Model Parameters**
  - **Context:** BACKLOG.md Medium Priority Features
  - **Action:**
    1. Allow custom system prompts via CLI flags/config
    2. Support fine-tuning of generation parameters
    3. Implement validation for all parameters
  - **Done-when:** Users can customize prompts and parameters

- [ ] **T020: Implement Token Counting & Cost Estimation**
  - **Context:** BACKLOG.md Medium Priority Features
  - **Action:**
    1. Track token usage with provider APIs
    2. Provide warnings for approaching limits
    3. Log estimated request costs
  - **Done-when:** Token usage is tracked with cost estimates

### Technical Excellence

- [ ] **T021: Improve Registry API Service Testability**
  - **Context:** BACKLOG.md Technical Excellence
  - **Action:**
    1. Apply Dependency Inversion Principle
    2. Refactor to use interfaces for dependencies
    3. Eliminate internal mocking
  - **Done-when:** Registry API is testable without policy violations

- [ ] **T022: Audit and Fix `context.Context` Propagation**
  - **Context:** BACKLOG.md Technical Excellence
  - **Action:**
    1. Ensure correct context handling throughout codebase
    2. Add race detection to CI
    3. Fix any race conditions
  - **Done-when:** `context.Context` is correctly used and propagated

- [ ] **T023: Consolidate Test Mock Implementations**
  - **Context:** BACKLOG.md Technical Debt Reduction
  - **Action:**
    1. Move common mocks to `internal/testutil`
    2. Update test files to use shared mocks
  - **Done-when:** Code duplication in tests is reduced

## Roadmap for Future Consideration

### Advanced Features
- [ ] **T024: Add Context Preprocessing (Summarization)**
- [ ] **T025: Auto-Select Models Based on Task/Context**
- [ ] **T026: Integrate AST Parsing for Code Understanding**
- [ ] **T027: Add Git Integration for Context**
- [ ] **T028: Add Code Generation Mode & Patch Output**

### Provider & Model Support
- [ ] **T029: Add Gemini Grounding Support**
- [ ] **T030: Add Grok API Support**

### Operational Excellence
- [ ] **T031: Use Distinct Exit Codes for Different Outcomes**
- [ ] **T032: Implement Metrics and Tracing**
- [ ] **T033: Advanced Cost Tracking**

## Task Validation Guidelines

For PR approval and task completion, verify:
1. ✅ Automated tests pass in CI pipeline
2. ✅ Linters run with strict configuration and pass
3. ✅ Tool versions are pinned and documented
4. ✅ Pre-commit hooks are correctly installed and functional
5. ✅ Vulnerability scans pass with no critical/high findings
6. ✅ E2E tests pass as part of CI
7. ✅ Coverage thresholds are met or exceeded
8. ✅ Documentation is updated and accurate
