# TODO - Critical Issues for Current PR

## BLOCKING ISSUES (Must Fix Before Merge)

- [x] **T001 · Bugfix · P1: Fix main.go functional regression**
    - **Context:** main.go uses `exec.Command("go", "run", ...)` which breaks standard Go deployment model
    - **Action:**
        1. Revert main.go to direct function delegation pattern
        2. Import `github.com/phrazzld/thinktank/cmd/thinktank` and call `thinktank.Main()`
        3. Remove the `exec.Command` approach entirely
    - **Done-when:**
        1. main.go directly calls thinktank.Main() without subprocess execution
        2. Binary can be deployed and run in environments without Go toolchain
        3. No performance degradation from recompilation on every execution
    - **Verification:**
        1. Build binary with `go build .` and run without Go installed
        2. Verify execution time is consistent (no recompilation delay)
    - **Depends-on:** none

- [x] **T002 · Bugfix · P1: Add error handling to quality gate config action**
    - **Context:** `.github/actions/read-quality-gate-config/action.yml` installs yq without error handling
    - **Action:**
        1. Add error handling to `wget` command for yq download
        2. Add error handling to `chmod` command for yq permissions
        3. Ensure script exits immediately if yq installation fails
    - **Done-when:**
        1. Script fails fast if yq download fails
        2. Script fails fast if yq chmod fails
        3. Clear error messages indicate installation failure cause
    - **Verification:**
        1. Test action with network disconnected to verify wget failure handling
        2. Test action with read-only filesystem to verify chmod failure handling
    - **Depends-on:** none

- [x] **T003 · Security · P1: Pin CI tool dependencies to specific versions**
    - **Context:** CI workflows use `@latest` for govulncheck, go-licenses, gosec creating supply chain risk
    - **Action:**
        1. Pin govulncheck to specific version in `.github/workflows/ci.yml`
        2. Pin go-licenses to specific version in `.github/workflows/security-gates.yml`
        3. Pin gosec to specific version in `.github/workflows/security-gates.yml`
        4. Document version update process in DEPENDENCY_UPDATES.md
    - **Done-when:**
        1. All security tools use pinned versions (no @latest)
        2. CI builds are reproducible with same tool versions
        3. Documentation explains how to update pinned versions
    - **Verification:**
        1. Run CI twice and verify identical tool versions are used
        2. Check that no `@latest` remains in workflow files
    - **Depends-on:** none

- [x] **T006 · Bugfix · P1: Fix missing checkout step in Go CI workflow read-config job**
    - **Context:** `.github/workflows/ci.yml` read-config job uses local action without checking out repository first
    - **Action:**
        1. Add `actions/checkout@v4` step as first step in read-config job
        2. Ensure step is positioned before `uses: ./.github/actions/read-quality-gate-config`
        3. Use consistent naming pattern with other jobs (`Checkout code`)
    - **Done-when:**
        1. read-config job includes checkout step before local action usage
        2. Go CI workflow can successfully execute read-config job
        3. Quality gate configuration outputs are properly set for dependent jobs
    - **Verification:**
        1. Verify read-config job passes in CI run
        2. Confirm dependent jobs (lint, test, build) can start properly
    - **Depends-on:** none

- [x] **CI-001 · Bugfix · P1: Fix TruffleHog duplicate --fail flag in security-gates.yml**
    - **Context:** PR #79 CI failure - TruffleHog receives duplicate --fail flag causing "flag 'fail' cannot be repeated" error
    - **Action:**
        1. Edit `.github/workflows/security-gates.yml` line 115
        2. Change `extra_args: --only-verified --fail` to `extra_args: --only-verified`
        3. Remove duplicate --fail flag (TruffleHog action includes it by default)
    - **Done-when:**
        1. Secret Detection Scan job passes without flag conflict errors
        2. TruffleHog still properly fails on actual secret detection
        3. Security scanning pipeline fully operational
    - **Verification:**
        1. Run Secret Detection Scan job and verify success
        2. Test with dummy secret to ensure scan still catches secrets
    - **Depends-on:** none

- [x] **CI-002 · Infrastructure · P1: Create missing docker/e2e-test.Dockerfile for E2E testing**
    - **Context:** PR #79 CI failure - Test job fails because `docker/e2e-test.Dockerfile` doesn't exist
    - **Action:**
        1. Create `docker/` directory if it doesn't exist
        2. Create `docker/e2e-test.Dockerfile` with multi-stage Go build
        3. Include Go environment, dependencies, and thinktank binary
        4. Optimize for CI execution speed and environment isolation
    - **Done-when:**
        1. Docker image builds successfully in CI environment
        2. Container can execute thinktank binary
        3. E2E tests run successfully inside container
        4. Test environment variables properly accessible
    - **Verification:**
        1. Local Docker build succeeds: `docker build -f docker/e2e-test.Dockerfile -t thinktank-e2e:latest .`
        2. CI Test job completes E2E test phase without errors
    - **Depends-on:** none

- [x] **CI-003 · Investigation · P1: Fix gosec SAST internal error with internal/cli package**
    - **Context:** PR #79 CI failure - gosec reports "package internal/cli without types was imported" internal error
    - **Action:**
        1. Investigate if `internal/cli` package exists in codebase
        2. Check for import cycles or missing build constraints
        3. Test gosec locally with verbose output for debugging
        4. Fix package structure issues or gosec configuration
    - **Done-when:**
        1. gosec completes analysis without internal errors
        2. SAST report generated successfully (JSON + text formats)
        3. Security findings properly categorized and actionable
    - **Verification:**
        1. Local gosec run: `gosec -fmt json -out gosec-report.json -severity medium ./...`
        2. CI Static Application Security Testing job passes
    - **Depends-on:** CI investigation to determine root cause

