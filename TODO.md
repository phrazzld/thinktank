# TODO

## Test Suite Enhancement

### Configure Test Coverage Reporting
- [x] Update Jest configuration
  - Description: Add coverage settings to jest.config.js (coverageDirectory, reporters, thresholds)
  - Dependencies: None
  - Priority: High
  - Status: Completed - Added coverageDirectory, reporters, and initial thresholds

- [x] Add coverage script to package.json
  - Description: Create a dedicated test:cov script for running tests with coverage
  - Dependencies: Updated Jest configuration
  - Priority: High
  - Status: Completed - Added test:cov script that runs jest with --coverage flag

- [x] Verify coverage directory is in .gitignore
  - Description: Ensure 'coverage/' is listed in .gitignore
  - Dependencies: None
  - Priority: High
  - Status: Completed - coverage/ was already in .gitignore

### Fix CLI Tests
- [ ] Add E2E testing dependencies
  - Description: Install execa or similar library for subprocess execution
  - Dependencies: None
  - Priority: High

- [ ] Create cli.e2e.test.ts
  - Description: Implement E2E tests that run the CLI as a subprocess
  - Dependencies: E2E testing dependencies
  - Priority: High

- [ ] Implement test input/output handling
  - Description: Setup temporary directories and files for E2E test scenarios
  - Dependencies: E2E testing dependencies
  - Priority: High

- [ ] Write basic CLI execution tests
  - Description: Test success and error scenarios for CLI options
  - Dependencies: E2E test implementation
  - Priority: Medium

- [ ] Deprecate or refactor existing CLI tests
  - Description: Update or remove the existing heavily-mocked CLI tests
  - Dependencies: Working E2E tests
  - Priority: Medium

### Improve Date/Time Dependent Tests
- [ ] Refactor generateRunDirectoryName tests
  - Description: Replace complex Date mocking with Jest's fake timers
  - Dependencies: None
  - Priority: High

- [ ] Test reliability verification
  - Description: Run tests repeatedly to ensure consistent results
  - Dependencies: Refactored date/time tests
  - Priority: Medium

### Standardize Mocking Approach
- [ ] Create TESTING.md document
  - Description: Document standard patterns for different mocking scenarios
  - Dependencies: None
  - Priority: Medium

- [ ] Review and refactor unit test mocking
  - Description: Apply consistent patterns to unit tests based on guidelines
  - Dependencies: Mocking guidelines document
  - Priority: Medium

- [ ] Improve integration test mocking
  - Description: Reduce excessive mocking in integration tests where appropriate
  - Dependencies: Mocking guidelines document
  - Priority: Medium

### Improve Test Coverage
- [ ] Run initial coverage analysis
  - Description: Generate a baseline coverage report to identify gaps
  - Dependencies: Coverage configuration
  - Priority: Medium

- [ ] Add tests for critical uncovered paths
  - Description: Write new tests targeting important uncovered code
  - Dependencies: Coverage analysis
  - Priority: Medium

## Feature Enhancements (Future)

- [ ] Add option to disable file writing
  - Description: Allow running without writing output files (for performance)
  - Dependencies: None
  - Priority: Low

- [ ] Add customizable output filename format
  - Description: Allow users to specify output filename patterns
  - Dependencies: None
  - Priority: Low

- [ ] Add option to disable timestamped subdirectories
  - Description: Allow writing directly to the output directory
  - Dependencies: None
  - Priority: Low

- [ ] Add console display option
  - Description: Add a flag to display model responses in console
  - Dependencies: None
  - Priority: Low

## Assumptions/Questions

1. We're prioritizing the test infrastructure improvements over new features
2. The current thresholds (50% branches, 60% functions/lines/statements) are appropriate starting points
3. E2E testing is preferred over heavily mocked unit tests for CLI testing
4. The team agrees with the standardized mocking approach outlined in the plan