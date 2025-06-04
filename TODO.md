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

- [x] **T007: Fix PR #24 Commit Message Format (Migration to Go-based Validator)**
  - **Context:** Node.js commitlint dependency was causing CI failures, replaced with Go validator
  - **Action:**
    1. ✅ Create cmd/commitvalidate/ with Go-based conventional commit validator
    2. ✅ Add internal/commitvalidate/ with validation logic and tests
    3. ✅ Remove package.json and Node.js dependency
    4. ✅ Remove .commitlintrc-local.js configuration  
    5. ✅ Remove scripts/commit.sh and scripts/ci/validate-baseline-commits.sh
    6. ✅ Update CI workflows to use Go validator instead of npm commitlint
    7. ✅ Update pre-commit config to use new Go validator
    8. ✅ Update documentation to reflect Go-based validation
    9. ✅ Maintain baseline-aware validation (commits after May 18, 2025)
  - **Status:** READY TO COMMIT - Eliminates Node.js dependency causing CI failures
  - **Done-when:** Go-based commit validator is deployed and Node.js dependency removed

- [x] **T004: Add Automated Vulnerability Scanning to CI**
  - **Context:** Missing `govulncheck` step in CI/CD pipeline
  - **Action:**
    1. Add `govulncheck` installation to CI/Release workflows
    2. Configure CI to fail on Critical/High vulnerabilities
    3. Document in `CONTRIBUTING.md`
  - **Done-when:** CI/Release builds fail if Critical/High vulnerabilities are found

- [ ] **T005: Enforce Robust Pre-commit Hook Installation**
  - **Context:** Inconsistent hook installation can lead to CI failures
  - **Action:**
    1. ✅ Enhance `scripts/setup.sh` to verify pre-commit CLI installation
    2. ✅ Update Makefile to install all required hooks (commit-msg, pre-push, post-commit)
    3. ✅ Make hook installation mandatory in documentation
    4. ✅ Add troubleshooting guide for hook issues
    5. ✅ Add network connectivity checks
    6. ✅ Improve version validation with explicit comparison
  - **Status:** READY TO COMMIT - All implementation complete, blocked by T009 linter issues
  - **Done-when:** Pre-commit hooks are consistently installed for all contributors

### Security & Stability
- [x] **T006: Audit Test Mocks for Race Conditions**
  - **Context:** `MockFileWriter` needed mutex protection, other mocks likely have similar issues
  - **Action:**
    1. ✅ Review all test mocks for shared mutable state
    2. ✅ Add mutex protection where necessary
    3. ✅ Ensure all constructors initialize mutexes correctly
    4. ✅ Run `go test -race ./...` to verify
  - **Done-when:** Race detector reports no issues
  - **Completed:** Fixed race conditions in MockEnvironmentProvider, MockTimeProvider, BaseMockOutputWriter, MockAuditLogger, and MockFileWriter. All tests pass with race detector.

- [x] **T007: Fix PR #24 Commit Message Format**
  - **Context:** CI Resolution Tasks - Invalid Commit Message Format in PR #24
  - **Action:**
    1. Create backup branch of current `feature/automated-semantic-versioning`
    2. Use interactive rebase to fix commit with invalid format
    3. Fix commit with footer warning (blank line between body/footer)
    4. Force-push corrected branch with `--force-with-lease`
  - **Done-when:** All commits follow Conventional Commits format and CI passes

## High Priority (P1)

### Testing & Quality
- [x] **T008: Restore and Enforce Coverage Thresholds**
  - **Context:** Coverage thresholds temporarily reduced to 64%
  - **Action:**
    1. Raise overall threshold to 75% in all scripts and CI
    2. Implement per-package coverage thresholds for critical packages
    3. Integrate E2E tests as required blocking CI step
    4. Document plan to reach 90% coverage
  - **Done-when:** CI fails below 75% overall and critical package thresholds

