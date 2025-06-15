# Coverage Measurement Strategy

This document defines the test coverage measurement strategy for the thinktank project, including what is included/excluded from coverage calculations and the rationale behind these decisions.

## Overview

The thinktank project uses test coverage as a quality gate to ensure adequate testing of business logic and critical functionality. Coverage measurements focus on meaningful code that directly contributes to the application's core functionality.

## Included in Coverage

Coverage calculations **include** the following:

- **Core business logic packages** (`internal/thinktank`, `internal/llm`, etc.)
- **Provider implementations** (`internal/providers/*`)
- **Registry and configuration** (`internal/registry`, `internal/config`)
- **Utility packages** (`internal/fileutil`, `internal/logutil`, etc.)
- **All production code** in `cmd/` packages

## Excluded from Coverage

Coverage calculations **exclude** the following categories:

### 1. Test Infrastructure Packages

- **`internal/testutil`** - Test utility functions, mocks, and helpers
- **Test helper files** - Files matching patterns:
  - `*_test_helpers.go`
  - `*_test_utils.go`
  - `mock_*.go`
  - `mocks.go`

**Rationale**: Test infrastructure has different testing characteristics and is not core business logic. Coverage of test utilities is less meaningful since they are testing tools themselves.

### 2. Test-Only Packages

- **`internal/integration`** - Integration test package
- **`internal/e2e`** - End-to-end test package

**Rationale**: These packages contain test code that validates the overall system behavior rather than implementing core functionality.

### 3. Disabled Code

- **`/disabled/`** - Any packages or files in disabled directories

**Rationale**: Disabled code is not active in the application and should not affect coverage metrics.

## Coverage Thresholds

### Overall Project Threshold

- **Minimum**: 90% overall coverage (enforced in CI)
- **Quality Gate**: Hard requirement for all pull requests

### Package-Specific Thresholds

Critical packages have differentiated requirements based on their complexity:

| Package | Threshold | Classification | Rationale |
|---------|-----------|----------------|-----------|
| `internal/llm` | 95% | Critical | Core LLM interface and error handling |
| `internal/providers` | 80% | Critical | Provider abstraction layer |
| `internal/registry` | 75% | Critical | Model registry and configuration |
| `internal/thinktank` | 70% | Application | Complex orchestration logic with comprehensive testing |

## Implementation

### Coverage Scripts

The project uses several scripts to enforce coverage requirements:

1. **`scripts/check-coverage.sh`** - Overall coverage validation
2. **`scripts/check-package-coverage.sh`** - Per-package coverage reporting
3. **`scripts/ci/check-package-specific-coverage.sh`** - CI-specific package thresholds

All scripts consistently exclude test utility packages and helper files.

### CI Integration

Coverage is enforced in the CI pipeline:

- **Generation**: Coverage data is generated during the `test` job
- **Validation**: Both overall and package-specific thresholds are checked
- **Override**: Emergency override labels can bypass coverage requirements
- **Artifacts**: Coverage reports are uploaded for analysis

### Coverage Generation Process

1. **Package Selection**: Identify packages to include (excluding test utilities)
2. **Test Execution**: Run tests with `go test -coverprofile`
3. **File Filtering**: Remove test helper files from coverage data
4. **Threshold Validation**: Check against defined thresholds
5. **Reporting**: Generate human-readable coverage reports

## Rationale for Exclusions

### Test Utility Exclusion

Test utilities (`internal/testutil`, mock files) are excluded because:

1. **Different Purpose**: They support testing rather than implement business logic
2. **Testing Characteristics**: Test infrastructure has different coverage patterns
3. **Focus on Value**: Excluding them focuses metrics on meaningful business code
4. **Industry Practice**: Common practice to exclude test infrastructure from coverage

### Integration Test Exclusion

Integration and E2E test packages are excluded because:

1. **Test Code**: They contain test implementations, not production functionality
2. **System Validation**: They test overall system behavior, not individual components
3. **Different Metrics**: Integration tests measure system-level coverage differently

## Benefits

This strategy provides:

1. **Focused Metrics**: Coverage percentages reflect business logic quality
2. **Meaningful Thresholds**: Thresholds can be set appropriately for production code
3. **Developer Focus**: Encourages testing of core functionality over test infrastructure
4. **Consistent Measurement**: All coverage scripts use the same exclusion rules

## Maintenance

This strategy should be reviewed when:

1. **New Packages**: Adding packages that may need classification
2. **Threshold Adjustment**: When packages achieve higher coverage consistently
3. **Tool Changes**: If coverage tools or measurement approaches change
4. **Team Feedback**: Based on developer experience with coverage requirements

## References

- [Quality Gate Feature Flags](./QUALITY_GATE_FEATURE_FLAGS.md)
- [Testing Philosophy](../DEVELOPMENT_PHILOSOPHY.md)
- Coverage scripts in `scripts/` directory
- CI configuration in `.github/workflows/ci.yml`
