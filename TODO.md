# Todo

## Phase 1: Foundation Enhancement
- [x] **T001 · Refactor · P1: enforce fail-fast behavior in ci scripts**
    - **Context:** Phase 1, Step 1.1: Implement True Fail-Fast Behavior
    - **Action:**
        1. Add `set -e` or `set -eo pipefail` to all shell scripts in `scripts/ci/` and related CI execution paths.
        2. Remove all `|| true` and `|| echo` patterns that suppress failures.
    - **Done‑when:**
        1. All CI shell scripts are configured to exit immediately on any command failure.
        2. A PR with a failing script command correctly fails the CI job.
    - **Verification:**
        1. Introduce a failing command (e.g., `false`) into a script and confirm the CI job halts immediately.
    - **Depends‑on:** none

- [x] **T002 · Refactor · P1: implement explicit job dependency graph in ci workflow**
    - **Context:** Phase 1, Step 1.1: Implement explicit job dependencies with `needs` relationships
    - **Action:**
        1. Update `.github/workflows/ci.yml` to use the `needs` keyword to enforce the sequential `Quality Gate Hierarchy` defined in the plan.
    - **Done‑when:**
        1. The GitHub Actions workflow graph correctly visualizes the stage dependencies.
        2. A failure in a `Stage 1` job (e.g., Code Quality) prevents all `Stage 2` jobs (e.g., Testing) from starting.
    - **Verification:**
        1. Intentionally fail an early-stage job and confirm that all dependent downstream jobs are skipped.
    - **Depends‑on:** none

- [x] **T003 · Refactor · P1: update overall test coverage threshold to 90%**
    - **Context:** Phase 1, Step 1.2: Restore Coverage Threshold Enforcement
    - **Action:**
        1. Change the coverage threshold from 64% to 90% in all relevant files, including `.github/workflows/ci.yml` and `scripts/check-coverage.sh`.
    - **Done‑when:**
        1. The CI pipeline's coverage gate fails if overall code coverage is below 90%.
    - **Verification:**
        1. Push a commit that lowers coverage below 90% and confirm the build fails at the coverage check.
    - **Depends‑on:** T001

- [x] **T004 · Feature · P2: implement package-specific coverage enforcement**
    - **Context:** Phase 1, Step 1.2: Implement package-specific coverage enforcement (critical packages: 95%)
    - **Action:**
        1. Create or update `scripts/ci/check-package-specific-coverage.sh` to parse a configuration of critical packages and their required thresholds (e.g., 95%).
        2. Integrate this script into the CI workflow as a new job in the Testing Gates stage.
    - **Done‑when:**
        1. The CI pipeline fails if any package defined as "critical" has coverage below 95%.
    - **Verification:**
        1. Lower the coverage of a critical package below 95% and confirm the new CI job fails.
    - **Depends‑on:** T003

- [ ] **T005 · Feature · P1: add trufflehog secret scanning to pre-commit and ci**
    - **Context:** Phase 1, Step 1.3: Add pre-commit secret scanning with truffleHog
    - **Action:**
        1. Integrate `truffleHog` into the project's pre-commit hook configuration.
        2. Add a `truffleHog` scanning job to a new `.github/workflows/security-gates.yml` workflow.
    - **Done‑when:**
        1. Committing a file containing a secret pattern is blocked by the pre-commit hook.
        2. A PR containing a secret pattern fails the new CI security gate.
    - **Verification:**
        1. Attempt to commit a file with a fake API key and confirm the commit is blocked.
    - **Depends‑on:** none

- [ ] **T006 · Feature · P2: implement dependency license compliance checking**
    - **Context:** Phase 1, Step 1.3: Implement dependency license compliance checking
    - **Action:**
        1. Integrate a license checking tool (e.g., `go-licenses`) into the security workflow.
        2. Configure an allow-list of approved licenses and fail the build on any violations.
    - **Done‑when:**
        1. The CI pipeline fails if a dependency with a non-compliant license is introduced.
    - **Verification:**
        1. Add a dependency with a known disallowed license and confirm the security workflow fails.
    - **Depends‑on:** none