- [x] **CI-004 · Verification · P2: Verify Secret Detection Scan effectiveness after TruffleHog fix**
    - **Context:** Ensure TruffleHog flag fix maintains security scanning effectiveness
    - **Action:**
        1. Run Secret Detection Scan job after CI-001 fix
        2. Test scan with dummy secret to verify detection still works
        3. Verify JSON and text reports generated correctly
    - **Done-when:**
        1. Scan passes on clean code without configuration errors
        2. Scan properly detects and fails on test secrets
        3. Security scanning maintains same effectiveness as before
    - **Verification:**
        1. CI Secret Detection Scan job shows green status
        2. Test secret detection triggers proper failure behavior
    - **Depends-on:** CI-001

- [x] **CI-005 · Verification · P2: Test Docker E2E build and execution in CI environment**
    - **Context:** Validate Docker configuration works properly in CI after creating Dockerfile
    - **Action:**
        1. Monitor CI Test job after CI-002 implementation
        2. Verify Docker image builds successfully
        3. Ensure E2E tests execute properly in container
        4. Check test output capture and environment variable access
    - **Done-when:**
        1. Docker build completes without errors in CI
        2. E2E tests run successfully inside container
        3. Test results properly reported back to CI
    - **Verification:**
        1. CI Test job passes E2E test phase
        2. Container execution logs show successful test runs
    - **Depends-on:** CI-002

- [x] **CI-006 · Verification · P2: Validate SAST scan completion after package fix**
    - **Context:** Ensure gosec SAST analysis works properly after resolving internal/cli package issues
    - **Action:**
        1. Run Static Application Security Testing job after CI-003 fix
        2. Verify gosec completes without internal errors
        3. Check SAST report generation and findings categorization
    - **Done-when:**
        1. gosec analysis completes successfully
        2. SAST reports generated in both JSON and text formats
        3. Security findings properly actionable and accurate
    - **Verification:**
        1. CI SAST job shows successful completion
        2. Generated reports contain meaningful security analysis
    - **Depends-on:** CI-003

## HIGH PRIORITY ISSUES (Fix in Next Sprint)

- [x] **T004 · Bugfix · P2: Fix race conditions in benchmark tests**
    - **Context:** `cmd/thinktank/cli_benchmark_test.go` mutates global os.Args in parallel benchmarks
    - **Action:**
        1. Refactor ParseFlags function to accept args slice parameter
        2. Update benchmarks to pass explicit args instead of mutating os.Args
        3. Ensure thread safety for parallel benchmark execution
    - **Done-when:**
        1. Benchmarks run reliably in parallel mode
        2. No global state mutation in benchmark loops
        3. ParseFlags accepts args parameter instead of reading os.Args
    - **Verification:**
        1. Run `go test -bench . -count=10` and verify consistent results
        2. Run benchmarks with `-parallel` flag without race conditions
    - **Depends-on:** none

- [x] **T005 · Refactor · P2: Add context propagation to FileWriter interface**
    - **Context:** FileWriter.SaveToFile doesn't accept context, breaking correlation ID traceability
    - **Action:**
        1. Update FileWriter interface to accept `ctx context.Context` as first parameter
        2. Update all FileWriter implementations to propagate context to audit logging
        3. Update all FileWriter callers to pass context
    - **Done-when:**
        1. FileWriter.SaveToFile signature includes context parameter
        2. Audit logs from file operations include correlation IDs
        3. End-to-end traceability maintained through file operations
    - **Verification:**
        1. Verify correlation IDs appear in file operation audit logs
        2. Test context deadline propagation through file operations
    - **Depends-on:** none

## MAINTENANCE TASKS (Lower Priority)

- [x] **CI-007 · Enhancement · P3: Add CI configuration validation to pre-commit hooks**
    - **Context:** Prevent future CI configuration errors like duplicate flags or missing files
    - **Action:**
        1. Create validation script for GitHub Actions workflow files
        2. Check for common configuration issues (duplicate flags, missing files)
        3. Add script to pre-commit hooks configuration
        4. Document validation rules and how to extend them
    - **Done-when:**
        1. Pre-commit hook validates workflow syntax and references
        2. Common configuration errors caught before commit
        3. Documentation explains validation rules
    - **Verification:**
        1. Pre-commit hook catches intentional configuration errors
        2. Valid configurations pass validation without issues
    - **Depends-on:** CI-001, CI-002, CI-003

- [x] **CI-008 · Cleanup · P3: Remove temporary CI analysis files**
    - **Context:** Clean up analysis files created during CI failure investigation
    - **Action:**
        1. Remove CI-FAILURE-SUMMARY.md after issues resolved
        2. Remove CI-RESOLUTION-PLAN.md after implementation complete
        3. Update .gitignore if needed to prevent future temporary files
    - **Done-when:**
        1. Temporary analysis files removed from repository
        2. CI issues fully resolved and verified
        3. No temporary investigation artifacts remain
    - **Verification:**
        1. Files no longer present in repository
        2. All CI jobs passing consistently
    - **Depends-on:** CI-004, CI-005, CI-006
