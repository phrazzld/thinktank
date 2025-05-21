# Code Coverage Improvement Roadmap

This document outlines the detailed plan for improving code coverage in the thinktank project, including package-specific targets, timelines, and implementation strategies.

## Current Coverage Status

As of May 2025, the project has an overall coverage of approximately 71.9%, which is below our target threshold of 90%. Coverage varies significantly across different packages:

| Package | Current Coverage | Target | Gap |
|---------|-----------------|--------|-----|
| `internal/thinktank` | ~18.3% | 90% | 71.7% |
| `internal/fileutil` | ~59.1% | 90% | 30.9% |
| `internal/gemini` | ~63.1% | 90% | 26.9% |
| `internal/logutil` | ~47.5% | 90% | 42.5% |
| `internal/openai` | ~61.5% | 90% | 28.5% |
| `internal/providers/gemini` | ~72.4% | 90% | 17.6% |
| `internal/providers/openai` | ~75.5% | 90% | 14.5% |
| `internal/providers/openrouter` | ~80.3% | 90% | 9.7% |
| `internal/registry` | ~73.2% | 90% | 16.8% |
| `internal/thinktank/modelproc` | ~60.2% | 90% | 29.8% |
| `internal/llm` | ~98.1% | 90% | ✅ |
| `internal/thinktank/prompt` | 100.0% | 90% | ✅ |

## Prioritization Strategy

We'll focus our test coverage improvements in the following order:

1. **High Priority** (Month 1-2)
   - `internal/thinktank` (18.3% → 50%)
   - `internal/logutil` (47.5% → 70%)
   - `internal/fileutil` (59.1% → 75%)

2. **Medium Priority** (Month 3-4)
   - `internal/thinktank` (50% → 75%)
   - `internal/gemini` (63.1% → 80%)
   - `internal/openai` (61.5% → 80%)
   - `internal/thinktank/modelproc` (60.2% → 80%)

3. **Lower Priority** (Month 5-6)
   - All packages to reach 85%+
   - Focus on reaching 90% overall coverage

## Implementation Approach

### Testing Strategies

1. **Unit Tests**
   - Focus on comprehensive unit testing for all packages
   - Use table-driven tests for thorough coverage of edge cases
   - Ensure all error paths are tested

2. **Mocking**
   - Use consistent mocking approach across the codebase
   - Avoid mocking internal collaborators (per development philosophy)
   - Refactor code to improve testability where necessary

3. **Integration Tests**
   - Enhance integration tests for critical components
   - Focus on testing component interactions

### Specific Approaches by Package

1. **internal/thinktank (Priority: Highest)**
   - Create comprehensive mock implementations for external dependencies
   - Refactor code to improve testability
   - Focus on core orchestration functionality
   - Add test coverage for error handling paths

2. **internal/logutil (Priority: High)**
   - Add tests for sanitizing and secret detection
   - Improve test coverage for buffer logger
   - Test stream separation functionality

3. **internal/fileutil (Priority: High)**
   - Improve test coverage for file processing
   - Add tests for error handling paths
   - Use memory-based filesystem for faster tests

## Measurement and Tracking

We will track coverage improvements using:

1. **Weekly Coverage Reports**
   - Run coverage analysis every week
   - Track improvements by package

2. **CI Integration**
   - Gradually increase thresholds in CI as coverage improves
   - Update package-specific thresholds based on progress

3. **Pull Request Reviews**
   - Enforce test coverage requirements for new code
   - All new code should aim for 90%+ coverage

## Threshold Adjustment Timeline

As coverage improves, we'll adjust thresholds in the following phases:

### Phase 1 (Month 1-2)
- Overall threshold: 64% → 70%
- `internal/thinktank`: 50% → 60%
- `internal/logutil`: 47.5% → 60%

### Phase 2 (Month 3-4)
- Overall threshold: 70% → 80%
- `internal/thinktank`: 60% → 75%
- All other packages: 75%+ minimum

### Phase 3 (Month 5-6)
- Overall threshold: 80% → 90%
- All packages: 85%+ minimum, aiming for 90%

## Challenges and Mitigation

1. **Complex External Dependencies**
   - Challenge: Testing code with external API calls
   - Mitigation: Improve mock implementations, use test adapters

2. **Concurrency Testing**
   - Challenge: Testing race conditions and concurrent code
   - Mitigation: Use race detector, implement synchronization tests

3. **Test Performance**
   - Challenge: Slow test execution as coverage increases
   - Mitigation: Organize tests by speed, use short flag for CI

## Success Criteria

The code coverage improvement initiative will be considered successful when:

1. Overall project coverage reaches 90%+
2. All packages have at least 85% coverage
3. Core packages have 90%+ coverage
4. CI enforces the target thresholds
5. All new code consistently meets coverage requirements
