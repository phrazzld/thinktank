# Test Coverage Remediation Plan

## Overview

This ticket outlines the necessary test coverage improvements needed after the dead code elimination PR (#22). The primary focus is on improving test coverage for the Registry API Service implementation, which now serves as the primary API provider mechanism but lacks sufficient test coverage.

## Current Status

Current overall code coverage: 64.4% (CI threshold: 75%)

## Key Files Requiring Test Coverage

1. `internal/thinktank/registry_api.go` - Several methods have 0% coverage:
   - `NewRegistryAPIService`: 0%
   - `InitLLMClient`: 0%
   - `GetModelParameters`: 0%
   - `ValidateModelParameter`: 0%
   - `GetModelDefinition`: 0%
   - `GetModelTokenLimits`: 0%
   - `ProcessLLMResponse`: 0%
   - `IsEmptyResponseError`: 0%
   - `IsSafetyBlockedError`: 0%
   - `getEnvVarNameForProvider`: 0%
   - `GetErrorDetails`: 0%

2. `internal/thinktank/adapters.go` - Adapter methods with low/no coverage:
   - `InitLLMClient`: 0%
   - `ProcessLLMResponse`: 0%
   - `GetErrorDetails`: 0%
   - `IsEmptyResponseError`: 0%
   - `IsSafetyBlockedError`: 0%
   - `GetModelParameters`: 0%
   - `ValidateModelParameter`: 0%
   - `GetModelDefinition`: 0%
   - `GetModelTokenLimits`: 0%
   - `interfacesToInternalContextStats`: 0%
   - `internalToInterfacesContextStats`: 0%
   - `internalToInterfacesGatherConfig`: 0%
   - `DisplayDryRunInfo`: 0%
   - `SaveToFile`: 0%
   - `GatherContext`: 0%

## Action Plan

1. Create test suite specifically for registry_api.go:
   - Add comprehensive test cases for each public method
   - Focus on edge cases and error handling
   - Create appropriate mocks for Registry dependencies

2. Improve adapter tests:
   - Test adapter conversion methods
   - Ensure proper type conversions are tested
   - Verify error propagation

3. Fix or exclude test helper files:
   - Examine test inclusion/exclusion patterns in coverage scripts
   - Consider excluding pure test utilities from coverage metrics

4. Create integration tests:
   - Add tests that verify Registry API and adapters work together
   - Test with real-world workflows

## Implementation Approach

1. Create `registry_api_test.go` with test cases for each method:
   - Mock Registry interface for unit testing
   - Test successful and error paths
   - Verify API key resolution behavior
   - Test parameter validation
   - Test error handling

2. Create `adapters_test.go` to test conversion functions:
   - Test conversion in both directions
   - Test edge cases
   - Test with realistic data

3. Update coverage scripts if needed:
   - Determine if threshold adjustment is necessary
   - Evaluate patterns for excluding test utilities

## Success Criteria

- Overall code coverage returns to at least 75%
- No business logic functions with 0% coverage
- Test coverage extends to error handling scenarios
- CI passes with standard coverage thresholds

## Timeline

This work should be completed before the next feature branch to ensure ongoing code quality.