- [ ] **T007 · Feature · P2: add sast scanning to the security workflow**
    - **Context:** Phase 1, Step 1.3: Add SAST (Static Application Security Testing) scanning
    - **Action:**
        1. Add a SAST tool (e.g., CodeQL or gosec) to the `.github/workflows/security-gates.yml` workflow.
        2. Configure it to fail the build for high-severity findings.
    - **Done‑when:**
        1. A PR with a new, high-severity vulnerability is blocked by the SAST check.
    - **Verification:**
        1. Introduce code with a known vulnerability pattern and confirm the SAST job fails.
    - **Depends‑on:** none

## Phase 2: Quality Gate Orchestration
- [ ] **T008 · Feature · P2: implement the quality gate controller binary**
    - **Context:** Phase 2, Step 2.1: Implement Quality Gate Controller
    - **Action:**
        1. Implement the Go interfaces and types for gates, results, and reports in `internal/cicd/quality_gates.go`.
        2. Implement the core orchestration logic in `internal/cicd/gate_orchestrator.go`.
        3. Create the command-line entrypoint in `cmd/quality-gates/main.go`.
    - **Done‑when:**
        1. A `quality-gates` binary can be built and executed.
        2. Unit test coverage for the new `internal/cicd` package exceeds 95%.
    - **Verification:**
        1. Run the compiled binary with a sample configuration and confirm it executes mock gates correctly.
    - **Depends‑on:** none

- [ ] **T009 · Feature · P2: implement performance regression detection gate**
    - **Context:** Phase 2, Step 2.2: Add Performance Regression Detection
    - **Action:**
        1. Implement a Go benchmark execution framework and a script (`scripts/performance/run-benchmarks.sh`) to run it.
        2. Implement baseline storage (e.g., as a CI artifact) and comparison logic in `internal/benchmarks/performance_gates.go`.
        3. Create a new CI workflow (`.github/workflows/performance-gates.yml`) that fails if performance degrades by more than 5%.
    - **Done‑when:**
        1. A new CI workflow runs benchmarks and compares them against a stored baseline.
        2. A PR that introduces a >5% performance degradation is blocked by the new workflow.
    - **Verification:**
        1. Intentionally slow down a benchmarked function and confirm the performance gate fails.
    - **Depends‑on:** none

- [ ] **T010 · Bugfix · P1: fix e2e test binary execution errors in ci**
    - **Context:** Phase 2, Step 2.3: Fix E2E Test Execution
    - **Action:**
        1. Investigate the root cause of the `exec format error` for the E2E binary in the CI environment.
        2. Apply the necessary fix, likely by setting `GOOS`/`GOARCH` correctly during the build or using a cross-compiler.
    - **Done‑when:**
        1. The E2E test job no longer fails with `exec format error` and can successfully start the test binary.
    - **Verification:**
        1. Observe a successful E2E test run in CI without execution format errors.
    - **Depends‑on:** none

- [ ] **T011 · Feature · P2: containerize the e2e test execution environment**
    - **Context:** Phase 2, Step 2.3: Fix E2E Test Execution
    - **Action:**
        1. Create a `docker/e2e-test.Dockerfile` that installs all necessary dependencies for running the E2E test suite.
        2. Modify the E2E job in `.github/workflows/ci.yml` to build this image and run the tests inside the container.
    - **Done‑when:**
        1. E2E tests execute successfully and reliably within the new Docker container in the CI environment.
    - **Verification:**
        1. The CI logs for the E2E job show that steps are being executed inside the Docker container.
    - **Depends‑on:** T010

## Phase 3: Advanced Quality Features
- [ ] **T012 · Feature · P2: implement emergency override system with audit trail**
    - **Context:** Phase 3, Step 3.1: Implement Emergency Override System
    - **Action:**
        1. Implement logic in `internal/cicd/emergency_overrides.go` to detect an override signal (e.g., a specific PR label) and bypass failed gates.
        2. Ensure every use of the override mechanism generates a structured audit log entry.
    - **Done‑when:**
        1. A PR with a specific label can bypass a failing quality gate.
        2. An audit log is produced as a CI artifact for every override event.
    - **Verification:**
        1. Add the override label to a failing PR and confirm it merges. Check the CI artifacts for the audit log.
    - **Depends‑on:** none

