# TODO: Fix CI Test Failures - Immediate Action Required

## Critical Path: Resolve TestMainConfigurationOptions Failures

**Current Issue**: PR #98 CI failing due to test flag contamination in subprocess tests
**Failed Tests**: All 5 subtests in TestMainConfigurationOptions
**Root Cause**: ParseFlags() tries to parse Go test flags like `-test.run` as application flags

---

## Phase 1: Implementation - Fix ParseFlags Function (URGENT)

### Core Fix: Filter Test Flags in ParseFlags
- [x] **Modify ParseFlags function** in `internal/cli/flags.go` (lines 85-89)
- [x] **Add test flag filtering logic** before calling ParseFlagsWithEnv
- [x] **Import strings package** if not already imported
- [x] **Preserve all existing functionality** - only filter out `-test.*` flags

#### Implementation Details:
```go
func ParseFlags() (*config.CliConfig, error) {
    // Filter out test flags from os.Args
    var filteredArgs []string
    for _, arg := range os.Args[1:] {
        if !strings.HasPrefix(arg, "-test.") {
            filteredArgs = append(filteredArgs, arg)
        }
    }

    flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    return ParseFlagsWithEnv(flagSet, filteredArgs, os.Getenv)
}
```

### Code Location:
- **File**: `/Users/phaedrus/Development/thinktank/internal/cli/flags.go`
- **Function**: `ParseFlags()` (lines 85-89)
- **Change**: Replace `os.Args[1:]` with `filteredArgs`

---

## Phase 2: Testing & Validation (CRITICAL)

### Local Testing
- [x] **Run individual failing test** to verify fix:
  ```bash
  go test -v -run TestMainConfigurationOptions ./internal/cli
  ```
- [x] **Run all CLI tests** to ensure no regressions:
  ```bash
  go test -v ./internal/cli
  ```
- [x] **Verify specific subtests pass**:
  - [x] main_with_custom_timeout
  - [x] main_with_rate_limiting
  - [x] main_with_custom_permissions
  - [x] main_with_multiple_models
  - [x] main_with_file_filtering

### Edge Case Testing
- [x] **Test with various test flags** to ensure robust filtering:
  ```bash
  go test -v -test.v -test.count=1 ./internal/cli
  ```
- [x] **Test production behavior** remains unchanged:
  ```bash
  go run cmd/thinktank/main.go --help
  ```

### Full Test Suite Validation
- [x] **Run complete test suite** to ensure no regressions:
  ```bash
  go test ./...
  ```
- [x] **Verify coverage maintained** (should still be 80%+):
  ```bash
  go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | tail -1
  ```

---

## Phase 3: CI Validation (FINAL STEP)

### Pre-Commit Verification
- [x] **Run linting** to ensure code quality:
  ```bash
  golangci-lint run ./...
  ```
- [x] **Run formatting** to ensure consistency:
  ```bash
  go fmt ./...
  ```
- [x] **Run race detection** on affected tests:
  ```bash
  go test -race ./internal/cli
  ```

### Commit and Push
- [ ] **Create conventional commit** with proper message:
  ```bash
  git add internal/cli/flags.go
  git commit -m "fix: filter test flags in ParseFlags to prevent subprocess test failures

  - Add test flag filtering logic to ParseFlags function
  - Prevents Go test flags like -test.run from being parsed as application flags
  - Fixes TestMainConfigurationOptions subprocess test failures
  - Maintains backward compatibility and production behavior

  Resolves CI test failures in PR #98"
  ```

### CI Monitoring
- [ ] **Push changes** and monitor CI pipeline:
  ```bash
  git push origin fix/multi-model-reliability
  ```
- [ ] **Verify CI test job passes** - specifically check TestMainConfigurationOptions
- [ ] **Confirm all other CI jobs remain green** (Build, Lint, Security, etc.)
- [ ] **Validate PR #98 shows green status** and is ready for merge

---

## Phase 4: Post-Fix Validation (VERIFICATION)

### Success Criteria Verification
- [ ] **All 5 TestMainConfigurationOptions subtests pass** ✅
- [ ] **No regression in other existing tests** ✅
- [ ] **CI pipeline completes successfully** ✅
- [ ] **Production functionality remains unaffected** ✅
- [ ] **Solution is maintainable and future-proof** ✅

### Documentation Updates
- [ ] **Update CI-FAILURE-SUMMARY.md** with resolution status
- [ ] **Update CI-RESOLUTION-PLAN.md** with implementation results
- [ ] **Clean up temporary analysis files** (CI-FAILURE-SUMMARY.md, CI-RESOLUTION-PLAN.md)

---

## Risk Mitigation & Rollback Plan

### Low-Risk Change Verification
- ✅ **Change is isolated** to argument preprocessing only
- ✅ **No impact on core application logic**
- ✅ **Easy to revert** if issues arise
- ✅ **Well-defined scope** (only affects test flag handling)

### Rollback Procedure (if needed)
```bash
# If issues arise, revert the change:
git revert <commit-hash>
git push origin fix/multi-model-reliability
```

### Alternative Solutions (backup plans)
1. **Environment-only test approach** - Remove `-test.run` flags entirely
2. **Test restructuring** - Modify tests to avoid subprocess with test flags
3. **Build tag approach** - Create test-specific ParseFlags version

---

## Expected Timeline

- **Phase 1 (Implementation)**: 15 minutes
- **Phase 2 (Testing)**: 30 minutes
- **Phase 3 (CI Validation)**: 15 minutes
- **Phase 4 (Verification)**: 10 minutes
- **Total Estimated Time**: ~1.25 hours

---

## Success Metrics

1. **Primary Goal**: All TestMainConfigurationOptions subtests pass ✅
2. **Secondary Goal**: No test regressions in full suite ✅
3. **Tertiary Goal**: CI pipeline green for PR #98 ✅
4. **Quality Goal**: Solution prevents future test flag issues ✅
