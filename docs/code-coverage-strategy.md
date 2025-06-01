# Code Coverage Strategy

This document outlines the code coverage strategy for the thinktank project, including current thresholds, future targets, and implementation details.

## Current Coverage Thresholds

The project implements a **gradual rollout approach** to code coverage thresholds:

1. **Overall Coverage Threshold**: 
   - **Target**: 75% (T008 compliance)
   - **Current**: 49% (gradual rollout)
   - **Implementation**: Environment variable controlled rollout
2. **Package-Specific Thresholds**: Different thresholds for critical packages

### Gradual Rollout Implementation

To prevent CI instability while meeting T008 requirements, we use environment variables:
- `COVERAGE_GRADUAL_ROLLOUT=true`: Enables gradual rollout mode
- `COVERAGE_THRESHOLD_CURRENT=49`: Current achievable threshold
- `COVERAGE_THRESHOLD_TARGET=75`: Target threshold per T008

### Package-Specific Thresholds

The following package-specific thresholds are currently enforced:

| Package | Current Threshold | Target Threshold | Rationale |
|---------|------------------|-----------------|-----------|
| `internal/thinktank` | 50% | 90% | Lower initial target due to current coverage level of ~18.3% |
| `internal/providers` | 85% | 90% | Higher due to current coverage of ~86.2% |
| `internal/registry` | 80% | 90% | Current coverage is ~80.9% |
| `internal/llm` | 85% | 90% | Current coverage is ~87.6% |

## Implementation

The coverage thresholds are implemented in two key scripts:

1. `scripts/check-coverage.sh`: Verifies the overall code coverage meets the specified threshold
2. `scripts/ci/check-package-specific-coverage.sh`: Enforces package-specific thresholds

These scripts are used in both the main CI workflow (`ci.yml`) and the release workflow (`release.yml`).

## Test Coverage Improvement Roadmap (T008 Implementation)

The project aims to gradually increase test coverage to reach the target threshold of **75%** (T008) and ultimately **90%** across all packages. This will be achieved through:

1. **Phase 1 - T008 Compliance (Current)**:
   - **Status**: Infrastructure implemented with gradual rollout
   - **Current Coverage**: 49.9%
   - **Target**: 75% (T008 requirement)
   - **Strategy**: Gradual rollout prevents CI instability
   - **E2E Tests**: Integrated as blocking CI step per T008

2. **Phase 2 - Short-term (1-2 months)**:
   - Focus on increasing coverage in low-coverage packages (especially `internal/thinktank`)
   - Maintain or improve coverage in high-coverage packages
   - Gradually increase threshold from 49% to 60%

2. **Medium-term (3-4 months)**:
   - Reach at least 75% coverage in all packages
   - Increase overall threshold to 80-85%
   - Normalize package-specific thresholds to a minimum of 80%

3. **Long-term (5-6 months)**:
   - Reach the target threshold of 90% across all packages
   - Standardize thresholds across the codebase

## Coverage Calculation Implementation

The coverage calculation excludes certain files and packages to ensure accurate and meaningful metrics:

- Integration and E2E tests are excluded as they have different coverage characteristics
- Test helper files and mocks are excluded
- Disabled code is excluded

## Workflow Integration

Both CI and Release workflows use the same coverage threshold strategy:

- The overall threshold is set to 64% in both workflows
- Package-specific thresholds are enforced in both workflows
- Clear TODOs are present to restore the target 90% threshold once coverage improves

## Updating Thresholds

When updating coverage thresholds:

1. Modify the threshold in both `ci.yml` and `release.yml` workflow files
2. Update package-specific thresholds in `check-package-specific-coverage.sh`
3. Update this documentation to reflect the changes and rationale
4. Update the TODO.md file with the new coverage targets