- [ ] **T013 · Feature · P3: create automatic tech-debt issue for overrides**
    - **Context:** Phase 3, Step 3.1: Create automatic follow-up issue creation for technical debt
    - **Action:**
        1. Create an issue template at `.github/ISSUE_TEMPLATE/quality-gate-override.yml`.
        2. Create a script or GitHub Action that, upon detecting an override, uses the GitHub API to create a new issue from the template.
    - **Done‑when:**
        1. Using a quality gate override automatically creates a new, trackable technical debt issue in the repository.
    - **Verification:**
        1. Use the override system and confirm that a corresponding issue is created.
    - **Depends‑on:** T012

- [ ] **T014 · Feature · P2: create quality dashboard generation and deployment workflow**
    - **Context:** Phase 3, Step 3.2: Create Quality Dashboard
    - **Action:**
        1. Implement a script (`scripts/quality/generate-dashboard.sh`) to collect quality metrics (coverage, lint/test results, etc.) from CI artifacts.
        2. Create a static HTML template (`docs/quality-dashboard/index.html`) to display the metrics.
        3. Create a new workflow (`.github/workflows/quality-dashboard.yml`) to run the script and deploy the resulting dashboard to GitHub Pages.
    - **Done‑when:**
        1. A quality dashboard is accessible via a GitHub Pages URL.
        2. The dashboard is automatically updated with data from recent CI runs.
    - **Verification:**
        1. Navigate to the GitHub Pages URL and verify that it displays quality metrics and historical trends.
    - **Depends‑on:** none

- [ ] **T015 · Chore · P2: configure dependabot for automatic dependency updates**
    - **Context:** Phase 3, Step 3.3: Implement Automatic Dependency Updates
    - **Action:**
        1. Create `.github/dependabot.yml` to enable version updates for Go modules.
        2. Create `.github/workflows/dependency-updates.yml` to run the full quality gate suite on PRs created by Dependabot.
        3. Configure auto-merge for passing security patch updates.
    - **Done‑when:**
        1. Dependabot creates PRs for outdated dependencies.
        2. These PRs are automatically tested by the CI pipeline.
        3. Passing security patches are automatically merged.
    - **Verification:**
        1. Observe that a Dependabot PR is created, CI runs, and a passing patch PR is merged without manual intervention.
    - **Depends‑on:** none

## Cross-Cutting Concerns
- [ ] **T016 · Refactor · P2: implement structured json logging for all new components**
    - **Context:** Logging & Observability Strategy: Structured Logging Implementation
    - **Action:**
        1. Ensure all new Go components (quality gate controller, override system) use a structured logging library (e.g., `slog`).
        2. Include correlation IDs in logs to trace execution flow.
    - **Done‑when:**
        1. Log output from the new Go binaries is in JSON format and includes contextual fields.
    - **Verification:**
        1. Run a new component and inspect its stdout to confirm logs are structured JSON.
    - **Depends‑on:** T008

- [ ] **T017 · Feature · P2: implement feature flags for rolling out new quality gates**
    - **Context:** Risk Analysis & Mitigation: "Quality gate implementation breaks existing workflows"
    - **Action:**
        1. Implement a mechanism (e.g., using repository variables or a config file) to enable or disable new, potentially disruptive quality gates.
        2. Refactor CI workflows to check these flags before running the corresponding jobs.
    - **Done‑when:**
        1. A new quality gate (e.g., Performance Regression) can be disabled via a configuration change without modifying the workflow YAML.
    - **Verification:**
        1. Disable a new gate via its feature flag and confirm the corresponding CI job is skipped.
    - **Depends‑on:** none

---
### Clarifications & Assumptions
- [ ] **Issue:** Finalize the list of team members who will have emergency override authority.
    - **Context:** Open Questions & Decisions Required #1
    - **Blocking?:** yes (for T012 permissions)

- [ ] **Issue:** Define the process for establishing and updating the initial performance benchmark baseline.
    - **Context:** Open Questions & Decisions Required #2
    - **Blocking?:** yes (for T009)

- [ ] **Issue:** Identify the initial list of legacy packages that will receive temporary exemptions from the 90% coverage threshold.
    - **Context:** Open Questions & Decisions Required #3
    - **Blocking?:** no

- [ ] **Issue:** Confirm GitHub Pages is the desired hosting platform for the quality dashboard.
    - **Context:** Open Questions & Decisions Required #4
    - **Blocking?:** no

- [ ] **Issue:** Identify any critical integration points with existing local development tools or team workflows.
    - **Context:** Open Questions & Decisions Required #5
    - **Blocking?:** no
