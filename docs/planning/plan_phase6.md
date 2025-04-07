# Thinktank Test Suite Refactoring - Phase 6

## Phase 6: Refine E2E Tests

**Objective:** Ensure E2E tests focus on CLI interaction using the real filesystem effectively.

### Steps:

1. **Review E2E Test Scope:**
   - Verify `cli.e2e.test.ts` and `runThinktank.e2e.test.ts` use real compiled CLI
   - Ensure they utilize `e2eTestUtils.ts` for temporary directories/files
   - Focus assertions on:
     - Command execution (arguments, options)
     - Standard output/error
     - Created output files and their contents
     - Exit codes (0 for success, non-zero for errors)
   - Remove internal logic checks; treat CLI as a black box

2. **Ensure Proper Cleanup:**
   - Verify `afterAll`/`afterEach` hooks use `cleanupTestDir` 
   - Ensure cleanup works even if tests fail