### Developer Experience
- [~] **T009: Fix golangci-lint Issues Identified by Expanded Linter Set**
  - **Context:** Expanded linter set in T002 revealed 749 remaining issues (reduced from 786)
  - **Action:**
    **Phase 1: Test Helper Issues (53 issues) - COMPLETED**
    - ✅ Fix thelper violations: Add t.Helper() to test helper functions
    - ✅ Fix thelper parameter naming: Rename testing.TB parameters to 'tb'

    **Phase 2: Security Issues (47 issues) - COMPLETED**  
    - ✅ G301: Fix directory permissions (expect 0750 or less) - 20 issues
    - ✅ G304: Fix potential file inclusion via variable - 22 issues
    - ✅ G306: Fix WriteFile permissions (expect 0600 or less) - 19 issues
    - ✅ G302: Fix file permissions (expect 0600 or less) - 6 issues
    - ✅ G204: Fix subprocess with tainted input - 2 issues
    - ✅ G101: Fix potential hardcoded credentials - 1 issue

    **Phase 3: Code Quality Issues (377 issues) - PARTIAL COMPLETION**
    - ✅ package-comments (revive): Add proper package comments - 9 issues FIXED
    - ✅ empty-block (revive): Remove empty blocks - 1 issue FIXED  
    - ✅ var-naming (revive): Fix variable naming conventions - 3 issues FIXED
    - ✅ var-declaration (revive): Fix variable declarations - 1 issue FIXED
    - ✅ redefines-builtin-id (revive): Fix builtin function redefinition - 3 issues FIXED
    - 🔄 indent-error-flow (revive): Fix if-else indentation - 2/4 issues FIXED
    - 🔄 unused-parameter (revive): Remove or rename unused parameters - 14/360 issues FIXED
    - ❌ exported (revive): Fix exported types/functions - 26 issues REMAIN
    - ❌ unexported-return (revive): Fix unexported return types - 2 issues REMAIN

    **Phase 4: Function Length Issues (167 issues)**
    - funlen: Split functions longer than 60 lines or 40 statements - 167 issues

    **Phase 5: Complexity Issues (14 issues)**
    - gocyclo: Reduce cyclomatic complexity > 30 - 14 issues

    **Phase 6: String Constants (29 issues)**
    - goconst: Extract repeated strings to constants - 29 issues
  - **Done-when:** All 749 issues are fixed and golangci-lint passes in CI and pre-commit hooks

- [x] **T010: Implement Automated Pre-commit Hook Installation** `Simple` `Priority: Normal` `Infrastructure` `Completed`
  - **Context:** CI Resolution Tasks - Prevention Measures
  - **Action:**
    1. ✅ Modify `scripts/setup.sh` for automatic hook installation
    2. ✅ Update Makefile with target for hooks  
    3. ✅ Add validation during Git operations
  - **Status:** Completed - discovered that comprehensive automated hook installation was already fully implemented in scripts/setup.sh and Makefile, meeting all success criteria
  - **Done-when:** Running setup automatically installs all required hooks

- [x] **T010a: Implement Commit Message Validation for Branch Commits**
  - **Context:** CI Resolution Tasks - Prevention Measures
  - **Action:**
    1. ✅ Create `scripts/validate-pr-commits.sh` for local validation
    2. ✅ Add documentation on usage
    3. ✅ Ensure same rules as CI workflow
  - **Done-when:** Developers can validate branch history before pushing

- [x] **T011: Add Pre-push Hook for Commit Message Validation**
  - **Context:** CI Resolution Tasks - Prevention Measures
  - **Action:**
    1. ✅ Update `.pre-commit-config.yaml` for pre-push validation
    2. ✅ Configure to run validation from branch point
    3. ✅ Document in developer docs
  - **Done-when:** Pushing invalid commits is blocked before reaching remote

- [x] **T011a: Fix golangci-lint pre-commit hook configuration**
  - **Context:** golangci-lint runs on entire codebase even when no Go files are modified
  - **Action:**
    1. ✅ Update .pre-commit-config.yaml to only run golangci-lint on changed Go files
    2. ✅ Ensure hook respects SKIP environment variable
    3. ✅ Consider updating to match CI behavior (only on relevant files)
  - **Done-when:** golangci-lint only runs when Go files are actually modified
  - **Completed:** Added `types: [go]` filter to golangci-lint hook configuration

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
