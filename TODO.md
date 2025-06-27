# 🚨 URGENT CI FAILURE RESOLUTION: TEST COVERAGE REGRESSION

## Critical Issue: Coverage Dropped to 87.8% (Below 90% Threshold)

**Root Cause**: Added 200+ lines of new functionality without corresponding tests
**Impact**: CI quality gates properly blocking PR merge (#100)
**Required Action**: Add comprehensive tests to restore 90% coverage threshold

---

## PHASE 1: IMMEDIATE COVERAGE RECOVERY (Target: 90%+)

### Priority 1: Fix internal/models Package Coverage (62.1% → 90%)
*Highest impact - Most untested new code*

#### Task 1: Add Tests for Token Estimation Functions
**Priority: Critical** | **Package**: `internal/models` | **Estimated Lines**: ~40
- **Target Functions**:
  - `EstimateTokensFromText(text string) int`
  - `EstimateTokensFromStats(charCount int, instructionsText string) int`
- **Test Requirements**:
  ```go
  func TestEstimateTokensFromText(t *testing.T) {
      tests := []struct {
          name     string
          text     string
          expected int
      }{
          {"empty text", "", 1000}, // overhead only
          {"simple text", "hello world", 1008}, // ~8 chars * 0.75 + overhead
          {"realistic text", strings.Repeat("code ", 200), 1750}, // real scenario
          {"unicode text", "测试文本", 1012}, // Unicode handling
          {"large text", strings.Repeat("x", 10000), 8500}, // Large input
      }
      // Implementation with precise calculations
  }
  ```
- **Verification**: `go test -cover ./internal/models` should show functions covered
- **Exit Criteria**: Both functions 100% covered with edge cases

#### Task 2: Add Tests for Model Selection Functions
**Priority: Critical** | **Package**: `internal/models` | **Estimated Lines**: ~60
- **Target Functions**:
  - `GetModelsWithMinContextWindow(minTokens int) []string`
  - `SelectModelsForInput(estimatedTokens int, availableProviders []string) []string`
  - `GetLargestContextModel(modelNames []string) string`
- **Test Requirements**:
  ```go
  func TestSelectModelsForInput(t *testing.T) {
      tests := []struct {
          name              string
          estimatedTokens   int
          availableProviders []string
          expectedModels    []string
      }{
          {
              name: "small input, all providers",
              estimatedTokens: 5000,
              availableProviders: []string{"openai", "gemini", "openrouter"},
              expectedModels: []string{"gemini-2.5-pro", "gpt-4.1", "o3", "o4-mini"}, // All large models
          },
          {
              name: "large input, limited models",
              estimatedTokens: 150000,
              availableProviders: []string{"gemini"},
              expectedModels: []string{"gemini-2.5-pro"}, // Only largest context models
          },
      }
      // Implementation with realistic test scenarios
  }
  ```
- **Mock Requirements**: Need to mock/control available models for testing
- **Verification**: Functions handle edge cases (empty providers, very large inputs)
- **Exit Criteria**: All model selection logic 100% covered

#### Task 3: Add Tests for Provider Detection Functions
**Priority: Critical** | **Package**: `internal/models` | **Estimated Lines**: ~30
- **Target Functions**:
  - `GetAvailableProviders() []string`
  - `GetAPIKeyEnvVar(provider string) string`
- **Test Requirements**:
  ```go
  func TestGetAvailableProviders(t *testing.T) {
      tests := []struct {
          name     string
          envVars  map[string]string
          expected []string
      }{
          {"no api keys", map[string]string{}, []string{}},
          {"openai only", map[string]string{"OPENAI_API_KEY": "sk-test"}, []string{"openai"}},
          {"all providers", map[string]string{
              "OPENAI_API_KEY": "sk-test",
              "GEMINI_API_KEY": "test-key",
              "OPENROUTER_API_KEY": "or-test",
          }, []string{"openai", "gemini", "openrouter"}},
      }
      // Implementation with environment variable mocking
  }
  ```
- **Environment Setup**: Use test environment variable control
- **Verification**: Provider detection works with various API key combinations
- **Exit Criteria**: 100% coverage for provider detection logic

### Priority 2: Fix internal/cli Package Coverage (78.5% → 90%)
*Complex business logic requiring careful testing*

#### Task 4: Add Tests for Model Selection Logic
**Priority: High** | **Package**: `internal/cli` | **Estimated Lines**: ~80
- **Target Function**: `selectModelsForConfig(simplifiedConfig *SimplifiedConfig) ([]string, string)`
- **Test Requirements**:
  ```go
  func TestSelectModelsForConfig(t *testing.T) {
      // Create temporary instruction files for testing
      smallInstructions := createTempFile(t, "small instructions")
      largeInstructions := createTempFile(t, strings.Repeat("complex task ", 1000))

      tests := []struct {
          name             string
          configFlags      uint8
          instructionFile  string
          expectedModels   []string
          expectedSynthesis string
      }{
          {
              name: "small input, no synthesis flag",
              configFlags: 0,
              instructionFile: smallInstructions,
              expectedModels: []string{"gemini-2.5-flash"},
              expectedSynthesis: "",
          },
          {
              name: "large input, multiple models",
              configFlags: 0,
              instructionFile: largeInstructions,
              expectedModels: []string{"gemini-2.5-pro", "gpt-4.1", "o3"},
              expectedSynthesis: "gemini-2.5-pro",
          },
          {
              name: "forced synthesis mode",
              configFlags: FlagSynthesis,
              instructionFile: smallInstructions,
              expectedModels: []string{"gemini-2.5-flash"},
              expectedSynthesis: "gemini-2.5-pro",
          },
      }
      // Implementation with file system mocking and API key setup
  }
  ```
- **Mock Requirements**: Control available API keys and file system access
- **File Setup**: Create temporary instruction files with controlled content
- **Verification**: Logic correctly estimates tokens and selects appropriate models
- **Exit Criteria**: Core model selection logic 100% covered

#### Task 5: Add Tests for Logger Routing Logic
**Priority: High** | **Package**: `internal/cli` | **Estimated Lines**: ~40
- **Target Function**: `createLoggerWithRouting(cfg *config.MinimalConfig, outputDir string) (logutil.LoggerInterface, *LoggerWrapper)`
- **Test Requirements**:
  ```go
  func TestCreateLoggerWithRouting(t *testing.T) {
      tempDir := t.TempDir()

      tests := []struct {
          name         string
          config       *config.MinimalConfig
          outputDir    string
          expectFile   bool
          expectStderr bool
      }{
          {
              name: "json logs to console when verbose",
              config: &config.MinimalConfig{Verbose: true},
              outputDir: "",
              expectFile: false,
              expectStderr: true,
          },
          {
              name: "json logs to file by default",
              config: &config.MinimalConfig{},
              outputDir: tempDir,
              expectFile: true,
              expectStderr: false,
          },
          {
              name: "json logs to console when flag set",
              config: &config.MinimalConfig{JsonLogs: true},
              outputDir: tempDir,
              expectFile: false,
              expectStderr: true,
          },
      }
      // Implementation with file system verification
  }
  ```
- **File System Testing**: Verify actual log file creation in temp directories
- **Logger Verification**: Test that correct logger type is returned
- **Cleanup Testing**: Verify LoggerWrapper.Close() works correctly
- **Exit Criteria**: All logging routing paths 100% covered

#### Task 6: Add Tests for Enhanced CLI Flag Parsing
**Priority: Medium** | **Package**: `internal/cli` | **Estimated Lines**: ~30
- **Target Functions**: Enhanced methods in `SimplifiedConfig` and flag parsing
- **Test Requirements**:
  ```go
  func TestSimplifiedConfigFlags(t *testing.T) {
      tests := []struct {
          name     string
          flags    uint8
          expected map[string]bool
      }{
          {
              name: "no flags set",
              flags: 0,
              expected: map[string]bool{
                  "dry_run": false, "verbose": false, "debug": false,
                  "quiet": false, "json_logs": false, "no_progress": false,
              },
          },
          {
              name: "all flags set",
              flags: FlagDryRun | FlagVerbose | FlagDebug | FlagQuiet | FlagJsonLogs | FlagNoProgress,
              expected: map[string]bool{
                  "dry_run": true, "verbose": true, "debug": true,
                  "quiet": true, "json_logs": true, "no_progress": true,
              },
          },
      }
      // Implementation testing all new flag constants
  }
  ```
- **Flag Validation**: Test all new flag constants work correctly
- **Bitfield Operations**: Verify flag manipulation functions
- **Exit Criteria**: All new CLI flag functionality 100% covered

---

## PHASE 2: COVERAGE VERIFICATION & CI RECOVERY

#### Task 7: Run Comprehensive Coverage Analysis
**Priority: High** | **Timeline**: After Task 1-6 completion
- **Commands**:
  ```bash
  # Generate detailed coverage report
  go test -coverprofile=coverage.out ./...
  go tool cover -html=coverage.out -o coverage.html

  # Check specific package coverage
  go test -cover ./internal/models
  go test -cover ./internal/cli
  go test -cover ./internal/config
  ```
- **Analysis Requirements**:
  - Identify any remaining uncovered lines
  - Verify overall coverage ≥90%
  - Check package-specific coverage meets standards
- **Documentation**: Save coverage.html for review and future reference
- **Exit Criteria**: Coverage report shows ≥90% overall, no critical uncovered paths

#### Task 8: Execute Full CI Test Suite Locally
**Priority: High** | **Dependencies**: Task 7 completed
- **Commands**:
  ```bash
  # Run all test categories that CI runs
  go test -v -race -short -parallel 4 ./internal/integration/...
  go test -v ./internal/cli/... ./internal/models/... ./internal/config/...
  go test -race ./... -count=1

  # Verify coverage threshold
  ./scripts/check-coverage.sh 90
  ```
- **Verification Steps**:
  - All tests pass without race conditions
  - Coverage threshold check passes
  - No new lint violations introduced
- **Performance Check**: Ensure test execution time reasonable (<5 minutes)
- **Exit Criteria**: Local test suite matches CI requirements and passes

#### Task 9: Commit Coverage Recovery Changes
**Priority: High** | **Dependencies**: Task 8 passes
- **Pre-commit Validation**:
  ```bash
  # Verify pre-commit hooks pass
  pre-commit run --all-files

  # Double-check coverage one more time
  go test -cover ./... | grep -E "(models|cli|config)"
  ```
- **Commit Message**: Follow conventional commit format
  ```
  test: add comprehensive tests for intelligent model selection and logging

  - Add 150+ lines of tests for internal/models package functions
  - Add 100+ lines of tests for internal/cli package functions
  - Restore test coverage from 87.8% to 90%+ to meet CI requirements
  - Include edge case testing for token estimation and model selection
  - Add file system and environment variable mocking for robust testing

  Fixes CI coverage regression introduced in feat: implement intelligent model selection
  ```
- **Verification**: Push and monitor CI status
- **Exit Criteria**: CI coverage check passes

---

## PHASE 3: PROCESS IMPROVEMENTS (PREVENT FUTURE REGRESSIONS)

#### Task 10: Implement Pre-commit Coverage Check
**Priority: Medium** | **Timeline**: After CI recovery
- **Implementation**:
  ```bash
  # Create scripts/check-coverage-delta.sh
  #!/bin/bash
  CURRENT=$(go test -cover ./... 2>/dev/null | grep -o 'coverage: [0-9.]*%' | grep -o '[0-9.]*' | tail -1)
  if (( $(echo "$CURRENT < 90" | bc -l) )); then
      echo "❌ Coverage $CURRENT% below 90% threshold"
      echo "🔧 Run: go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out"
      exit 1
  fi
  echo "✅ Coverage $CURRENT% meets 90% threshold"
  ```
- **Integration**: Add to `.pre-commit-config.yaml`
- **Testing**: Verify pre-commit hook prevents low-coverage commits
- **Exit Criteria**: Pre-commit hook blocks commits with <90% coverage

#### Task 11: Add Coverage Monitoring to CI
**Priority: Medium** | **Dependencies**: Pre-commit check working
- **Implementation**: Enhance CI to track coverage trends
- **Alerting**: Set up notifications for coverage regressions >5%
- **Dashboard**: Consider coverage trend visualization
- **Exit Criteria**: CI provides clear coverage feedback on every PR

#### Task 12: Document Testing Requirements
**Priority: Low** | **Timeline**: After process automation
- **Create**: `docs/testing-guidelines.md` with coverage requirements
- **PR Template**: Add coverage checklist to PR template
- **Developer Guide**: Update development workflow documentation
- **Exit Criteria**: Clear guidelines prevent future coverage regressions

---

## PHASE 4: CLEANUP & VERIFICATION

#### Task 13: Remove Temporary Analysis Files
**Priority: Low** | **Dependencies**: All critical tasks completed
- **Files to Remove**:
  - `CI-FAILURE-SUMMARY.md`
  - `CI-RESOLUTION-PLAN.md`
  - Any temporary test files or coverage reports
- **Git Status**: Ensure clean working directory
- **Exit Criteria**: Repository clean, only permanent improvements remain

#### Task 14: Update Documentation
**Priority: Low** | **Timeline**: Final cleanup
- **Update**: Project documentation reflects testing improvements
- **Examples**: Add testing examples to README if appropriate
- **Migration**: Update any documentation affected by new test patterns
- **Exit Criteria**: Documentation accurately reflects current testing approach

---

## SUCCESS METRICS & MONITORING

### Immediate Success (Tasks 1-9)
- [ ] Overall test coverage ≥90%
- [ ] `internal/models` package coverage ≥90%
- [ ] `internal/cli` package coverage ≥90%
- [ ] CI pipeline passes all checks
- [ ] PR #100 unblocked for merge

### Process Success (Tasks 10-12)
- [ ] Pre-commit coverage validation active
- [ ] Coverage regression alerts configured
- [ ] Zero coverage regressions in subsequent PRs
- [ ] Developer workflow includes coverage validation

### Quality Metrics
- [ ] Test execution time <5 minutes for full suite
- [ ] No flaky tests introduced
- [ ] All tests pass consistently with race detection
- [ ] Coverage quality maintained (not just quantity)

---

## EMERGENCY ESCALATION

If coverage cannot be restored within 2 days:
1. **Root Cause Analysis**: Deep dive into testability issues
2. **Architecture Review**: Consider refactoring for better testability
3. **Scope Reduction**: Identify minimum viable coverage increase
4. **Technical Debt**: Document coverage debt and timeline for resolution

**Note**: Emergency override NOT recommended per leyline guidance - this represents a "broken window" requiring proper resolution.

---

## ESTIMATED TIMELINE

- **Phase 1 (Critical)**: 1-2 days (Tasks 1-6)
- **Phase 2 (Recovery)**: 0.5 days (Tasks 7-9)
- **Phase 3 (Process)**: 1-2 days (Tasks 10-12)
- **Phase 4 (Cleanup)**: 0.5 days (Tasks 13-14)

**Total Estimated Time**: 3-5 days to full resolution with process improvements
