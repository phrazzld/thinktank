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

- [ ] **T005 · Refactor · P2: Add context propagation to FileWriter interface**
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
