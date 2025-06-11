# TODO - E2E CI Failure Resolution

## CRITICAL ISSUES (Must Fix Before Merge)

- [x] **E2E-001 · Bugfix · P1: Fix Docker E2E container configuration for models.yaml**
    - **Context:** E2E tests fail because Docker container missing models.yaml at `/home/thinktank/.config/thinktank/models.yaml`
    - **Root Cause:** Binary expects user config directory structure, but Docker container doesn't create it
    - **Error:** `Failed to load configuration: configuration file not found at /home/thinktank/.config/thinktank/models.yaml`
    - **Action:**
        1. Modify `docker/e2e-test.Dockerfile` to create user config directory structure
        2. Copy `config/models.yaml` to `/home/thinktank/.config/thinktank/models.yaml`
        3. Set proper ownership with `chown -R thinktank:thinktank /home/thinktank`
        4. Position changes after user creation but before switching to thinktank user
    - **Done-when:**
        1. Docker image builds successfully with config directory structure
        2. `models.yaml` accessible to thinktank user in container at expected path
        3. TestBasicExecution finds "Gathering context" and "Generating plan" outputs
        4. E2E tests pass without configuration errors
    - **Verification:**
        1. Local Docker build: `docker build -f docker/e2e-test.Dockerfile -t thinktank-e2e:latest .`
        2. Test config access: `docker run --rm thinktank-e2e:latest ls -la /home/thinktank/.config/thinktank/`
        3. CI Test job passes E2E test phase
    - **Depends-on:** none

- [x] **E2E-002 · Verification · P1: Validate E2E tests pass after Docker configuration fix**
    - **Context:** Verify that Docker configuration fix resolves CI failure completely
    - **Action:**
        1. Trigger CI run after E2E-001 implementation
        2. Monitor Test job "Run E2E tests in Docker container" step
        3. Verify TestBasicExecution passes with expected outputs
        4. Confirm no exit code 4 configuration errors
    - **Done-when:**
        1. All CI checks pass (14/14)
        2. Test job completes without failures
        3. E2E test outputs include "Gathering context" and "Generating plan"
        4. No configuration file not found errors in logs
    - **Verification:**
        1. CI Status shows all green checkmarks
        2. E2E test logs show successful binary execution
        3. Output file `output/gemini-test-model.md` created as expected
    - **Depends-on:** E2E-001

## CLEANUP TASKS

- [x] **E2E-003 · Cleanup · P2: Remove temporary CI analysis files**
    - **Context:** Clean up CI failure analysis files after resolution
    - **Action:**
        1. Remove `CI-FAILURE-SUMMARY.md` after E2E tests pass
        2. Remove `CI-RESOLUTION-PLAN.md` after implementation complete
        3. Verify `.gitignore` patterns prevent future CI analysis file commits
    - **Done-when:**
        1. Temporary analysis files removed from repository
        2. CI issues fully resolved and verified
        3. No temporary investigation artifacts remain
    - **Verification:**
        1. Files no longer present in repository
        2. All CI jobs passing consistently
    - **Depends-on:** E2E-002

## ENHANCEMENT TASKS (Future Improvements)

- [x] **E2E-004 · Enhancement · P3: Add configuration fallback mechanisms**
    - **Context:** Make application more resilient for containerized environments
    - **Action:**
        1. Add environment variable-based configuration override capability
        2. Implement default configuration when file missing
        3. Improve error messages for configuration issues
        4. Add configuration validation and better diagnostics
    - **Done-when:**
        1. Application can run with environment-based config
        2. Graceful handling when models.yaml missing
        3. Clear error messages guide users on configuration setup
        4. Both file-based and env-based config tested
    - **Verification:**
        1. Binary runs successfully with env vars instead of file
        2. Helpful error messages when config invalid or missing
        3. Backward compatibility maintained with existing config files
    - **Depends-on:** E2E-002

- [x] **E2E-005 · Testing · P3: Add comprehensive configuration testing**
    - **Context:** Ensure robust configuration handling across scenarios
    - **Action:**
        1. Add tests for missing configuration file scenarios
        2. Add tests for invalid configuration content
        3. Add tests for environment variable overrides
        4. Add tests for configuration loading in different environments
    - **Done-when:**
        1. Configuration edge cases covered by tests
        2. Environment variable configuration tested
        3. Error handling for config issues validated
        4. Container vs local config loading tested
    - **Verification:**
        1. Test suite covers config loading scenarios
        2. Tests pass in both local and container environments
        3. Configuration errors properly caught and handled
    - **Depends-on:** E2E-004

## IMPLEMENTATION NOTES

### Critical Path
1. **E2E-001** (Docker fix) → **E2E-002** (Verification) → Merge ready
2. **E2E-003** (Cleanup) → Post-merge cleanup

### Enhancement Path
3. **E2E-004** (Config robustness) → **E2E-005** (Testing) → Future releases

### Key Files
- `docker/e2e-test.Dockerfile` - Primary fix target
- `config/models.yaml` - Source configuration file
- `internal/e2e/cli_basic_test.go` - Test validation
- `.github/workflows/ci.yml` - CI pipeline execution

### Success Criteria
- ✅ CI shows 14/14 checks passing
- ✅ E2E tests complete successfully in Docker container
- ✅ Configuration loading works in container environment
- ✅ No regression in other test suites
